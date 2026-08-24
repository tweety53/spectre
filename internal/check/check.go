// Package check holds spectre's validation rules. Every rule is a pure
// function of parsed content, so its tests are fixture trees.
package check

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/tree"
)

// Finding is one validation failure, located in a file.
type Finding struct {
	File string // path relative to the tree root
	Line int
	Msg  string
}

func (f Finding) String() string { return fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Msg) }

var placeholders = []string{"TBD", "TODO"}

var (
	wellFormedReq  = regexp.MustCompile(`^- R\d+: `)
	wellFormedTask = regexp.MustCompile(`^- \[[ x]\] \d+\. `)
)

// malformedReqFindings reports bullets inside "## Requirements" that the
// parser cannot read as requirements, which would otherwise vanish silently.
func malformedReqFindings(rel string, raw []byte) []Finding {
	var out []Finding
	inReqs := false
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "## ") {
			inReqs = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Requirements"
			continue
		}
		if inReqs && strings.HasPrefix(line, "- ") && !wellFormedReq.MatchString(line) {
			out = append(out, Finding{File: rel, Line: i + 1, Msg: "malformed requirement bullet, want \"- R<n>: ... SHALL ...\""})
		}
	}
	return out
}

func headingFindings(rel string, raw []byte, want []string) []Finding {
	var out []Finding
	for _, h := range want {
		if !bytes.Contains(raw, []byte("\n"+h+"\n")) && !bytes.HasPrefix(raw, []byte(h+"\n")) {
			out = append(out, Finding{File: rel, Line: 1, Msg: fmt.Sprintf("missing %q", h)})
		}
	}
	return out
}

func placeholderFindings(rel string, raw []byte) []Finding {
	var out []Finding
	for i, line := range strings.Split(string(raw), "\n") {
		for _, p := range placeholders {
			if strings.Contains(line, p) {
				out = append(out, Finding{File: rel, Line: i + 1, Msg: fmt.Sprintf("placeholder %q", p)})
			}
		}
	}
	return out
}

// SpecFindings applies every spec rule to one capability file.
func SpecFindings(rel string, s model.Spec, raw []byte) []Finding {
	out := headingFindings(rel, raw, []string{"# " + s.Capability, "## Purpose", "## Requirements"})
	out = append(out, placeholderFindings(rel, raw)...)
	out = append(out, malformedReqFindings(rel, raw)...)

	seen := map[string]bool{}
	for i, r := range s.Reqs {
		if !strings.Contains(r.Text, " SHALL ") {
			out = append(out, Finding{File: rel, Line: r.Line, Msg: fmt.Sprintf("requirement %s has no SHALL clause", r.ID)})
		}
		if seen[r.ID] {
			out = append(out, Finding{File: rel, Line: r.Line, Msg: fmt.Sprintf("duplicate requirement id %s", r.ID)})
		} else if r.Num != i+1 {
			out = append(out, Finding{File: rel, Line: r.Line, Msg: fmt.Sprintf("requirement id %s out of sequence, expected R%d", r.ID, i+1)})
		}
		seen[r.ID] = true
	}
	return out
}

// ProposalFindings applies every proposal rule.
func ProposalFindings(rel string, raw []byte) []Finding {
	out := headingFindings(rel, raw, []string{"## Why", "## What changes"})
	return append(out, placeholderFindings(rel, raw)...)
}

// TaskFindings applies every tasks.md rule.
func TaskFindings(rel string, raw []byte, ts []model.Task) []Finding {
	var out []Finding
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "- [") && !wellFormedTask.MatchString(line) {
			out = append(out, Finding{File: rel, Line: i + 1, Msg: "malformed task line, want \"- [ ] <n>. ...\""})
		}
	}
	seen := map[int]bool{}
	expected := 1
	for _, task := range ts {
		if seen[task.Num] {
			out = append(out, Finding{File: rel, Line: task.Line, Msg: fmt.Sprintf("duplicate task number %d", task.Num)})
			continue
		}
		if task.Num != expected {
			out = append(out, Finding{File: rel, Line: task.Line, Msg: fmt.Sprintf("task number %d out of sequence, expected %d", task.Num, expected)})
		}
		seen[task.Num] = true
		expected++
	}
	return out
}

// Structural checks every spec and change in the tree, or one change when
// changeID is non-empty.
func Structural(t *tree.Tree, changeID string) ([]Finding, error) {
	var out []Finding

	if changeID == "" {
		specs, err := t.Specs()
		if err != nil {
			return nil, err
		}
		for _, s := range specs {
			raw, err := os.ReadFile(s.Path)
			if err != nil {
				return nil, err
			}
			out = append(out, SpecFindings(rel(t, s.Path), s, raw)...)
		}
	}

	changes, err := t.Changes(false)
	if err != nil {
		return nil, err
	}
	for _, c := range changes {
		if changeID != "" && c.ID != changeID {
			continue
		}
		proposal := filepath.Join(c.Dir, "proposal.md")
		raw, err := os.ReadFile(proposal)
		switch {
		case errors.Is(err, os.ErrNotExist):
			out = append(out, Finding{File: rel(t, proposal), Line: 1, Msg: "missing proposal.md"})
		case err != nil:
			return nil, err
		default:
			out = append(out, ProposalFindings(rel(t, proposal), raw)...)
		}
		tasksPath := filepath.Join(c.Dir, "tasks.md")
		tasksRaw, err := os.ReadFile(tasksPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		out = append(out, TaskFindings(rel(t, tasksPath), tasksRaw, c.Tasks)...)
	}
	return out, nil
}

func rel(t *tree.Tree, path string) string {
	r, err := filepath.Rel(t.Root, path)
	if err != nil {
		return path
	}
	return r
}
