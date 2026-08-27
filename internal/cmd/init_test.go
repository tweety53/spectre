package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/config"
)

// TestInitCreatesTree also pins task 9: init prints created paths relative
// to the working directory rather than the tree's absolute root, matching
// the docs (docs/example.md, README.md) this task makes true. base is
// created under the working directory (not t.TempDir(), which is not) so
// every "created" line must be relative — asserted with an exact stdout
// comparison, never a substring match that would also accept the absolute
// form.
func TestInitCreatesTree(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	base, err := os.MkdirTemp(wd, "spectre-init-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })
	rel, err := filepath.Rel(wd, base)
	if err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer

	if code := Init([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}

	root := filepath.Join(base, "spectre")
	for _, want := range []string{"specs", "changes", "config.md"} {
		if _, err := os.Stat(filepath.Join(root, want)); err != nil {
			t.Errorf("%s: %v", want, err)
		}
	}

	wantOut := strings.Join([]string{
		"created " + filepath.Join(rel, "spectre", "specs"),
		"created " + filepath.Join(rel, "spectre", "changes"),
		"created " + filepath.Join(rel, "spectre", "config.md"),
		"",
	}, "\n")
	if got := out.String(); got != wantOut {
		t.Errorf("stdout = %q, want %q", got, wantOut)
	}
}

// TestInitRejectsPositionalArgs pins the fix for the usage-line bug:
// init takes no positional argument, only --root, so an unexpected
// positional must be a usage error rather than being silently ignored
// (which previously let `init <path>` exit 0 while creating the tree in
// the working directory instead of at <path>).
func TestInitRejectsPositionalArgs(t *testing.T) {
	base := t.TempDir()
	var out, errBuf bytes.Buffer

	if code := Init([]string{"--root", base, "somewhere"}, &out, &errBuf); code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
	if errBuf.Len() == 0 {
		t.Error("stderr is empty, want a diagnostic")
	}
	if _, err := os.Stat(filepath.Join(base, "spectre")); !os.IsNotExist(err) {
		t.Error("tree was created despite the usage error")
	}
}

func TestInitFillsMissingDir(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.md"), []byte("# spectre config\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer

	if code := Init([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(root, "changes")); err != nil {
		t.Errorf("changes: %v", err)
	}
}

// TestInitRejectsFileWhereDirBelongs pins the fix for the existence check
// that used to test only os.Stat's error, never fi.IsDir(): a plain file
// sitting where specs/ or changes/ belongs must refuse with Fail, not be
// treated as an already-present directory and papered over while the rest
// of the tree is scaffolded around it.
func TestInitRejectsFileWhereDirBelongs(t *testing.T) {
	tests := []struct {
		name string
		path func(root string) string
	}{
		{"specs is a file", func(root string) string { return filepath.Join(root, "specs") }},
		{"changes is a file", func(root string) string { return filepath.Join(root, "changes") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			root := filepath.Join(base, "spectre")
			if err := os.MkdirAll(root, 0o755); err != nil {
				t.Fatal(err)
			}
			blocked := tt.path(root)
			if err := os.WriteFile(blocked, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			var out, errBuf bytes.Buffer

			if code := Init([]string{"--root", base}, &out, &errBuf); code != Fail {
				t.Fatalf("exit = %d, want %d (Fail), stderr = %s", code, Fail, errBuf.String())
			}
			if !strings.Contains(errBuf.String(), blocked) {
				t.Errorf("stderr = %q, want it to name %s", errBuf.String(), blocked)
			}
			if _, err := os.Stat(filepath.Join(root, "config.md")); !os.IsNotExist(err) {
				t.Error("config.md was created despite the refusal, want a half-made tree never to be left behind")
			}
			fi, err := os.Stat(blocked)
			if err != nil {
				t.Fatal(err)
			}
			if fi.IsDir() {
				t.Error("blocked path was replaced with a directory, want it untouched")
			}
		})
	}
}

func TestInitPreservesExistingConfig(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "spectre")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := []byte("# sentinel config\n\ndo not touch\n")
	configPath := filepath.Join(root, "config.md")
	if err := os.WriteFile(configPath, sentinel, 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer

	if code := Init([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, sentinel) {
		t.Errorf("config.md changed:\ngot:  %q\nwant: %q", got, sentinel)
	}
}

func TestInitRootResolution(t *testing.T) {
	tests := []struct {
		name string
		root func(base string) string
	}{
		{
			name: "basename not spectre gains it",
			root: func(base string) string { return base },
		},
		{
			name: "path already ending in spectre is unchanged",
			root: func(base string) string { return filepath.Join(base, "spectre") },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			var out, errBuf bytes.Buffer
			if code := Init([]string{"--root", tt.root(base)}, &out, &errBuf); code != OK {
				t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
			}
			if _, err := os.Stat(filepath.Join(base, "spectre")); err != nil {
				t.Errorf("spectre dir not created at expected location: %v", err)
			}
		})
	}
}

func TestInitIsIdempotent(t *testing.T) {
	base := t.TempDir()
	var out1, errBuf1 bytes.Buffer
	if code := Init([]string{"--root", base}, &out1, &errBuf1); code != OK {
		t.Fatalf("first run exit = %d, stderr = %s", code, errBuf1.String())
	}

	var out2, errBuf2 bytes.Buffer
	if code := Init([]string{"--root", base}, &out2, &errBuf2); code != OK {
		t.Fatalf("second run exit = %d, stderr = %s", code, errBuf2.String())
	}
	if strings.Contains(out2.String(), "created") {
		t.Errorf("second run stdout = %q, want no mention of anything created", out2.String())
	}
}

func TestInitUnwritableParent(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	base := t.TempDir()
	dir := filepath.Join(base, "locked")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	var out, errBuf bytes.Buffer
	code := Init([]string{"--root", dir}, &out, &errBuf)
	if code != Usage {
		t.Fatalf("exit = %d, want %d", code, Usage)
	}
	if errBuf.Len() == 0 {
		t.Error("stderr is empty, want a diagnostic")
	}
}

// TestInitScaffoldParsesToDefaults pins the scaffolded config.md's *content*
// to config.Default(), not just to config.Load's fenced-content-is-ignored
// behaviour. Loading the scaffold as written proves nothing on its own —
// config.Parse skips every fenced line unconditionally, so config.Load
// returns config.Default() no matter what the fence contains. To actually
// exercise the values inside the fence, this test strips the fence markers
// from _configTemplate and parses what's left: that is the content a user
// would activate by moving a line out of its fence.
//
// A parsed-equality check alone still can't catch every regression:
// config.Parse starts from config.Default() and only overrides keys it
// sees, so deleting a rule line whose value already equals the default (or
// reordering two rule lines) leaves the parsed result identical. The
// second assertion below pins the Rules block's line order directly
// against config.RuleNames, which is what actually catches a deleted or
// reordered rule.
func TestInitScaffoldParsesToDefaults(t *testing.T) {
	unfenced := unfence(_configTemplate)

	got, err := config.Parse(unfenced)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Default()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unfenced scaffold parses to %+v, want config.Default() = %+v", got, want)
	}

	wantRules := make([]string, len(config.RuleNames))
	for i, name := range config.RuleNames {
		wantRules[i] = "- " + name + ": error"
	}
	if gotRules := rulesSectionLines(string(unfenced)); !reflect.DeepEqual(gotRules, wantRules) {
		t.Errorf("Rules section lines = %v, want %v (config.RuleNames order)", gotRules, wantRules)
	}
}

// unfence strips the ``` fence-marker lines from a template, leaving the
// settings they wrapped as live config.Parse input.
func unfence(template string) []byte {
	var out []string
	for _, line := range strings.Split(template, "\n") {
		if strings.TrimSpace(line) == "```" {
			continue
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

// rulesSectionLines returns the "- <key>: <value>" lines under the "##
// Rules" heading of an un-fenced config.md body, in file order.
func rulesSectionLines(body string) []string {
	var lines []string
	inRules := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inRules = trimmed == "## Rules"
			continue
		}
		if inRules && strings.HasPrefix(trimmed, "- ") {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func TestInitThenNewThenList(t *testing.T) {
	base := t.TempDir()
	var out, errBuf bytes.Buffer
	if code := Init([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("init exit = %d, stderr = %s", code, errBuf.String())
	}

	out.Reset()
	errBuf.Reset()
	if code := New([]string{"--root", base, "kan-1-example"}, &out, &errBuf); code != OK {
		t.Fatalf("new exit = %d, stderr = %s", code, errBuf.String())
	}

	out.Reset()
	errBuf.Reset()
	if code := List([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("list exit = %d, stderr = %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "kan-1-example") {
		t.Errorf("list stdout = %q, want it to contain the new change", out.String())
	}
}
