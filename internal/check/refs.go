package check

import (
	"fmt"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/tree"
)

// index maps capability name to the set of requirement ids it declares.
type index map[string]map[string]bool

func indexOf(specs []model.Spec) index {
	idx := index{}
	for _, s := range specs {
		ids := map[string]bool{}
		for _, r := range s.Reqs {
			ids[r.ID] = true
		}
		idx[s.Capability] = ids
	}
	return idx
}

// RefFindings resolves every reference in specs against this tree (root)
// and peers, keyed by declared peer name and already resolved — Resolution
// == PeerFound entries carry their specs, PeerNotPresent/PeerUnreadable
// entries carry their classification. RefFindings does no I/O of its own:
// a name absent from peers is treated as not declared.
func RefFindings(root string, specs []model.Spec, peers map[string]tree.ResolvedPeer) []Finding {
	self := indexOf(specs)
	peerIdx := map[string]index{} // peer name -> its index, built lazily from rp.Specs

	var out []Finding
	for _, s := range specs {
		for _, r := range s.Reqs {
			for _, ref := range r.Refs {
				at := Finding{File: relTo(root, s.Path), Line: ref.Line}

				if ref.Peer == "" {
					capName := ref.Capability
					if capName == "" {
						capName = s.Capability
					}
					ids, ok := self[capName]
					if !ok {
						at.Msg = fmt.Sprintf("%s: no capability %q in this tree", ref.Raw, capName)
						out = append(out, at)
						continue
					}
					if !ids[ref.ID] {
						at.Msg = fmt.Sprintf("%s: no requirement %s in %s", ref.Raw, ref.ID, capName)
						out = append(out, at)
					}
					continue
				}

				if ref.Capability == "" {
					at.Msg = fmt.Sprintf("%s: malformed reference, want <peer>:<capability>#<id>", ref.Raw)
					out = append(out, at)
					continue
				}

				rp, declared := peers[ref.Peer]
				if !declared {
					at.Msg = fmt.Sprintf("%s: peer %q is not declared in peers", ref.Raw, ref.Peer)
					out = append(out, at)
					continue
				}
				switch rp.Resolution {
				case tree.PeerNotPresent:
					at.Msg = fmt.Sprintf("%s: peer %q is declared but its tree is not present at %s",
						ref.Raw, ref.Peer, rp.Path)
					out = append(out, at)
					continue
				case tree.PeerUnreadable:
					at.Msg = fmt.Sprintf("%s: peer %q could not be read at %s: %v",
						ref.Raw, ref.Peer, rp.Path, rp.Err)
					out = append(out, at)
					continue
				}

				idx, cached := peerIdx[ref.Peer]
				if !cached {
					idx = indexOf(rp.Specs)
					peerIdx[ref.Peer] = idx
				}
				ids, ok := idx[ref.Capability]
				if !ok {
					at.Msg = fmt.Sprintf("%s: no capability %q in peer %q", ref.Raw, ref.Capability, ref.Peer)
					out = append(out, at)
					continue
				}
				if !ids[ref.ID] {
					at.Msg = fmt.Sprintf("%s: no requirement %s in %s:%s",
						ref.Raw, ref.ID, ref.Peer, ref.Capability)
					out = append(out, at)
				}
			}
		}
	}
	return out
}
