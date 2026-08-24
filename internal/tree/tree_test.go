package tree

import (
	"errors"
	"os"
	"path/filepath"
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
