package tree

import (
	"errors"
	"os"
	"path/filepath"
)

// PeerResolution classifies the outcome of resolving a declared peer name.
type PeerResolution int

const (
	// PeerFound means the peer is declared, present, and opened successfully.
	PeerFound PeerResolution = iota
	// PeerNotDeclared means name has no entry in the peers map.
	PeerNotDeclared
	// PeerNotPresent means the declared path does not exist on disk.
	PeerNotPresent
	// PeerUnreadable means the declared path exists but could not be opened
	// as a tree; Err holds the underlying cause.
	PeerUnreadable
)

// ResolvedPeer is the outcome of resolving one declared peer name.
type ResolvedPeer struct {
	Name       string
	Path       string // declared path; empty when Resolution == PeerNotDeclared
	Resolution PeerResolution
	Tree       *Tree // non-nil only when Resolution == PeerFound
	Err        error // non-nil only when Resolution == PeerUnreadable
}

// ResolvePeer resolves name against peers (as returned by Tree.Peers),
// classifying failure as not declared, not present, or present but
// unreadable, distinguishing a missing path from a real I/O error.
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
	return ResolvedPeer{Name: name, Path: path, Resolution: PeerFound, Tree: pt}
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
