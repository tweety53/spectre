package parse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTasksFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.md")
	body := `# Tasks

- [x] 1. Parse the spec file
- [ ] 2. Render it back
`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := TasksFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if !got[0].Done || got[0].Num != 1 || got[0].Text != "Parse the spec file" || got[0].Line != 3 {
		t.Errorf("task 1 = %+v", got[0])
	}
	if got[1].Done || got[1].Num != 2 || got[1].Text != "Render it back" {
		t.Errorf("task 2 = %+v", got[1])
	}
}

func TestTasksFileEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(p, []byte("# Tasks\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := TasksFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}
