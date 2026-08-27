package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// chdir switches the process working directory to dir for the duration of
// the test and restores the original on cleanup.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatal(err)
		}
	})
}

// TestDisplayPath pins displayPath's contract: relativise against the
// working directory when the result reads better than the absolute form,
// and never touch the filesystem to do it.
func TestDisplayPath(t *testing.T) {
	t.Run("under the working directory renders relative", func(t *testing.T) {
		base := t.TempDir()
		chdir(t, base)
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(wd, "spectre", "specs")

		if got, want := displayPath(p), filepath.Join("spectre", "specs"); got != want {
			t.Errorf("displayPath(%q) = %q, want %q", p, got, want)
		}
	})

	t.Run("outside the working directory stays absolute", func(t *testing.T) {
		base := t.TempDir()
		chdir(t, base)
		other := t.TempDir()
		p := filepath.Join(other, "spectre", "specs")

		if got := displayPath(p); got != p {
			t.Errorf("displayPath(%q) = %q, want the absolute path unchanged", p, got)
		}
	})

	t.Run("the working directory itself renders as dot", func(t *testing.T) {
		base := t.TempDir()
		chdir(t, base)
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := displayPath(wd), "."; got != want {
			t.Errorf("displayPath(%q) = %q, want %q", wd, got, want)
		}
	})

	// A path reached through a symlinked working directory must not be
	// silently rewritten into something that resolves elsewhere:
	// displayPath must relativise lexically, not resolve symlinks. After
	// os.Chdir into a symlink, os.Getwd returns the physically resolved
	// directory (no symlink components), which lexically shares no
	// prefix with a path built through the symlink — so filepath.Rel
	// cannot relate them and displayPath must fall back to the absolute
	// path exactly as given, not a resolved variant of it.
	t.Run("a path reached through a symlink is not resolved", func(t *testing.T) {
		base := t.TempDir()
		real := filepath.Join(base, "real")
		if err := os.MkdirAll(real, 0o755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(base, "link")
		if err := os.Symlink(real, link); err != nil {
			t.Skipf("symlinks unsupported: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(real, "child"), 0o755); err != nil {
			t.Fatal(err)
		}
		chdir(t, link)
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if wd == link {
			t.Skip("os.Getwd did not resolve the symlink on this platform")
		}

		p := filepath.Join(link, "child")
		if got := displayPath(p); got != p {
			t.Errorf("displayPath(%q) = %q, want the path unchanged, not resolved", p, got)
		}
	})
}

// TestExitCodeContract pins the numeric exit codes every command relies on.
// Every assertion in this package compares an exit code against these
// constants by name, so nothing catches OK/Fail/Usage's numeric values
// drifting — changing Usage from 2 to 3 leaves the whole suite green. This
// test is the one place that pins the actual numbers.
func TestExitCodeContract(t *testing.T) {
	if OK != 0 {
		t.Errorf("OK = %d, want 0", OK)
	}
	if Fail != 1 {
		t.Errorf("Fail = %d, want 1", Fail)
	}
	if Usage != 2 {
		t.Errorf("Usage = %d, want 2", Usage)
	}
}
