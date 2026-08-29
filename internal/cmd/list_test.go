package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/spectre/internal/testtree"
)

func seedTree(t *testing.T) string {
	t.Helper()
	base := emptyTree(t)
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(base, "spectre", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("specs/auth.md", "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n- R2: The system SHALL b.\n")
	write("changes/kan-1-first/proposal.md", "# kan-1-first\n\n## Why\nBecause.\n\n## What changes\n- a thing\n")
	write("changes/kan-1-first/tasks.md", "# kan-1-first\n\n- [x] 1. One\n- [ ] 2. Two\n- [ ] 3. Three\n")
	return base
}

func TestListChanges(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", seedTree(t)}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	want := "kan-1-first  1/3\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestListSpecs(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", seedTree(t), "--specs"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	want := "auth  2 requirements\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestListJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", seedTree(t), "--json"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	var got struct {
		Changes []struct {
			ID    string `json:"id"`
			Done  int    `json:"done"`
			Total int    `json:"total"`
		} `json:"changes"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json: %v (%s)", err, out.String())
	}
	if len(got.Changes) != 1 || got.Changes[0].ID != "kan-1-first" || got.Changes[0].Done != 1 || got.Changes[0].Total != 3 {
		t.Errorf("json = %+v", got.Changes)
	}
}

func TestListEmptyTree(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", emptyTree(t)}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	if out.String() != "" {
		t.Errorf("stdout = %q, want empty", out.String())
	}
}

// canonicalTasks is the tasks.md body for a canonical change with 2 of 3
// tasks done, used by every fixture below that needs real progress to
// resolve through.
const canonicalTasks = "# kan-1-first\n\n- [x] 1. One\n- [x] 2. Two\n- [ ] 3. Three\n"

// satelliteLink builds a "## Part of" link.md naming peer:changeID, with
// "## Branch" carried along since design.md's grammar requires it.
func satelliteLink(peer, changeID string) string {
	return "## Part of\n`" + peer + ":" + changeID + "`\n\n## Branch\nmain\n"
}

// canonicalLink builds a "## Parts", "## Merge order" and "## Branch"
// link.md naming one part.
func canonicalLink(peer, changeID string) string {
	return "## Parts\n`" + peer + ":" + changeID + "`\n\n" +
		"## Merge order\n1. `" + peer + "`\n2. `.`\n\n" +
		"## Branch\nmain\n"
}

func TestListSatellite(t *testing.T) {
	parent := t.TempDir()
	canBase := testtree.Build(t, parent, "can", nil, "sat ../sat")
	testtree.Change(t, canBase, "kan-1-first", false, map[string]string{
		"proposal.md": "# kan-1-first\n\n## Why\nBecause.\n\n## What changes\n- a thing\n",
		"tasks.md":    canonicalTasks,
	})
	satBase := testtree.Build(t, parent, "sat", nil, "can ../can")
	testtree.Change(t, satBase, "kan-1-first", false, map[string]string{
		"link.md": satelliteLink("can", "kan-1-first"),
	})

	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", satBase}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	want := "kan-1-first  ->  can  2/3\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestListSatellitePeerAbsent(t *testing.T) {
	parent := t.TempDir()
	// "can" is declared in peers but never built on disk, so it is
	// PeerNotPresent, not PeerFound.
	satBase := testtree.Build(t, parent, "sat", nil, "can ../can")
	testtree.Change(t, satBase, "kan-1-first", false, map[string]string{
		"link.md": satelliteLink("can", "kan-1-first"),
	})

	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", satBase}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	want := "kan-1-first  ->  can  -\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestListJSONLinks(t *testing.T) {
	parent := t.TempDir()
	canBase := testtree.Build(t, parent, "can", nil, "sat ../sat")
	testtree.Change(t, canBase, "kan-1-first", false, map[string]string{
		"proposal.md": "# kan-1-first\n\n## Why\nBecause.\n\n## What changes\n- a thing\n",
		"tasks.md":    canonicalTasks,
		"link.md":     canonicalLink("sat", "kan-1-first"),
	})
	testtree.Change(t, canBase, "kan-2-plain", false, map[string]string{
		"proposal.md": "# kan-2-plain\n\n## Why\nBecause.\n\n## What changes\n- a thing\n",
		"tasks.md":    "# kan-2-plain\n\n- [ ] 1. One\n",
	})
	satBase := testtree.Build(t, parent, "sat", nil, "can ../can")
	testtree.Change(t, satBase, "kan-1-first", false, map[string]string{
		"link.md": satelliteLink("can", "kan-1-first"),
	})

	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", canBase, "--json"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	var gotCan struct {
		Changes []map[string]any `json:"changes"`
	}
	if err := json.Unmarshal(out.Bytes(), &gotCan); err != nil {
		t.Fatalf("json: %v (%s)", err, out.String())
	}
	var first, second map[string]any
	for _, row := range gotCan.Changes {
		switch row["id"] {
		case "kan-1-first":
			first = row
		case "kan-2-plain":
			second = row
		}
	}
	if first == nil || second == nil {
		t.Fatalf("json = %+v", gotCan.Changes)
	}
	if _, ok := first["parts"]; !ok {
		t.Errorf("canonical row missing %q key: %+v", "parts", first)
	}
	if _, ok := first["partOf"]; ok {
		t.Errorf("canonical row must not carry %q key: %+v", "partOf", first)
	}
	parts, ok := first["parts"].([]any)
	if !ok || len(parts) != 1 {
		t.Fatalf("parts = %+v", first["parts"])
	}
	part := parts[0].(map[string]any)
	if part["peer"] != "sat" || part["changeId"] != "kan-1-first" {
		t.Errorf("part = %+v", part)
	}
	if _, ok := second["parts"]; ok {
		t.Errorf("unlinked row must not carry %q key: %+v", "parts", second)
	}
	if _, ok := second["partOf"]; ok {
		t.Errorf("unlinked row must not carry %q key: %+v", "partOf", second)
	}

	out.Reset()
	errBuf.Reset()
	if code := List([]string{"--root", satBase, "--json"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	var gotSat struct {
		Changes []map[string]any `json:"changes"`
	}
	if err := json.Unmarshal(out.Bytes(), &gotSat); err != nil {
		t.Fatalf("json: %v (%s)", err, out.String())
	}
	if len(gotSat.Changes) != 1 {
		t.Fatalf("changes = %+v", gotSat.Changes)
	}
	satRow := gotSat.Changes[0]
	if _, ok := satRow["partOf"]; !ok {
		t.Errorf("satellite row missing %q key: %+v", "partOf", satRow)
	}
	if _, ok := satRow["parts"]; ok {
		t.Errorf("satellite row must not carry %q key: %+v", "parts", satRow)
	}
	partOf, ok := satRow["partOf"].(map[string]any)
	if !ok || partOf["peer"] != "can" || partOf["changeId"] != "kan-1-first" {
		t.Errorf("partOf = %+v", satRow["partOf"])
	}
	if _, ok := satRow["progressUnknown"]; ok {
		t.Errorf("resolved satellite row must not carry %q key: %+v", "progressUnknown", satRow)
	}
}

// TestListMalformedLinkDegrades covers design.md's list-degrades-and-says-so:
// a syntactically broken link.md in one change directory must not abort the
// whole listing (the failure mode task 4's review found) — that row renders
// without link information, and every other change still renders.
func TestListMalformedLinkDegrades(t *testing.T) {
	parent := t.TempDir()
	base := testtree.Build(t, parent, "sat", nil, "")
	testtree.Change(t, base, "kan-1-broken", false, map[string]string{
		// "## Part of" with a body line that isn't a single backtick-
		// delimited span: LinkFile errors on this.
		"link.md": "## Part of\nnot-a-code-span\n\n## Branch\nmain\n",
	})
	testtree.Change(t, base, "kan-2-other", false, map[string]string{
		"proposal.md": "# kan-2-other\n\n## Why\nBecause.\n\n## What changes\n- a thing\n",
		"tasks.md":    "# kan-2-other\n\n- [x] 1. One\n",
	})

	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	want := "kan-1-broken  0/0\nkan-2-other  1/1\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

// TestListMalformedPeersDegrades covers the same decision for a malformed
// peers file: it must not stop list from rendering every change, including
// a satellite whose progress simply can't be resolved without it.
func TestListMalformedPeersDegrades(t *testing.T) {
	parent := t.TempDir()
	base := testtree.Build(t, parent, "sat", nil, "")
	if err := os.WriteFile(filepath.Join(base, "spectre", "peers"), []byte("this is not name-and-path\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testtree.Change(t, base, "kan-1-first", false, map[string]string{
		"link.md": satelliteLink("can", "kan-1-first"),
	})

	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", base}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	want := "kan-1-first  ->  can  -\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

// TestListJSONProgressUnknown covers design.md's list-degrades-and-says-so
// JSON half: a satellite whose progress could not be resolved must carry an
// explicit "progressUnknown" marker rather than serialising indistinguishably
// from a canonical change with genuinely zero tasks.
func TestListJSONProgressUnknown(t *testing.T) {
	parent := t.TempDir()
	// "can" is declared but never built on disk: PeerNotPresent.
	base := testtree.Build(t, parent, "sat", nil, "can ../can")
	testtree.Change(t, base, "kan-1-first", false, map[string]string{
		"link.md": satelliteLink("can", "kan-1-first"),
	})

	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", base, "--json"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	var got struct {
		Changes []map[string]any `json:"changes"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json: %v (%s)", err, out.String())
	}
	if len(got.Changes) != 1 {
		t.Fatalf("changes = %+v", got.Changes)
	}
	row := got.Changes[0]
	if row["done"] != float64(0) || row["total"] != float64(0) {
		t.Errorf("done/total = %v/%v, want 0/0", row["done"], row["total"])
	}
	unknown, ok := row["progressUnknown"]
	if !ok {
		t.Fatalf("row missing %q key: %+v", "progressUnknown", row)
	}
	if unknown != true {
		t.Errorf("progressUnknown = %v, want true", unknown)
	}
}
