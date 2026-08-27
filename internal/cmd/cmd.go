// Package cmd implements one function per spectre subcommand. Each returns
// the process exit code and writes nothing to the real stdout or stderr.
package cmd

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/tree"
)

// Exit codes shared by every command.
const (
	OK    = 0
	Fail  = 1
	Usage = 2
)

// flagSet builds a silent flag set carrying the shared --root flag.
func flagSet(name string, stderr io.Writer) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	usage := "path to the spectre tree (default: search upwards from the working directory)"
	root := fs.String("root", "", usage)
	return fs, root
}

// resolve opens the tree named by --root, or searches upwards from the
// working directory.
func resolve(root string) (*tree.Tree, error) {
	if root != "" {
		return tree.Open(root)
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return tree.Find(wd)
}

// displayPath renders p for a human to read: relative to the working
// directory when that reads better, the absolute path p otherwise. This is
// a display concern only — os.Getwd or filepath.Rel failing, or p sitting
// outside the working directory, all fall back to p unchanged. It never
// returns an error and never touches what p names on disk.
func displayPath(p string) string {
	wd, err := os.Getwd()
	if err != nil {
		return p
	}
	rel, err := filepath.Rel(wd, p)
	if err != nil {
		return p
	}
	if !filepath.IsLocal(rel) {
		return p
	}
	return rel
}

// hasChange reports whether id names one of changes.
func hasChange(changes []model.Change, id string) bool {
	for _, c := range changes {
		if c.ID == id {
			return true
		}
	}
	return false
}

// resolvePeers resolves every name declared in peers, so callers pass
// already-resolved peers into check's pure functions instead of those
// functions resolving peers themselves.
func resolvePeers(peers map[string]string) map[string]tree.ResolvedPeer {
	resolved := make(map[string]tree.ResolvedPeer, len(peers))
	for name := range peers {
		resolved[name] = tree.ResolvePeer(peers, name)
	}
	return resolved
}
