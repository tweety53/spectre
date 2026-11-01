package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/spectre/internal/config"
)

func writeSpec(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSpecFile(t *testing.T) {
	p := writeSpec(t, `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token before expiry.
  Refresh happens on the request path.
- R2: The picker SHALL show enabled plans (@web:plans#R7)
`)

	parser := New(config.Default())
	got, err := parser.SpecFile(p)
	if err != nil {
		t.Fatalf("SpecFile: %v", err)
	}
	if got.Capability != "auth" {
		t.Errorf("Capability = %q, want auth", got.Capability)
	}
	if got.Purpose != "Sessions and tokens." {
		t.Errorf("Purpose = %q", got.Purpose)
	}
	if len(got.Reqs) != 2 {
		t.Fatalf("len(Reqs) = %d, want 2", len(got.Reqs))
	}

	r1 := got.Reqs[0]
	if r1.ID != "R1" || r1.Num != 1 {
		t.Errorf("r1 id/num = %q/%d", r1.ID, r1.Num)
	}
	if r1.Text != "The system SHALL refresh the token before expiry." {
		t.Errorf("r1 text = %q", r1.Text)
	}
	if len(r1.Notes) != 1 || r1.Notes[0] != "Refresh happens on the request path." {
		t.Errorf("r1 notes = %#v", r1.Notes)
	}
	if r1.Line != 7 {
		t.Errorf("r1 line = %d, want 7", r1.Line)
	}

}

// TestSpecFileSkipsFencedRequirements pins the final-review fix: a
// requirement-shaped bullet inside a fenced code block must not be read as
// a real requirement, the same convention internal/check already applies
// to headings and the malformed-bullet check.
func TestSpecFileSkipsFencedRequirements(t *testing.T) {
	p := writeSpec(t, "# auth\n\n"+
		"## Purpose\nSessions.\n\n"+
		"## Requirements\n"+
		"- R1: The system SHALL refresh the token.\n"+
		"```\n"+
		"- R99: This is just an example bullet, not a real requirement.\n"+
		"```\n"+
		"- R2: The system SHALL revoke stale tokens.\n")

	parser := New(config.Default())
	got, err := parser.SpecFile(p)
	if err != nil {
		t.Fatalf("SpecFile: %v", err)
	}
	if len(got.Reqs) != 2 {
		t.Fatalf("len(Reqs) = %d, want 2 (R99 inside the fence must be skipped): %+v", len(got.Reqs), got.Reqs)
	}
	if got.Reqs[0].ID != "R1" || got.Reqs[1].ID != "R2" {
		t.Errorf("Reqs = %+v, want R1 then R2 (R99 excluded)", got.Reqs)
	}
}

func TestSpecFileMissing(t *testing.T) {
	parser := New(config.Default())
	if _, err := parser.SpecFile(filepath.Join(t.TempDir(), "nope.md")); err == nil {
		t.Fatal("want error for missing file")
	}
}
