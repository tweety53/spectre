package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/testtree"
	"github.com/tweety53/spectre/internal/tree"
)

// linkedTrees builds two peer trees, "sat" and "can", each declaring the
// other under the name the peer half of a link.md reference uses, so a
// link written between them resolves both ways. satFiles and canFiles are
// the files (by name) to write into changes/x/ (or changes/archive/x/,
// per satArchived/canArchived) in each tree; a nil map skips that
// change's directory entirely, reproducing "no counterpart at all".
func linkedTrees(t *testing.T, satFiles, canFiles map[string]string, satArchived, canArchived bool) (sat, can *tree.Tree) {
	t.Helper()
	parent := t.TempDir()
	satBase := testtree.Build(t, parent, "sat", nil, "can ../can\n")
	canBase := testtree.Build(t, parent, "can", nil, "sat ../sat\n")
	if satFiles != nil {
		testtree.Change(t, satBase, "x", satArchived, satFiles)
	}
	if canFiles != nil {
		testtree.Change(t, canBase, "x", canArchived, canFiles)
	}
	var err error
	sat, err = tree.Open(satBase)
	if err != nil {
		t.Fatal(err)
	}
	can, err = tree.Open(canBase)
	if err != nil {
		t.Fatal(err)
	}
	return sat, can
}

// findings runs Structural over tr's whole tree, resolving tr's own
// declared peers first — the same sequence internal/cmd/validate.go's
// whole-tree route uses.
func findings(t *testing.T, tr *tree.Tree) []Finding {
	t.Helper()
	declared, err := tr.Peers()
	if err != nil {
		t.Fatal(err)
	}
	peers := map[string]tree.ResolvedPeer{}
	for name := range declared {
		peers[name] = tree.ResolvePeer(declared, name)
	}
	got, err := Structural(tr, "", peers)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func findingMsgs(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.String())
		b.WriteString("\n")
	}
	return b.String()
}

const goodProposal = "# x\n\n## Why\nB.\n\n## What changes\nC.\n"
const oneTaskTasksFile = "# x\n\n- [ ] 1. a\n"

func satLinkMD(branch string, tasksHere string) string {
	return "## Part of\n\n`can:x`\n\n## Branch\n\n" + branch + "\n\n## Tasks here\n\n" + tasksHere + "\n"
}

func canLinkMD(branch, mergeOrder string) string {
	return "## Parts\n\n`sat:x`\n\n## Branch\n\n" + branch + "\n\n## Merge order\n\n" + mergeOrder + "\n"
}

const goodMergeOrder = "1. `sat`\n2. `.`\n"

