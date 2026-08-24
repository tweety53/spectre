package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/config"
	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/tree"
)

const noModalSpec = "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: the system does a thing.\n"

func specWithoutModal() model.Spec {
	return model.Spec{
		Capability: "auth",
		Purpose:    "P.",
		Reqs:       []model.Requirement{{ID: "R1", Num: 1, Text: "the system does a thing.", Line: 7}},
	}
}

func TestShallClauseRuleOff(t *testing.T) {
	cfg := config.Default()
	if got := msgs(SpecFindings(cfg, "specs/auth.md", specWithoutModal(), []byte(noModalSpec))); !strings.Contains(got, "no SHALL clause") {
		t.Fatalf("default config should flag it, got:\n%s", got)
	}
	cfg = cfg.WithRule("shall-clause", false)
	if got := msgs(SpecFindings(cfg, "specs/auth.md", specWithoutModal(), []byte(noModalSpec))); got != "" {
		t.Errorf("rule off should be silent, got:\n%s", got)
	}
}

func TestModalVocabulary(t *testing.T) {
	cfg := config.Default()
	cfg.Modal = "MUST"
	s := model.Spec{
		Capability: "auth",
		Purpose:    "P.",
		Reqs:       []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7}},
	}
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n")
	got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw))
	if !strings.Contains(got, "no MUST clause") {
		t.Errorf("want a MUST finding, got:\n%s", got)
	}
}

func TestIDPrefixInMessages(t *testing.T) {
	cfg := config.Default()
	cfg.IDPrefix = "REQ-"
	s := model.Spec{
		Capability: "auth",
		Purpose:    "P.",
		Reqs:       []model.Requirement{{ID: "REQ-5", Num: 5, Text: "The system SHALL a.", Line: 7}},
	}
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- REQ-5: The system SHALL a.\n")
	got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw))
	if !strings.Contains(got, "expected REQ-1") {
		t.Errorf("want the configured prefix in the message, got:\n%s", got)
	}
	if strings.Contains(got, "malformed requirement bullet") {
		t.Errorf("REQ- bullets must not read as malformed:\n%s", got)
	}
}

func TestPlaceholderAndHeadingRulesOff(t *testing.T) {
	cfg := config.Default().WithRule("placeholders", false).WithRule("headings", false)
	raw := []byte("# auth\n\n## Requirements\n- R1: The system SHALL do TODO work.\n")
	s := model.Spec{Capability: "auth", Reqs: []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL do TODO work.", Line: 4}}}
	if got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw)); got != "" {
		t.Errorf("both rules off should be silent, got:\n%s", got)
	}
}

func TestTaskSequenceRuleOff(t *testing.T) {
	cfg := config.Default().WithRule("task-sequence", false)
	raw := []byte("# Tasks\n\n- [ ] 5. c\n")
	ts := []model.Task{{Num: 5, Text: "c", Line: 3}}
	if got := msgs(TaskFindings(cfg, "changes/x/tasks.md", raw, ts)); got != "" {
		t.Errorf("rule off should be silent, got:\n%s", got)
	}
}

func TestMalformedBulletRuleOff(t *testing.T) {
	cfg := config.Default()
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1 The system SHALL a.\n")
	s := model.Spec{Capability: "auth"}
	if got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw)); !strings.Contains(got, "malformed requirement bullet") {
		t.Fatalf("default config should flag it, got:\n%s", got)
	}
	cfg = cfg.WithRule("malformed-bullet", false)
	if got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw)); got != "" {
		t.Errorf("rule off should be silent, got:\n%s", got)
	}
}

func TestIDSequenceRuleOff(t *testing.T) {
	cfg := config.Default()
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n- R1: The system SHALL b.\n- R5: The system SHALL c.\n")
	s := model.Spec{
		Capability: "auth",
		Purpose:    "P.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7},
			{ID: "R1", Num: 1, Text: "The system SHALL b.", Line: 8},
			{ID: "R5", Num: 5, Text: "The system SHALL c.", Line: 9},
		},
	}
	got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw))
	for _, want := range []string{"duplicate requirement id R1", "out of sequence"} {
		if !strings.Contains(got, want) {
			t.Fatalf("default config should flag %q, got:\n%s", want, got)
		}
	}
	cfg = cfg.WithRule("id-sequence", false)
	if got := msgs(SpecFindings(cfg, "specs/auth.md", s, raw)); got != "" {
		t.Errorf("rule off should be silent, got:\n%s", got)
	}
}

func TestRefsRuleOff(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@auth#R9).\n"
	if err := os.WriteFile(filepath.Join(root, "specs", "auth.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	tr := tree.At(root, cfg)
	got, err := Structural(tr, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msgs(got), "no requirement R9") {
		t.Fatalf("default config should flag the bad ref, got:\n%s", msgs(got))
	}

	tr = tree.At(root, cfg.WithRule("refs", false))
	got, err = Structural(tr, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(msgs(got), "no requirement R9") {
		t.Errorf("rule off should silence the ref finding, got:\n%s", msgs(got))
	}
}
