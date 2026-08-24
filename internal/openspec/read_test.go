package openspec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSpec = `# myflow-build-green Specification

## Purpose
Tasks declare whether they leave the build green.

## Requirements

### Requirement: Every task in a plan declares its build state

The plan SHALL tag every task with green or red.

#### Scenario: A task cannot leave the build green

- **WHEN** a task cannot leave the project building
- **THEN** the plan tags it red

### Requirement: A guard enforces the tags

A guard script SHALL fail the run when a task has no tag.
`

func TestReadSpec(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(p, []byte(sampleSpec), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSpec(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Purpose != "Tasks declare whether they leave the build green." {
		t.Errorf("Purpose = %q", got.Purpose)
	}
	if len(got.Reqs) != 2 {
		t.Fatalf("len(Reqs) = %d, want 2", len(got.Reqs))
	}
	if got.Reqs[0].Title != "Every task in a plan declares its build state" {
		t.Errorf("title = %q", got.Reqs[0].Title)
	}
	body := strings.Join(got.Reqs[0].Body, "\n")
	for _, want := range []string{
		"The plan SHALL tag every task with green or red.",
		"#### Scenario: A task cannot leave the build green",
		"- **WHEN** a task cannot leave the project building",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q:\n%s", want, body)
		}
	}
	if len(got.Reqs[1].Body) == 0 {
		t.Error("second requirement lost its body")
	}
}

func TestConvertTasksFlattens(t *testing.T) {
	raw := []byte("## 1. Implementation\n\n- [x] 1.1 Write the parser\n- [ ] 1.2 Write the renderer\n- [ ] 2. Ship it\n")
	got := ConvertTasks(raw)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if !got[0].Done || got[0].Text != "Write the parser" {
		t.Errorf("task 1 = %+v", got[0])
	}
	if got[2].Text != "Ship it" || got[2].Done {
		t.Errorf("task 3 = %+v", got[2])
	}
}

func TestConvertProposalRenamesHeading(t *testing.T) {
	got := string(ConvertProposal([]byte("# x\n\n## Why\nB.\n\n## What Changes\n- a thing\n")))
	if strings.Contains(got, "## What Changes") || !strings.Contains(got, "## What changes") {
		t.Errorf("got:\n%s", got)
	}
}

func TestReadChangesFindsDeltas(t *testing.T) {
	base := t.TempDir()
	mk := func(rel, body string) {
		p := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("kan-9/proposal.md", "# kan-9\n")
	mk("kan-9/specs/auth/spec.md", "## ADDED Requirements\n")
	mk("archive/kan-8/proposal.md", "# kan-8\n")

	got, err := ReadChanges(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "kan-9" {
		t.Fatalf("changes = %+v", got)
	}
	if len(got[0].Deltas) != 1 || !strings.HasSuffix(got[0].Deltas[0], filepath.Join("auth", "spec.md")) {
		t.Errorf("deltas = %v", got[0].Deltas)
	}
}
