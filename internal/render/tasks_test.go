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
	want := `# Tasks

- [x] 1. Parse the spec file
- [ ] 2. Render it back
`
	if got := string(Tasks(ts)); got != want {
		t.Errorf("Tasks() =\n%q\nwant\n%q", got, want)
	}
}

func TestTasksEmpty(t *testing.T) {
	if got := string(Tasks(nil)); got != "# Tasks\n\n" {
		t.Errorf("Tasks(nil) = %q, want %q", got, "# Tasks\n\n")
	}
	if got := string(Tasks([]model.Task{})); got != "# Tasks\n\n" {
		t.Errorf("Tasks([]model.Task{}) = %q, want %q", got, "# Tasks\n\n")
	}
}

func TestTasksRenumbers(t *testing.T) {
	ts := []model.Task{{Num: 7, Text: "First"}, {Num: 9, Text: "Second"}}
	want := "# Tasks\n\n- [ ] 1. First\n- [ ] 2. Second\n"
	if got := string(Tasks(ts)); got != want {
		t.Errorf("Tasks() = %q, want %q", got, want)
	}
}
