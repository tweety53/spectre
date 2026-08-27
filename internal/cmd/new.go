package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/tweety53/spectre/internal/render"
	"github.com/tweety53/spectre/internal/tree"
)

const _proposalTemplate = `# %s

## Why

<!-- the problem this change solves -->

## What changes

<!-- the observable difference once this lands -->
`

const _designTemplate = `## Context

<!-- what constrains this change and why it is one change -->

## Decisions

<!-- what was chosen, what was considered, and why -->
`

// New scaffolds changes/<id>/ with proposal, tasks and design templates.
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
		tree.TasksFile:    render.Tasks(id, nil),
		tree.DesignFile:   []byte(_designTemplate),
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
	fmt.Fprintf(stdout, "created %s\n", displayPath(dir))
	return OK
}

// isValidID checks that id is a single flat directory name with no path
// traversal and no control characters. The check rejects the whole
// Unicode control category (unicode.IsControl), not just \n, \r and \t:
// id becomes both a required heading (see ProposalFindings and
// TaskFindings) and a directory name, and any control character can
// break one of those two roles — a newline splits headingFindings'
// line-by-line scan and validate's file:line: message output, a NUL
// breaks the directory name outright, and stray control bytes in a
// heading serve no purpose a caller could intend. Rejecting the whole
// category is one rule instead of an allowlist of "the control
// characters we happened to think of".
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
	if strings.ContainsFunc(id, unicode.IsControl) {
		return false
	}
	return true
}
