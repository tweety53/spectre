package tree

import (
	"os"
	"path/filepath"
	"testing"
)

// treeWithConfig writes a tree whose layout and id prefix are non-default.
func treeWithConfig(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	mk := func(rel, body string) {
		t.Helper()
		p := filepath.Join(base, "spectre", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("config.md", "# spectre config\n\n## Vocabulary\n- id-prefix: REQ-\n\n## Layout\n- specs: docs/specs\n- changes: docs/changes\n- extension: .markdown\n")
	mk("docs/specs/auth.markdown", "# auth\n\n## Purpose\nP.\n\n## Requirements\n- REQ-1: The system SHALL a.\n- REQ-2: The system SHALL b (@auth#REQ-1).\n")
	mk("docs/changes/kan-1/tasks.md", "# Tasks\n\n- [x] 1. One\n")
	return base
}

func TestConfiguredLayout(t *testing.T) {
	tr, err := Find(treeWithConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(tr.SpecsDir()) != "specs" || filepath.Base(filepath.Dir(tr.SpecsDir())) != "docs" {
		t.Errorf("SpecsDir = %q", tr.SpecsDir())
	}
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 || specs[0].Capability != "auth" {
		t.Fatalf("specs = %+v", specs)
	}
	changes, err := tr.Changes()
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].ID != "kan-1" {
		t.Errorf("changes = %+v", changes)
	}
}

func TestConfiguredIDPrefix(t *testing.T) {
	tr, err := Find(treeWithConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	s, err := tr.Spec("auth")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Reqs) != 2 {
		t.Fatalf("len(Reqs) = %d, want 2", len(s.Reqs))
	}
	if s.Reqs[0].ID != "REQ-1" || s.Reqs[0].Num != 1 {
		t.Errorf("req 1 = %q/%d", s.Reqs[0].ID, s.Reqs[0].Num)
	}
	if len(s.Reqs[1].Refs) != 1 || s.Reqs[1].Refs[0].ID != "REQ-1" {
		t.Errorf("refs = %+v", s.Reqs[1].Refs)
	}
}

func TestDefaultTreeStillWorks(t *testing.T) {
	tr := seed(t) // the fixture from read_test.go: default layout, R-prefixed ids
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 || specs[0].Reqs[0].ID != "R1" {
		t.Errorf("default tree changed shape: %+v", specs)
	}
}

func TestConfiguredExtensionExcludesOtherFiles(t *testing.T) {
	base := treeWithConfig(t)
	stray := filepath.Join(base, "spectre", "docs", "specs", "note.md")
	if err := os.WriteFile(stray, []byte("not a spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := Find(base)
	if err != nil {
		t.Fatal(err)
	}
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 || specs[0].Capability != "auth" {
		t.Errorf("specs = %+v, want only auth.markdown — a .md file must not be read as a spec in a .markdown tree", specs)
	}
}

func TestOpenLoadsConfigLikeFind(t *testing.T) {
	base := treeWithConfig(t)
	tr, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Cfg.IDPrefix != "REQ-" || tr.Cfg.Extension != ".markdown" {
		t.Fatalf("Cfg = %+v, want the tree's own config.md loaded", tr.Cfg)
	}
	s, err := tr.Spec("auth")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Reqs) != 2 || s.Reqs[0].ID != "REQ-1" {
		t.Errorf("Open-resolved spec = %+v, want the same result Find gives", s)
	}
}

func TestBadConfigIsAnError(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "spectre", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "spectre", "config.md"), []byte("## Rules\n- nope: error\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Find(base); err == nil {
		t.Fatal("want an error for an invalid config.md")
	}
}
