package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/testtree"
)

func TestValidateClean(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := Validate([]string{"--root", seedTree(t)}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stdout = %s", code, out.String())
	}
	if !strings.Contains(out.String(), "no findings") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestValidateReportsFindings(t *testing.T) {
	base := seedTree(t)
	bad := filepath.Join(base, "spectre", "specs", "broken.md")
	if err := os.WriteFile(bad, []byte("# broken\n\n## Requirements\n- R2: TODO write this.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Validate([]string{"--root", base}, &out, &errBuf)
	if code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
	}
	got := out.String()
	for _, want := range []string{
		"specs/broken.md:1: missing \"## Purpose\"",
		"specs/broken.md:4: placeholder \"TODO\"",
		"specs/broken.md:4: requirement R2 has no SHALL clause",
		"specs/broken.md:4: requirement id R2 out of sequence, expected R1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestValidateUnknownChange(t *testing.T) {
	var out, errBuf bytes.Buffer
	// A typo'd change id must not read as success: it exits Usage (2), not
	// OK, and must not print "no findings".
	code := Validate([]string{"--root", seedTree(t), "typo-of-my-change-id"}, &out, &errBuf)
	if code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
	if strings.Contains(out.String(), "no findings") {
		t.Errorf("stdout = %q, must not read a missing change as success", out.String())
	}
	if !strings.Contains(errBuf.String(), "no such change \"typo-of-my-change-id\"") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestValidateReportsUnreadablePeer(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	parent := t.TempDir()
	app := testtree.Build(t, parent, "app", map[string]string{
		"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@web:billing#R1).\n",
	}, "web ../web\n")
	web := testtree.Build(t, parent, "web", map[string]string{
		"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n",
	}, "")
	// Block traversal into web so Stat(web/spectre) fails with a
	// permission error rather than not-exist — a different failure mode
	// than the existing tests, which chmod a spec file inside an otherwise
	// reachable peer tree.
	if err := os.Chmod(web, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(web, 0o755) })

	var out, errBuf bytes.Buffer
	code := Validate([]string{"--root", app}, &out, &errBuf)
	if code != Fail {
		t.Fatalf("exit = %d, want %d, stdout=%s stderr=%s", code, Fail, out.String(), errBuf.String())
	}
	if !strings.Contains(out.String(), "could not be read") {
		t.Errorf("stdout = %q, want a finding naming the peer unreadable, not not-present", out.String())
	}
}

func TestValidateSingleChange(t *testing.T) {
	base := seedTree(t)
	dir := filepath.Join(base, "spectre", "changes", "kan-1-first")
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("# kan-1-first\n\n## Why\nB.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	if code := Validate([]string{"--root", base, "kan-1-first"}, &out, &errBuf); code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(out.String(), "missing \"## What changes\"") {
		t.Errorf("stdout = %q", out.String())
	}
}

// TestValidateSingleChangeResolvesPeers pins the task 3 fix: before it,
// validate.go resolved peers only when changeID == "", so
// "spectre validate <change-id>" — the exact form /flow's implement step 1
// runs — checked every link.md against a nil peer map and silently
// reported nothing. A single-change validate must report the same link
// finding a whole-tree validate reports.
func TestValidateSingleChangeResolvesPeers(t *testing.T) {
	satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
	// Write only the satellite side by hand: a one-sided link, since the
	// canonical never links back.
	dir := filepath.Join(satBase, "spectre", "changes", "kan-9-cross")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "## Part of\n\n`can:kan-9-cross`\n\n## Branch\n\nfeature-cross\n"
	if err := os.WriteFile(filepath.Join(dir, "link.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var wholeOut, wholeErr bytes.Buffer
	if code := Validate([]string{"--root", satBase}, &wholeOut, &wholeErr); code != Fail {
		t.Fatalf("whole-tree validate exit = %d, want %d (Fail); stderr=%s", code, Fail, wholeErr.String())
	}
	if !strings.Contains(wholeOut.String(), "one-sided link") {
		t.Fatalf("whole-tree validate stdout = %q, want a one-sided link finding", wholeOut.String())
	}

	var singleOut, singleErr bytes.Buffer
	if code := Validate([]string{"--root", satBase, "kan-9-cross"}, &singleOut, &singleErr); code != Fail {
		t.Fatalf("single-change validate exit = %d, want %d (Fail); stderr=%s", code, Fail, singleErr.String())
	}
	if !strings.Contains(singleOut.String(), "one-sided link") {
		t.Fatalf("single-change validate stdout = %q, want the same one-sided link finding "+
			"a nil peer map would silently drop", singleOut.String())
	}
}
