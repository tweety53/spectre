package tree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tweety53/spectre/internal/model"
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
		if e.IsDir() || !strings.HasSuffix(e.Name(), t.Cfg.Extension) {
			continue
		}
		s, err := t.P.SpecFile(filepath.Join(t.SpecsDir(), e.Name()))
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
	p := filepath.Join(t.SpecsDir(), capability+t.Cfg.Extension)
	if _, err := os.Stat(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.Spec{}, fmt.Errorf("no such capability %q in %s", capability, t.SpecsDir())
		}
		return model.Spec{}, fmt.Errorf("capability %q in %s: %w", capability, t.SpecsDir(), err)
	}
	return t.P.SpecFile(p)
}

// Changes reads open change folders under changes/, sorted by id.
func (t *Tree) Changes() ([]model.Change, error) {
	return t.changes(false)
}

// changes reads change folders, sorted by id. archived selects
// changes/archive/ instead of changes/. Unexported: nothing outside the
// package needs archived changes today, so there is no public
// ArchivedChanges wrapper — add one when a real caller needs it.
func (t *Tree) changes(archived bool) ([]model.Change, error) {
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
		tasks, err := t.P.TasksFile(filepath.Join(c.Dir, TasksFile))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		c.Tasks = tasks
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
