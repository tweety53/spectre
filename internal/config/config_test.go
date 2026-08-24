package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.Modal != "SHALL" || c.IDPrefix != "R" {
		t.Errorf("vocabulary = %q/%q", c.Modal, c.IDPrefix)
	}
	if c.SpecsDir != "specs" || c.ChangesDir != "changes" || c.Extension != ".md" {
		t.Errorf("layout = %q/%q/%q", c.SpecsDir, c.ChangesDir, c.Extension)
	}
	for _, name := range RuleNames {
		if !c.Rules[name] {
			t.Errorf("rule %q defaults off, want on", name)
		}
	}
}

func TestParseFull(t *testing.T) {
	raw := []byte(`# spectre config

## Rules
- shall-clause: off
- placeholders: error

## Vocabulary
- modal: MUST
- id-prefix: REQ-

## Layout
- specs: docs/specs
- changes: docs/changes
- extension: .markdown
`)
	c, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if c.Rules["shall-clause"] {
		t.Error("shall-clause should be off")
	}
	if !c.Rules["placeholders"] || !c.Rules["id-sequence"] {
		t.Error("unmentioned and error rules should be on")
	}
	if c.Modal != "MUST" || c.IDPrefix != "REQ-" {
		t.Errorf("vocabulary = %q/%q", c.Modal, c.IDPrefix)
	}
	if c.SpecsDir != "docs/specs" || c.ChangesDir != "docs/changes" || c.Extension != ".markdown" {
		t.Errorf("layout = %q/%q/%q", c.SpecsDir, c.ChangesDir, c.Extension)
	}
}

func TestParseRejects(t *testing.T) {
	cases := map[string]string{
		"unknown rule":      "## Rules\n- shal-clause: off\n",
		"unknown value":     "## Rules\n- shall-clause: warn\n",
		"unknown key":       "## Vocabulary\n- tone: formal\n",
		"unknown section":   "## Behaviour\n- speed: fast\n",
		"malformed bullet":  "## Rules\n- shall-clause off\n",
		"modal with spaces": "## Vocabulary\n- modal: SHALL NOT\n",
		"bad id prefix":     "## Vocabulary\n- id-prefix: 1R\n",
		"absolute specs":    "## Layout\n- specs: /etc/specs\n",
		"escaping specs":    "## Layout\n- specs: ../outside\n",
		"bad extension":     "## Layout\n- extension: md\n",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(raw)); err == nil {
				t.Fatalf("want error for %s", name)
			}
		})
	}
}

func TestParseErrorNamesTheLine(t *testing.T) {
	_, err := Parse([]byte("# spectre config\n\n## Rules\n- shall-clause: error\n- nope: error\n"))
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "config.md:5:") {
		t.Errorf("err = %v, want it to name config.md:5", err)
	}
}

func TestLoadAbsentFileIsDefault(t *testing.T) {
	c, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Modal != "SHALL" {
		t.Errorf("Modal = %q, want the default", c.Modal)
	}
}

func TestLoadReadsFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.md"), []byte("## Vocabulary\n- modal: MUST\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Modal != "MUST" {
		t.Errorf("Modal = %q, want MUST", c.Modal)
	}
}

func TestParseDuplicateRule(t *testing.T) {
	raw := []byte("## Rules\n- shall-clause: off\n- shall-clause: error\n")
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("want error for duplicate rule")
	}
	if !strings.Contains(err.Error(), "config.md:3:") {
		t.Errorf("err = %v, want it to name config.md:3 (the duplicate line)", err)
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("err = %v, want it to name line 2 (where first set)", err)
	}
}

func TestParseDuplicateVocabularyKey(t *testing.T) {
	raw := []byte("## Vocabulary\n- modal: MUST\n- modal: SHALL\n")
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("want error for duplicate vocabulary key")
	}
	if !strings.Contains(err.Error(), "config.md:3:") {
		t.Errorf("err = %v, want it to name config.md:3 (the duplicate line)", err)
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("err = %v, want it to name line 2 (where first set)", err)
	}
}

func TestParseFencedExampleIgnored(t *testing.T) {
	raw := []byte("## Vocabulary\n- modal: MUST\n\n```\n## Rules\n- shall-clause: off\n```\n")
	c, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if c.Modal != "MUST" {
		t.Errorf("Modal = %q, want MUST (unaffected by the fenced example)", c.Modal)
	}
	if !c.Rules["shall-clause"] {
		t.Error("shall-clause should still be on; the fenced example must not turn it off")
	}
}
