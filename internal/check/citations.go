package check

import (
	"github.com/tweety53/spectre/internal/model"
)

// Citation is one place a requirement is cited from.
type Citation struct {
	Peer string // "" for a citation found in this tree
	Path string // relative to the citing tree's root
	Line int
	From string // the citing requirement's ID
	Raw  string // the raw reference text, e.g. "@auth#R1"
}

// PeerCitationSource is one already-resolved peer's specs, plus the names
// by which that peer itself refers back to this tree (needed to recognize
// whether one of its references names us).
type PeerCitationSource struct {
	Name     string
	Root     string // the peer's absolute tree root, for relative-path display
	Specs    []model.Spec
	OurNames map[string]bool
}

// Citations finds every citation of capability#id: in this tree's specs
// (a same-file or same-tree reference resolving to it), and in each peer
// source (a reference naming this tree by any of its OurNames). root is
// this tree's absolute root, used only to render display paths — Citations
// performs no I/O of its own.
func Citations(
	root, capName, reqID string, specs []model.Spec, peers []PeerCitationSource,
) []Citation {
	var out []Citation
	for _, s := range specs {
		for _, r := range s.Reqs {
			for _, ref := range r.Refs {
				if ref.Peer != "" || ref.ID != reqID {
					continue
				}
				if ref.Capability == capName || (ref.Capability == "" && s.Capability == capName) {
					out = append(out, Citation{Path: relTo(root, s.Path), Line: ref.Line, From: r.ID, Raw: ref.Raw})
				}
			}
		}
	}
	for _, p := range peers {
		for _, s := range p.Specs {
			for _, r := range s.Reqs {
				for _, ref := range r.Refs {
					if p.OurNames[ref.Peer] && ref.Capability == capName && ref.ID == reqID {
						out = append(out, Citation{
							Peer: p.Name, Path: relTo(p.Root, s.Path), Line: ref.Line, From: r.ID, Raw: ref.Raw,
						})
					}
				}
			}
		}
	}
	return out
}
