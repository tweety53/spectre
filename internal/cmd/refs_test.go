package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/testtree"
)

// twoTreesCmd builds app/ and gymie/ as siblings, each a spectre tree,
// declaring each other as peers, and returns app's base directory.
func twoTreesCmd(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	app := testtree.Build(t, parent, "app", map[string]string{
		"auth":  "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n",
		"plans": "# plans\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL b.\n- R2: The system SHALL c (@auth#R1).\n",
	}, "gymie ../gymie\n")
	testtree.Build(t, parent, "gymie", map[string]string{
		"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d (@app:auth#R1).\n",
	}, "app ../app\n")
	return app
}

func TestRefsFindsCitations(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", twoTreesCmd(t), "auth#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{
		"specs/plans.md:8: R2 cites @auth#R1",
		"gymie:specs/billing.md:7: R1 cites @app:auth#R1",
		"scanned: this tree, gymie",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRefsNoCitations(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", twoTreesCmd(t), "plans#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out.String(), "no citations") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestRefsReportsAbsentPeer(t *testing.T) {
	base := twoTreesCmd(t)
	if err := os.WriteFile(filepath.Join(base, "spectre", "peers"), []byte("ghost ../ghost\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", base, "auth#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out.String(), "ghost (not present)") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestRefsBadArgument(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", twoTreesCmd(t), "auth"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
	if !strings.Contains(errBuf.String(), "<capability>#<id>") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRefsPeerBadPeersFile(t *testing.T) {
	base := twoTreesCmd(t)
	gymiePeers := filepath.Join(filepath.Dir(base), "gymie", "spectre", "peers")
	if err := os.WriteFile(gymiePeers, []byte("app\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", base, "auth#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "specs/plans.md:8: R2 cites @auth#R1") {
		t.Errorf("missing local citation in:\n%s", got)
	}
	if !strings.Contains(got, "gymie (unreadable:") {
		t.Errorf("missing unreadable marker in:\n%s", got)
	}
}

func TestRefsPeerUnreadablePath(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	base := twoTreesCmd(t)
	// Block traversal into gymie itself, a different failure mode than
	// TestRefsPeerUnreadableSpec's chmod of one file inside an otherwise
	// reachable tree: this makes Stat(gymie/spectre) fail with a
	// permission error rather than not-exist.
	gymieDir := filepath.Join(filepath.Dir(base), "gymie")
	if err := os.Chmod(gymieDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(gymieDir, 0o755) })

	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", base, "auth#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "specs/plans.md:8: R2 cites @auth#R1") {
		t.Errorf("missing local citation in:\n%s", got)
	}
	if !strings.Contains(got, "gymie (unreadable:") {
		t.Errorf("missing unreadable marker in:\n%s, want unreadable not not-present", got)
	}
	if strings.Contains(got, "gymie (not present)") {
		t.Errorf("got %q, a permission error must not be reported as not present", got)
	}
}

func TestRefsPeerUnreadableSpec(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	base := twoTreesCmd(t)
	billing := filepath.Join(filepath.Dir(base), "gymie", "spectre", "specs", "billing.md")
	if err := os.Chmod(billing, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(billing, 0o644) })
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", base, "auth#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "specs/plans.md:8: R2 cites @auth#R1") {
		t.Errorf("missing local citation in:\n%s", got)
	}
	if !strings.Contains(got, "gymie (unreadable:") {
		t.Errorf("missing unreadable marker in:\n%s", got)
	}
}
