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

// RefFindings resolves every reference in specs against this tree and its
// declared peers.
func RefFindings(t *tree.Tree, specs []model.Spec) ([]Finding, error) {
	peers, err := t.Peers()
	if err != nil {
		return nil, err
	}
	self := indexOf(specs)
	peerIdx := map[string]index{}              // peer name -> its index
	resolved := map[string]tree.ResolvedPeer{} // peer name -> its resolution, cached

	var out []Finding
	for _, s := range specs {
		for _, r := range s.Reqs {
			for _, ref := range r.Refs {
				at := Finding{File: rel(t, s.Path), Line: ref.Line}

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

				rp, cached := resolved[ref.Peer]
				if !cached {
					rp = tree.ResolvePeer(peers, ref.Peer)
					resolved[ref.Peer] = rp
				}
				switch rp.Resolution {
				case tree.PeerNotDeclared:
					at.Msg = fmt.Sprintf("%s: peer %q is not declared in peers", ref.Raw, ref.Peer)
					out = append(out, at)
					continue
				case tree.PeerNotPresent:
					at.Msg = fmt.Sprintf("%s: peer %q is declared but its tree is not present at %s", ref.Raw, ref.Peer, rp.Path)
					out = append(out, at)
					continue
				case tree.PeerUnreadable:
					at.Msg = fmt.Sprintf("%s: peer %q could not be read at %s: %v", ref.Raw, ref.Peer, rp.Path, rp.Err)
					out = append(out, at)
					continue
				}

				idx, loaded := peerIdx[ref.Peer]
				if !loaded {
					pSpecs, err := rp.Tree.Specs()
					if err != nil {
						resolved[ref.Peer] = tree.ResolvedPeer{Name: ref.Peer, Path: rp.Path, Resolution: tree.PeerUnreadable, Err: err}
						at.Msg = fmt.Sprintf("%s: peer %q could not be read at %s: %v", ref.Raw, ref.Peer, rp.Path, err)
						out = append(out, at)
						continue
					}
					idx = indexOf(pSpecs)
					peerIdx[ref.Peer] = idx
				}
				ids, ok := idx[ref.Capability]
				if !ok {
					at.Msg = fmt.Sprintf("%s: no capability %q in peer %q", ref.Raw, ref.Capability, ref.Peer)
					out = append(out, at)
					continue
				}
				if !ids[ref.ID] {
					at.Msg = fmt.Sprintf("%s: no requirement %s in %s:%s", ref.Raw, ref.ID, ref.Peer, ref.Capability)
					out = append(out, at)
				}
			}
		}
	}
	return out, nil
}
