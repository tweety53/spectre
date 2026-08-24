package tree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/parse"
)

// Specs reads every capability file, sorted by capability name.
func (t *Tree) Specs() ([]model.Spec, error) {
	ents, err := os.ReadDir(t.SpecsDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []model.Spec
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		s, err := parse.SpecFile(filepath.Join(t.SpecsDir(), e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Capability < out[j].Capability })
	return out, nil
}

// Spec reads one capability by name.
func (t *Tree) Spec(capability string) (model.Spec, error) {
	p := filepath.Join(t.SpecsDir(), capability+".md")
	if _, err := os.Stat(p); err != nil {
		return model.Spec{}, fmt.Errorf("no such capability %q in %s", capability, t.SpecsDir())
	}
	return parse.SpecFile(p)
}

// Changes reads change folders, sorted by id. archived selects
// changes/archive/ instead of changes/.
func (t *Tree) Changes(archived bool) ([]model.Change, error) {
	dir := t.ChangesDir()
	if archived {
		dir = t.ArchiveDir()
	}
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []model.Change
	for _, e := range ents {
		if !e.IsDir() || (!archived && e.Name() == "archive") {
			continue
		}
		c := model.Change{ID: e.Name(), Dir: filepath.Join(dir, e.Name()), Archived: archived}
		tasks, err := parse.TasksFile(filepath.Join(c.Dir, "tasks.md"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		c.Tasks = tasks
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
