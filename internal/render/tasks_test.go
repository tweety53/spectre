package render

import (
	"testing"

	"github.com/tweety53/spectre/internal/model"
)

func TestTasks(t *testing.T) {
	ts := []model.Task{
		{Num: 1, Text: "Parse the spec file", Done: true},
		{Num: 2, Text: "Render it back"},
	}
	want := `# my-change

- [x] 1. Parse the spec file
- [ ] 2. Render it back
`
	if got := string(Tasks("my-change", ts)); got != want {
		t.Errorf("Tasks() =\n%q\nwant\n%q", got, want)
	}
}

func TestTasksEmpty(t *testing.T) {
	if got := string(Tasks("my-change", nil)); got != "# my-change\n\n" {
		t.Errorf("Tasks(nil) = %q, want %q", got, "# my-change\n\n")
	}
	if got := string(Tasks("my-change", []model.Task{})); got != "# my-change\n\n" {
		t.Errorf("Tasks([]model.Task{}) = %q, want %q", got, "# my-change\n\n")
	}
}

func TestTasksRenumbers(t *testing.T) {
	ts := []model.Task{{Num: 7, Text: "First"}, {Num: 9, Text: "Second"}}
	want := "# my-change\n\n- [ ] 1. First\n- [ ] 2. Second\n"
	if got := string(Tasks("my-change", ts)); got != want {
		t.Errorf("Tasks() = %q, want %q", got, want)
	}
}

// TestTasksTitle pins the title: Tasks writes "# <id>" as the first
// line, matching proposal.md's title, and the task-rendering behaviour
// otherwise is unchanged.
func TestTasksTitle(t *testing.T) {
	ts := []model.Task{{Num: 1, Text: "Do the thing"}}
	want := "# kan-351-quick-start-guide\n\n- [ ] 1. Do the thing\n"
	if got := string(Tasks("kan-351-quick-start-guide", ts)); got != want {
		t.Errorf("Tasks() =\n%q\nwant\n%q", got, want)
	}
}
