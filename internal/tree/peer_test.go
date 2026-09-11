package tree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runGit runs a git command in dir, failing the test on a non-zero exit.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// initRepo makes dir a git repository with one commit, so a worktree can be
// added to it — Peers' worktree-resolution path (repoParent) has nothing to
// resolve against otherwise.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-q", "-m", "initial")
}

func TestResolvePeerNotDeclared(t *testing.T) {
	rp := ResolvePeer(map[string]string{}, "web")
	if rp.Resolution != PeerNotDeclared {
		t.Errorf("Resolution = %v, want PeerNotDeclared", rp.Resolution)
	}
}

func TestResolvePeerNotPresent(t *testing.T) {
	parent := t.TempDir()
	rp := ResolvePeer(map[string]string{"web": filepath.Join(parent, "web", "spectre")}, "web")
	if rp.Resolution != PeerNotPresent {
		t.Errorf("Resolution = %v, want PeerNotPresent", rp.Resolution)
	}
}

func TestResolvePeerUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	parent := t.TempDir()
	path := makeTree(t, filepath.Join(parent, "web"))
	root := filepath.Join(path, "spectre")
	// Block traversal into path so Stat(root) fails with a permission error
	// rather than not-exist.
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })

	rp := ResolvePeer(map[string]string{"web": root}, "web")
	if rp.Resolution != PeerUnreadable {
		t.Errorf("Resolution = %v, want PeerUnreadable", rp.Resolution)
	}
	if rp.Err == nil {
		t.Error("Err = nil, want the underlying open error")
	}
}

func TestResolvePeerFound(t *testing.T) {
	parent := t.TempDir()
	path := makeTree(t, filepath.Join(parent, "web"))
	root := filepath.Join(path, "spectre")
	spec := "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"
	if err := os.WriteFile(filepath.Join(root, "specs", "billing.md"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	rp := ResolvePeer(map[string]string{"web": root}, "web")
	if rp.Resolution != PeerFound {
		t.Fatalf("Resolution = %v, want PeerFound", rp.Resolution)
	}
	if rp.Tree == nil || rp.Tree.Root != root {
		t.Errorf("Tree = %+v, want Root %q", rp.Tree, root)
	}
	if len(rp.Specs) != 1 || rp.Specs[0].Capability != "billing" {
		t.Errorf("Specs = %+v, want one spec named billing", rp.Specs)
	}
}

func TestResolvePeerSpecsUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	parent := t.TempDir()
	path := makeTree(t, filepath.Join(parent, "web"))
	root := filepath.Join(path, "spectre")
	spec := "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"
	specPath := filepath.Join(root, "specs", "billing.md")
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(specPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(specPath, 0o644) })

	// The path stats fine and Open succeeds; only reading the spec file
	// fails, so ResolvePeer must classify this as PeerUnreadable itself
	// rather than leaving Specs()-load failures for the caller to detect.
	rp := ResolvePeer(map[string]string{"web": root}, "web")
	if rp.Resolution != PeerUnreadable {
		t.Errorf("Resolution = %v, want PeerUnreadable", rp.Resolution)
	}
	if rp.Err == nil {
		t.Error("Err = nil, want the underlying read error")
	}
}

func TestPeerResolutionString(t *testing.T) {
	for _, tt := range []struct {
		r    PeerResolution
		want string
	}{
		{PeerFound, "found"},
		{PeerNotDeclared, "not declared"},
		{PeerNotPresent, "not present"},
		{PeerUnreadable, "unreadable"},
	} {
		if got := tt.r.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.r, got, tt.want)
		}
	}
}

