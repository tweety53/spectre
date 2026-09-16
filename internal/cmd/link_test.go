package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// commitAll commits everything currently in the git repo at dir, so a
// fixture state created after gitInit reads as clean to
// hasUncommittedModifications.
func commitAll(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"add", "-A"},
		{"-c", "user.email=a@b.c", "-c", "user.name=t", "commit", "-q", "-m", "fixture"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
}

// gitInit initializes a git repo at dir and commits everything currently
// there, so later modifications show as uncommitted in `git status`.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "-A"},
		{"-c", "user.email=a@b.c", "-c", "user.name=t", "commit", "-q", "-m", "seed"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
}

// gitCheckoutBranch checks out a new branch name in the git repo at dir.
func gitCheckoutBranch(t *testing.T, dir, name string) {
	t.Helper()
	c := exec.Command("git", "checkout", "-q", "-b", name)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b %s: %v (%s)", name, err, out)
	}
}

// linkFixture builds two sibling repositories, "sat" (the satellite tree,
// where spectre link is run) and "can" (the canonical tree, holding the
// real plan), declaring each other under the names "can" and "sat". can
// carries a full change scaffold named canonicalID unless canFiles is
// given, in which case those files replace the scaffold — used to build
// the "canonical is itself a satellite" fixture. Both repositories are
// separate real git repos with everything committed; sat is checked out
// on satBranch.
func linkFixture(t *testing.T, canonicalID, satBranch string, canFiles map[string]string) (satBase, canBase string) {
	t.Helper()
	parent := t.TempDir()
	satBase = filepath.Join(parent, "sat")
	canBase = filepath.Join(parent, "can")
	for _, base := range []string{satBase, canBase} {
		for _, sub := range []string{"specs", "changes"} {
			if err := os.MkdirAll(filepath.Join(base, "spectre", sub), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.WriteFile(filepath.Join(satBase, "spectre", "peers"), []byte("can ../can\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canBase, "spectre", "peers"), []byte("sat ../sat\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if canFiles == nil {
		canFiles = map[string]string{
			"proposal.md": "# " + canonicalID + "\n\n## Why\nBecause.\n\n## What changes\n- a thing\n",
			"tasks.md":    "# " + canonicalID + "\n\n- [ ] 1. One\n",
		}
	}
	dir := filepath.Join(canBase, "spectre", "changes", canonicalID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range canFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	gitInit(t, satBase)
	gitCheckoutBranch(t, satBase, satBranch)
	gitInit(t, canBase)
	return satBase, canBase
}

// TestLinkCounterpartOnlyInPeerWorktree pins the link command's own worktree
// fallback: the canonical change exists only under can's
// .worktrees/kan-9-cross/spectre — never in the primary checkout's changes/ —
// because before either side of a cross-repo link has landed on a primary
// checkout, that worktree is the only place the counterpart exists (KAN-518:
// the invoking-worktree half bc9f03b's resolveCounterpart and NamesFor fixes
// left in the link command itself). Without the fallback Link refuses with
// "does not exist" and the run escapes through link --force plus a
// hand-written link.md.
func TestLinkCounterpartOnlyInPeerWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
	wtChanges := filepath.Join(canBase, ".worktrees", "kan-9-cross", "spectre", "changes", "kan-9-cross")
	if err := os.MkdirAll(filepath.Dir(wtChanges), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(canBase, "spectre", "changes", "kan-9-cross"), wtChanges); err != nil {
		t.Fatal(err)
	}
	commitAll(t, canBase)

	var out, errBuf bytes.Buffer
	if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	canLink, err := os.ReadFile(filepath.Join(wtChanges, "link.md"))
	if err != nil {
		t.Fatalf("canonical link.md not written into the peer worktree: %v", err)
	}
	if !strings.Contains(string(canLink), "## Parts") || !strings.Contains(string(canLink), "`sat:kan-9-cross`") {
		t.Errorf("canonical link.md = %s, want Parts sat:kan-9-cross", canLink)
	}
	// The primary checkout's changes/ stays empty: the fallback must write
	// the counterpart where it lives, not create the change on the primary
	// checkout.
	if _, statErr := os.Stat(filepath.Join(canBase, "spectre", "changes", "kan-9-cross")); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("primary changes/kan-9-cross exists (%v), want the counterpart confined to the worktree", statErr)
	}
}

func TestLink(t *testing.T) {
	satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)

	var out, errBuf bytes.Buffer
	if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	satLink, err := os.ReadFile(filepath.Join(satBase, "spectre", "changes", "kan-9-cross", "link.md"))
	if err != nil {
		t.Fatalf("satellite link.md not written: %v", err)
	}
	if !strings.Contains(string(satLink), "## Part of") || !strings.Contains(string(satLink), "`can:kan-9-cross`") {
		t.Errorf("satellite link.md = %s, want Part of can:kan-9-cross", satLink)
	}
	if !strings.Contains(string(satLink), "## Branch") || !strings.Contains(string(satLink), "feature-cross") {
		t.Errorf("satellite link.md = %s, want Branch feature-cross", satLink)
	}

	canLink, err := os.ReadFile(filepath.Join(canBase, "spectre", "changes", "kan-9-cross", "link.md"))
	if err != nil {
		t.Fatalf("canonical link.md not written: %v", err)
	}
	if !strings.Contains(string(canLink), "## Parts") || !strings.Contains(string(canLink), "`sat:kan-9-cross`") {
		t.Errorf("canonical link.md = %s, want Parts sat:kan-9-cross", canLink)
	}
	if !strings.Contains(string(canLink), "## Merge order") ||
		!strings.Contains(string(canLink), "`.`") || !strings.Contains(string(canLink), "`sat`") {
		t.Errorf("canonical link.md = %s, want Merge order . then sat", canLink)
	}

	// A clean two-tree link must round-trip through task 2's checks with
	// no findings in either tree.
	for _, root := range []string{satBase, canBase} {
		var vOut, vErr bytes.Buffer
		if code := Validate([]string{"--root", root}, &vOut, &vErr); code != OK {
			t.Errorf("validate --root %s: exit=%d stdout=%s stderr=%s", root, code, vOut.String(), vErr.String())
		}
	}
}

func TestLinkRefuses(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T) (satBase string, arg string)
		wantSub string
	}{
		{
			name: "peer not declared",
			setup: func(t *testing.T) (string, string) {
				satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				return satBase, "nope:kan-9-cross"
			},
			wantSub: "not declared",
		},
		{
			name: "peer declared but not present",
			setup: func(t *testing.T) (string, string) {
				satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				peersPath := filepath.Join(satBase, "spectre", "peers")
				if err := os.WriteFile(peersPath, []byte("can ../can\nghost ../ghost\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				return satBase, "ghost:kan-9-cross"
			},
			wantSub: "not present",
		},
		{
			name: "canonical change does not exist",
			setup: func(t *testing.T) (string, string) {
				satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				return satBase, "can:kan-missing"
			},
			wantSub: "does not exist",
		},
		{
			name: "canonical is itself a satellite",
			setup: func(t *testing.T) (string, string) {
				satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", map[string]string{
					"link.md": "## Part of\n\n`other:some-id`\n\n## Branch\n\nfeature-cross\n",
				})
				return satBase, "can:kan-9-cross"
			},
			wantSub: "one level deep",
		},
		{
			name: "link already exists",
			setup: func(t *testing.T) (string, string) {
				satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				var out, errBuf bytes.Buffer
				if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out, &errBuf); code != OK {
					t.Fatalf("setup link failed: exit=%d stderr=%s", code, errBuf.String())
				}
				return satBase, "can:kan-9-cross"
			},
			wantSub: "already exists",
		},
		{
			// guarded-cross-tree-write's reasoning applies to the local
			// side too: if the satellite tree already holds an unrelated
			// change of the same id — its own proposal.md/tasks.md, never
			// written by spectre link — writing link.md alongside them
			// would silently mix an unrelated change with the peer's
			// naming rather than refuse.
			name: "local change directory already exists with unrelated content",
			setup: func(t *testing.T) (string, string) {
				satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				localDir := filepath.Join(satBase, "spectre", "changes", "kan-9-cross")
				if err := os.MkdirAll(localDir, 0o755); err != nil {
					t.Fatal(err)
				}
				body := "# kan-9-cross\n\n## Why\nUnrelated.\n\n## What changes\n- something else\n"
				if err := os.WriteFile(filepath.Join(localDir, "proposal.md"), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
				return satBase, "can:kan-9-cross"
			},
			wantSub: "already exists",
		},
		{
			name: "peer change directory has uncommitted modifications",
			setup: func(t *testing.T) (string, string) {
				satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				dirty := filepath.Join(canBase, "spectre", "changes", "kan-9-cross", "tasks.md")
				body := "# kan-9-cross\n\n- [ ] 1. One\n- [ ] 2. Two\n"
				if err := os.WriteFile(dirty, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
				return satBase, "can:kan-9-cross"
			},
			wantSub: "uncommitted modifications",
		},
		{
			// Pins link.go's sixth, self-imposed refusal: writing a valid
			// "## Parts" entry into the peer's link.md requires a name the
			// peer itself declares for this tree (tree.NamesFor against
			// its own peers file). Erasing the peer's peers file removes
			// that name even though the local side still declares the
			// peer correctly.
			name: "peer has no declared name for this tree",
			setup: func(t *testing.T) (string, string) {
				satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
				peersPath := filepath.Join(canBase, "spectre", "peers")
				if err := os.WriteFile(peersPath, []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
				return satBase, "can:kan-9-cross"
			},
			wantSub: "no declared name",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			satBase, arg := tc.setup(t)
			var out, errBuf bytes.Buffer
			if code := Link([]string{"--root", satBase, arg}, &out, &errBuf); code != Fail {
				t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
			}
			if !strings.Contains(errBuf.String(), tc.wantSub) {
				t.Errorf("stderr = %q, want it to contain %q", errBuf.String(), tc.wantSub)
			}
		})
	}
}

// TestLinkForce pins guarded-cross-tree-write's scope: --force overrides
// the uncommitted-modifications refusal alone.
func TestLinkForce(t *testing.T) {
	t.Run("overrides uncommitted modifications", func(t *testing.T) {
		satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
		dirty := filepath.Join(canBase, "spectre", "changes", "kan-9-cross", "tasks.md")
		body := "# kan-9-cross\n\n- [ ] 1. One\n- [ ] 2. Two\n"
		if err := os.WriteFile(dirty, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "--force", "can:kan-9-cross"}, &out, &errBuf); code != OK {
			t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
		}
	})

	t.Run("does not override peer not declared", func(t *testing.T) {
		satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "--force", "nope:kan-9-cross"}, &out, &errBuf); code != Fail {
			t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
		}
	})

	t.Run("does not override canonical does not exist", func(t *testing.T) {
		satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "--force", "can:kan-missing"}, &out, &errBuf); code != Fail {
			t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
		}
	})

	t.Run("does not override canonical is itself a satellite", func(t *testing.T) {
		satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", map[string]string{
			"link.md": "## Part of\n\n`other:some-id`\n\n## Branch\n\nfeature-cross\n",
		})
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "--force", "can:kan-9-cross"}, &out, &errBuf); code != Fail {
			t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
		}
	})

	t.Run("does not override link already exists", func(t *testing.T) {
		satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out, &errBuf); code != OK {
			t.Fatalf("setup link failed: exit=%d stderr=%s", code, errBuf.String())
		}
		out.Reset()
		errBuf.Reset()
		if code := Link([]string{"--root", satBase, "--force", "can:kan-9-cross"}, &out, &errBuf); code != Fail {
			t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
		}
	})

	t.Run("does not override peer has no declared name", func(t *testing.T) {
		satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
		peersPath := filepath.Join(canBase, "spectre", "peers")
		if err := os.WriteFile(peersPath, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "--force", "can:kan-9-cross"}, &out, &errBuf); code != Fail {
			t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
		}
	})

	// guarded-cross-tree-write: --force overrides the uncommitted-
	// modifications refusal alone. This pins that it does not also widen
	// the local-directory-already-exists refusal added for F2/F7 — the
	// same scope guarantee TestLinkRefuses's other subtests pin for the
	// refusals that predate it.
	t.Run("does not override local change directory already exists", func(t *testing.T) {
		satBase, _ := linkFixture(t, "kan-9-cross", "feature-cross", nil)
		localDir := filepath.Join(satBase, "spectre", "changes", "kan-9-cross")
		if err := os.MkdirAll(localDir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "# kan-9-cross\n\n## Why\nUnrelated.\n\n## What changes\n- something else\n"
		if err := os.WriteFile(filepath.Join(localDir, "proposal.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		var out, errBuf bytes.Buffer
		if code := Link([]string{"--root", satBase, "--force", "can:kan-9-cross"}, &out, &errBuf); code != Fail {
			t.Fatalf("exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf.String())
		}
	})
}

// TestLinkWriteOrder pins the peer-side-first ordering: when the local
// write fails, the peer side must already have landed. The local write is
// blocked by making the satellite's already-created change directory
// read-only, so Link's earlier "does this link already exist" check
// (which only stats link.md, not the directory) still passes and the
// failure happens at the actual write.
func TestLinkWriteOrder(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
	localDir := filepath.Join(satBase, "spectre", "changes", "kan-9-cross")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(localDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(localDir, 0o755) })

	var out, errBuf bytes.Buffer
	if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out, &errBuf); code == OK {
		t.Fatalf("exit = OK, want a failure since the local write was blocked")
	}

	canLink := filepath.Join(canBase, "spectre", "changes", "kan-9-cross", "link.md")
	if _, err := os.Stat(canLink); err != nil {
		t.Fatalf("peer side must be written before the local write is attempted: %v", err)
	}
}

// TestLinkRetryAfterPartialWriteRefusesDuplicatePart reproduces a
// partial-write-then-retry: the peer side lands, the local write is
// blocked (as in TestLinkWriteOrder), and the operator retries the exact
// same command once the local side is unblocked. The retry must be
// refused by the canonical-side "already exists" guard (the loop over
// canLink.Parts in link.go, distinct from the local os.Stat(localLinkPath)
// check above it) rather than appending a second "sat:kan-9-cross" entry
// to the peer's "## Parts".
func TestLinkRetryAfterPartialWriteRefusesDuplicatePart(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
	localDir := filepath.Join(satBase, "spectre", "changes", "kan-9-cross")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(localDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(localDir, 0o755) })

	var out1, errBuf1 bytes.Buffer
	if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out1, &errBuf1); code == OK {
		t.Fatalf("first attempt exit = OK, want a failure since the local write was blocked")
	}
	canLinkPath := filepath.Join(canBase, "spectre", "changes", "kan-9-cross", "link.md")
	before, err := os.ReadFile(canLinkPath)
	if err != nil {
		t.Fatalf("peer side must have landed after the first attempt: %v", err)
	}

	if err := os.Chmod(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var out2, errBuf2 bytes.Buffer
	if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out2, &errBuf2); code != Fail {
		t.Fatalf("retry exit = %d, want %d (Fail); stderr=%s", code, Fail, errBuf2.String())
	}
	if !strings.Contains(errBuf2.String(), "already exists") {
		t.Errorf("stderr = %q, want the retry refused as already existing", errBuf2.String())
	}

	after, err := os.ReadFile(canLinkPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("canonical link.md changed on the refused retry:\nbefore=%s\nafter=%s", before, after)
	}
	if n := strings.Count(string(after), "`sat:kan-9-cross`"); n != 1 {
		t.Errorf("canonical link.md has %d \"sat:kan-9-cross\" entries, want exactly 1:\n%s", n, after)
	}
}
