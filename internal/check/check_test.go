package check

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/config"
	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/tree"
)

func msgs(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.String())
		b.WriteString("\n")
	}
	return b.String()
}

func TestSpecFindingsClean(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n")
	s := model.Spec{Capability: "auth", Purpose: "P.", Raw: raw, Reqs: []model.Requirement{
		{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7},
	}}
	if got := SpecFindings(config.Default(), "specs/auth.md", s, s.Raw); len(got) != 0 {
		t.Errorf("want clean, got:\n%s", msgs(got))
	}
}

func TestSpecFindingsRequirementRules(t *testing.T) {
	raw := []byte("# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: no modal verb here.\n- R1: The system SHALL b.\n- R5: The system SHALL c.\n")
	s := model.Spec{Capability: "auth", Purpose: "P.", Raw: raw, Reqs: []model.Requirement{
		{ID: "R1", Num: 1, Text: "no modal verb here.", Line: 7},
		{ID: "R1", Num: 1, Text: "The system SHALL b.", Line: 8},
		{ID: "R5", Num: 5, Text: "The system SHALL c.", Line: 9},
	}}
	got := msgs(SpecFindings(config.Default(), "specs/auth.md", s, s.Raw))
	for _, want := range []string{
		"specs/auth.md:7: requirement R1 has no SHALL clause",
		"specs/auth.md:8: duplicate requirement id R1",
		"specs/auth.md:9: requirement id R5 out of sequence, expected R3",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSpecFindingsTable(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		reqs         []model.Requirement
		wantContains []string
		wantAbsent   []string
	}{
		{
			name: "missing heading",
			raw:  "# auth\n\n## Requirements\n- R1: The system SHALL a.\n",
			reqs: []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 4}},
			wantContains: []string{
				"missing \"## Purpose\"",
			},
		},
		{
			name:         "placeholder",
			raw:          "# auth\n\n## Purpose\nTBD\n\n## Requirements\n- R1: The system SHALL a.\n",
			reqs:         []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7}},
			wantContains: []string{"specs/auth.md:4: placeholder \"TBD\""},
		},
		{
			name:         "malformed bullet",
			raw:          "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1 The system SHALL a.\n",
			wantContains: []string{"specs/auth.md:7: malformed requirement bullet"},
		},
		{
			name:         "heading in fence not recognized",
			raw:          "# auth\n\n```\n## Purpose\n```\n\n## Requirements\n- R1: The system SHALL a.\n",
			reqs:         []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 8}},
			wantContains: []string{"missing \"## Purpose\""},
		},
		{
			name:       "heading as file's last line, no trailing newline",
			raw:        "# auth\n\n## Purpose\nP.\n\n## Requirements",
			wantAbsent: []string{"missing \"## Requirements\""},
		},
		{
			name:       "malformed bullet in fence ignored",
			raw:        "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n\n```\n- R1 no colon\n```\n",
			reqs:       []model.Requirement{{ID: "R1", Num: 1, Text: "The system SHALL a.", Line: 7}},
			wantAbsent: []string{"malformed requirement bullet"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := model.Spec{Capability: "auth", Raw: []byte(tt.raw), Reqs: tt.reqs}
			got := msgs(SpecFindings(config.Default(), "specs/auth.md", s, s.Raw))
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q in:\n%s", want, got)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("unwanted %q in:\n%s", absent, got)
				}
			}
		})
	}
}

func TestStructuralMissingTasks(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	changeDir := filepath.Join(root, "changes", "x")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# x\n\n## Why\nB.\n\n## What changes\nC.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := tree.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Structural(tr, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msgs(got), "changes/x/tasks.md:1: missing tasks.md") {
		t.Errorf("got:\n%s", msgs(got))
	}
}

func TestProposalFindings(t *testing.T) {
	got := msgs(ProposalFindings(config.Default(), "x", "changes/x/proposal.md", []byte("# x\n\n## Why\nBecause.\n")))
	if !strings.Contains(got, "missing \"## What changes\"") {
		t.Errorf("got:\n%s", got)
	}
	clean := ProposalFindings(config.Default(), "x", "changes/x/proposal.md", []byte("# x\n\n## Why\nB.\n\n## What changes\nC.\n"))
	if len(clean) != 0 {
		t.Errorf("want clean, got:\n%s", msgs(clean))
	}
}

