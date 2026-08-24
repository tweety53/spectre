package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/spectre/internal/config"
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
	parser := New(config.Default())
	got, err := parser.TasksFile(p)
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

// TestTasksFileSkipsFencedExample pins the second-round final-review fix:
// a task-shaped line inside a fenced example must not be read as a real
// task, the same fence convention parser.SpecFile already applies.
func TestTasksFileSkipsFencedExample(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.md")
	body := "# Tasks\n\n" +
		"- [x] 1. Parse the spec file\n" +
		"- [ ] 2. Render it back\n" +
		"```\n" +
		"- [ ] 1. Not a real task, just an example\n" +
		"```\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	parser := New(config.Default())
	got, err := parser.TasksFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (the fenced example line must be skipped): %+v", len(got), got)
	}
}

func TestTasksFileMissing(t *testing.T) {
	parser := New(config.Default())
	if _, err := parser.TasksFile(filepath.Join(t.TempDir(), "nope.md")); err == nil {
		t.Fatal("want error for missing file")
	}
}

func TestTasksFileEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(p, []byte("# Tasks\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	parser := New(config.Default())
	got, err := parser.TasksFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}
