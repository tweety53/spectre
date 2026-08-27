package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tweety53/spectre/internal/tree"
)

const _configTemplate = `# spectre config

Every key below is a default, shown inside a fence so spectre ignores it.
Move a line out of the fence to make it take effect.

## Rules

` + "```" + `
- headings: error
- placeholders: error
- shall-clause: error
- malformed-bullet: error
- id-sequence: error
- task-sequence: error
- refs: error
` + "```" + `

## Vocabulary

` + "```" + `
- modal: SHALL
- id-prefix: R
` + "```" + `

## Layout

` + "```" + `
- specs: specs
- changes: changes
- extension: .md
` + "```" + `
`

// Init creates the specs/, changes/ and config.md that make a directory a
// spectre tree. It never overwrites what is already there, so it is safe
// to run against an existing tree to fill in what is missing.
func Init(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("init", stderr)
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: spectre init [--root <path>]")
		return Usage
	}

	rootPath, err := tree.RootPath(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	created := false

	dirs := []string{filepath.Join(rootPath, "specs"), filepath.Join(rootPath, "changes")}
	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err == nil {
			if !fi.IsDir() {
				// A file sitting where specs/ or changes/ belongs blocks the
				// tree the same way an already-occupied change id blocks
				// `new` (new.go's own "already exists" check): a content
				// refusal, not an I/O failure, so it gets Fail rather than
				// Usage, and init must not paper over it by scaffolding the
				// rest of the tree around it.
				fmt.Fprintf(stderr, "%s already exists and is not a directory\n", dir)
				return Fail
			}
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		fmt.Fprintf(stdout, "created %s\n", dir)
		created = true
	}

	configPath := filepath.Join(rootPath, "config.md")
	if _, err := os.Stat(configPath); err != nil {
		if err := os.WriteFile(configPath, []byte(_configTemplate), 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		fmt.Fprintf(stdout, "created %s\n", configPath)
		created = true
	}

	if !created {
		fmt.Fprintln(stdout, "already complete")
	}
	return OK
}
