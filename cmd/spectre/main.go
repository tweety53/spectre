// Command spectre manages a markdown tree of capability specs and changes.
package main

import (
	"fmt"
	"os"

	"github.com/tweety53/spectre/internal/cmd"
)

const _usage = `spectre — spec-driven change tracking in markdown

usage: spectre <command> [flags] [args]

commands:
  new <change-id>              scaffold changes/<id>/
  list [--specs] [--json]      list open changes, or capabilities
  validate [change-id]         check the tree, or one change
  refs <capability>#<id>       find citations of a requirement
  archive <change-id>          move a finished change into changes/archive/
  migrate <openspec-dir>       convert an OpenSpec tree into a new spectre tree

every command except migrate accepts --root <path> to name the tree
explicitly; migrate takes a source path and --out instead
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, _usage)
		os.Exit(cmd.Usage)
	}
	args := os.Args[2:]
	switch os.Args[1] {
	case "new":
		os.Exit(cmd.New(args, os.Stdout, os.Stderr))
	case "list":
		os.Exit(cmd.List(args, os.Stdout, os.Stderr))
	case "validate":
		os.Exit(cmd.Validate(args, os.Stdout, os.Stderr))
	case "refs":
		os.Exit(cmd.Refs(args, os.Stdout, os.Stderr))
	case "archive":
		os.Exit(cmd.Archive(args, os.Stdout, os.Stderr))
	case "migrate":
		os.Exit(cmd.Migrate(args, os.Stdout, os.Stderr))
	case "-h", "--help", "help":
		fmt.Fprint(os.Stdout, _usage)
		os.Exit(cmd.OK)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], _usage)
		os.Exit(cmd.Usage)
	}
}
