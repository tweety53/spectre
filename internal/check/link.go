package check

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/parse"
	"github.com/tweety53/spectre/internal/tree"
)

// LinkFile is the fixed filename a change's pointer tree uses, mirroring
// tree.ProposalFile, tree.TasksFile and tree.DesignFile.
const LinkFile = "link.md"

// IsSatellite reports whether dir is a pointer tree: it holds link.md and
// nothing else (design.md's pointer-tree-not-full-tree). "Nothing else"
// means design.md too — a directory that also carries one keeps every
// check a full scaffold gets, same as proposal.md and tasks.md. It is
// exported so archive reads the same definition Structural does, rather
// than a copy that would drift the way it already has once, when review
// widened "and nothing else" to include design.md.
func IsSatellite(dir string) bool {
	return fileExists(filepath.Join(dir, LinkFile)) &&
		!fileExists(filepath.Join(dir, tree.ProposalFile)) &&
		!fileExists(filepath.Join(dir, tree.TasksFile)) &&
		!fileExists(filepath.Join(dir, tree.DesignFile))
}

// LinkFindings checks change c's parsed link.md (link) against peers, per
// design.md's pointer-tree-not-full-tree, independent-archive and
// peer-absence-is-not-a-finding decisions:
//
//   - a peer named by "## Part of" or a "## Parts" entry must be declared
//     in peers — not declared is a finding;
//   - a peer that is declared but not present on disk is not checked
//     further and produces no finding (unreadable-peer-is-a-finding
//     narrows this: a declared path that exists but cannot be opened as a
//     tree, or whose specs cannot be read, is a finding — that peer is
//     there and broken, not merely absent);
//   - a peer that is declared and present must resolve a counterpart
//     change (searched in changes/<id>/ then changes/archive/<id>/) whose
//     own link.md names this change back — a one-sided link is a finding;
//   - a counterpart resolved under changes/archive/ while this side is
//     open is an archive-skew finding, never a refusal;
//   - once the counterpart resolves and names back, both sides' "##
//     Branch" must agree, and — on the satellite side — every "## Tasks
//     here" number must match a task in the canonical's tasks.md;
//   - "## Merge order", when present, must name "." exactly once and
//     every "## Parts" entry exactly once — independent of any peer's
//     resolution, since it is checked against link's own Parts list.
//
// relPath is link.md's path relative to the tree root, for Finding.File.
func LinkFindings(t *tree.Tree, c model.Change, link model.Link, relPath string, peers map[string]tree.ResolvedPeer) []Finding {
	var out []Finding

	if link.PartOf != (model.LinkRef{}) {
		out = append(out, linkRefFindings(t, c, relPath, link.PartOf, link, peers, true)...)
	}
	for _, part := range link.Parts {
		out = append(out, linkRefFindings(t, c, relPath, part, link, peers, false)...)
	}
	if len(link.MergeOrder) > 0 {
		out = append(out, mergeOrderFindings(relPath, link)...)
	}

	return out
}

// linkRefFindings checks one LinkRef: link.PartOf when isPartOf, or one
// entry of link.Parts otherwise.
func linkRefFindings(t *tree.Tree, c model.Change, relPath string, ref model.LinkRef, link model.Link, peers map[string]tree.ResolvedPeer, isPartOf bool) []Finding {
	rp, declared := peers[ref.Peer]
	if !declared || rp.Resolution == tree.PeerNotDeclared {
		return []Finding{{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"peer %q is not declared in peers", ref.Peer)}}
	}
	if rp.Resolution == tree.PeerUnreadable {
		// unreadable-peer-is-a-finding narrows peer-absence-is-not-a-
		// finding: a peer tree that is there and broken is not the
		// ordinary absent-worktree case, so it is reported, the same way
		// RefFindings already reports it for a citation.
		return []Finding{{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"peer %q could not be read at %s: %v", ref.Peer, rp.Path, rp.Err)}}
	}
	if rp.Resolution != tree.PeerFound {
		// peer-absence-is-not-a-finding: declared but not present is
		// reported as not checked, never as a finding.
		return nil
	}

	dir, archived, found := resolveCounterpart(rp.Tree, ref.ChangeID)
	var cLink model.Link
	reciprocal := false
	if found {
		cp := parse.New(rp.Tree.Cfg)
		if l, err := cp.LinkFile(filepath.Join(dir, LinkFile)); err == nil {
			cLink = l
			reciprocal = namesBack(t, rp.Tree, c.ID, cLink, isPartOf)
		}
	}
	if !found || !reciprocal {
		return []Finding{{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"one-sided link: %s names %s:%s, but %s:%s does not link back to it",
			c.ID, ref.Peer, ref.ChangeID, ref.Peer, ref.ChangeID)}}
	}

	var out []Finding
	if archived {
		out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"archive skew: %s:%s is archived, %s is not", ref.Peer, ref.ChangeID, c.ID)})
	}
	if link.Branch != cLink.Branch {
		out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"branch mismatch: %q here, %q in %s:%s", link.Branch, cLink.Branch, ref.Peer, ref.ChangeID)})
	}
	if isPartOf {
		want := map[int]bool{}
		cp := parse.New(rp.Tree.Cfg)
		if ts, err := cp.TasksFile(filepath.Join(dir, tree.TasksFile)); err == nil {
			for _, task := range ts {
				want[task.Num] = true
			}
		}
		for _, n := range link.TasksHere {
			if !want[n] {
				out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf(
					"tasks here: %d is not a task in %s:%s", n, ref.Peer, ref.ChangeID)})
			}
		}
	}
	return out
}

