package tree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePeerNotDeclared(t *testing.T) {
	rp := ResolvePeer(map[string]string{}, "gymie")
	if rp.Resolution != PeerNotDeclared {
		t.Errorf("Resolution = %v, want PeerNotDeclared", rp.Resolution)
	}
}

func TestResolvePeerNotPresent(t *testing.T) {
	parent := t.TempDir()
	rp := ResolvePeer(map[string]string{"gymie": filepath.Join(parent, "gymie", "spectre")}, "gymie")
	if rp.Resolution != PeerNotPresent {
		t.Errorf("Resolution = %v, want PeerNotPresent", rp.Resolution)
	}
}

func TestResolvePeerUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	parent := t.TempDir()
	path := makeTree(t, filepath.Join(parent, "gymie"))
	root := filepath.Join(path, "spectre")
	// Block traversal into path so Stat(root) fails with a permission error
	// rather than not-exist.
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })

	rp := ResolvePeer(map[string]string{"gymie": root}, "gymie")
	if rp.Resolution != PeerUnreadable {
		t.Errorf("Resolution = %v, want PeerUnreadable", rp.Resolution)
	}
	if rp.Err == nil {
		t.Error("Err = nil, want the underlying open error")
	}
}

func TestResolvePeerFound(t *testing.T) {
	parent := t.TempDir()
	path := makeTree(t, filepath.Join(parent, "gymie"))
	root := filepath.Join(path, "spectre")

	rp := ResolvePeer(map[string]string{"gymie": root}, "gymie")
	if rp.Resolution != PeerFound {
		t.Fatalf("Resolution = %v, want PeerFound", rp.Resolution)
	}
	if rp.Tree == nil || rp.Tree.Root != root {
		t.Errorf("Tree = %+v, want Root %q", rp.Tree, root)
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
