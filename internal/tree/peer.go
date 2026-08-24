package tree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tweety53/spectre/internal/model"
)

// PeerResolution classifies the outcome of resolving a declared peer name.
type PeerResolution int

const (
	// PeerFound means the peer is declared, present, and its specs loaded.
	PeerFound PeerResolution = iota
	// PeerNotDeclared means name has no entry in the peers map.
	PeerNotDeclared
	// PeerNotPresent means the declared path does not exist on disk.
	PeerNotPresent
	// PeerUnreadable means the declared path exists but either could not be
	// opened as a tree, or its specs could not be read; Err holds the
	// underlying cause either way.
	PeerUnreadable
)

// String names r for diagnostics and log output.
func (r PeerResolution) String() string {
	switch r {
	case PeerFound:
		return "found"
	case PeerNotDeclared:
		return "not declared"
	case PeerNotPresent:
		return "not present"
	case PeerUnreadable:
		return "unreadable"
	default:
		return fmt.Sprintf("PeerResolution(%d)", int(r))
	}
}

// ResolvedPeer is the outcome of resolving one declared peer name.
type ResolvedPeer struct {
	Name       string
	Path       string // declared path; empty when Resolution == PeerNotDeclared
	Resolution PeerResolution
	Tree       *Tree        // non-nil only when Resolution == PeerFound
	Specs      []model.Spec // set only when Resolution == PeerFound
	Err        error        // non-nil only when Resolution == PeerUnreadable
}

// ResolvePeer resolves name against peers (as returned by Tree.Peers) and,
// when found, loads its specs. Failure is classified as not declared, not
// present, or unreadable — a path that doesn't stat, a path that isn't a
// spectre tree, and a tree whose specs failed to load are all
// PeerUnreadable, with Err set to the underlying cause. Callers never need
// to load a resolved peer's specs or classify its failures themselves.
func ResolvePeer(peers map[string]string, name string) ResolvedPeer {
	path, declared := peers[name]
	if !declared {
		return ResolvedPeer{Name: name, Resolution: PeerNotDeclared}
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ResolvedPeer{Name: name, Path: path, Resolution: PeerNotPresent}
		}
		return ResolvedPeer{Name: name, Path: path, Resolution: PeerUnreadable, Err: err}
	}
	pt, err := Open(path)
	if err != nil {
		return ResolvedPeer{Name: name, Path: path, Resolution: PeerUnreadable, Err: err}
	}
	specs, err := pt.Specs()
	if err != nil {
		return ResolvedPeer{Name: name, Path: path, Resolution: PeerUnreadable, Err: err}
	}
	return ResolvedPeer{Name: name, Path: path, Resolution: PeerFound, Tree: pt, Specs: specs}
}

// NamesFor returns the subset of declared (a peer's own declared-peers map)
// whose path resolves to the same directory as root — i.e. the name(s) by
// which that peer refers back to root.
func NamesFor(declared map[string]string, root string) map[string]bool {
	out := map[string]bool{}
	for name, path := range declared {
		if sameDir(path, root) {
			out[name] = true
		}
	}
	return out
}

func sameDir(a, b string) bool {
	ra, erra := filepath.EvalSymlinks(a)
	rb, errb := filepath.EvalSymlinks(b)
	if erra != nil || errb != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}
