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

	"github.com/tweety53/spectre/internal/config"
	"github.com/tweety53/spectre/internal/parse"
)

// ErrNoRoot reports that no spectre/ directory was found.
var ErrNoRoot = errors.New("no spectre/ directory found")

// ProposalFile, TasksFile and DesignFile are the fixed filenames a change
// uses when it carries them, the one source of truth every package that
// creates, reads, or validates a change agrees on. ProposalFile and
// TasksFile are required on every change; DesignFile is optional.
const (
	ProposalFile = "proposal.md"
	TasksFile    = "tasks.md"
	DesignFile   = "design.md"
)

// Tree is a resolved spectre tree, carrying its own configuration. The
// zero value is not valid: its SpecsDir and ChangesDir would resolve to
// Root itself and its parser would be nil. Build one with At, Find or
// Open instead.
type Tree struct {
	Root string // absolute path of the spectre/ directory
	Cfg  config.Config
	p    *parse.Parser
}

// At builds a Tree for root under cfg, without reading config.md itself
// — for a caller that has already resolved cfg on its own. Find and
// Open resolve cfg themselves via config.Load and are the usual way to
// open an existing tree.
func At(root string, cfg config.Config) *Tree {
	return &Tree{Root: root, Cfg: cfg, p: parse.New(cfg)}
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
			return load(cand)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf(
				"%w (searched upwards from %s); run \"spectre init\" to create one",
				ErrNoRoot, start,
			)
		}
		dir = parent
	}
}

// RootPath resolves root to the path of a spectre tree directory: a
// path whose basename is not literally "spectre" gains it. The
// directory need not exist, which is what lets Init create one; Open
// stats the result, Init creates it.
func RootPath(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if filepath.Base(abs) != "spectre" {
		abs = filepath.Join(abs, "spectre")
	}
	return abs, nil
}

// Open uses an explicit root, which may be the tree directory itself or
// its parent.
func Open(root string) (*Tree, error) {
	abs, err := RootPath(root)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(abs)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("%w at %s; run \"spectre init\" to create one", ErrNoRoot, abs)
	}
	return load(abs)
}

// load reads the tree's configuration and builds its parser.
func load(root string) (*Tree, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	return At(root, cfg), nil
}

// SpecsDir is the tree's capability-specs subdirectory.
func (t *Tree) SpecsDir() string { return filepath.Join(t.Root, t.Cfg.SpecsDir) }

// ChangesDir is the tree's open-changes subdirectory.
func (t *Tree) ChangesDir() string { return filepath.Join(t.Root, t.Cfg.ChangesDir) }

// ArchiveDir is the tree's archived-changes subdirectory.
func (t *Tree) ArchiveDir() string { return filepath.Join(t.ChangesDir(), "archive") }

// isTree reports whether p is already a spectre tree directory rather than
// the repository that contains one, by probing for changes/ inside it. A
// peers entry names the repository, so Peers appends the tree leaf only
// when this is false — but a repository whose own directory happens to be
// named "spectre" (as spectre's own is) makes a basename comparison see the
// leaf name where there is none, and resolve one level short. changes/,
// not the bare directory itself, is what spectre init actually creates and
// what every caller goes on to read, so a hit here is evidence a caller can
// use — the same convention <agents repo>/scripts/lib/spec-root.sh applies
// to the same question.
func isTree(p string) bool {
	fi, err := os.Stat(filepath.Join(p, "changes"))
	return err == nil && fi.IsDir()
}

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
		if !isTree(p) {
			p = filepath.Join(p, "spectre")
		}
		out[name] = filepath.Clean(p)
	}
	return out, sc.Err()
}
