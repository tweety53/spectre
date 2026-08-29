package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/spectre/internal/config"
	"github.com/tweety53/spectre/internal/model"
)

func writeLink(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "link.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseLink(t *testing.T) {
	cases := []struct {
		name string
		body string
		want model.Link
	}{
		{
			name: "satellite side",
			body: "## Part of\n\n`agents:kan-363-example`\n\n## Branch\n\nkan-363-example\n\n" +
				"## Tasks here\n\n1-5\n",
			want: model.Link{
				PartOf:    model.LinkRef{Peer: "agents", ChangeID: "kan-363-example"},
				Branch:    "kan-363-example",
				TasksHere: []int{1, 2, 3, 4, 5},
			},
		},
		{
			name: "canonical side",
			body: "## Parts\n\n`spectre:kan-363-example`\n\n## Branch\n\nkan-363-example\n\n" +
				"## Merge order\n\n1. `spectre`\n2. `.`\n",
			want: model.Link{
				Parts:      []model.LinkRef{{Peer: "spectre", ChangeID: "kan-363-example"}},
				Branch:     "kan-363-example",
				MergeOrder: []string{"spectre", "."},
			},
		},
		{
			name: "canonical side with multiple parts",
			body: "## Parts\n\n`spectre:kan-363-example`\n`web:kan-363-example`\n\n## Branch\n\n" +
				"kan-363-example\n\n## Merge order\n\n1. `.`\n2. `spectre`\n3. `web`\n",
			want: model.Link{
				Parts: []model.LinkRef{
					{Peer: "spectre", ChangeID: "kan-363-example"},
					{Peer: "web", ChangeID: "kan-363-example"},
				},
				Branch:     "kan-363-example",
				MergeOrder: []string{".", "spectre", "web"},
			},
		},
		{
			name: "tasks here: comma list and ranges",
			body: "## Part of\n\n`agents:kan-363-example`\n\n## Branch\n\nkan-363-example\n\n" +
				"## Tasks here\n\n1,3,7-9\n",
			want: model.Link{
				PartOf:    model.LinkRef{Peer: "agents", ChangeID: "kan-363-example"},
				Branch:    "kan-363-example",
				TasksHere: []int{1, 3, 7, 8, 9},
			},
		},
		{
			name: "branch shape allows dots, underscores, slashes, hyphens",
			body: "## Part of\n\n`agents:kan-363-example`\n\n## Branch\n\n" +
				"feature/kan-363_example.v2\n\n## Tasks here\n\n1\n",
			want: model.Link{
				PartOf:    model.LinkRef{Peer: "agents", ChangeID: "kan-363-example"},
				Branch:    "feature/kan-363_example.v2",
				TasksHere: []int{1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := writeLink(t, tc.body)
			parser := New(config.Default())
			got, err := parser.LinkFile(p)
			if err != nil {
				t.Fatalf("LinkFile: %v", err)
			}
			if got.PartOf != tc.want.PartOf {
				t.Errorf("PartOf = %+v, want %+v", got.PartOf, tc.want.PartOf)
			}
			if len(got.Parts) != len(tc.want.Parts) {
				t.Fatalf("len(Parts) = %d, want %d", len(got.Parts), len(tc.want.Parts))
			}
			for i := range got.Parts {
				if got.Parts[i] != tc.want.Parts[i] {
					t.Errorf("Parts[%d] = %+v, want %+v", i, got.Parts[i], tc.want.Parts[i])
				}
			}
			if got.Branch != tc.want.Branch {
				t.Errorf("Branch = %q, want %q", got.Branch, tc.want.Branch)
			}
			if len(got.MergeOrder) != len(tc.want.MergeOrder) {
				t.Fatalf("len(MergeOrder) = %d, want %d", len(got.MergeOrder), len(tc.want.MergeOrder))
			}
			for i := range got.MergeOrder {
				if got.MergeOrder[i] != tc.want.MergeOrder[i] {
					t.Errorf("MergeOrder[%d] = %q, want %q", i, got.MergeOrder[i], tc.want.MergeOrder[i])
				}
			}
			if len(got.TasksHere) != len(tc.want.TasksHere) {
				t.Fatalf("len(TasksHere) = %d, want %d", len(got.TasksHere), len(tc.want.TasksHere))
			}
			for i := range got.TasksHere {
				if got.TasksHere[i] != tc.want.TasksHere[i] {
					t.Errorf("TasksHere[%d] = %d, want %d", i, got.TasksHere[i], tc.want.TasksHere[i])
				}
			}
		})
	}
}

func TestParseLinkRejects(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{
			name: "duplicate heading",
			body: "## Parts\n\n`spectre:kan-363-example`\n\n## Parts\n\n`web:kan-363-example`\n",
		},
		{
			name: "part of with more than one line",
			body: "## Part of\n\n`agents:kan-363-example`\n`agents:kan-363-other`\n",
		},
		{
			name: "part of missing colon",
			body: "## Part of\n\n`agents-kan-363-example`\n",
		},
		{
			name: "part of not a code span",
			body: "## Part of\n\nagents:kan-363-example\n",
		},
		{
			name: "part of invalid peer characters",
			body: "## Part of\n\n`Agents:kan-363-example`\n",
		},
		{
			name: "part of invalid change id: path traversal",
			body: "## Part of\n\n`agents:..`\n",
		},
		{
			name: "part of invalid change id: contains slash",
			body: "## Part of\n\n`agents:kan/363`\n",
		},
		{
			name: "parts section empty",
			body: "## Parts\n\n## Branch\n\nkan-363-example\n",
		},
		{
			name: "branch empty",
			body: "## Branch\n\n## Tasks here\n\n1\n",
		},
		{
			name: "branch more than one line",
			body: "## Branch\n\nkan-363-example\nkan-363-other\n",
		},
		{
			name: "branch invalid shape: leading slash",
			body: "## Branch\n\n/kan-363-example\n",
		},
		{
			name: "branch invalid shape: space",
			body: "## Branch\n\nkan 363\n",
		},
		{
			name: "merge order item missing code span",
			body: "## Merge order\n\n1. spectre\n",
		},
		{
			name: "merge order item invalid peer name",
			body: "## Merge order\n\n1. `Spectre`\n",
		},
		{
			name: "tasks here descending range",
			body: "## Tasks here\n\n12-4\n",
		},
		{
			// A range names more than one task; a single task is written
			// as a bare number. parseTaskToken's doc comment says lo must
			// be strictly less than hi — this pins that "5-5" is rejected
			// rather than silently accepted as task 5 alone.
			name: "tasks here equal bounds range",
			body: "## Tasks here\n\n5-5\n",
		},
		{
			name: "tasks here not ascending",
			body: "## Tasks here\n\n3,1\n",
		},
		{
			name: "tasks here overlapping range",
			body: "## Tasks here\n\n2-5,4\n",
		},
		{
			// A bare number equal to the previous range's own high bound
			// duplicates that task rather than continuing the sequence.
			// Pins "lo <= max", not "lo < max": the boundary itself is a
			// duplicate.
			name: "tasks here duplicates range's own high bound",
			body: "## Tasks here\n\n1-5,5\n",
		},
		{
			name: "tasks here zero is not positive",
			body: "## Tasks here\n\n0\n",
		},
		{
			name: "part of and parts both present",
			body: "## Part of\n\n`agents:kan-363-example`\n\n## Parts\n\n`spectre:kan-363-example`\n",
		},
		{
			name: "part of and merge order both present",
			body: "## Part of\n\n`agents:kan-363-example`\n\n## Merge order\n\n1. `.`\n",
		},
		{
			name: "parts and tasks here both present",
			body: "## Parts\n\n`spectre:kan-363-example`\n\n## Tasks here\n\n1\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := writeLink(t, tc.body)
			parser := New(config.Default())
			if _, err := parser.LinkFile(p); err == nil {
				t.Fatalf("LinkFile: got nil error, want a parse error")
			}
		})
	}
}

// TestIsValidLinkChangeIDRejectsBacktick is a direct unit test of
// isValidLinkChangeID, the parse-side hand-kept duplicate of
// internal/cmd's isValidID. TestParseLinkRejects cannot exercise a
// backtick-bearing id through LinkFile: linkCodeSpanRe's `[^`]*` already
// stops a backtick from ever being captured as code-span content, so a
// mutation deleting isValidLinkChangeID's own backtick check would leave
// every parse-path test green. This test calls the function directly, the
// only way to prove it independently of that other guard.
func TestIsValidLinkChangeIDRejectsBacktick(t *testing.T) {
	if isValidLinkChangeID("weird`id") {
		t.Fatal("isValidLinkChangeID(\"weird`id\") = true, want false")
	}
}
