package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tweety53/spectre/internal/check"
	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/parse"
	"github.com/tweety53/spectre/internal/tree"
)

// Link writes both sides of a cross-tree link in one run
// (design.md's one-command-writes-both-sides): the peer's canonical
// change gains "## Parts" and "## Merge order", and the local (satellite)
// tree gains "## Part of" naming it. Run in the satellite tree; the
// satellite change directory is created under the same id as the
// canonical change, since link.md records the same change viewed from
// the other repository.
//
// Refused (exit Fail, each naming the failed check) unless: the peer is
// declared in peers and present, the canonical change exists in the peer
// tree, the canonical change is not itself a satellite, the link does not
// already exist on either side, and the peer's own change directory has
// no uncommitted modifications. --force overrides the last refusal only
// (design.md's guarded-cross-tree-write). The peer side is written first
// and the local side second, so a failed local write leaves the peer side
// as the only thing landed rather than nothing at all. A malformed
// invocation is exit Usage.
func Link(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("link", stderr)
	force := fs.Bool("force", false, "override the uncommitted-modifications refusal")
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: spectre link [--root <path>] [--force] <peer>:<canonical-id>")
		return Usage
	}
	arg := fs.Arg(0)
	peerName, canonicalID, ok := strings.Cut(arg, ":")
	if !ok || peerName == "" || canonicalID == "" {
		fmt.Fprintf(stderr, "usage: spectre link [--root <path>] [--force] <peer>:<canonical-id>; got %q\n", arg)
		return Usage
	}
	if !isValidID(canonicalID) {
		fmt.Fprintf(stderr, "invalid change id: %q\n", canonicalID)
		return Usage
	}

	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	declared, err := t.Peers()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	rp := tree.ResolvePeer(declared, peerName)
	if rp.Resolution != tree.PeerFound {
		fmt.Fprintf(stderr, "peer %q is %s\n", peerName, rp.Resolution)
		return Fail
	}
	peerTree := rp.Tree

	peerChanges, err := peerTree.Changes()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	if !hasChange(peerChanges, canonicalID) {
		fmt.Fprintf(stderr, "canonical change %q does not exist in peer %q (%s)\n",
			canonicalID, peerName, peerTree.ChangesDir())
		return Fail
	}
	canonicalDir := filepath.Join(peerTree.ChangesDir(), canonicalID)
	canonicalLinkPath := filepath.Join(canonicalDir, check.LinkFile)

	peerParser := parse.New(peerTree.Cfg)
	canLink, err := readLinkIfExists(peerParser, canonicalLinkPath)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", canonicalID, err)
		return Usage
	}
	if canLink.PartOf != (model.LinkRef{}) {
		fmt.Fprintf(stderr, "canonical change %q is itself a satellite (links are one level deep)\n", canonicalID)
		return Fail
	}

	localDir := filepath.Join(t.ChangesDir(), canonicalID)
	localLinkPath := filepath.Join(localDir, check.LinkFile)
	if _, statErr := os.Stat(localLinkPath); statErr == nil {
		fmt.Fprintf(stderr, "this link already exists: %s\n", localLinkPath)
		return Fail
	} else if !errors.Is(statErr, os.ErrNotExist) {
		fmt.Fprintln(stderr, statErr)
		return Usage
	}
	// guarded-cross-tree-write applies to the local side too: a change
	// directory that already exists here — even without a link.md of its
	// own — is an unrelated change sharing the same id, and writing
	// link.md alongside its content would silently mix the two rather
	// than refuse.
	if entries, err := os.ReadDir(localDir); err == nil {
		if len(entries) > 0 {
			fmt.Fprintf(stderr, "%s: change directory already exists\n", localDir)
			return Fail
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	peerDeclared, err := peerTree.Peers()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	namesForLocal := tree.NamesFor(peerDeclared, t.Root)
	if len(namesForLocal) == 0 {
		fmt.Fprintf(stderr, "peer %q has no declared name for this tree in its own peers file\n", peerName)
		return Fail
	}
	nameForLocal := firstSorted(namesForLocal)
	for _, part := range canLink.Parts {
		if namesForLocal[part.Peer] && part.ChangeID == canonicalID {
			fmt.Fprintf(stderr, "this link already exists: %s:%s already names %s:%s back\n",
				peerName, canonicalID, part.Peer, canonicalID)
			return Fail
		}
	}

	dirty, err := hasUncommittedModifications(canonicalDir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	if dirty && !*force {
		fmt.Fprintf(stderr, "%s: peer's change directory has uncommitted modifications (use --force to override)\n",
			canonicalDir)
		return Fail
	}

	satBranch, err := currentBranch(t.Root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	newParts := append(append([]model.LinkRef{}, canLink.Parts...), model.LinkRef{Peer: nameForLocal, ChangeID: canonicalID})
	newMergeOrder := canLink.MergeOrder
	if len(newMergeOrder) == 0 {
		newMergeOrder = []string{".", nameForLocal}
	} else {
		newMergeOrder = append(append([]string{}, newMergeOrder...), nameForLocal)
	}
	newBranch := canLink.Branch
	if newBranch == "" {
		newBranch = satBranch
	}
	newCanLink := model.Link{Parts: newParts, Branch: newBranch, MergeOrder: newMergeOrder}

	// The peer side is written first: a failure writing the local side
	// below then leaves only the peer side landed, never the reverse
	// (design.md's guarded-cross-tree-write).
	if err := os.WriteFile(canonicalLinkPath, renderLink(newCanLink), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	if err := os.MkdirAll(localDir, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	localLink := model.Link{PartOf: model.LinkRef{Peer: peerName, ChangeID: canonicalID}, Branch: satBranch}
	if err := os.WriteFile(localLinkPath, renderLink(localLink), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	fmt.Fprintf(stdout, "linked %s to %s:%s\n", displayPath(localDir), peerName, canonicalID)
	return OK
}

// readLinkIfExists parses path's link.md when it exists, and returns the
// zero Link (never an error) when it does not.
func readLinkIfExists(p *parse.Parser, path string) (model.Link, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.Link{}, nil
		}
		return model.Link{}, err
	}
	return p.LinkFile(path)
}

// firstSorted returns the lexicographically first key of names, for a
// deterministic choice when a peer tree declares more than one name for
// the same directory.
func firstSorted(names map[string]bool) string {
	keys := make([]string, 0, len(names))
	for k := range names {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys[0]
}

// hasUncommittedModifications reports whether dir, inside a git working
// tree, carries any uncommitted change — tracked or untracked — under it.
func hasUncommittedModifications(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf(
			"checking %s for uncommitted modifications requires a git repository; git status failed: %w\n%s",
			dir, err, out)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// currentBranch reports the branch checked out in the git working tree
// containing dir.
func currentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"reading the current branch of %s requires a git repository; git rev-parse failed: %w\n%s",
			dir, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// renderLink writes l as link.md text: the satellite shape ("## Part of"
// + "## Branch") when l.PartOf is set, the canonical shape ("## Parts" +
// "## Branch" + "## Merge order") when l.Parts is set. spectre link never
// writes "## Tasks here" — task 6's own worked example sets it by hand
// afterwards.
func renderLink(l model.Link) []byte {
	var b strings.Builder
	if l.PartOf != (model.LinkRef{}) {
		fmt.Fprintf(&b, "## Part of\n\n`%s:%s`\n\n", l.PartOf.Peer, l.PartOf.ChangeID)
	}
	if len(l.Parts) > 0 {
		b.WriteString("## Parts\n\n")
		for _, p := range l.Parts {
			fmt.Fprintf(&b, "`%s:%s`\n", p.Peer, p.ChangeID)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "## Branch\n\n%s\n", l.Branch)
	if len(l.MergeOrder) > 0 {
		b.WriteString("\n## Merge order\n\n")
		for i, m := range l.MergeOrder {
			fmt.Fprintf(&b, "%d. `%s`\n", i+1, m)
		}
	}
	return []byte(b.String())
}
