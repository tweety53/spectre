// Package check holds spectre's validation rules. Every rule is a pure
// function of parsed content, so its tests are fixture trees.
package check

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tweety53/spectre/internal/config"
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

var _placeholders = []string{"TBD", "TODO"}

var _wellFormedTask = regexp.MustCompile(`^- \[[ x]\] \d+\. `)

// malformedReqFindings reports bullets inside "## Requirements" that the
// parser cannot read as requirements, which would otherwise vanish silently.
// Fenced code blocks are skipped, so an example bullet inside a ``` fence
// is not mistaken for a malformed requirement. The well-formed pattern and
// the message are built from cfg's configured id prefix and modal verb, so
// a bullet using the tree's own vocabulary is never reported as malformed.
func malformedReqFindings(cfg config.Config, relPath string, raw []byte) []Finding {
	prefix, modal := cfg.IDPrefix, cfg.Modal
	wellFormedReq := regexp.MustCompile(`^- ` + regexp.QuoteMeta(prefix) + `\d+: `)
	var out []Finding
	inReqs := false
	inFence := false
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			inReqs = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Requirements"
			continue
		}
		if inReqs && strings.HasPrefix(line, "- ") && !wellFormedReq.MatchString(line) {
			msg := fmt.Sprintf("malformed requirement bullet, want \"- %s<n>: ... %s ...\"", prefix, modal)
			out = append(out, Finding{File: relPath, Line: i + 1, Msg: msg})
		}
	}
	return out
}

// headingFindings reports headings from want that never appear as their own
// line outside a fenced code block, and headings that do appear but out of
// the order want states. A line scan (rather than substring matching) means
// a heading that is the file's last line, with no trailing newline, is
// still recognized, and a heading that appears only inside a ``` fence or
// as quoted example text is not. Ordering is judged only across the
// headings that are actually present: a missing heading produces the
// missing finding above and is excluded from the order check, so one cause
// never produces two findings.
func headingFindings(relPath string, raw []byte, want []string) []Finding {
	firstLine := map[string]int{}
	inFence := false
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		trimmed := strings.TrimRight(line, " \t\r")
		if _, ok := firstLine[trimmed]; !ok {
			firstLine[trimmed] = i
		}
	}

	var out []Finding
	prevHeading, prevLine := "", -1
	for _, h := range want {
		line, ok := firstLine[h]
		if !ok {
			out = append(out, Finding{File: relPath, Line: 1, Msg: fmt.Sprintf("missing %q", h)})
			continue
		}
		if prevLine >= 0 && line < prevLine {
			msg := fmt.Sprintf("heading %q appears before %q, out of the template's order", h, prevHeading)
			out = append(out, Finding{File: relPath, Line: 1, Msg: msg})
		}
		prevHeading, prevLine = h, line
	}
	return out
}

// specHeadings, proposalHeadings, taskHeadings and _designHeadings are the
// single source for the heading order SpecFindings, ProposalFindings,
// TaskFindings and DesignFindings check. TestDocumentedHeadingsMatchChecker
// (check_test.go) checks README.md's "File templates" table and
// docs/example.md's prompts against these same lists, so changing one here
// without updating both documents now fails that test.

// specHeadings returns the headings a capability spec must carry, in
// order.
func specHeadings(capability string) []string {
	return []string{"# " + capability, "## Purpose", "## Requirements"}
}

// proposalHeadings returns the headings a proposal.md must carry, in
// order. id is the owning change's id.
func proposalHeadings(id string) []string {
	return []string{"# " + id, "## Why", "## What changes"}
}

// taskHeadings returns the headings a tasks.md must carry, in order. id is
// the owning change's id.
func taskHeadings(id string) []string {
	return []string{"# " + id}
}

// _designHeadings are the headings a design.md must carry, in order, when
// the file exists.
var _designHeadings = []string{"## Context", "## Decisions"}

func placeholderFindings(relPath string, raw []byte) []Finding {
	var out []Finding
	for i, line := range strings.Split(string(raw), "\n") {
		for _, p := range _placeholders {
			if strings.Contains(line, p) {
				out = append(out, Finding{File: relPath, Line: i + 1, Msg: fmt.Sprintf("placeholder %q", p)})
			}
		}
	}
	return out
}

// SpecFindings applies every spec rule to one capability file, under cfg's
// rule gating and vocabulary. raw supplies the pre-parse text the
// regex-based checks need; s carries the parsed requirements.
func SpecFindings(cfg config.Config, relPath string, s model.Spec, raw []byte) []Finding {
	var out []Finding
	if cfg.RuleOn("headings") {
		out = append(out, headingFindings(relPath, raw, specHeadings(s.Capability))...)
	}
	if cfg.RuleOn("placeholders") {
		out = append(out, placeholderFindings(relPath, raw)...)
	}
	if cfg.RuleOn("malformed-bullet") {
		out = append(out, malformedReqFindings(cfg, relPath, raw)...)
	}

	seen := map[string]bool{}
	for i, r := range s.Reqs {
		if cfg.RuleOn("shall-clause") && !strings.Contains(r.Text, " "+cfg.Modal+" ") {
			msg := fmt.Sprintf("requirement %s has no %s clause", r.ID, cfg.Modal)
			out = append(out, Finding{File: relPath, Line: r.Line, Msg: msg})
		}
		if cfg.RuleOn("id-sequence") {
			if seen[r.ID] {
				msg := fmt.Sprintf("duplicate requirement id %s", r.ID)
				out = append(out, Finding{File: relPath, Line: r.Line, Msg: msg})
			} else if r.Num != i+1 {
				msg := fmt.Sprintf("requirement id %s out of sequence, expected %s%d", r.ID, cfg.IDPrefix, i+1)
				out = append(out, Finding{File: relPath, Line: r.Line, Msg: msg})
			}
		}
		seen[r.ID] = true
	}
	return out
}

