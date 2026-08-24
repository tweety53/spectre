package check

import (
	"fmt"
	"os"

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
	peerIdx := map[string]index{}    // peer name -> its index
	missing := map[string]bool{}     // peer name -> declared but absent on disk
	unreadable := map[string]error{} // peer name -> declared, present, but could not be opened/read

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

				path, declared := peers[ref.Peer]
				if !declared {
					at.Msg = fmt.Sprintf("%s: peer %q is not declared in peers", ref.Raw, ref.Peer)
					out = append(out, at)
					continue
				}
				if missing[ref.Peer] {
					at.Msg = fmt.Sprintf("%s: peer %q is declared but its tree is not present at %s", ref.Raw, ref.Peer, path)
					out = append(out, at)
					continue
				}
				if uerr, failed := unreadable[ref.Peer]; failed {
					at.Msg = fmt.Sprintf("%s: peer %q could not be read at %s: %v", ref.Raw, ref.Peer, path, uerr)
					out = append(out, at)
					continue
				}
				idx, loaded := peerIdx[ref.Peer]
				if !loaded {
					if _, err := os.Stat(path); err != nil {
						missing[ref.Peer] = true
						at.Msg = fmt.Sprintf("%s: peer %q is declared but its tree is not present at %s", ref.Raw, ref.Peer, path)
						out = append(out, at)
						continue
					}
					pt, openErr := tree.Open(path)
					var loadErr error
					if openErr != nil {
						loadErr = openErr
					} else {
						var pSpecs []model.Spec
						pSpecs, loadErr = pt.Specs()
						if loadErr == nil {
							idx = indexOf(pSpecs)
							peerIdx[ref.Peer] = idx
						}
					}
					if loadErr != nil {
						unreadable[ref.Peer] = loadErr
						at.Msg = fmt.Sprintf("%s: peer %q could not be read at %s: %v", ref.Raw, ref.Peer, path, loadErr)
						out = append(out, at)
						continue
					}
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
