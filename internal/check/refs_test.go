package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/testtree"
	"github.com/tweety53/spectre/internal/tree"
)

// twoTrees builds app/ and gymie/ side by side, with app declaring gymie
// as a peer, and returns the app tree.
func twoTrees(t *testing.T, appSpecs, gymieSpecs map[string]string, peers string) *tree.Tree {
	t.Helper()
	parent := t.TempDir()
	app := testtree.Build(t, parent, "app", appSpecs, peers)
	testtree.Build(t, parent, "gymie", gymieSpecs, "")
	tr, err := tree.Find(app)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

func findingsFor(t *testing.T, tr *tree.Tree) string {
	t.Helper()
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	got := RefFindings(tr.Root, specs, resolvePeers(t, tr))
	return msgs(got)
}

// resolvePeers mirrors what cmd.Validate does: resolve every declared peer
// name before calling the pure RefFindings.
func resolvePeers(t *testing.T, tr *tree.Tree) map[string]tree.ResolvedPeer {
	t.Helper()
	peers, err := tr.Peers()
	if err != nil {
		t.Fatal(err)
	}
	resolved := map[string]tree.ResolvedPeer{}
	for name := range peers {
		resolved[name] = tree.ResolvePeer(peers, name)
	}
	return resolved
}

func TestRefFindingsResolves(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{
			"auth":  "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@plans#R1).\n- R2: The system SHALL b (@gymie:billing#R1).\n",
			"plans": "# plans\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL c (@R1).\n",
		},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"},
		"gymie ../gymie\n")
	if got := findingsFor(t, tr); got != "" {
		t.Errorf("want clean, got:\n%s", got)
	}
}

func TestRefFindingsUndeclaredPeer(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@gymie:billing#R1).\n"},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"},
		"")
	if got := findingsFor(t, tr); !strings.Contains(got, "peer \"gymie\" is not declared in peers") {
		t.Errorf("got:\n%s", got)
	}
}

func TestRefFindingsMissingPeerPath(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@ghost:billing#R1).\n"},
		map[string]string{},
		"ghost ../ghost\n")
	if got := findingsFor(t, tr); !strings.Contains(got, "peer \"ghost\" is declared but its tree is not present") {
		t.Errorf("got:\n%s", got)
	}
}

func TestRefFindingsUnknownCapabilityAndID(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@gymie:nope#R1) (@gymie:billing#R9) (@R7).\n"},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"},
		"gymie ../gymie\n")
	got := findingsFor(t, tr)
	for _, want := range []string{
		"no capability \"nope\" in peer \"gymie\"",
		"no requirement R9 in gymie:billing",
		"no requirement R7 in auth",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRefFindingsUnknownCapabilityInThisTree(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{
			"auth":  "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@other-cap#R1).\n",
			"plans": "# plans\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL b.\n",
		},
		map[string]string{}, "")
	if got := findingsFor(t, tr); !strings.Contains(got, "no capability \"other-cap\" in this tree") {
		t.Errorf("got:\n%s", got)
	}
}

func TestRefFindingsMalformedPeerReference(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@gymie:R4).\n"},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"},
		"gymie ../gymie\n")
	if got := findingsFor(t, tr); !strings.Contains(got, "@gymie:R4: malformed reference, want <peer>:<capability>#<id>") {
		t.Errorf("got:\n%s", got)
	}
}

// TestRefFindingsCrossTreeDifferentIDPrefix pins the case the config
// design calls out explicitly: app (default "R" prefix) cites a
// requirement in its peer gymie, which is configured with a different
// id prefix ("REQ-"). Each tree parses its own files by its own
// configuration, and the citation carries gymie's id exactly as
// written, so the reference must still resolve.
func TestRefFindingsCrossTreeDifferentIDPrefix(t *testing.T) {
	parent := t.TempDir()
	writeTree := func(name, configBody string, specs map[string]string, peers string) string {
		t.Helper()
		base := filepath.Join(parent, name)
		if err := os.MkdirAll(filepath.Join(base, "spectre", "specs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if configBody != "" {
			if err := os.WriteFile(filepath.Join(base, "spectre", "config.md"), []byte(configBody), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		for capName, body := range specs {
			p := filepath.Join(base, "spectre", "specs", capName+".md")
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if peers != "" {
			if err := os.WriteFile(filepath.Join(base, "spectre", "peers"), []byte(peers), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return base
	}

	app := writeTree("app", "", map[string]string{
		"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@gymie:billing#REQ-1).\n",
	}, "gymie ../gymie\n")
	writeTree("gymie", "## Vocabulary\n- id-prefix: REQ-\n", map[string]string{
		"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- REQ-1: The system SHALL d.\n",
	}, "")

	tr, err := tree.Find(app)
	if err != nil {
		t.Fatal(err)
	}

	// Confirm the reference was actually recognized and extracted, not
	// silently dropped by a parser anchored to app's own "R" prefix — a
	// dropped reference and a resolved one both produce zero findings,
	// so an empty findingsFor result alone would not distinguish them.
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 || len(specs[0].Reqs) != 1 || len(specs[0].Reqs[0].Refs) != 1 {
		t.Fatalf("want one extracted reference, got specs = %+v", specs)
	}
	if ref := specs[0].Reqs[0].Refs[0]; ref.Peer != "gymie" || ref.Capability != "billing" || ref.ID != "REQ-1" {
		t.Errorf("extracted ref = %+v, want peer gymie, capability billing, id REQ-1", ref)
	}

	if got := findingsFor(t, tr); got != "" {
		t.Errorf("cross-tree reference to a peer with a different id prefix should resolve, got:\n%s", got)
	}
}

func TestRefFindingsPeerUnreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: file modes are not enforced")
	}
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@gymie:billing#R1).\n"},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"},
		"gymie ../gymie\n")
	specPath := filepath.Join(filepath.Dir(tr.Root), "..", "gymie", "spectre", "specs", "billing.md")
	if err := os.Chmod(specPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(specPath, 0o644) })

	if got := findingsFor(t, tr); !strings.Contains(got, "peer \"gymie\" could not be read at") {
		t.Errorf("got:\n%s", got)
	}
}
