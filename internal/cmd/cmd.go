// Package cmd implements one function per spectre subcommand. Each returns
// the process exit code and writes nothing to the real stdout or stderr.
package cmd

import (
	"flag"
	"io"
	"os"

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

// hasChange reports whether id names one of changes.
func hasChange(changes []model.Change, id string) bool {
	for _, c := range changes {
		if c.ID == id {
			return true
		}
	}
	return false
}
