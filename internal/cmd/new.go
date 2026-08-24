package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tweety53/spectre/internal/render"
	"github.com/tweety53/spectre/internal/tree"
)

const _proposalTemplate = `# %s

## Why

<!-- the problem this change solves -->

## What changes

<!-- the observable difference once this lands -->
`

// New scaffolds changes/<id>/ with proposal and task templates.
func New(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("new", stderr)
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: spectre new <change-id>")
		return Usage
	}
	id := fs.Arg(0)

	if !isValidID(id) {
		fmt.Fprintf(stderr, "invalid change id: %q\n", id)
		return Usage
	}

	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	// The archive directory's own name is reserved: a change scaffolded
	// under it would sit inside changes/archive/ but be invisible to
	// list, validate and archive, which all treat that name specially.
	// Derived from the tree rather than hardcoded, so a configured
	// changes layout still rejects the right name.
	if archiveName := filepath.Base(t.ArchiveDir()); id == archiveName {
		fmt.Fprintf(stderr, "invalid change id: %q is reserved for the archive directory\n", id)
		return Usage
	}

	dir := filepath.Join(t.ChangesDir(), id)
	if _, err := os.Stat(dir); err == nil {
		fmt.Fprintf(stderr, "%s already exists\n", dir)
		return Fail
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	files := map[string][]byte{
		tree.ProposalFile: []byte(fmt.Sprintf(_proposalTemplate, id)),
		tree.TasksFile:    render.Tasks(nil),
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			if rmErr := os.RemoveAll(dir); rmErr != nil {
				fmt.Fprintf(stderr, "and cleanup of %s also failed: %v; remove it by hand\n", dir, rmErr)
			}
			return Usage
		}
	}
	fmt.Fprintf(stdout, "created %s\n", dir)
	return OK
}

// isValidID checks that id is a single flat directory name with no path traversal.
func isValidID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	// Check for path separators
	if strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return false
	}
	if filepath.Clean(id) != id {
		return false
	}
	if filepath.IsAbs(id) {
		return false
	}
	return true
}
