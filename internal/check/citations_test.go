package check

import (
	"testing"

	"github.com/tweety53/spectre/internal/model"
)

func TestCitationsThisTree(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{
			"auth":  "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n",
			"plans": "# plans\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL b (@auth#R1).\n",
		},
		map[string]string{},
		"")
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	got := Citations(tr, "auth", "R1", specs, nil)
	if len(got) != 1 || got[0].Path != "specs/plans.md" || got[0].From != "R1" {
		t.Errorf("got %+v", got)
	}
}

func TestCitationsPeer(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n"},
		map[string]string{"billing": "# billing\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL d (@app:auth#R1).\n"},
		"gymie ../gymie\n")
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	peerSpecs := []model.Spec{
		{Capability: "billing", Reqs: []model.Requirement{
			{ID: "R1", Line: 7, Refs: []model.Ref{{Peer: "app", Capability: "auth", ID: "R1", Raw: "@app:auth#R1", Line: 7}}},
		}, Path: "/peer/spectre/specs/billing.md"},
	}
	sources := []PeerCitationSource{{
		Name:     "gymie",
		Root:     "/peer/spectre",
		Specs:    peerSpecs,
		OurNames: map[string]bool{"app": true},
	}}
	got := Citations(tr, "auth", "R1", specs, sources)
	if len(got) != 1 || got[0].Peer != "gymie" || got[0].Path != "specs/billing.md" {
		t.Errorf("got %+v", got)
	}
}

func TestCitationsNone(t *testing.T) {
	tr := twoTrees(t,
		map[string]string{"auth": "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n"},
		map[string]string{}, "")
	specs, err := tr.Specs()
	if err != nil {
		t.Fatal(err)
	}
	if got := Citations(tr, "auth", "R1", specs, nil); len(got) != 0 {
		t.Errorf("got %+v", got)
	}
}
