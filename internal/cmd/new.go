package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tweety53/spectre/internal/render"
)

const proposalTemplate = `# %s

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

	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
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
		"proposal.md": []byte(fmt.Sprintf(proposalTemplate, id)),
		"tasks.md":    render.Tasks(nil),
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
	}
	fmt.Fprintf(stdout, "created %s\n", dir)
	return OK
}