func TestLinkFindings(t *testing.T) {
	cases := []struct {
		name        string
		satFiles    map[string]string
		canFiles    map[string]string
		satArchived bool
		canArchived bool
		side        string // "sat" or "can": which tree to run Structural over
		want        string
	}{
		{
			name:     "peer not declared",
			satFiles: map[string]string{"link.md": satLinkMD("b", "1")},
			side:     "sat-no-peers", // handled specially below
			want:     `peer "can" is not declared`,
		},
		{
			name:     "counterpart missing entirely",
			satFiles: map[string]string{"link.md": satLinkMD("b", "1")},
			canFiles: nil,
			side:     "sat",
			want:     "one-sided link",
		},
		{
			name:     "counterpart exists but does not link back",
			satFiles: map[string]string{"link.md": satLinkMD("b", "1")},
			canFiles: map[string]string{"proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			side:     "sat",
			want:     "one-sided link",
		},
		{
			name:        "archive skew: canonical archived, satellite open",
			satFiles:    map[string]string{"link.md": satLinkMD("b", "1")},
			canFiles:    map[string]string{"link.md": canLinkMD("b", goodMergeOrder), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			canArchived: true,
			side:        "sat",
			want:        "archive skew",
		},
		{
			name:     "merge order missing an entry",
			canFiles: map[string]string{"link.md": canLinkMD("b", "1. `sat`\n"), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			side:     "can",
			want:     `merge order: missing "."`,
		},
		{
			name:     "merge order duplicated entry",
			canFiles: map[string]string{"link.md": canLinkMD("b", "1. `sat`\n2. `.`\n3. `sat`\n"), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			side:     "can",
			want:     `merge order: "sat" appears 2 times`,
		},
		{
			name:     "merge order unknown entry",
			canFiles: map[string]string{"link.md": canLinkMD("b", "1. `sat`\n2. `.`\n3. `ghost`\n"), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			side:     "can",
			want:     `merge order: "ghost" is not "." or a declared part`,
		},
		{
			name:     "branch mismatch",
			satFiles: map[string]string{"link.md": satLinkMD("sat-branch", "1")},
			canFiles: map[string]string{"link.md": canLinkMD("can-branch", goodMergeOrder), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			side:     "sat",
			want:     `branch mismatch: "sat-branch" here, "can-branch" in can:x`,
		},
		{
			name:     "tasks here cites a task the canonical plan does not have",
			satFiles: map[string]string{"link.md": satLinkMD("b", "1,2")},
			canFiles: map[string]string{"link.md": canLinkMD("b", goodMergeOrder), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			side:     "sat",
			want:     `tasks here: 2 is not a task in can:x`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sat, can *tree.Tree
			if tc.side == "sat-no-peers" {
				parent := t.TempDir()
				satBase := testtree.Build(t, parent, "sat", nil, "")
				testtree.Change(t, satBase, "x", false, tc.satFiles)
				var err error
				sat, err = tree.Open(satBase)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				sat, can = linkedTrees(t, tc.satFiles, tc.canFiles, tc.satArchived, tc.canArchived)
			}

			var got []Finding
			switch {
			case tc.side == "sat-no-peers", tc.side == "sat":
				got = findings(t, sat)
			case tc.side == "can":
				got = findings(t, can)
			default:
				t.Fatalf("unknown side %q", tc.side)
			}

			if msgs := findingMsgs(got); !strings.Contains(msgs, tc.want) {
				t.Errorf("findings =\n%s\nwant a finding containing %q", msgs, tc.want)
			}
		})
	}
}

func TestLinkFindingsQuiet(t *testing.T) {
	t.Run("peer declared but absent", func(t *testing.T) {
		parent := t.TempDir()
		satBase := testtree.Build(t, parent, "sat", nil, "can ../can\n") // "can" never built
		testtree.Change(t, satBase, "x", false, map[string]string{"link.md": satLinkMD("b", "1")})
		sat, err := tree.Open(satBase)
		if err != nil {
			t.Fatal(err)
		}
		if got := findings(t, sat); len(got) != 0 {
			t.Errorf("want no findings for a declared-but-absent peer, got:\n%s", findingMsgs(got))
		}
	})

	// "canonical archived with the part archived too" is quiet for a
	// reason outside LinkFindings' own logic: tree.Changes() (read.go)
	// reads only changes/, never changes/archive/, so once a change is
	// archived Structural never visits it as c and LinkFindings never
	// runs for it at all — on either side. Calling LinkFindings directly
	// cannot reproduce that quietness: the function has no "this side is
	// archived" input, so archived-counterpart-while-checked-as-open
	// always trips the archive-skew finding (see "archive skew" in
	// TestLinkFindings), regardless of whether the side being checked is
	// itself archived. The quietness is Structural's iteration boundary,
	// not a rule LinkFindings evaluates, so this pins that boundary
	// directly rather than staging a call that cannot exercise it.
	t.Run("canonical archived with the part archived too (structural quietness, not a link rule)", func(t *testing.T) {
		sat, can := linkedTrees(t,
			map[string]string{"link.md": satLinkMD("b", "1")},
			map[string]string{"link.md": canLinkMD("b", goodMergeOrder), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			true, true, // both archived
		)
		satChanges, err := sat.Changes()
		if err != nil {
			t.Fatal(err)
		}
		if len(satChanges) != 0 {
			t.Fatalf("want the archived satellite invisible to Changes(), got %v", satChanges)
		}
		canChanges, err := can.Changes()
		if err != nil {
			t.Fatal(err)
		}
		if len(canChanges) != 0 {
			t.Fatalf("want the archived canonical invisible to Changes(), got %v", canChanges)
		}
		if got := findings(t, sat); len(got) != 0 {
			t.Errorf("sat: want no findings, got:\n%s", findingMsgs(got))
		}
		if got := findings(t, can); len(got) != 0 {
			t.Errorf("can: want no findings, got:\n%s", findingMsgs(got))
		}
	})

	t.Run("link-only satellite directory, clean round trip", func(t *testing.T) {
		sat, can := linkedTrees(t,
			map[string]string{"link.md": satLinkMD("b", "1")},
			map[string]string{"link.md": canLinkMD("b", goodMergeOrder), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
			false, false,
		)
		if got := findings(t, sat); len(got) != 0 {
			t.Errorf("sat: want no findings, got:\n%s", findingMsgs(got))
		}
		if got := findings(t, can); len(got) != 0 {
			t.Errorf("can: want no findings, got:\n%s", findingMsgs(got))
		}
	})
}

// TestStructuralLinkPlusDesignKeepsDesignChecks pins that "nothing else"
// in the satellite definition (design.md's pointer-tree-not-full-tree)
// includes design.md: a change directory carrying link.md and a design.md
// is not "link.md and nothing else", so DesignFindings must still run for
// it rather than being silently skipped as satellite classification would
// do if it checked only proposal.md and tasks.md.
func TestStructuralLinkPlusDesignKeepsDesignChecks(t *testing.T) {
	sat, _ := linkedTrees(t,
		map[string]string{"link.md": satLinkMD("b", "1"), "design.md": "nothing here\n"},
		nil,
		false, false,
	)
	got := findingMsgs(findings(t, sat))
	for _, want := range []string{`missing "## Context"`, `missing "## Decisions"`} {
		if !strings.Contains(got, want) {
			t.Errorf("a link.md + design.md change must still run DesignFindings, got:\n%s\nwant %q", got, want)
		}
	}
}

// TestLinkFindingsPeerUnreadable pins unreadable-peer-is-a-finding: a
// declared peer whose path exists but cannot be traversed is a finding,
// not silence — the same technique
// internal/cmd/validate_test.go's TestValidateReportsUnreadablePeer uses
// to force PeerUnreadable for RefFindings' identical distinction.
func TestLinkFindingsPeerUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	sat, can := linkedTrees(t,
		map[string]string{"link.md": satLinkMD("b", "1")},
		map[string]string{"link.md": canLinkMD("b", goodMergeOrder), "proposal.md": goodProposal, "tasks.md": oneTaskTasksFile},
		false, false,
	)
	canBase := filepath.Dir(can.Root)
	if err := os.Chmod(canBase, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(canBase, 0o755) })

	got := findingMsgs(findings(t, sat))
	if !strings.Contains(got, `peer "can" could not be read`) {
		t.Errorf("want a finding naming peer %q unreadable, got:\n%s", "can", got)
	}
}

// TestLinkFindingsReciprocalWithDifferentPeerNames guards namesBack's use
// of tree.NamesFor: sat and can deliberately declare each other under
// different names, so an implementation that compared ref.Peer against
// the checking side's own peer name (assuming both trees use the same
// name for each other) would wrongly call this link one-sided.
func TestLinkFindingsReciprocalWithDifferentPeerNames(t *testing.T) {
	parent := t.TempDir()
	satBase := testtree.Build(t, parent, "sat", nil, "canonical-repo ../can\n")
	canBase := testtree.Build(t, parent, "can", nil, "satellite-repo ../sat\n")
	testtree.Change(t, satBase, "x", false, map[string]string{
		"link.md": "## Part of\n\n`canonical-repo:x`\n\n## Branch\n\nb\n\n## Tasks here\n\n1\n",
	})
	testtree.Change(t, canBase, "x", false, map[string]string{
		"link.md":     "## Parts\n\n`satellite-repo:x`\n\n## Branch\n\nb\n\n## Merge order\n\n1. `satellite-repo`\n2. `.`\n",
		"proposal.md": goodProposal,
		"tasks.md":    oneTaskTasksFile,
	})
	sat, err := tree.Open(satBase)
	if err != nil {
		t.Fatal(err)
	}
	can, err := tree.Open(canBase)
	if err != nil {
		t.Fatal(err)
	}
	if got := findings(t, sat); len(got) != 0 {
		t.Errorf("sat: want a clean round trip across differently-named peers, got:\n%s", findingMsgs(got))
	}
	if got := findings(t, can); len(got) != 0 {
		t.Errorf("can: want a clean round trip across differently-named peers, got:\n%s", findingMsgs(got))
	}
}

// TestLinkFindingsStaleParticipantNameIsNotReciprocal is the case
// TestLinkFindingsReciprocalWithDifferentPeerNames does not catch: a bare
// ChangeID comparison in namesBack (ignoring tree.NamesFor entirely) would
// still pass that test, because both directions there use whatever name
// each side's own peers file happens to declare. Here can's own peers file
// renames the satellite to "othersat", but can's "## Parts" entry still
// says "sat:x" — the stale name. A correct namesBack resolves "sat" is not
// a name can's own peers file gives the satellite, so it must NOT be
// reciprocal; a bare ChangeID comparison would ignore the name entirely
// and wrongly call it reciprocal.
func TestLinkFindingsStaleParticipantNameIsNotReciprocal(t *testing.T) {
	parent := t.TempDir()
	satBase := testtree.Build(t, parent, "sat", nil, "can ../can\n")
	canBase := testtree.Build(t, parent, "can", nil, "othersat ../sat\n")
	testtree.Change(t, satBase, "x", false, map[string]string{
		"link.md": "## Part of\n\n`can:x`\n\n## Branch\n\nb\n\n## Tasks here\n\n1\n",
	})
	testtree.Change(t, canBase, "x", false, map[string]string{
		"link.md":     "## Parts\n\n`sat:x`\n\n## Branch\n\nb\n\n## Merge order\n\n1. `sat`\n2. `.`\n",
		"proposal.md": goodProposal,
		"tasks.md":    oneTaskTasksFile,
	})
	sat, err := tree.Open(satBase)
	if err != nil {
		t.Fatal(err)
	}
	got := findings(t, sat)
	if len(got) != 1 || !strings.Contains(got[0].String(), "one-sided link") {
		t.Errorf("sat: want exactly one one-sided-link finding, got:\n%s", findingMsgs(got))
	}
}

// TestIsSatellite reproduces each case as real directories on disk rather
// than asserting against a stubbed filesystem.
func TestIsSatellite(t *testing.T) {
	write := func(t *testing.T, names ...string) string {
		t.Helper()
		d := filepath.Join(t.TempDir(), "x")
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if err := os.WriteFile(filepath.Join(d, name), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return d
	}

	tests := []struct {
		name string
		dir  func(t *testing.T) string
		want bool
	}{
		{"link.md alone", func(t *testing.T) string {
			return write(t, "link.md")
		}, true},
		{"link.md plus proposal.md", func(t *testing.T) string {
			return write(t, "link.md", "proposal.md")
		}, false},
		{"link.md plus tasks.md", func(t *testing.T) string {
			return write(t, "link.md", "tasks.md")
		}, false},
		{"link.md plus design.md", func(t *testing.T) string {
			return write(t, "link.md", "design.md")
		}, false},
		{"no link.md", func(t *testing.T) string {
			return write(t, "proposal.md", "tasks.md")
		}, false},
		{"empty directory", func(t *testing.T) string {
			return write(t)
		}, false},
		{"directory does not exist", func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "does-not-exist")
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.dir(t)
			if got := IsSatellite(dir); got != tt.want {
				t.Errorf("IsSatellite(%s) = %v, want %v", dir, got, tt.want)
			}
		})
	}
}

func TestStructuralSkipsHeadingsForSatellite(t *testing.T) {
	sat, _ := linkedTrees(t,
		map[string]string{"link.md": satLinkMD("b", "1")},
		nil,
		false, false,
	)
	got := findingMsgs(findings(t, sat))
	for _, absent := range []string{"missing proposal.md", "missing tasks.md", "missing design.md"} {
		if strings.Contains(got, absent) {
			t.Errorf("satellite must skip proposal/task/design checks entirely, got:\n%s", got)
		}
	}
}
