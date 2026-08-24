package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// openspecTree builds a small OpenSpec tree and returns its directory.
func openspecTree(t *testing.T) string {
	t.Helper()
	base := filepath.Join(t.TempDir(), "openspec")
	mk := func(rel, body string) {
		t.Helper()
		p := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("specs/auth/spec.md", "# auth Specification\n\n## Purpose\nSessions.\n\n## Requirements\n\n### Requirement: The system SHALL refresh tokens\n\nRefresh happens before expiry.\n\n#### Scenario: token near expiry\n\n- **WHEN** the token is near expiry\n- **THEN** it is refreshed\n")
	mk("changes/kan-9-add-picker/proposal.md", "# kan-9\n\n## Why\nBecause.\n\n## What Changes\n- adds a picker\n")
	mk("changes/kan-9-add-picker/tasks.md", "## 1. Work\n\n- [x] 1.1 First\n- [ ] 1.2 Second\n")
	mk("changes/archive/kan-8-old/proposal.md", "# kan-8\n\n## Why\nHistory.\n\n## What Changes\n- done\n")
	return base
}

func TestMigrateConvertsSpecsAndChanges(t *testing.T) {
	src := openspecTree(t)
	out := filepath.Join(t.TempDir(), "spectre")

	var stdout, stderr bytes.Buffer
	if code := Migrate([]string{"--out", out, src}, &stdout, &stderr); code != OK {
		t.Fatalf("exit = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}

	spec, err := os.ReadFile(filepath.Join(out, "specs", "auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# auth\n",
		"## Purpose\nSessions.\n",
		"- R1: The system SHALL refresh tokens\n",
		"  Refresh happens before expiry.\n",
		"  #### Scenario: token near expiry\n",
		"  - **WHEN** the token is near expiry\n",
	} {
		if !strings.Contains(string(spec), want) {
			t.Errorf("spec missing %q:\n%s", want, spec)
		}
	}

	tasks, err := os.ReadFile(filepath.Join(out, "changes", "kan-9-add-picker", "tasks.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(tasks) != "# Tasks\n\n- [x] 1. First\n- [ ] 2. Second\n" {
		t.Errorf("tasks = %q", tasks)
	}

	proposal, err := os.ReadFile(filepath.Join(out, "changes", "kan-9-add-picker", "proposal.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(proposal), "## What changes") {
		t.Errorf("proposal = %s", proposal)
	}

	if _, err := os.Stat(filepath.Join(out, "changes", "archive", "kan-8-old", "proposal.md")); err != nil {
		t.Errorf("archived change not copied: %v", err)
	}
	if !strings.Contains(stdout.String(), "1 spec") {
		t.Errorf("report = %q", stdout.String())
	}
}

func TestMigrateIsNonDestructive(t *testing.T) {
	src := openspecTree(t)
	out := filepath.Join(t.TempDir(), "spectre")
	var stdout, stderr bytes.Buffer
	if code := Migrate([]string{"--out", out, src}, &stdout, &stderr); code != OK {
		t.Fatalf("exit = %d", code)
	}
	if _, err := os.Stat(filepath.Join(src, "specs", "auth", "spec.md")); err != nil {
		t.Errorf("source tree was modified: %v", err)
	}
	if code := Migrate([]string{"--out", out, src}, &stdout, &stderr); code != Fail {
		t.Fatalf("second run exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(stderr.String(), "already exists") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestMigrateWarnsOnDeltasAndExitsOne(t *testing.T) {
	src := openspecTree(t)
	delta := filepath.Join(src, "changes", "kan-9-add-picker", "specs", "auth", "spec.md")
	if err := os.MkdirAll(filepath.Dir(delta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(delta, []byte("## ADDED Requirements\n\n### Requirement: new thing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "spectre")

	var stdout, stderr bytes.Buffer
	if code := Migrate([]string{"--out", out, src}, &stdout, &stderr); code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
	}
	copied := filepath.Join(out, "changes", "kan-9-add-picker", "deltas-auth.md")
	if _, err := os.Stat(copied); err != nil {
		t.Fatalf("delta not copied verbatim: %v", err)
	}
	if !strings.Contains(stdout.String(), "deltas-auth.md") {
		t.Errorf("report missing the delta warning:\n%s", stdout.String())
	}
}

func TestMigrateWarnsOnMissingSHALL(t *testing.T) {
	src := openspecTree(t)
	if err := os.WriteFile(filepath.Join(src, "specs", "auth", "spec.md"),
		[]byte("# auth Specification\n\n## Purpose\nP.\n\n## Requirements\n\n### Requirement: token refresh\n\nBody prose.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "spectre")
	var stdout, stderr bytes.Buffer
	if code := Migrate([]string{"--out", out, src}, &stdout, &stderr); code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(stdout.String(), "no SHALL clause") {
		t.Errorf("report = %s", stdout.String())
	}
}

func TestMigrateForceClearsStaleFiles(t *testing.T) {
	base := filepath.Join(t.TempDir(), "openspec")
	mk := func(rel, body string) {
		t.Helper()
		p := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("specs/auth/spec.md", "# auth Specification\n\n## Purpose\nSessions.\n\n## Requirements\n\n### Requirement: The system SHALL refresh tokens\n\nRefresh happens before expiry.\n")
	mk("specs/beta/spec.md", "# beta Specification\n\n## Purpose\nOther.\n\n## Requirements\n\n### Requirement: The system SHALL do a thing\n\nBody.\n")

	out := filepath.Join(t.TempDir(), "spectre")
	var stdout, stderr bytes.Buffer
	if code := Migrate([]string{"--out", out, base}, &stdout, &stderr); code != OK {
		t.Fatalf("first migrate exit = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(out, "specs", "beta.md")); err != nil {
		t.Fatalf("expected specs/beta.md after first migrate: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(base, "specs", "beta")); err != nil {
		t.Fatal(err)
	}

	stdout.Reset()
	stderr.Reset()
	if code := Migrate([]string{"--out", out, "--force", base}, &stdout, &stderr); code != OK {
		t.Fatalf("forced migrate exit = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(out, "specs", "beta.md")); !os.IsNotExist(err) {
		t.Errorf("stale specs/beta.md survived --force: err = %v", err)
	}
	if !strings.Contains(stdout.String(), "cleared") {
		t.Errorf("report does not name what --force cleared:\n%s", stdout.String())
	}
}

func TestMigrateWarnsOnPlaceholderPurpose(t *testing.T) {
	src := openspecTree(t)
	if err := os.WriteFile(filepath.Join(src, "specs", "auth", "spec.md"),
		[]byte("# auth Specification\n\n## Purpose\nTBD\n\n## Requirements\n\n### Requirement: The system SHALL refresh tokens\n\nRefresh happens before expiry.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "spectre")
	var stdout, stderr bytes.Buffer
	if code := Migrate([]string{"--out", out, src}, &stdout, &stderr); code != Fail {
		t.Fatalf("exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(stdout.String(), `placeholder "TBD"`) {
		t.Errorf("report missing the placeholder finding:\n%s", stdout.String())
	}
}
