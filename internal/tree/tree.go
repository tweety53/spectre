// Package tree locates and reads a spectre tree. The tree is the only
// state: nothing here caches, indexes or writes sidecar files.
package tree

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoRoot reports that no spectre/ directory was found.
var ErrNoRoot = errors.New("no spectre/ directory found")

// ProposalFile and TasksFile are the fixed filenames every change carries,
// the one source of truth every package that creates, reads, or validates
// a change agrees on.
const (
	ProposalFile = "proposal.md"
	TasksFile    = "tasks.md"
)

// Tree is a resolved spectre tree.
type Tree struct {
	Root string // absolute path of the spectre/ directory
}

// Find walks up from start looking for a spectre/ directory.
func Find(start string) (*Tree, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	for {
		cand := filepath.Join(dir, "spectre")
		if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
			return &Tree{Root: cand}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("%w (searched upwards from %s)", ErrNoRoot, start)
		}
		dir = parent
	}
}

// Open uses an explicit root, which may be the tree directory itself or
// its parent.
func Open(root string) (*Tree, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if filepath.Base(abs) != "spectre" {
		abs = filepath.Join(abs, "spectre")
	}
	fi, err := os.Stat(abs)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("%w at %s", ErrNoRoot, abs)
	}
	return &Tree{Root: abs}, nil
}

// SpecsDir is the tree's capability-specs subdirectory.
func (t *Tree) SpecsDir() string { return filepath.Join(t.Root, "specs") }

// ChangesDir is the tree's open-changes subdirectory.
func (t *Tree) ChangesDir() string { return filepath.Join(t.Root, "changes") }

// ArchiveDir is the tree's archived-changes subdirectory.
func (t *Tree) ArchiveDir() string { return filepath.Join(t.Root, "changes", "archive") }

// Peers reads the peers file: "<name> <relative-path>" lines, blank lines
// and # comments ignored. An absent file is not an error.
func (t *Tree) Peers() (map[string]string, error) {
	f, err := os.Open(filepath.Join(t.Root, "peers"))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	seen := map[string]int{} // track line numbers for duplicate detection
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) != 2 {
			return nil, fmt.Errorf("peers:%d: want \"<name> <path>\", got %q", line, text)
		}
		name := fields[0]
		if prevLine, exists := seen[name]; exists {
			return nil, fmt.Errorf("peers:%d: duplicate name %q (first seen on line %d)",
				line, name, prevLine)
		}
		seen[name] = line

		p := fields[1]
		if !filepath.IsAbs(p) {
			p = filepath.Join(filepath.Dir(t.Root), p)
		}
		if filepath.Base(p) != "spectre" {
			p = filepath.Join(p, "spectre")
		}
		out[name] = filepath.Clean(p)
	}
	return out, sc.Err()
}