// ProposalFindings applies every proposal rule, under cfg's rule gating.
// id is the owning change's id: proposal.md's title must be "# <id>".
func ProposalFindings(cfg config.Config, id, relPath string, raw []byte) []Finding {
	var out []Finding
	if cfg.RuleOn("headings") {
		out = append(out, headingFindings(relPath, raw, proposalHeadings(id))...)
	}
	if cfg.RuleOn("placeholders") {
		out = append(out, placeholderFindings(relPath, raw)...)
	}
	return out
}

// DesignFindings applies every design.md rule, under cfg's rule gating.
// Unlike ProposalFindings and TaskFindings, design.md carries no id
// heading requirement — only the section order.
func DesignFindings(cfg config.Config, relPath string, raw []byte) []Finding {
	var out []Finding
	if cfg.RuleOn("headings") {
		out = append(out, headingFindings(relPath, raw, _designHeadings)...)
	}
	if cfg.RuleOn("placeholders") {
		out = append(out, placeholderFindings(relPath, raw)...)
	}
	return out
}

// TaskFindings applies every tasks.md rule, under cfg's rule gating. id is
// the owning change's id: tasks.md's title must be "# <id>", gated on
// "headings" the same way ProposalFindings gates its own title check. The
// malformed-task-line check, the empty-tasks check, duplicate task numbers
// and out-of-sequence task numbers are gated as one unit on the
// "task-sequence" rule, so switching that rule off switches all four off
// together.
func TaskFindings(cfg config.Config, id, relPath string, raw []byte, ts []model.Task) []Finding {
	var out []Finding
	if cfg.RuleOn("headings") {
		out = append(out, headingFindings(relPath, raw, taskHeadings(id))...)
	}
	if !cfg.RuleOn("task-sequence") {
		return out
	}
	inFence := false
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(line, "- [") && !_wellFormedTask.MatchString(line) {
			msg := "malformed task line, want \"- [ ] <n>. ...\""
			out = append(out, Finding{File: relPath, Line: i + 1, Msg: msg})
		}
	}
	if len(ts) == 0 {
		out = append(out, Finding{File: relPath, Line: 1, Msg: "no tasks"})
	}
	seen := map[int]bool{}
	for i, task := range ts {
		if seen[task.Num] {
			msg := fmt.Sprintf("duplicate task number %d", task.Num)
			out = append(out, Finding{File: relPath, Line: task.Line, Msg: msg})
		} else if task.Num != i+1 {
			msg := fmt.Sprintf("task number %d out of sequence, expected %d", task.Num, i+1)
			out = append(out, Finding{File: relPath, Line: task.Line, Msg: msg})
		}
		seen[task.Num] = true
	}
	return out
}

// Structural checks every spec and change in the tree, or one change when
// changeID is non-empty. peers is this tree's declared peers, already
// resolved (see tree.ResolvePeer); it is only consulted when changeID is
// empty, and may be nil otherwise.
func Structural(t *tree.Tree, changeID string, peers map[string]tree.ResolvedPeer) ([]Finding, error) {
	var out []Finding

	if changeID == "" {
		specs, err := t.Specs()
		if err != nil {
			return nil, err
		}
		for _, s := range specs {
			out = append(out, SpecFindings(t.Cfg, rel(t, s.Path), s, s.Raw)...)
		}

		if t.Cfg.RuleOn("refs") {
			out = append(out, RefFindings(t.Root, specs, peers)...)
		}
	}

	changes, err := t.Changes()
	if err != nil {
		return nil, err
	}
	for _, c := range changes {
		if changeID != "" && c.ID != changeID {
			continue
		}
		proposal := filepath.Join(c.Dir, tree.ProposalFile)
		raw, err := os.ReadFile(proposal)
		switch {
		case errors.Is(err, os.ErrNotExist):
			out = append(out, Finding{File: rel(t, proposal), Line: 1, Msg: "missing " + tree.ProposalFile})
		case err != nil:
			return nil, err
		default:
			out = append(out, ProposalFindings(t.Cfg, c.ID, rel(t, proposal), raw)...)
		}
		tasksPath := filepath.Join(c.Dir, tree.TasksFile)
		tasksRaw, err := os.ReadFile(tasksPath)
		switch {
		case errors.Is(err, os.ErrNotExist):
			out = append(out, Finding{File: rel(t, tasksPath), Line: 1, Msg: "missing " + tree.TasksFile})
		case err != nil:
			return nil, err
		default:
			out = append(out, TaskFindings(t.Cfg, c.ID, rel(t, tasksPath), tasksRaw, c.Tasks)...)
		}
		// design.md is optional: an absent file is not a finding, but any
		// other read error propagates exactly as the two branches above do.
		designPath := filepath.Join(c.Dir, tree.DesignFile)
		designRaw, err := os.ReadFile(designPath)
		switch {
		case errors.Is(err, os.ErrNotExist):
			// optional, nothing to report
		case err != nil:
			return nil, err
		default:
			out = append(out, DesignFindings(t.Cfg, rel(t, designPath), designRaw)...)
		}
	}
	return out, nil
}

// rel returns path relative to t's root; see relTo.
func rel(t *tree.Tree, path string) string {
	return relTo(t.Root, path)
}

// relTo returns path relative to root, falling back to path itself on the
// (unreachable in practice) case that path isn't a descendant of root —
// every path passed in is always constructed from the tree's own root.
func relTo(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}
