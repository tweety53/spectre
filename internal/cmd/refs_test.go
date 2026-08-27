package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/testtree"
)

// twoTreesCmd builds app/ and web/ as siblings, each a spectre tree,
// declaring each other as peers, and returns app's base directory.
func twoTreesCmd(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	app := testtree.Build(t, parent, "app", map[string]string{
		"auth":  "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n",
		"plans": "# plans\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL b.\n- R2: The system SHALL c (@auth#R1).\n",
	}, "web ../web\n")
	testtree.Build(t, parent, "web", map[string]string{
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
		"web:specs/billing.md:7: R1 cites @app:auth#R1",
		"scanned: this tree, web",
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

// TestRefsRejectsUnknownCapability pins the final-review fix: a typo'd
// capability like "auht#R1" must not print "no citations of auht#R1" and
// exit 0, indistinguishable from a real empty answer. It is a wrong
// invocation, so it exits Usage (2) and names the unknown capability.
func TestRefsRejectsUnknownCapability(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", twoTreesCmd(t), "auht#R1"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d, stdout = %q", code, Usage, out.String())
	}
	if !strings.Contains(errBuf.String(), `"auht"`) {
		t.Errorf("stderr = %q, want it to name the unknown capability", errBuf.String())
	}
	if strings.Contains(out.String(), "no citations") {
		t.Errorf("stdout = %q, an unknown capability must not print a real-answer-shaped message", out.String())
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
	webPeers := filepath.Join(filepath.Dir(base), "web", "spectre", "peers")
	if err := os.WriteFile(webPeers, []byte("app\n"), 0o644); err != nil {
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
	if !strings.Contains(got, "web (unreadable:") {
		t.Errorf("missing unreadable marker in:\n%s", got)
	}
}

func TestRefsPeerUnreadablePath(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	base := twoTreesCmd(t)
	// Block traversal into web itself, a different failure mode than
	// TestRefsPeerUnreadableSpec's chmod of one file inside an otherwise
	// reachable tree: this makes Stat(web/spectre) fail with a
	// permission error rather than not-exist.
	webDir := filepath.Join(filepath.Dir(base), "web")
	if err := os.Chmod(webDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(webDir, 0o755) })

	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", base, "auth#R1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "specs/plans.md:8: R2 cites @auth#R1") {
		t.Errorf("missing local citation in:\n%s", got)
	}
	if !strings.Contains(got, "web (unreadable:") {
		t.Errorf("missing unreadable marker in:\n%s, want unreadable not not-present", got)
	}
	if strings.Contains(got, "web (not present)") {
		t.Errorf("got %q, a permission error must not be reported as not present", got)
	}
}

// TestRefsHonorsConfiguredIDPrefix pins the behaviour this task actually
// added to cmd.Refs: the <capability>#<id> argument pattern is built from
// the resolved tree's own configured id prefix, not a hardcoded "R\d+".
func TestRefsHonorsConfiguredIDPrefix(t *testing.T) {
	base := testtree.Build(t, t.TempDir(), "app", map[string]string{
		"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- REQ-1: The system SHALL a.\n- REQ-2: The system SHALL b (@auth#REQ-1).\n",
	}, "")
	if err := os.WriteFile(filepath.Join(base, "spectre", "config.md"), []byte("## Vocabulary\n- id-prefix: REQ-\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	if code := Refs([]string{"--root", base, "auth#REQ-1"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "specs/auth.md:8: REQ-2 cites @auth#REQ-1") {
		t.Errorf("missing citation in:\n%s", out.String())
	}

	out.Reset()
	errBuf.Reset()
	if code := Refs([]string{"--root", base, "auth#R1"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d (R1 does not match this tree's REQ- prefix)", code, Usage)
	}
	if !strings.Contains(errBuf.String(), `"REQ-"`) {
		t.Errorf("stderr = %q, want it to name the configured prefix", errBuf.String())
	}
}

func TestRefsPeerUnreadableSpec(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	base := twoTreesCmd(t)
	billing := filepath.Join(filepath.Dir(base), "web", "spectre", "specs", "billing.md")
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
	if !strings.Contains(got, "web (unreadable:") {
		t.Errorf("missing unreadable marker in:\n%s", got)
	}
}
