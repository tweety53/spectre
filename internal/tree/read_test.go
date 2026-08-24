package tree

import (
	"os"
	"path/filepath"
	"testing"
)

func seed(t *testing.T) *Tree {
	t.Helper()
	base := makeTree(t, t.TempDir())
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(base, "spectre", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("specs/auth.md", "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n- R2: The system SHALL b.\n")
	write("specs/plans.md", "# plans\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL c.\n")
	write("changes/kan-2-second/tasks.md", "# Tasks\n\n- [x] 1. One\n- [ ] 2. Two\n")
	write("changes/kan-1-first/tasks.md", "# Tasks\n\n- [x] 1. One\n")
	write("changes/archive/kan-0-old/tasks.md", "# Tasks\n\n- [x] 1. Done\n")
	tr, err := Find(base)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

func TestSpecsSorted(t *testing.T) {
	specs, err := seed(t).Specs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("len = %d, want 2", len(specs))
	}
	if specs[0].Capability != "auth" || specs[1].Capability != "plans" {
		t.Errorf("order = %q, %q", specs[0].Capability, specs[1].Capability)
	}
	if len(specs[0].Reqs) != 2 {
		t.Errorf("auth reqs = %d, want 2", len(specs[0].Reqs))
	}
}

func TestSpecByName(t *testing.T) {
	tr := seed(t)
	s, err := tr.Spec("plans")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Reqs) != 1 {
		t.Errorf("reqs = %d, want 1", len(s.Reqs))
	}
	if _, err := tr.Spec("nope"); err == nil {
		t.Error("want error for unknown capability")
	}
}

func TestChangesOpenAndArchived(t *testing.T) {
	tr := seed(t)
	open, err := tr.Changes(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 2 {
		t.Fatalf("open = %d, want 2", len(open))
	}
	if open[0].ID != "kan-1-first" || open[1].ID != "kan-2-second" {
		t.Errorf("order = %q, %q", open[0].ID, open[1].ID)
	}
	if open[1].DoneCount() != 1 || len(open[1].Tasks) != 2 {
		t.Errorf("kan-2 progress = %d/%d", open[1].DoneCount(), len(open[1].Tasks))
	}
	if open[0].Archived {
		t.Error("open change marked archived")
	}

	arch, err := tr.Changes(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(arch) != 1 || arch[0].ID != "kan-0-old" || !arch[0].Archived {
		t.Errorf("archived = %+v", arch)
	}
}
