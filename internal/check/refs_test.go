package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/spectre/internal/tree"
)

// twoTrees builds app/ and gymie/ side by side, with app declaring gymie
// as a peer, and returns the app tree.
func twoTrees(t *testing.T, appSpecs, gymieSpecs map[string]string, peers string) *tree.Tree {
	t.Helper()
	parent := t.TempDir()
	build := func(name string, specs map[string]string) string {
		base := filepath.Join(parent, name)
		if err := os.MkdirAll(filepath.Join(base, "spectre", "specs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(base, "spectre", "changes"), 0o755); err != nil {
			t.Fatal(err)
		}
		for cap, body := range specs {
			if err := os.WriteFile(filepath.Join(base, "spectre", "specs", cap+".md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return base
	}
	app := build("app", appSpecs)
	build("gymie", gymieSpecs)
	if peers != "" {
		if err := os.WriteFile(filepath.Join(app, "spectre", "peers"), []byte(peers), 0o644); err != nil {
			t.Fatal(err)
		}
	}
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
	got, err := RefFindings(tr, specs)
	if err != nil {
		t.Fatal(err)
	}
	return msgs(got)
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

func TestRefFindingsMalformedPeerReference(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a (@gymie:R4).\n"},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d.\n"},
		"gymie ../gymie\n")
	if got := findingsFor(t, tr); !strings.Contains(got, "@gymie:R4: malformed reference, want <peer>:<capability>#<id>") {
		t.Errorf("got:\n%s", got)
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

	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RefFindings(tr, specs)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if msg := msgs(got); !strings.Contains(msg, "peer \"gymie\" could not be read at") {
		t.Errorf("got:\n%s", msg)
	}
}
