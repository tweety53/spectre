package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/tree"
)

func msgs(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.String())
		b.WriteString("\n")
	}
	return b.String()
}

func TestSpecFindingsClean(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n")
	s := model.Spec{Capability: "auth", Purpose: "P.", Reqs: []model.Requirement{
		{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7},
	}}
	if got := SpecFindings("specs/auth.md", s, raw); len(got) != 0 {
		t.Errorf("want clean, got:\n%s", msgs(got))
	}
}

func TestSpecFindingsMissingHeading(t *testing.T) {
	raw := []byte("# auth\n\n## Requirements\n- R1: The system SHALL a.\n")
	s := model.Spec{Capability: "auth", Reqs: []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 4}}}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	if !strings.Contains(got, "missing \"## Purpose\"") {
		t.Errorf("got:\n%s", got)
	}
}

func TestSpecFindingsPlaceholder(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nTBD\n\n## Requirements\n- R1: The system SHALL a.\n")
	s := model.Spec{Capability: "auth", Purpose: "TBD", Reqs: []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7}}}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	if !strings.Contains(got, "specs/auth.md:4: placeholder \"TBD\"") {
		t.Errorf("got:\n%s", got)
	}
}

func TestSpecFindingsRequirementRules(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: no modal verb here.\n- R1: The system SHALL b.\n- R5: The system SHALL c.\n")
	s := model.Spec{Capability: "auth", Purpose: "P.", Reqs: []model.Requirement{
		{ID: "R1", Num: 1, Text: "no modal verb here.", Line: 7},
		{ID: "R1", Num: 1, Text: "The system SHALL b.", Line: 8},
		{ID: "R5", Num: 5, Text: "The system SHALL c.", Line: 9},
	}}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	for _, want := range []string{
		"specs/auth.md:7: requirement R1 has no SHALL clause",
		"specs/auth.md:8: duplicate requirement id R1",
		"specs/auth.md:9: requirement id R5 out of sequence, expected R3",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSpecFindingsMalformedBullet(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1 The system SHALL a.\n")
	s := model.Spec{Capability: "auth", Purpose: "P."}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	if !strings.Contains(got, "specs/auth.md:7: malformed requirement bullet") {
		t.Errorf("got:\n%s", got)
	}
}

func TestSpecFindingsHeadingInFence(t *testing.T) {
	raw := []byte("# auth\n\n```\n## Purpose\n```\n\n## Requirements\n- R1: The system SHALL a.\n")
	s := model.Spec{Capability: "auth", Reqs: []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 8}}}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	if !strings.Contains(got, "missing \"## Purpose\"") {
		t.Errorf("got:\n%s", got)
	}
}

func TestSpecFindingsHeadingLastLineNoNewline(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements")
	s := model.Spec{Capability: "auth", Purpose: "P."}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	if strings.Contains(got, "missing \"## Requirements\"") {
		t.Errorf("got:\n%s", got)
	}
}

func TestSpecFindingsMalformedBulletInFenceIgnored(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n\n```\n- R1 no colon\n```\n")
	s := model.Spec{Capability: "auth", Purpose: "P.", Reqs: []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7}}}
	got := msgs(SpecFindings("specs/auth.md", s, raw))
	if strings.Contains(got, "malformed requirement bullet") {
		t.Errorf("got:\n%s", got)
	}
}

func TestStructuralMissingTasks(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	changeDir := filepath.Join(root, "changes", "x")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# x\n\n## Why\nB.\n\n## What changes\nC.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := tree.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Structural(tr, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msgs(got), "changes/x/tasks.md:1: missing tasks.md") {
		t.Errorf("got:\n%s", msgs(got))
	}
}

func TestProposalFindings(t *testing.T) {
	got := msgs(ProposalFindings("changes/x/proposal.md", []byte("# x\n\n## Why\nBecause.\n")))
	if !strings.Contains(got, "missing \"## What changes\"") {
		t.Errorf("got:\n%s", got)
	}
	clean := ProposalFindings("changes/x/proposal.md", []byte("# x\n\n## Why\nB.\n\n## What changes\nC.\n"))
	if len(clean) != 0 {
		t.Errorf("want clean, got:\n%s", msgs(clean))
	}
}

func TestTaskFindings(t *testing.T) {
	raw := []byte("# Tasks\n\n- [x] 1. a\n- [ ] 1. b\n- [ ] 5. c\n")
	ts := []model.Task{{Num: 1, Text: "a", Line: 3}, {Num: 1, Text: "b", Line: 4}, {Num: 5, Text: "c", Line: 5}}
	got := msgs(TaskFindings("changes/x/tasks.md", raw, ts))
	for _, want := range []string{
		"changes/x/tasks.md:4: duplicate task number 1",
		"changes/x/tasks.md:5: task number 5 out of sequence, expected 3",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestTaskFindingsMalformedLine(t *testing.T) {
	raw := []byte("# Tasks\n\n- [ ] Write the parser\n")
	got := msgs(TaskFindings("changes/x/tasks.md", raw, nil))
	if !strings.Contains(got, "changes/x/tasks.md:3: malformed task line") {
		t.Errorf("got:\n%s", got)
	}
}
