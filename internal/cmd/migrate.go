package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/openspec"
	"github.com/tweety53/spectre/internal/render"
)

// Migrate converts an OpenSpec tree into a new spectre tree. It never
// modifies the source.
func Migrate(args []string, stdout, stderr io.Writer) int {
	fs, _ := flagSet("migrate", stderr)
	out := fs.String("out", "", "target tree (default: ./spectre)")
	force := fs.Bool("force", false, "write into an existing target")
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: spectre migrate [--out <dir>] [--force] <openspec-dir>")
		return Usage
	}
	src := fs.Arg(0)
	dst := *out
	if dst == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		dst = filepath.Join(wd, "spectre")
	}
	if _, err := os.Stat(dst); err == nil && !*force {
		fmt.Fprintf(stderr, "%s already exists (use --force to write into it)\n", dst)
		return Fail
	}

	var warnings []string

	specs, err := openspec.ReadSpecs(filepath.Join(src, "specs"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	if err := os.MkdirAll(filepath.Join(dst, "specs"), 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	for _, s := range specs {
		converted := model.Spec{Capability: s.Capability, Purpose: s.Purpose}
		for i, r := range s.Reqs {
			if !strings.Contains(r.Title, " SHALL ") {
				warnings = append(warnings, fmt.Sprintf("specs/%s.md: R%d (%q) has no SHALL clause — edit before validating", s.Capability, i+1, r.Title))
			}
			converted.Reqs = append(converted.Reqs, model.Requirement{
				ID:    fmt.Sprintf("R%d", i+1),
				Num:   i + 1,
				Text:  r.Title,
				Notes: r.Body,
			})
		}
		if err := os.WriteFile(filepath.Join(dst, "specs", s.Capability+".md"), render.Spec(converted), 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
	}

	changes, err := openspec.ReadChanges(filepath.Join(src, "changes"))
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	for _, c := range changes {
		target := filepath.Join(dst, "changes", c.ID)
		if err := os.MkdirAll(target, 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		if raw, err := os.ReadFile(filepath.Join(c.Dir, "proposal.md")); err == nil {
			if err := os.WriteFile(filepath.Join(target, "proposal.md"), openspec.ConvertProposal(raw), 0o644); err != nil {
				fmt.Fprintln(stderr, err)
				return Usage
			}
		}
		if raw, err := os.ReadFile(filepath.Join(c.Dir, "tasks.md")); err == nil {
			if err := os.WriteFile(filepath.Join(target, "tasks.md"), render.Tasks(openspec.ConvertTasks(raw)), 0o644); err != nil {
				fmt.Fprintln(stderr, err)
				return Usage
			}
		}
		if raw, err := os.ReadFile(filepath.Join(c.Dir, "design.md")); err == nil {
			if err := os.WriteFile(filepath.Join(target, "design.md"), raw, 0o644); err != nil {
				fmt.Fprintln(stderr, err)
				return Usage
			}
		}
		for _, delta := range c.Deltas {
			capName := filepath.Base(filepath.Dir(delta))
			name := "deltas-" + capName + ".md"
			raw, err := os.ReadFile(delta)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return Usage
			}
			if err := os.WriteFile(filepath.Join(target, name), raw, 0o644); err != nil {
				fmt.Fprintln(stderr, err)
				return Usage
			}
			warnings = append(warnings, fmt.Sprintf("changes/%s/%s: delta spec copied verbatim — apply it to specs/%s.md by hand", c.ID, name, capName))
		}
	}

	archived, err := copyDir(filepath.Join(src, "changes", "archive"), filepath.Join(dst, "changes", "archive"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	fmt.Fprintf(stdout, "migrated %d spec(s), %d open change(s), %d archived file(s) into %s\n", len(specs), len(changes), archived, dst)
	for _, w := range warnings {
		fmt.Fprintf(stdout, "warning: %s\n", w)
	}
	if len(warnings) > 0 {
		fmt.Fprintf(stdout, "%d warning(s) — the migration is incomplete\n", len(warnings))
		return Fail
	}
	return OK
}

// copyDir copies src to dst verbatim, returning the number of files
// written. A missing source is not an error.
func copyDir(src, dst string) (int, error) {
	n := 0
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return 0, nil
	}
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		n++
		return os.WriteFile(target, raw, 0o644)
	})
	return n, err
}
