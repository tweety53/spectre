package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tweety53/spectre/internal/tree"
)

// Archive moves a finished change into changes/archive/ with git mv. It
// does not commit.
func Archive(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("archive", stderr)
	force := fs.Bool("force", false, "archive even when tasks are unchecked or missing")
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: spectre archive <change-id>")
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
	changes, err := t.Changes()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	for _, c := range changes {
		if c.ID != id {
			continue
		}
		tasksPath := filepath.Join(c.Dir, tree.TasksFile)
		if _, err := os.Stat(tasksPath); errors.Is(err, os.ErrNotExist) && !*force {
			fmt.Fprintf(stderr, "%s: no tasks.md (use --force to archive anyway)\n", id)
			return Fail
		}
		// `new` always writes a tasks.md, just an empty one, so the
		// missing-file check above never catches "spectre new X &&
		// spectre archive X": zero tasks trivially reads as zero unchecked,
		// which the guard below would let through. Guard on having no
		// tasks at all, the same content refusal --force overrides.
		if len(c.Tasks) == 0 && !*force {
			fmt.Fprintf(stderr, "%s: tasks.md has no tasks (use --force to archive anyway)\n", id)
			return Fail
		}
		if left := len(c.Tasks) - c.DoneCount(); left > 0 && !*force {
			fmt.Fprintf(stderr, "%s: %d of %d tasks are unchecked (use --force to archive anyway)\n",
				id, left, len(c.Tasks))
			return Fail
		}
		dst := filepath.Join(t.ArchiveDir(), id)
		if _, err := os.Stat(dst); err == nil {
			fmt.Fprintf(stderr, "%s: destination %s already exists\n", id, dst)
			return Fail
		} else if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		if err := os.MkdirAll(t.ArchiveDir(), 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		mv := exec.Command("git", "mv", c.Dir, dst)
		mv.Dir = t.Root
		if out, err := mv.CombinedOutput(); err != nil {
			fmt.Fprintf(stderr, "git mv failed: %v\n%s", err, out)
			return Usage
		}
		fmt.Fprintf(stdout, "archived %s\n", id)
		return OK
	}

	fmt.Fprintf(stderr, "no open change %q in %s\n", id, t.ChangesDir())
	return Usage
}
