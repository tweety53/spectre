package cmd

import (
	"fmt"
	"io"

	"github.com/tweety53/spectre/internal/check"
)

// Validate checks the whole tree, or one change when an id is given.
func Validate(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("validate", stderr)
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "usage: spectre validate [change-id]")
		return Usage
	}
	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	changeID := fs.Arg(0)
	if changeID != "" {
		changes, err := t.Changes()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		if !hasChange(changes, changeID) {
			fmt.Fprintf(stderr, "no such change %q in %s\n", changeID, t.ChangesDir())
			return Usage
		}
	}

	findings, err := check.Structural(t, changeID)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	if len(findings) == 0 {
		fmt.Fprintln(stdout, "no findings")
		return OK
	}
	for _, f := range findings {
		fmt.Fprintln(stdout, f)
	}
	fmt.Fprintf(stdout, "%d finding(s)\n", len(findings))
	return Fail
}