// resolveCounterpart locates a change directory in peer tree p, searching
// changes/<id>/ before changes/archive/<id>/ (independent-archive), then —
// when neither exists — a worktree of p's own repository named after id,
// via tree.CounterpartWorktree. Peers() always resolves a peer name to the
// repository's primary checkout, regardless of which worktree issued the
// lookup, so p is never itself a worktree tree; before either side of a
// cross-repo link has landed on its primary checkout, that worktree is the
// only place the counterpart exists at all, and without this fallback it
// reads as one-sided on both sides for as long as the change lives only in
// worktrees.
func resolveCounterpart(p *tree.Tree, id string) (dir string, archived, found bool) {
	if dir, archived, found := resolveCounterpartInTree(p, id); found {
		return dir, archived, found
	}
	wt, ok := tree.CounterpartWorktree(p, id)
	if !ok {
		return "", false, false
	}
	return resolveCounterpartInTree(wt, id)
}

// resolveCounterpartInTree is resolveCounterpart's search of one tree,
// shared between the peer's primary checkout and its worktree fallback.
func resolveCounterpartInTree(p *tree.Tree, id string) (dir string, archived, found bool) {
	open := filepath.Join(p.ChangesDir(), id)
	if fi, err := os.Stat(open); err == nil && fi.IsDir() {
		return open, false, true
	}
	arch := filepath.Join(p.ArchiveDir(), id)
	if fi, err := os.Stat(arch); err == nil && fi.IsDir() {
		return arch, true, true
	}
	return "", false, false
}

// namesBack reports whether cLink — the counterpart's own parsed link.md,
// reached from tree t's change ourID via peer tree p — names that change
// back. p's own peers file may not use the same name for t that t uses
// for p, so the comparison goes through tree.NamesFor rather than
// assuming the names agree.
func namesBack(t *tree.Tree, p *tree.Tree, ourID string, cLink model.Link, isPartOf bool) bool {
	declared, err := p.Peers()
	if err != nil {
		return false
	}
	names := tree.NamesFor(declared, t.Root)
	if len(names) == 0 {
		return false
	}
	if isPartOf {
		// We are the satellite; the counterpart is canonical and must
		// list us in its "## Parts".
		for _, part := range cLink.Parts {
			if names[part.Peer] && part.ChangeID == ourID {
				return true
			}
		}
		return false
	}
	// We are canonical; the counterpart is a satellite and must carry a
	// "## Part of" naming us.
	return names[cLink.PartOf.Peer] && cLink.PartOf.ChangeID == ourID
}

// mergeOrderFindings checks link.MergeOrder against link.Parts: "." must
// appear exactly once and every Parts peer name exactly once, and nothing
// else. Missing, duplicated and unknown entries are three distinct
// findings, reported in a deterministic (sorted) order.
func mergeOrderFindings(relPath string, link model.Link) []Finding {
	want := map[string]bool{".": true}
	for _, p := range link.Parts {
		want[p.Peer] = true
	}

	seen := map[string]int{}
	var out []Finding
	for _, item := range link.MergeOrder {
		seen[item]++
		if !want[item] {
			out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf(
				"merge order: %q is not \".\" or a declared part", item)})
		}
	}

	var missing []string
	for item := range want {
		if seen[item] == 0 {
			missing = append(missing, item)
		}
	}
	sort.Strings(missing)
	for _, item := range missing {
		out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"merge order: missing %q", item)})
	}

	var dup []string
	for item, n := range seen {
		if n > 1 {
			dup = append(dup, item)
		}
	}
	sort.Strings(dup)
	for _, item := range dup {
		out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf(
			"merge order: %q appears %d times", item, seen[item])})
	}

	return out
}
