package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitTree seeds a tree inside a real git repository with everything added.
func gitTree(t *testing.T) string {
	t.Helper()
	base := seedTree(t)
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
	return base
}

func TestArchiveMovesFinishedChange(t *testing.T) {
	base := gitTree(t)
	tasks := filepath.Join(base, "spectre", "changes", "kan-1-first", "tasks.md")
	if err := os.WriteFile(tasks, []byte("# Tasks\n\n- [x] 1. One\n- [x] 2. Two\n- [x] 3. Three\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "kan-1-first"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "kan-1-first", "tasks.md")); err != nil {
		t.Fatalf("change not archived: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "kan-1-first")); !os.IsNotExist(err) {
		t.Error("original folder still present")
	}
}

func TestArchiveRefusesUncheckedTasks(t *testing.T) {
	base := gitTree(t)
	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "kan-1-first"}, &out, &errBuf); code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(errBuf.String(), "2 of 3 tasks are unchecked") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestArchiveForce(t *testing.T) {
	base := gitTree(t)
	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "--force", "kan-1-first"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "kan-1-first")); err != nil {
		t.Fatalf("change not archived: %v", err)
	}
}

func TestArchiveUnknownChange(t *testing.T) {
	var out, errBuf bytes.Buffer
	// An unknown change id is a wrong invocation, not a findings/content
	// refusal, so it exits Usage (2), not Fail (1).
	if code := Archive([]string{"--root", gitTree(t), "nope"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
	if !strings.Contains(errBuf.String(), "no open change \"nope\"") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// gitTreeNoTasks seeds a git repo with a change folder that holds only
// proposal.md — no tasks.md at all.
func gitTreeNoTasks(t *testing.T) string {
	t.Helper()
	base := emptyTree(t)
	dir := filepath.Join(base, "spectre", "changes", "kan-2-notasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	proposal := "# kan-2-notasks\n\n## Why\nBecause.\n\n## What changes\n- a thing\n"
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte(proposal), 0o644); err != nil {
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
	return base
}

func TestArchiveRefusesMissingTasksFile(t *testing.T) {
	base := gitTreeNoTasks(t)
	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "kan-2-notasks"}, &out, &errBuf); code != Fail {
		t.Fatalf("exit = %d, want %d, stderr = %s", code, Fail, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "kan-2-notasks") || !strings.Contains(errBuf.String(), "tasks.md") {
		t.Errorf("stderr = %q, want it to name the change and tasks.md", errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "kan-2-notasks")); err != nil {
		t.Fatalf("change folder should not have moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "kan-2-notasks")); !os.IsNotExist(err) {
		t.Error("change folder should not be archived")
	}
}

func TestArchiveRefusesDestinationCollision(t *testing.T) {
	base := gitTree(t)
	tasks := filepath.Join(base, "spectre", "changes", "kan-1-first", "tasks.md")
	if err := os.WriteFile(tasks, []byte("# Tasks\n\n- [x] 1. One\n- [x] 2. Two\n- [x] 3. Three\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(base, "spectre", "changes", "archive", "kan-1-first")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(dst, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("already archived\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	// --force must not override a destination collision: it means "archive
	// despite unfinished work", not "overwrite an existing archived change".
	if code := Archive([]string{"--root", base, "--force", "kan-1-first"}, &out, &errBuf); code != Fail {
		t.Fatalf("exit = %d, want %d, stderr = %s", code, Fail, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), dst) {
		t.Errorf("stderr = %q, want it to name the destination %q", errBuf.String(), dst)
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "kan-1-first")); err != nil {
		t.Fatalf("source folder should not have moved: %v", err)
	}
	if body, err := os.ReadFile(sentinel); err != nil || string(body) != "already archived\n" {
		t.Fatalf("existing archived folder was disturbed: body=%q, err=%v", body, err)
	}
	if _, err := os.Stat(filepath.Join(dst, "kan-1-first")); !os.IsNotExist(err) {
		t.Error("change was nested into the existing archived folder instead of being refused")
	}
}

// gitTreeEmptyTasks seeds a git repo with a change folder holding exactly
// what `spectre new` writes: proposal.md and a tasks.md with zero tasks —
// present, unlike gitTreeNoTasks, but empty.
func gitTreeEmptyTasks(t *testing.T) string {
	t.Helper()
	base := emptyTree(t)
	dir := filepath.Join(base, "spectre", "changes", "kan-3-empty")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	proposal := "# kan-3-empty\n\n## Why\nBecause.\n\n## What changes\n- a thing\n"
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte(proposal), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("# Tasks\n\n"), 0o644); err != nil {
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
	return base
}

// TestArchiveRefusesNoTasksAtAll pins the final-review fix: `spectre new X
// && spectre archive X` must not archive an empty change immediately. The
// Task 5/11 guard only caught a MISSING tasks.md, but `new` always writes
// one with zero tasks, which reads as zero unchecked and slipped past it.
func TestArchiveRefusesNoTasksAtAll(t *testing.T) {
	base := gitTreeEmptyTasks(t)
	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "kan-3-empty"}, &out, &errBuf); code != Fail {
		t.Fatalf("exit = %d, want %d, stderr = %s", code, Fail, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "kan-3-empty") || !strings.Contains(errBuf.String(), "no tasks") {
		t.Errorf("stderr = %q, want it to name the change and say it has no tasks", errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "kan-3-empty")); err != nil {
		t.Fatalf("change folder should not have moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "kan-3-empty")); !os.IsNotExist(err) {
		t.Error("change with no tasks should not be archived")
	}
}

func TestArchiveForceOverridesNoTasksAtAll(t *testing.T) {
	base := gitTreeEmptyTasks(t)
	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "--force", "kan-3-empty"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "kan-3-empty", "proposal.md")); err != nil {
		t.Fatalf("change not archived: %v", err)
	}
}

// TestFencedTaskExampleDoesNotFoolListValidateOrArchive pins the
// second-round final-review fix: a tasks.md with two real, finished tasks
// and a fenced example containing a task-shaped line must not let that
// fenced line count as a real task anywhere. The fenced line's number (1)
// deliberately reuses a real task's number, so that on the pre-fix
// fence-unaware parser it does not just inflate the count — it also trips
// check.TaskFindings' duplicate-task-number check, making all three
// assertions below genuine regressions on this one fixture: pre-fix, list
// mis-reports "c1  2/3", validate reports a false "duplicate task number 1"
// finding, and archive refuses a genuinely finished change with "1 of 3
// tasks are unchecked" — whose only escape, --force, also disarms the
// zero-task and unchecked-task content guards. This is the case that
// matters most: archive must accept the change WITHOUT --force.
func TestFencedTaskExampleDoesNotFoolListValidateOrArchive(t *testing.T) {
	base := emptyTree(t)
	dir := filepath.Join(base, "spectre", "changes", "c1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	proposal := "# c1\n\n## Why\nBecause.\n\n## What changes\n- a thing\n"
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte(proposal), 0o644); err != nil {
		t.Fatal(err)
	}
	tasks := "# Tasks\n\n" +
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

	// Assertion 3 (the one that matters most): archive accepts the change
	// WITHOUT --force.
	var arcOut, arcErr bytes.Buffer
	if code := Archive([]string{"--root", base, "c1"}, &arcOut, &arcErr); code != OK {
		t.Fatalf("archive exit = %d, stderr = %s", code, arcErr.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "c1")); err != nil {
		t.Fatalf("change not archived: %v", err)
	}
}

func TestArchiveForceOverridesMissingTasksFile(t *testing.T) {
	base := gitTreeNoTasks(t)
	var out, errBuf bytes.Buffer
	if code := Archive([]string{"--root", base, "--force", "kan-2-notasks"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "kan-2-notasks", "proposal.md")); err != nil {
		t.Fatalf("change not archived: %v", err)
	}
}
