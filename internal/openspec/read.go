// Package openspec reads an OpenSpec tree. It exists only to serve
// "spectre migrate" and is not used by any other command.
package openspec

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tweety53/spectre/internal/model"
)

// Requirement is one "### Requirement: <title>" block.
type Requirement struct {
	Title string
	Body  []string // non-empty lines beneath the heading, verbatim
}

// Spec is one OpenSpec capability file.
type Spec struct {
	Capability string
	Purpose    string
	Reqs       []Requirement
}

// Change is one folder under changes/, excluding archive/.
type Change struct {
	ID     string
	Dir    string
	Deltas []string // absolute paths of delta spec files
}

var reqHeadingRe = regexp.MustCompile(`^### Requirement:\s*(.*)$`)

// ReadSpec reads one OpenSpec spec.md. The capability name comes from the
// containing directory.
func ReadSpec(path string) (Spec, error) {
	f, err := os.Open(path)
	if err != nil {
		return Spec{}, err
	}
	defer f.Close()

	spec := Spec{Capability: filepath.Base(filepath.Dir(path))}
	var purpose []string
	section := ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()

		if m := reqHeadingRe.FindStringSubmatch(line); m != nil {
			spec.Reqs = append(spec.Reqs, Requirement{Title: strings.TrimSpace(m[1])})
			section = "req"
			continue
		}
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}

		switch section {
		case "Purpose":
			purpose = append(purpose, line)
		case "req":
			if t := strings.TrimRight(line, " \t"); strings.TrimSpace(t) != "" {
				last := &spec.Reqs[len(spec.Reqs)-1]
				last.Body = append(last.Body, t)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return Spec{}, err
	}
	spec.Purpose = strings.TrimSpace(strings.Join(purpose, "\n"))
	return spec, nil
}

// ReadSpecs reads every <specsDir>/<capability>/spec.md, sorted by name.
func ReadSpecs(specsDir string) ([]Spec, error) {
	ents, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}
	var out []Spec
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(specsDir, e.Name(), "spec.md")
		if _, err := os.Stat(p); err != nil {
			continue
		}
		s, err := ReadSpec(p)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Capability < out[j].Capability })
	return out, nil
}

// ReadChanges lists open changes and the delta spec files each carries.
func ReadChanges(changesDir string) ([]Change, error) {
	ents, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, err
	}
	var out []Change
	for _, e := range ents {
		if !e.IsDir() || e.Name() == "archive" {
			continue
		}
		c := Change{ID: e.Name(), Dir: filepath.Join(changesDir, e.Name())}
		specsDir := filepath.Join(c.Dir, "specs")
		if entries, err := os.ReadDir(specsDir); err == nil {
			for _, se := range entries {
				p := filepath.Join(specsDir, se.Name(), "spec.md")
				if _, err := os.Stat(p); err == nil {
					c.Deltas = append(c.Deltas, p)
				}
			}
		}
		sort.Strings(c.Deltas)
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

var osTaskRe = regexp.MustCompile(`^\s*- \[([ xX])\] ([0-9]+(?:\.[0-9]+)*)\.?\s+(.*)$`)

// ConvertTasks flattens OpenSpec's nested task numbering into a flat list.
func ConvertTasks(raw []byte) []model.Task {
	var out []model.Task
	for _, line := range strings.Split(string(raw), "\n") {
		m := osTaskRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		out = append(out, model.Task{
			Num:  len(out) + 1,
			Text: strings.TrimSpace(m[3]),
			Done: strings.EqualFold(m[1], "x"),
		})
	}
	return out
}

// ConvertProposal renames OpenSpec's headings to spectre's.
func ConvertProposal(raw []byte) []byte {
	return bytes.ReplaceAll(raw, []byte("## What Changes"), []byte("## What changes"))
}