func TestTaskFindings(t *testing.T) {
	raw := []byte("# x\n\n- [x] 1. a\n- [ ] 1. b\n- [ ] 5. c\n")
	ts := []model.Task{{Num: 1, Text: "a", Line: 3}, {Num: 1, Text: "b", Line: 4}, {Num: 5, Text: "c", Line: 5}}
	got := msgs(TaskFindings(config.Default(), "x", "changes/x/tasks.md", raw, ts))
	for _, want := range []string{
		"changes/x/tasks.md:4: duplicate task number 1",
		"changes/x/tasks.md:5: task number 5 out of sequence, expected 3",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// TestTaskFindingsIgnoresFencedLine pins the second-round final-review fix:
// a task-shaped line inside a fenced example must not be reported as
// malformed, the same fence convention this file's other findings
// functions already apply.
func TestTaskFindingsIgnoresFencedLine(t *testing.T) {
	raw := []byte("# x\n\n- [ ] 1. Real task\n\n```\n- [ ] Not a real task, just an example\n```\n")
	got := msgs(TaskFindings(config.Default(), "x", "changes/x/tasks.md", raw, []model.Task{{Num: 1, Text: "Real task", Line: 3}}))
	if got != "" {
		t.Errorf("want clean, got:\n%s", got)
	}
}

func TestTaskFindingsMalformedLine(t *testing.T) {
	raw := []byte("# x\n\n- [ ] Write the parser\n")
	got := msgs(TaskFindings(config.Default(), "x", "changes/x/tasks.md", raw, nil))
	if !strings.Contains(got, "changes/x/tasks.md:3: malformed task line") {
		t.Errorf("got:\n%s", got)
	}
}

// TestTasksHeading pins step 3 of task 2: tasks.md's title must be
// "# <change-id>", the same wanted-heading treatment ProposalFindings
// already gives proposal.md. The id has to reach TaskFindings as a
// parameter — Structural has c.ID in hand, so nothing re-derives it from
// a path.
func TestTasksHeading(t *testing.T) {
	ts := []model.Task{{Num: 1, Text: "a", Line: 3}}
	got := msgs(TaskFindings(config.Default(), "my-change", "changes/my-change/tasks.md", []byte("# wrong-title\n\n- [ ] 1. a\n"), ts))
	if !strings.Contains(got, `missing "# my-change"`) {
		t.Errorf("got:\n%s", got)
	}

	clean := TaskFindings(config.Default(), "my-change", "changes/my-change/tasks.md", []byte("# my-change\n\n- [ ] 1. a\n"), ts)
	if len(clean) != 0 {
		t.Errorf("want clean, got:\n%s", msgs(clean))
	}
}

// TestDesignFindings pins step 2 of task 2: DesignFindings mirrors
// ProposalFindings, wanting "## Context" then "## Decisions" in order, and
// reusing headingFindings' ordering check rather than reimplementing it.
func TestDesignFindings(t *testing.T) {
	got := msgs(DesignFindings(config.Default(), "changes/x/design.md", []byte("# x\n\n## Context\nC.\n")))
	if !strings.Contains(got, `missing "## Decisions"`) {
		t.Errorf("got:\n%s", got)
	}

	outOfOrder := msgs(DesignFindings(config.Default(), "changes/x/design.md", []byte("# x\n\n## Decisions\nD.\n\n## Context\nC.\n")))
	if !strings.Contains(outOfOrder, "out of the template's order") {
		t.Errorf("got:\n%s", outOfOrder)
	}

	clean := DesignFindings(config.Default(), "changes/x/design.md", []byte("# x\n\n## Context\nC.\n\n## Decisions\nD.\n\n## Extra\nE.\n"))
	if len(clean) != 0 {
		t.Errorf("want clean, got:\n%s", msgs(clean))
	}
}

// TestEmptyTasksReported pins step 5: a tasks.md that parses to zero tasks
// is a finding gated under the existing "task-sequence" rule, alongside
// the tasks.md gains from parsing.
func TestEmptyTasksReported(t *testing.T) {
	raw := []byte("# my-change\n\n")
	got := msgs(TaskFindings(config.Default(), "my-change", "changes/my-change/tasks.md", raw, nil))
	if !strings.Contains(got, "no tasks") {
		t.Errorf("got:\n%s", got)
	}

	off := config.Default().WithRule("task-sequence", false)
	if got := msgs(TaskFindings(off, "my-change", "changes/my-change/tasks.md", raw, nil)); got != "" {
		t.Errorf("rule off should be silent, got:\n%s", got)
	}
}

// TestStructuralDesignAndTasksTemplates reproduces task 2's behavior
// against real files on disk via Structural, rather than calling the leaf
// checkers with hand-built byte slices only: a change with no design.md
// produces zero findings (design.md is optional), and a design.md whose
// headings are out of order is caught the same way Structural already
// catches a swapped proposal.md.
func TestStructuralDesignAndTasksTemplates(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	changeDir := filepath.Join(root, "changes", "my-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(changeDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("proposal.md", "# my-change\n\n## Why\nB.\n\n## What changes\nC.\n")
	write("tasks.md", "# my-change\n\n- [ ] 1. a\n")

	tr, err := tree.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Structural(tr, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("design.md is optional: want no findings without it, got:\n%s", msgs(got))
	}

	write("design.md", "# my-change\n\n## Decisions\nD.\n\n## Context\nCtx.\n")
	tr, err = tree.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err = Structural(tr, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msgs(got), "out of the template's order") {
		t.Errorf("want an ordering finding for design.md, got:\n%s", msgs(got))
	}
}

// TestStructuralDesignUnreadablePropagatesError pins step 4 of task 2:
// an os.ErrNotExist reading design.md is not a finding (the file is
// optional), but any other read error must propagate exactly as the
// proposal.md and tasks.md branches already do, not be swallowed as if
// the file were merely absent.
func TestStructuralDesignUnreadablePropagatesError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	changeDir := filepath.Join(root, "changes", "my-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(changeDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("proposal.md", "# my-change\n\n## Why\nB.\n\n## What changes\nC.\n")
	write("tasks.md", "# my-change\n\n- [ ] 1. a\n")
	designPath := filepath.Join(changeDir, "design.md")
	write("design.md", "# my-change\n\n## Context\nC.\n\n## Decisions\nD.\n")
	if err := os.Chmod(designPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(designPath, 0o644) })

	tr, err := tree.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Structural(tr, "", nil)
	if err == nil {
		t.Fatal("want a non-nil error for an unreadable design.md, got nil")
	}
	if got != nil {
		t.Errorf("want no findings alongside a propagated error, got:\n%s", msgs(got))
	}
}

// TestHeadingOrder pins headingFindings' ordering check: required headings
// must appear in the order want states, not merely be present. A heading
// that is missing must produce only the existing missing-heading finding,
// never an ordering finding on top of it.
func TestHeadingOrder(t *testing.T) {
	want := []string{"## Why", "## What changes"}
	tests := []struct {
		name         string
		raw          string
		wantCount    int
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:      "correct order has no findings",
			raw:       "# x\n\n## Why\nB.\n\n## What changes\nC.\n",
			wantCount: 0,
		},
		{
			name:      "swapped headings report one ordering finding naming both",
			raw:       "# x\n\n## What changes\nC.\n\n## Why\nB.\n",
			wantCount: 1,
			wantContains: []string{
				`"## What changes"`,
				`"## Why"`,
			},
		},
		{
			name:      "a missing heading reports missing, not ordering",
			raw:       "# x\n\n## What changes\nC.\n",
			wantCount: 1,
			wantContains: []string{
				`missing "## Why"`,
			},
			wantAbsent: []string{"order"},
		},
		{
			name:      "extra headings interleaved do not affect order",
			raw:       "# x\n\n## Why\nB.\n\n## Extra\nE.\n\n## What changes\nC.\n",
			wantCount: 0,
		},
		{
			name:      "required headings inside a fence remain missing, not ordered",
			raw:       "# x\n\n```\n## What changes\n## Why\n```\n",
			wantCount: 2,
			wantContains: []string{
				`missing "## Why"`,
				`missing "## What changes"`,
			},
			wantAbsent: []string{"order"},
		},
		{
			// Regression for a mutant that survived the suite: deleting the
			// `continue` after the missing-heading append lets execution
			// fall through into the ordering check with a zero line value.
			// The earlier "missing, not ordering" case above lists the
			// missing heading first, so prevLine is still -1 there and the
			// fallthrough is indistinguishable from the continue. Listing
			// the missing heading second makes the fallen-through zero
			// line less than prevLine, which is what pins the guarantee.
			name:      "a missing heading listed second reports missing, not ordering",
			raw:       "# x\n\n## Why\nB.\n",
			wantCount: 1,
			wantContains: []string{
				`missing "## What changes"`,
			},
			wantAbsent: []string{"order"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := headingFindings("changes/x/proposal.md", []byte(tt.raw), want)
			if len(got) != tt.wantCount {
				t.Errorf("want %d findings, got %d:\n%s", tt.wantCount, len(got), msgs(got))
			}
			gotMsgs := msgs(got)
			for _, want := range tt.wantContains {
				if !strings.Contains(gotMsgs, want) {
					t.Errorf("missing %q in:\n%s", want, gotMsgs)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(gotMsgs, absent) {
					t.Errorf("unwanted %q in:\n%s", absent, gotMsgs)
				}
			}
		})
	}
}

// docPaths returns README.md and docs/example.md's paths, resolved from
// this test file's own location rather than the working directory: `go
// test` can be invoked from anywhere, and both files sit two directories
// above internal/check.
func docPaths(t *testing.T) (readmePath, examplePath string) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine this test file's own path")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(root, "README.md"), filepath.Join(root, "docs", "example.md")
}

func mustReadDoc(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(raw)
}

var backtickItem = regexp.MustCompile("`([^`]+)`")

// backtickList extracts every backtick-quoted item from s, in order.
func backtickList(s string) []string {
	var out []string
	for _, m := range backtickItem.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

// flattenQuotes strips docs/example.md's blockquote "> " line markers and
// collapses all runs of whitespace (including the newlines the prompts
// wrap on) to a single space, so a marker phrase split across several
// blockquote lines can still be found and sliced as one continuous string.
func flattenQuotes(doc string) string {
	var b strings.Builder
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimPrefix(line, "> ")
		if line == ">" {
			line = ""
		}
		b.WriteString(line)
		b.WriteString(" ")
	}
	return regexp.MustCompile(`\s+`).ReplaceAllString(b.String(), " ")
}

// headingsAfterMarker finds marker in doc and returns the backtick-quoted
// items in the text between the end of marker and the next sentence-ending
// period. It fails the test loudly (never skips) if marker isn't found, so
// a restructured document breaks this test instead of passing it
// vacuously.
func headingsAfterMarker(t *testing.T, doc, marker string) []string {
	t.Helper()
	idx := strings.Index(doc, marker)
	if idx < 0 {
		t.Fatalf("marker %q not found in document", marker)
	}
	rest := doc[idx+len(marker):]
	end := strings.IndexByte(rest, '.')
	if end < 0 {
		t.Fatalf("no sentence terminator after marker %q", marker)
	}
	return backtickList(rest[:end])
}

// tableRowHeadings finds the README "File templates" row for filename and
// returns the backtick-quoted headings in its "Required headings" column.
// It fails the test loudly if the row isn't found.
func tableRowHeadings(t *testing.T, readme, filename string) []string {
	t.Helper()
	marker := "| `" + filename + "` |"
	idx := strings.Index(readme, marker)
	if idx < 0 {
		t.Fatalf("README table row for %q not found", filename)
	}
	rest := readme[idx+len(marker):]
	end := strings.IndexByte(rest, '|')
	if end < 0 {
		t.Fatalf("no closing column delimiter for %q row", filename)
	}
	return backtickList(rest[:end])
}

// TestDocumentedHeadingsMatchChecker binds the heading order stated in
// README.md's "File templates" table and docs/example.md's prompts to the
// same lists SpecFindings, ProposalFindings, TaskFindings and
// DesignFindings actually check (specHeadings, proposalHeadings,
// taskHeadings, _designHeadings — see task 16 of
// changes/kan-351-quick-start-guide/tasks.md). A literal changed in
// check.go without updating both documents now fails this test instead of
// leaving every check green while the docs describe a different order.
func TestDocumentedHeadingsMatchChecker(t *testing.T) {
	readmePath, examplePath := docPaths(t)
	readme := mustReadDoc(t, readmePath)
	example := flattenQuotes(mustReadDoc(t, examplePath))

	t.Run("README file templates table", func(t *testing.T) {
		cases := []struct {
			file string
			want []string
		}{
			{"proposal.md", proposalHeadings("<change-id>")},
			{"tasks.md", taskHeadings("<change-id>")},
			{"design.md", _designHeadings},
			{"specs/<capability>.md", specHeadings("<capability>")},
		}
		for _, c := range cases {
			got := tableRowHeadings(t, readme, c.file)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%s row: got %v, want %v", c.file, got, c.want)
			}
		}
	})

	t.Run("docs/example.md prompts", func(t *testing.T) {
		cases := []struct {
			name   string
			marker string
			want   []string
		}{
			{
				"spec",
				"specs/notifications.md`. It must have exactly these headings, in this order:",
				specHeadings("notifications"),
			},
			{
				"proposal",
				"proposal.md`. It must have exactly these headings, in this order:",
				proposalHeadings("multi-channel-notifications"),
			},
			{
				"design",
				"design.md`. It must have exactly these headings, in this order:",
				_designHeadings,
			},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got := headingsAfterMarker(t, example, c.marker)
				if !reflect.DeepEqual(got, c.want) {
					t.Errorf("got %v, want %v", got, c.want)
				}
			})
		}

		t.Run("tasks", func(t *testing.T) {
			marker := "tasks.md`. The first line must be exactly `"
			idx := strings.Index(example, marker)
			if idx < 0 {
				t.Fatalf("marker %q not found in document", marker)
			}
			rest := example[idx+len(marker):]
			end := strings.IndexByte(rest, '`')
			if end < 0 {
				t.Fatalf("no closing backtick after marker %q", marker)
			}
			got := []string{rest[:end]}
			want := taskHeadings("multi-channel-notifications")
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	})
}
