// Package testtree builds fixture spectre trees on disk for tests in other
// packages, so the on-disk layout convention (spectre/{specs,changes},
// spec files named <capability>.md, an optional peers file) lives in one
// place instead of being hand-rolled per package.
package testtree

import (
	"os"
	"path/filepath"
	"testing"
)

// Build creates <parent>/<name>/spectre/{specs,changes}, writes one file
// per entry in specs (named <key>.md under specs/), and — when peers is
// non-empty — writes it as the tree's peers file. It returns the tree's
// base directory, <parent>/<name>.
func Build(t *testing.T, parent, name string, specs map[string]string, peers string) string {
	t.Helper()
	base := filepath.Join(parent, name)
	if err := os.MkdirAll(filepath.Join(base, "spectre", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, "spectre", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	for capName, body := range specs {
		p := filepath.Join(base, "spectre", "specs", capName+".md")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if peers != "" {
		p := filepath.Join(base, "spectre", "peers")
		if err := os.WriteFile(p, []byte(peers), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return base
}
