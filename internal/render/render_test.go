package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/parse"
)

func TestSpec(t *testing.T) {
	s := model.Spec{
		Capability: "auth",
		Purpose:    "Sessions and tokens.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL refresh the token.", Notes: []string{"On the request path."}},
			{ID: "R2", Num: 2, Text: "The picker SHALL show enabled plans (@gymie:plans#R7)"},
		},
	}

	want := `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  On the request path.
- R2: The picker SHALL show enabled plans (@gymie:plans#R7)
`
	if got := string(Spec(s)); got != want {
		t.Errorf("Spec() =\n%q\nwant\n%q", got, want)
	}
}

func TestSpecRoundTrip(t *testing.T) {
	src := `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  On the request path.
- R2: The picker SHALL show enabled plans (@gymie:plans#R7)
`
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.md")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := parse.SpecFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(Spec(parsed)); got != src {
		t.Errorf("round trip changed the file:\ngot\n%s\nwant\n%s", got, src)
	}
}

func TestSpecNoteWithNewline(t *testing.T) {
	s := model.Spec{
		Capability: "auth",
		Purpose:    "Sessions and tokens.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL refresh the token.", Notes: []string{"Line 1\nLine 2"}},
		},
	}

	want := `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  Line 1
  Line 2
`
	got := string(Spec(s))
	if got != want {
		t.Errorf("Spec() with newline in note =\n%q\nwant\n%q", got, want)
	}

	// Verify render → parse → render is stable
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.md")
	if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := parse.SpecFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if rerendered := string(Spec(parsed)); rerendered != got {
		t.Errorf("render → parse → render not stable:\ngot\n%s\nwant\n%s", rerendered, got)
	}
}

func TestSpecTextWithNewline(t *testing.T) {
	s := model.Spec{
		Capability: "auth",
		Purpose:    "Sessions and tokens.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL\nrefresh the token.", Notes: []string{"Note"}},
		},
	}

	want := `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  Note
`
	got := string(Spec(s))
	if got != want {
		t.Errorf("Spec() with newline in text =\n%q\nwant\n%q", got, want)
	}

	// Verify round trip: render → parse → render is stable
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.md")
	if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := parse.SpecFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if rerendered := string(Spec(parsed)); rerendered != got {
		t.Errorf("round trip not stable:\ngot\n%s\nwant\n%s", rerendered, got)
	}
}

func TestSpecCapabilityWithNewline(t *testing.T) {
	s := model.Spec{
		Capability: "auth\nservice",
		Purpose:    "Sessions and tokens.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL refresh the token.", Notes: []string{"Note"}},
		},
	}

	want := `# auth service

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  Note
`
	if got := string(Spec(s)); got != want {
		t.Errorf("Spec() with newline in capability =\n%q\nwant\n%q", got, want)
	}
}

func TestSpecNoteWithLeadingWhitespace(t *testing.T) {
	s := model.Spec{
		Capability: "auth",
		Purpose:    "Sessions and tokens.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL refresh the token.", Notes: []string{"Line 1\n  Line 2 indented"}},
		},
	}

	want := `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  Line 1
    Line 2 indented
`
	got := string(Spec(s))
	if got != want {
		t.Errorf("Spec() with leading whitespace in note =\n%q\nwant\n%q", got, want)
	}

	// Verify round trip: render → parse → render is stable
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.md")
	if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := parse.SpecFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if rerendered := string(Spec(parsed)); rerendered != got {
		t.Errorf("round trip not stable:\ngot\n%s\nwant\n%s", rerendered, got)
	}
}

func TestSpecNoteWithBlankLine(t *testing.T) {
	s := model.Spec{
		Capability: "auth",
		Purpose:    "Sessions and tokens.",
		Reqs: []model.Requirement{
			{ID: "R1", Num: 1, Text: "The system SHALL refresh the token.", Notes: []string{"Line 1\n\nLine 3"}},
		},
	}

	// The blank line is normalized away on first render
	want := `# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the token.
  Line 1
  Line 3
`
	got := string(Spec(s))
	if got != want {
		t.Errorf("Spec() with blank line in note =\n%q\nwant\n%q", got, want)
	}

	// Verify that the normalized output is stable on second pass
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.md")
	if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := parse.SpecFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if rerendered := string(Spec(parsed)); rerendered != got {
		t.Errorf("normalized output not stable on second pass:\ngot\n%s\nwant\n%s", rerendered, got)
	}
}
