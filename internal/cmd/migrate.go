package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tweety53/spectre/internal/check"
	"github.com/tweety53/spectre/internal/config"
	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/openspec"
	"github.com/tweety53/spectre/internal/render"
	"github.com/tweety53/spectre/internal/tree"
)

// Migrate converts an OpenSpec tree into a new spectre tree. It never
// modifies the source.
func Migrate(args []string, stdout, stderr io.Writer) int {
	// migrate does not use flagSet: it takes a source path and --out, not
	// a spectre tree to resolve via --root — --root registered and
	// discarded would contradict the top-level usage text's promise that
	// every command accepts it.
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(stderr)
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

	// Resolve dst exactly as tree.Open resolves a --root: when the given
	// path's basename isn't literally "spectre", append "spectre" to it.
	// Every other command's root resolution already does this, so writing
	// straight into a differently named --out would produce a tree no
	// other command could then open by passing that same --out value as
	// --root — an internal inconsistency between this command's write path
	// and every other command's read path, not something the caller should
	// have to work around by always naming the target "spectre" themselves.
	absDst, err := filepath.Abs(dst)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	if filepath.Base(absDst) != "spectre" {
		absDst = filepath.Join(absDst, "spectre")
	}
	dst = absDst

	// Stat the source before touching the destination at all: checking dst
	// first meant a mistyped src with an existing dst reported "<dst>
	// already exists (use --force to write into it)", pointing the user at
	// a flag that would clear their tree instead of naming their actual
	// typo.
	if _, err := os.Stat(src); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	// A tree's own config.md governs how that tree is read everywhere
	// else in spectre, so it governs how migrate writes into it too: an
	// existing config.md — which --force deliberately preserves — decides
	// the layout, extension and vocabulary migrate writes, not a
	// hardcoded default. A target with no config.md yet gets
	// config.Default(), which config.Load already returns for an absent
	// file.
	cfg, err := config.Load(dst)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	dstTree := tree.At(dst, cfg)

	if _, err := os.Stat(dst); err == nil {
		if !*force {
			fmt.Fprintf(stderr, "%s already exists (use --force to write into it)\n", dst)
			return Fail
		}
		cleared, err := clearForce(dstTree)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		if len(cleared) > 0 {
			fmt.Fprintf(stdout, "--force: cleared %s in %s\n", strings.Join(cleared, " and "), dst)
		}
	}

	var warnings []string

	specs, err := openspec.ReadSpecs(filepath.Join(src, "specs"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	if err := os.MkdirAll(dstTree.SpecsDir(), 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	for _, s := range specs {
		converted := model.Spec{Capability: s.Capability, Purpose: s.Purpose}
		for i, r := range s.Reqs {
			converted.Reqs = append(converted.Reqs, model.Requirement{
				ID:    fmt.Sprintf("%s%d", cfg.IDPrefix, i+1),
				Num:   i + 1,
				Text:  r.Title,
				Notes: r.Body,
			})
		}
		specPath := filepath.Join(dstTree.SpecsDir(), s.Capability+cfg.Extension)
		if err := os.WriteFile(specPath, render.Spec(converted), 0o644); err != nil {
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
		target := filepath.Join(dstTree.ChangesDir(), c.ID)
		if err := os.MkdirAll(target, 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		if raw, err := os.ReadFile(filepath.Join(c.Dir, "proposal.md")); err == nil {
			if err := os.WriteFile(filepath.Join(target, tree.ProposalFile), openspec.ConvertProposal(raw), 0o644); err != nil {
				fmt.Fprintln(stderr, err)
				return Usage
			}
		}
		if raw, err := os.ReadFile(filepath.Join(c.Dir, "tasks.md")); err == nil {
			if err := os.WriteFile(filepath.Join(target, tree.TasksFile), render.Tasks(openspec.ConvertTasks(raw)), 0o644); err != nil {
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

	archived, err := copyDir(filepath.Join(src, "changes", "archive"), dstTree.ArchiveDir())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	// The promise of a clean exit is that the written tree validates, so
	// migrate runs exactly the checks "spectre validate" runs over a whole
	// tree and reports every finding as a warning, rather than reimplementing
	// any of check's rules itself.
	declaredPeers, err := dstTree.Peers()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	findings, err := check.Structural(dstTree, "", resolvePeers(declaredPeers))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	// OpenSpec states a requirement's normative sentence in the body;
	// spectre's shall-clause rule requires it on the bullet itself, so a
	// straightforward conversion trips this one rule on nearly every
	// requirement — 249 of 278 warnings on a real corpus. The rule stays
	// strict rather than accepting the body: a requirement's normative
	// sentence belongs on its own line. noShall counts them separately so
	// the report can name the class once instead of leaving the user to
	// notice the pattern across hundreds of near-identical lines.
	noShall := 0
	shallSuffix := fmt.Sprintf("has no %s clause", cfg.Modal)
	for _, f := range findings {
		warnings = append(warnings, f.String())
		if strings.Contains(f.Msg, shallSuffix) {
			noShall++
		}
	}

	fmt.Fprintf(stdout, "migrated %d spec(s), %d open change(s), %d archived file(s) into %s\n", len(specs), len(changes), archived, dst)
	for _, w := range warnings {
		fmt.Fprintf(stdout, "warning: %s\n", w)
	}
	if noShall > 1 {
		fmt.Fprintf(stdout, "%d of the warnings above are missing-%s findings: OpenSpec states the modal in the requirement body, spectre states it in the bullet, so each needs a hand pass\n",
			noShall, cfg.Modal)
	}
	if len(warnings) > 0 {
		fmt.Fprintf(stdout, "%d warning(s) — the migration is incomplete\n", len(warnings))
		return Fail
	}
	return OK
}

// clearForce removes dstTree's configured specs and changes subdirectories,
// if present, so a --force re-run cannot leave a stale file behind from a
// capability or change that no longer exists in the source. It touches only
// those two directories, never dstTree.Root itself, so any other file a
// user has already placed there (a config.md, a peers file) survives. It
// returns the names cleared, relative to dstTree.Root, for the report.
func clearForce(dstTree *tree.Tree) ([]string, error) {
	var cleared []string
	for _, dir := range []string{dstTree.SpecsDir(), dstTree.ChangesDir()} {
		if _, err := os.Stat(dir); err == nil {
			if err := os.RemoveAll(dir); err != nil {
				return nil, err
			}
			rel, err := filepath.Rel(dstTree.Root, dir)
			if err != nil {
				rel = dir
			}
			cleared = append(cleared, rel+"/")
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return cleared, nil
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