// markerChange writes changes/<name>/tasks.md under treeRoot (a tree
// directory, e.g. .../spectre) so a test can prove resolution landed on the
// real tree by finding it, rather than merely by comparing path strings.
func markerChange(t *testing.T, treeRoot, name string) {
	t.Helper()
	dir := filepath.Join(treeRoot, "changes", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestPeersResolvesRepoNamedLikeTheTree covers a `peers` path that names a
// repository whose own directory happens to be called "spectre" — the
// reproduced defect: Peers() used to see the basename "spectre" and skip
// appending the tree leaf, landing one level short at the repository root
// rather than at <repo>/spectre. Contrasted with an ordinary peer, whose
// basename differs and must still gain the leaf exactly as before.
func TestPeersResolvesRepoNamedLikeTheTree(t *testing.T) {
	tests := []struct {
		name    string
		peerRel string                                   // path written into the peers file
		build   func(t *testing.T, parent string) string // returns the peer repo dir
	}{
		{
			name:    "repository directory named spectre, declared by its own name",
			peerRel: "../spectre",
			build: func(t *testing.T, parent string) string {
				return makeTree(t, filepath.Join(parent, "spectre"))
			},
		},
		{
			name:    "ordinary peer, declared by its own name",
			peerRel: "../web",
			build: func(t *testing.T, parent string) string {
				return makeTree(t, filepath.Join(parent, "web"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			me := makeTree(t, filepath.Join(parent, "app"))
			repoDir := tt.build(t, parent)
			wantTree := filepath.Join(repoDir, "spectre")
			markerChange(t, wantTree, "marker-change")

			body := "peer " + tt.peerRel + "\n"
			if err := os.WriteFile(filepath.Join(me, "spectre", "peers"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			tr, err := Find(me)
			if err != nil {
				t.Fatal(err)
			}
			peers, err := tr.Peers()
			if err != nil {
				t.Fatal(err)
			}
			want, _ := filepath.EvalSymlinks(wantTree)
			got, _ := filepath.EvalSymlinks(peers["peer"])
			if got != want {
				t.Fatalf("peers[peer] = %q, want %q", got, want)
			}

			rp := ResolvePeer(peers, "peer")
			if rp.Resolution != PeerFound {
				t.Fatalf("Resolution = %v, want PeerFound", rp.Resolution)
			}
			if _, err := os.Stat(filepath.Join(rp.Tree.ChangesDir(), "marker-change", "tasks.md")); err != nil {
				t.Errorf("marker change not visible through resolved tree: %v", err)
			}
		})
	}
}

// TestPeersResolvesPathAlreadyATree covers a `peers` path that already names
// the tree directory — which must gain no second "spectre" leaf — and a
// declared path that is neither a repository containing a tree nor a tree
// itself, which must resolve to PeerNotPresent rather than silently
// succeeding one level off.
func TestPeersResolvesPathAlreadyATree(t *testing.T) {
	t.Run("path already ending in spectre and already a tree gains no leaf", func(t *testing.T) {
		parent := t.TempDir()
		me := makeTree(t, filepath.Join(parent, "app"))
		repoDir := makeTree(t, filepath.Join(parent, "spectre"))
		wantTree := filepath.Join(repoDir, "spectre")
		markerChange(t, wantTree, "marker-change")

		body := "peer ../spectre/spectre\n"
		if err := os.WriteFile(filepath.Join(me, "spectre", "peers"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		tr, err := Find(me)
		if err != nil {
			t.Fatal(err)
		}
		peers, err := tr.Peers()
		if err != nil {
			t.Fatal(err)
		}
		want, _ := filepath.EvalSymlinks(wantTree)
		got, _ := filepath.EvalSymlinks(peers["peer"])
		if got != want {
			t.Fatalf("peers[peer] = %q, want %q", got, want)
		}

		rp := ResolvePeer(peers, "peer")
		if rp.Resolution != PeerFound {
			t.Fatalf("Resolution = %v, want PeerFound", rp.Resolution)
		}
		if _, err := os.Stat(filepath.Join(rp.Tree.ChangesDir(), "marker-change", "tasks.md")); err != nil {
			t.Errorf("marker change not visible through resolved tree: %v", err)
		}
	})

	t.Run("declared path that is neither a repo nor a tree resolves to PeerNotPresent", func(t *testing.T) {
		parent := t.TempDir()
		me := makeTree(t, filepath.Join(parent, "app"))
		// A bare directory with no changes/ inside it — not a tree, and not
		// a repository whose spectre/ subdirectory exists either.
		if err := os.MkdirAll(filepath.Join(parent, "empty"), 0o755); err != nil {
			t.Fatal(err)
		}

		body := "peer ../empty\n"
		if err := os.WriteFile(filepath.Join(me, "spectre", "peers"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		tr, err := Find(me)
		if err != nil {
			t.Fatal(err)
		}
		peers, err := tr.Peers()
		if err != nil {
			t.Fatal(err)
		}

		rp := ResolvePeer(peers, "peer")
		if rp.Resolution != PeerNotPresent {
			t.Fatalf("Resolution = %v, want PeerNotPresent", rp.Resolution)
		}
	})
}

// TestPeersResolvesFromInsideAWorktree covers the defect this test guards
// against: a relative peers entry resolved from inside a git worktree
// checked out under `<repo>/.worktrees/<name>/` used to land one level
// short — inside `.worktrees/` itself — because that directory sits two
// levels deeper than the repository the entry was authored against
// (`../peer`, written assuming siblings at the repository's own level).
// `spectre link` run from such a worktree therefore always saw the peer as
// PeerNotPresent, exactly the shape KAN-486 hit hand-writing link.md
// instead of using the CLI. repoParent's git-common-dir resolution fixes
// it: this test proves resolution from the worktree lands on the same real
// peer tree that resolution from the main checkout does.
func TestPeersResolvesFromInsideAWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	parent := t.TempDir()
	canonicalDir := filepath.Join(parent, "canonical")
	peerDir := filepath.Join(parent, "peer")

	initRepo(t, canonicalDir)

	peerTreeDir := makeTree(t, peerDir)
	initRepo(t, peerDir)
	peerRoot := filepath.Join(peerTreeDir, "spectre")
	markerChange(t, peerRoot, "marker-change")

	// A worktree of canonicalDir, nested two levels under it — the same
	// shape `git worktree add <repo>/.worktrees/<name>` produces. Its own
	// spectre/ tree is written directly (never committed): Find/Peers read
	// the filesystem, not git history, so nothing here needs to be tracked.
	worktreeDir := filepath.Join(canonicalDir, ".worktrees", "kan-486")
	runGit(t, canonicalDir, "worktree", "add", "--quiet", worktreeDir, "-b", "kan-486")
	makeTree(t, worktreeDir)

	body := "peer ../peer\n"
	if err := os.WriteFile(filepath.Join(worktreeDir, "spectre", "peers"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	tr, err := Find(worktreeDir)
	if err != nil {
		t.Fatal(err)
	}
	peers, err := tr.Peers()
	if err != nil {
		t.Fatal(err)
	}

	want, _ := filepath.EvalSymlinks(peerRoot)
	got, _ := filepath.EvalSymlinks(peers["peer"])
	if got != want {
		t.Fatalf("peers[peer] = %q, want %q (the real peer tree, not something under .worktrees/)", got, want)
	}

	rp := ResolvePeer(peers, "peer")
	if rp.Resolution != PeerFound {
		t.Fatalf("Resolution = %v, want PeerFound", rp.Resolution)
	}
	if _, err := os.Stat(filepath.Join(rp.Tree.ChangesDir(), "marker-change", "tasks.md")); err != nil {
		t.Errorf("marker change not visible through resolved tree: %v", err)
	}
}

// TestNamesForFromInsideAWorktree covers NamesFor's own half of the same
// defect TestPeersResolvesFromInsideAWorktree fixes for Peers: a peer's
// declared path always resolves against ITS repository's primary
// checkout (Peers' repoParent-based resolution), so when root is itself a
// worktree tree, a bare directory comparison against root never matches
// even the entry that genuinely names root's own repository — link.go's
// namesBack calls NamesFor(declared, t.Root) with exactly this shape, and
// without the fix a change confined to worktrees on both sides of a
// cross-repo link always reads as one-sided, on both sides, regardless of
// resolveCounterpart's own worktree fallback.
func TestNamesForFromInsideAWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	parent := t.TempDir()
	meDir := filepath.Join(parent, "me")
	initRepo(t, meDir)
	makeTree(t, meDir)

	otherDir := filepath.Join(parent, "other")
	initRepo(t, otherDir)

	worktreeDir := filepath.Join(meDir, ".worktrees", "kan-486")
	runGit(t, meDir, "worktree", "add", "--quiet", worktreeDir, "-b", "kan-486")
	makeTree(t, worktreeDir)

	worktreeRoot := filepath.Join(worktreeDir, "spectre")
	declared := map[string]string{"me": filepath.Join(meDir, "spectre")}

	if got := NamesFor(declared, worktreeRoot); !got["me"] {
		t.Errorf("NamesFor(declared, worktree root) = %v, want {\"me\": true} — "+
			"declared[\"me\"] names the same repository worktreeRoot belongs to", got)
	}

	// The ordinary, non-worktree case stays unchanged: declared["other"]
	// genuinely is a different repository, so it must still not match.
	declared["other"] = otherDir
	if got := NamesFor(declared, worktreeRoot); got["other"] {
		t.Errorf("NamesFor(declared, worktree root) = %v, want \"other\" absent — "+
			"it names an unrelated repository", got)
	}
}

func TestNamesFor(t *testing.T) {
	parent := t.TempDir()
	me := makeTree(t, filepath.Join(parent, "app"))
	meRoot := filepath.Join(me, "spectre")
	theirDeclared := map[string]string{
		"app-alias": meRoot,
		"someone":   filepath.Join(parent, "elsewhere", "spectre"),
	}
	got := NamesFor(theirDeclared, meRoot)
	if !got["app-alias"] || got["someone"] {
		t.Errorf("NamesFor = %+v", got)
	}
}
