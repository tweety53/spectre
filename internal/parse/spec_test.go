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
- R2: The picker SHALL show enabled plans (@gymie:plans#R7)
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

	r2 := got.Reqs[1]
	if len(r2.Refs) != 1 {
		t.Fatalf("len(r2.Refs) = %d, want 1", len(r2.Refs))
	}
	ref := r2.Refs[0]
	if ref.Peer != "gymie" || ref.Capability != "plans" || ref.ID != "R7" {
		t.Errorf("ref = %+v", ref)
	}
	if ref.Raw != "@gymie:plans#R7" {
		t.Errorf("ref.Raw = %q", ref.Raw)
	}
}

func TestRefsScopes(t *testing.T) {
	parser := New(config.Default())
	refs := parser.Refs("- R3: See (@R1) and (@plans#R2) and (@gymie:plans#R4)", 12)
	if len(refs) != 3 {
		t.Fatalf("len = %d, want 3", len(refs))
	}
	if refs[0].Peer != "" || refs[0].Capability != "" || refs[0].ID != "R1" {
		t.Errorf("same-file ref = %+v", refs[0])
	}
	if refs[1].Peer != "" || refs[1].Capability != "plans" || refs[1].ID != "R2" {
		t.Errorf("same-tree ref = %+v", refs[1])
	}
	if refs[2].Peer != "gymie" || refs[2].Capability != "plans" || refs[2].ID != "R4" {
		t.Errorf("cross-tree ref = %+v", refs[2])
	}
	for _, r := range refs {
		if r.Line != 12 {
			t.Errorf("ref line = %d, want 12", r.Line)
		}
	}
}

func TestSpecFileMissing(t *testing.T) {
	parser := New(config.Default())
	if _, err := parser.SpecFile(filepath.Join(t.TempDir(), "nope.md")); err == nil {
		t.Fatal("want error for missing file")
	}
}
