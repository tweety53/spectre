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

// TestNewRejectsArchiveID pins the final-review fix: a change id equal to
// the archive directory's own name must be rejected, since a change
// scaffolded there is invisible to list, validate and archive.
func TestNewRejectsArchiveID(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, "archive"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d, stderr = %s", code, Usage, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "archive") {
		t.Errorf("stderr = %q, want it to name the archive directory", errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "proposal.md")); !os.IsNotExist(err) {
		t.Error("scaffold files were written into the archive directory")
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive")); !os.IsNotExist(err) {
		t.Error("archive directory should not have been created")
	}
}

func TestNewRejectsInvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"escape parent", "../escape"},
		{"nested dir", "sub/dir"},
		{"parent dot dot", ".."},
		{"current dot", "."},
		{"absolute path", "/absolute"},
		{"backslash escape", "..\\escape"},
		{"empty string", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := emptyTree(t)
			var out, errBuf bytes.Buffer
			code := New([]string{"--root", base, tt.id}, &out, &errBuf)
			if code != Usage {
				t.Fatalf("exit = %d, want %d", code, Usage)
			}
			if !strings.Contains(errBuf.String(), "invalid change id") {
				t.Errorf("stderr = %q, want 'invalid change id'", errBuf.String())
			}
			// Verify nothing was created under changes/
			changesDir := filepath.Join(base, "spectre", "changes")
			entries, err := os.ReadDir(changesDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) > 0 {
				t.Errorf("created unwanted entries: %v", entries)
			}
		})
	}
}
