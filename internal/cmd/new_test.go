package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func emptyTree(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	for _, sub := range []string{"specs", "changes"} {
		if err := os.MkdirAll(filepath.Join(base, "spectre", sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return base
}

func TestNewScaffolds(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer

	if code := New([]string{"--root", base, "kan-7-add-picker"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}

	dir := filepath.Join(base, "spectre", "changes", "kan-7-add-picker")
	proposal, err := os.ReadFile(filepath.Join(dir, "proposal.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# kan-7-add-picker", "## Why", "## What changes"} {
		if !strings.Contains(string(proposal), want) {
			t.Errorf("proposal missing %q:\n%s", want, proposal)
		}
	}
	tasks, err := os.ReadFile(filepath.Join(dir, "tasks.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(tasks) != "# Tasks\n\n" {
		t.Errorf("tasks.md = %q", tasks)
	}
	if !strings.Contains(out.String(), "kan-7-add-picker") {
		t.Errorf("stdout = %q, want the change id", out.String())
	}
}

func TestNewRefusesExisting(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, "kan-7"}, &out, &errBuf); code != OK {
		t.Fatalf("first run exit = %d", code)
	}
	if code := New([]string{"--root", base, "kan-7"}, &out, &errBuf); code != Fail {
		t.Fatalf("second run exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(errBuf.String(), "already exists") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestNewRequiresID(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", emptyTree(t)}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
}
