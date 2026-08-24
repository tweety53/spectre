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
	if code := Archive([]string{"--root", gitTree(t), "nope"}, &out, &errBuf); code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
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
