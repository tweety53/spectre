package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/check"
	"github.com/tweety53/spectre/internal/tree"
)

// nextTopLevelHeading matches the start of the next numbered top-level
// section in docs/example.md ("## 3. ..."), as opposed to a nested heading
// such as "## Why" or "## Context" inside a fenced file body.
var nextTopLevelHeading = regexp.MustCompile(`\n## \d`)

func emptyTree(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	for _, sub := range []string{"specs", "changes"} {
		if err := os.MkdirAll(filepath.Join(base, "spectre", sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return base
}

// TestNewScaffolds also pins task 9: new prints the created change
// directory relative to the working directory, matching the docs
// (docs/example.md, README.md) this task makes true. base is created
// under the working directory (not t.TempDir(), which is not) so the
// "created" line must be relative — asserted with an exact stdout
// comparison, never a substring match that would also accept the
// absolute form.
func TestNewScaffolds(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	base, err := os.MkdirTemp(wd, "spectre-new-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })
	for _, sub := range []string{"specs", "changes"} {
		if err := os.MkdirAll(filepath.Join(base, "spectre", sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	rel, err := filepath.Rel(wd, base)
	if err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer

	if code := New([]string{"--root", base, "kan-7-add-picker"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}

	dir := filepath.Join(base, "spectre", "changes", "kan-7-add-picker")
	proposal, err := os.ReadFile(filepath.Join(dir, "proposal.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# kan-7-add-picker", "## Why", "## What changes"} {
		if !strings.Contains(string(proposal), want) {
			t.Errorf("proposal missing %q:\n%s", want, proposal)
		}
	}
	tasks, err := os.ReadFile(filepath.Join(dir, "tasks.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(tasks) != "# kan-7-add-picker\n\n" {
		t.Errorf("tasks.md = %q", tasks)
	}

	wantOut := "created " + filepath.Join(rel, "spectre", "changes", "kan-7-add-picker") + "\n"
	if got := out.String(); got != wantOut {
		t.Errorf("stdout = %q, want %q", got, wantOut)
	}
}

func TestNewRefusesExisting(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, "kan-7"}, &out, &errBuf); code != OK {
		t.Fatalf("first run exit = %d", code)
	}
	if code := New([]string{"--root", base, "kan-7"}, &out, &errBuf); code != Fail {
		t.Fatalf("second run exit = %d, want %d", code, Fail)
	}
	if !strings.Contains(errBuf.String(), "already exists") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestNewRequiresID(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", emptyTree(t)}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
}

// TestNewRejectsArchiveID pins the final-review fix: a change id equal to
// the archive directory's own name must be rejected, since a change
// scaffolded there is invisible to list, validate and archive.
func TestNewRejectsArchiveID(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, "archive"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d, stderr = %s", code, Usage, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "archive") {
		t.Errorf("stderr = %q, want it to name the archive directory", errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive", "proposal.md")); !os.IsNotExist(err) {
		t.Error("scaffold files were written into the archive directory")
	}
	if _, err := os.Stat(filepath.Join(base, "spectre", "changes", "archive")); !os.IsNotExist(err) {
		t.Error("archive directory should not have been created")
	}
}

// TestNewWritesDesign pins task 4 step 2: new scaffolds design.md
// alongside proposal.md and tasks.md, carrying both required headings.
func TestNewWritesDesign(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, "kan-7-add-picker"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}

	dir := filepath.Join(base, "spectre", "changes", "kan-7-add-picker")
	design, err := os.ReadFile(filepath.Join(dir, "design.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Context", "## Decisions"} {
		if !strings.Contains(string(design), want) {
			t.Errorf("design.md missing %q:\n%s", want, design)
		}
	}
	for _, bad := range []string{"TODO", "TBD"} {
		if strings.Contains(string(design), bad) {
			t.Errorf("design.md contains placeholder %q, which the placeholders rule reports:\n%s", bad, design)
		}
	}
}

// TestNewScaffoldValidates is the test that stops the scaffold and the
// rules drifting apart: a freshly scaffolded change, run through the same
// checkers validate uses, produces exactly the "no tasks" finding on
// tasks.md and nothing else. Zero findings would be a false claim — a
// scaffold genuinely has no tasks yet (design-scaffold-reports-no-tasks in
// design.md) — but any other finding means the scaffold and the rules it
// is scaffolded against have drifted.
func TestNewScaffoldValidates(t *testing.T) {
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, "kan-7-add-picker"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}

	tr, err := tree.Open(filepath.Join(base, "spectre"))
	if err != nil {
		t.Fatal(err)
	}
	findings, err := check.Structural(tr, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %v, want exactly one", findings)
	}
	got := findings[0]
	if !strings.HasSuffix(got.File, "tasks.md") || !strings.Contains(got.Msg, "no tasks") {
		t.Errorf("finding = %v, want the tasks.md \"no tasks\" finding", got)
	}
}

func TestNewRejectsInvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"escape parent", "../escape"},
		{"nested dir", "sub/dir"},
		{"parent dot dot", ".."},
		{"current dot", "."},
		{"absolute path", "/absolute"},
		{"backslash escape", "..\\escape"},
		{"empty string", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := emptyTree(t)
			var out, errBuf bytes.Buffer
			code := New([]string{"--root", base, tt.id}, &out, &errBuf)
			if code != Usage {
				t.Fatalf("exit = %d, want %d", code, Usage)
			}
			if !strings.Contains(errBuf.String(), "invalid change id") {
				t.Errorf("stderr = %q, want 'invalid change id'", errBuf.String())
			}
			// Verify nothing was created under changes/
			changesDir := filepath.Join(base, "spectre", "changes")
			entries, err := os.ReadDir(changesDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) > 0 {
				t.Errorf("created unwanted entries: %v", entries)
			}
		})
	}
}

// TestNewRejectsControlCharacterID pins task 10 step 1: an id containing a
// control character is refused. Task 2 made the change id a required
// heading, so an id with a newline can never be satisfied — headingFindings
// scans line by line, and a wanted heading spanning two lines matches
// nothing — and the finding it produces breaks the "file:line: message"
// contract validate's output holds.
func TestNewRejectsControlCharacterID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"newline", "weird\nid"},
		{"carriage return", "weird\rid"},
		{"tab", "weird\tid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := emptyTree(t)
			var out, errBuf bytes.Buffer
			code := New([]string{"--root", base, tt.id}, &out, &errBuf)
			if code != Usage {
				t.Fatalf("exit = %d, want %d, stderr = %s", code, Usage, errBuf.String())
			}
			if !strings.Contains(errBuf.String(), "invalid change id") {
				t.Errorf("stderr = %q, want 'invalid change id'", errBuf.String())
			}
			changesDir := filepath.Join(base, "spectre", "changes")
			entries, err := os.ReadDir(changesDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) > 0 {
				t.Errorf("created unwanted entries: %v", entries)
			}
		})
	}
}

