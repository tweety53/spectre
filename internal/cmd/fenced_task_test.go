package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFencedTaskExampleDoesNotFoolListOrValidate pins the
// second-round final-review fix: a tasks.md with two real, finished tasks
// and a fenced example containing a task-shaped line must not let that
// fenced line count as a real task anywhere. The fenced line's number (1)
// deliberately reuses a real task's number, so that on the pre-fix
// fence-unaware parser it does not just inflate the count — it also trips
// check.TaskFindings' duplicate-task-number check, making both
// assertions below genuine regressions on this one fixture: pre-fix, list
// mis-reports "c1  2/3", and validate reports a false "duplicate task
// number 1" finding.
func TestFencedTaskExampleDoesNotFoolListOrValidate(t *testing.T) {
	base := emptyTree(t)
	dir := filepath.Join(base, "spectre", "changes", "c1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	proposal := "# c1\n\n## Why\nBecause.\n\n## What changes\n- a thing\n"
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte(proposal), 0o644); err != nil {
		t.Fatal(err)
	}
	tasks := "# c1\n\n" +
		"- [x] 1. Real task one\n" +
		"- [x] 2. Real task two\n" +
		"```\n" +
		"- [ ] 1. Not a real task, just an example\n" +
		"```\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "-A"},
	} {
		c := exec.Command("git", args...)
		c.Dir = base
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}

	// Assertion 1: list counts 2/2, not 2/3.
	var listOut, listErr bytes.Buffer
	if code := List([]string{"--root", base}, &listOut, &listErr); code != OK {
		t.Fatalf("list exit = %d, stderr = %s", code, listErr.String())
	}
	if !strings.Contains(listOut.String(), "c1  2/2") {
		t.Errorf("list stdout = %q, want progress 2/2 (the fenced example line must not count)", listOut.String())
	}

	// Assertion 2: validate reports no findings for it.
	var valOut, valErr bytes.Buffer
	if code := Validate([]string{"--root", base}, &valOut, &valErr); code != OK {
		t.Fatalf("validate exit = %d, stdout = %s, stderr = %s", code, valOut.String(), valErr.String())
	}
	if !strings.Contains(valOut.String(), "no findings") {
		t.Errorf("validate stdout = %q, want no findings from the fenced example", valOut.String())
	}
}
