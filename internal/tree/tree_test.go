package tree

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeTree creates <dir>/spectre/{specs,changes} and returns <dir>.
func makeTree(t *testing.T, dir string) string {
	t.Helper()
	for _, sub := range []string{"specs", "changes"} {
		if err := os.MkdirAll(filepath.Join(dir, "spectre", sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFindWalksUp(t *testing.T) {
	base := makeTree(t, t.TempDir())
	deep := filepath.Join(base, "internal", "pkg")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	tr, err := Find(deep)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(base, "spectre"))
	got, _ := filepath.EvalSymlinks(tr.Root)
	if got != want {
		t.Errorf("Root = %q, want %q", got, want)
	}
}

func TestFindNoRoot(t *testing.T) {
	_, err := Find(t.TempDir())
	if !errors.Is(err, ErrNoRoot) {
		t.Fatalf("err = %v, want ErrNoRoot", err)
	}
}

func TestPeers(t *testing.T) {
	parent := t.TempDir()
	me := makeTree(t, filepath.Join(parent, "app"))
	other := makeTree(t, filepath.Join(parent, "gymie"))

	body := "# neighbours\ngymie ../gymie\n\n"
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
	if len(peers) != 1 {
		t.Fatalf("len = %d, want 1", len(peers))
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(other, "spectre"))
	got, _ := filepath.EvalSymlinks(peers["gymie"])
	if got != want {
		t.Errorf("peers[gymie] = %q, want %q", got, want)
	}
}

func TestPeersAbsent(t *testing.T) {
	tr, err := Find(makeTree(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	peers, err := tr.Peers()
	if err != nil {
		t.Fatalf("missing peers file must not error: %v", err)
	}
	if len(peers) != 0 {
		t.Errorf("len = %d, want 0", len(peers))
	}
}

func TestPeersDuplicateName(t *testing.T) {
	parent := t.TempDir()
	me := makeTree(t, filepath.Join(parent, "app"))
	makeTree(t, filepath.Join(parent, "gymie"))
	makeTree(t, filepath.Join(parent, "other"))

	body := "gymie ../gymie\nother ../other\ngymie ../other\n"
	if err := os.WriteFile(filepath.Join(me, "spectre", "peers"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := Find(me)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tr.Peers()
	if err == nil {
		t.Fatalf("Peers() should error on duplicate name, got nil")
	}
	// Pin the association, not just the presence of both digits: line 3 is
	// the current (duplicate) line, line 1 is where the name first appeared.
	if !strings.Contains(err.Error(), "peers:3:") || !strings.Contains(err.Error(), "first seen on line 1") {
		t.Errorf("error = %q, want it to name line 3 as current and line 1 as first seen", err.Error())
	}
}

func TestPeersAliases(t *testing.T) {
	parent := t.TempDir()
	me := makeTree(t, filepath.Join(parent, "app"))
	target := makeTree(t, filepath.Join(parent, "target"))

	body := "alias1 ../target\nalias2 ../target\n"
	if err := os.WriteFile(filepath.Join(me, "spectre", "peers"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := Find(me)
	if err != nil {
		t.Fatal(err)
	}
	peers, err := tr.Peers()
	if err != nil {
		t.Fatalf("Peers() should allow aliases: %v", err)
	}
	if len(peers) != 2 {
		t.Fatalf("len = %d, want 2", len(peers))
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(target, "spectre"))
	for name := range peers {
		got, _ := filepath.EvalSymlinks(peers[name])
		if got != want {
			t.Errorf("peers[%s] = %q, want %q", name, got, want)
		}
	}
}