// exampleDocPath resolves docs/example.md relative to this test file's own
// location, rather than assuming a working directory: `go test` can be
// invoked from anywhere, and docs/example.md sits two directories above
// internal/cmd.
func exampleDocPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine this test file's own path")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "docs", "example.md")
}

// scaffoldFence extracts one file's scaffolded body from the "## 2.
// Scaffold the change" section of docs/example.md: the fenced ```markdown
// block that immediately follows the line naming `spectre/changes/<id>/<name>`.
// It fails loudly (via error, never a silent skip) when the section, the
// filename label, or the fence cannot be found, so a restructured page
// breaks this test instead of passing it vacuously.
func scaffoldFence(page []byte, id, name string) ([]byte, error) {
	const sectionHeading = "## 2. Scaffold the change"
	_, section, found := bytes.Cut(page, []byte(sectionHeading))
	if !found {
		return nil, fmt.Errorf("section %q not found in docs/example.md", sectionHeading)
	}
	// Bound the section at the next numbered top-level heading ("## 3. ...").
	// A plain "\n## " would also match the nested "## Why" / "## Context"
	// headings inside the fenced bodies themselves, so this alone needs a
	// regexp rather than bytes.Cut.
	if loc := nextTopLevelHeading.FindIndex(section); loc != nil {
		section = section[:loc[0]]
	}

	label := fmt.Sprintf("spectre/changes/%s/%s", id, name)
	_, rest, found := bytes.Cut(section, []byte(label))
	if !found {
		return nil, fmt.Errorf("label %q not found in the scaffold section", label)
	}

	const fenceOpen = "```markdown\n"
	_, body, found := bytes.Cut(rest, []byte(fenceOpen))
	if !found {
		return nil, fmt.Errorf("no %q fence found after label %q", fenceOpen, label)
	}

	const fenceClose = "\n```"
	want, _, found := bytes.Cut(body, []byte(fenceClose))
	if !found {
		return nil, fmt.Errorf("unterminated fence after label %q", label)
	}
	// want shares body's backing array; clone it before appending the
	// newline the fenceClose cut consumed, so the append can never alias
	// into bytes still owned by page.
	return append(bytes.Clone(want), '\n'), nil
}

// TestExampleDocMatchesScaffold pins task 11 (F1 from the review panel's
// principles slot): the design.md scaffold body existed twice — compiled
// into _designTemplate and quoted in docs/example.md — with nothing
// binding the copies, so editing one left gofmt, go vet and go test all
// green while the page silently went stale. This test extracts the
// scaffolded bodies straight from the page's own text and compares them,
// byte for byte, against what cmd.New actually writes into a fresh
// tree — not against the template constants, which would still pass if
// New stopped using them.
func TestExampleDocMatchesScaffold(t *testing.T) {
	page, err := os.ReadFile(exampleDocPath(t))
	if err != nil {
		t.Fatalf("reading docs/example.md: %v", err)
	}

	const id = "multi-channel-notifications"
	base := emptyTree(t)
	var out, errBuf bytes.Buffer
	if code := New([]string{"--root", base, id}, &out, &errBuf); code != OK {
		t.Fatalf("cmd.New exit = %d, stderr = %s", code, errBuf.String())
	}
	dir := filepath.Join(base, "spectre", "changes", id)

	for _, name := range []string{"proposal.md", "tasks.md", "design.md"} {
		t.Run(name, func(t *testing.T) {
			want, err := scaffoldFence(page, id, name)
			if err != nil {
				t.Fatalf("extracting %s from docs/example.md: %v", name, err)
			}
			got, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("reading scaffolded %s: %v", name, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("docs/example.md's %s scaffold body does not match cmd.New's output\ngot (New):\n%s\nwant (docs/example.md):\n%s", name, got, want)
			}
		})
	}
}
