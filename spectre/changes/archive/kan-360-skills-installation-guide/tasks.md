# kan-360-skills-installation-guide

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Eight tasks. The first three were planned: the manifests exist before the README tells anyone to
install from them, and the test comes after both. Tasks 4 to 8 were added during the run, each
recording work a review-panel finding caused — they are written after the fact rather than left out,
so this plan describes what the branch did and not only what was intended. Every task leaves a green
build. The design this plan implements is
`docs/superpowers/specs/2026-08-28-kan-360-skills-installation-guide-design.md`.

**Verification, every task:**

```bash verified:run in this worktree at 6181228 on the unmodified tree
gofmt -l .        # must print nothing
go vet ./...      # must print nothing
go test ./...     # must be all ok
```

Test counts below are counts of test runs, subtests included, as produced by:

```bash verified:run in this worktree at 6181228
go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN'
```

The baseline was 197 at the branch point, when this plan expected to touch no Go file at all.
It is 205 at the branch head: task 3 added a test, and tasks 5 to 8 reshaped it.
<!-- measured: the command above @ 6181228, and again @ 2c94161 -->

**Global constraints, both tasks:**

- No new dependency, and no change to the `spectre` binary's own behaviour. **Task 3 adds a Go
  test** — the original "no Go code changes" constraint was overturned by the operator after the
  panel found four surviving mutants, per decision `manifest-validation-test`.
- **The two skills become one**, per decision `single-spectre-skill`:
  `.claude/skills/spectre-new/` is deleted and its scaffold-then-fill behaviour folded into
  `.claude/skills/spectre/SKILL.md`. This overturns the original "no skill file moves" constraint.
  `.claude/skills/spectre/` keeps its name, so the bare command stays `/spectre`; under a plugin
  install it is `/spectre:spectre`, because Claude Code namespaces every plugin skill by its plugin
  name and offers no bare form.
- `README.md`'s `## File templates` table is parsed by the subtest `README file templates table` in
  `internal/check/check_test.go`. Neither task may alter that table.
- README prose wraps at the width the file already uses. The widest existing prose line is 99
  bytes. <!-- measured: python3 -c "print(max(len(l.rstrip()) for l in open('README.md')))" @ 6181228 -->
- The manual install path names Claude Code's `~/.claude/skills/` only, and otherwise says "your
  agent's own skills directory" — no claim about any other product's path (`generic-non-claude-path`).

---

**Tasks 1 and 2 below record the plan as it was written, before the review panel ran.** Their step
bodies still show the two-skill text — two skill directories to copy, a `Skills (2)` inventory line,
manifest descriptions naming `/spectre` and `/spectre-new`. That is what was planned and what those
two commits first did. Task 4 records what superseded it, and decision `single-spectre-skill` in
`design.md` records why. The step bodies are left as written rather than rewritten, so the plan still
shows what was actually planned at the time; read task 4 for the end state.

- [x] 1. Add the Claude Code plugin and marketplace manifests

**Build:** green
**Files:** `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`
**Tests:** none — this task adds two static JSON manifests, and the repository has no harness that
could assert on Claude Code's plugin loader. Verification is Step 2's parse check and Step 3's real
install cycle, both run by hand.
**Regression:** reverting this commit removes both manifests, and `claude plugin marketplace add ./`
then fails with `Invalid marketplace source format` — there is no `.claude-plugin/marketplace.json`
for it to read.
**Baseline:** before=197 after=197
**Commit:** `feat(plugin): add Claude Code plugin and marketplace manifests`

The `skills` field points at the existing `.claude/skills/` rather than the plugin default
`skills/`, per decision `skills-stay-in-claude-skills`. That field *adds to* the default rather
than replacing it, so no directory has to exist at the default path.

  - [x] **Step 1: Write `.claude-plugin/plugin.json`.**

    ```json verified:schema at code.claude.com/docs/en/plugins-reference; this exact body installed and read back in Step 3
    {
      "$schema": "https://json.schemastore.org/claude-code-plugin-manifest.json",
      "name": "spectre",
      "description": "The /spectre and /spectre-new skills for the spectre CLI: spec-driven change tracking in markdown.",
      "version": "1.0.0",
      "author": {
        "name": "tweety53",
        "url": "https://github.com/tweety53"
      },
      "homepage": "https://github.com/tweety53/spectre#readme",
      "repository": "https://github.com/tweety53/spectre",
      "license": "MIT",
      "keywords": ["spectre", "specs", "change-tracking", "markdown"],
      "skills": "./.claude/skills/"
    }
    ```

  - [x] **Step 2: Write `.claude-plugin/marketplace.json`.** Its one plugin's `source` is this
    repository itself, so the catalogue and the plugin it lists are one checkout.

    ```json verified:schema at code.claude.com/docs/en/plugin-marketplaces; this exact body installed and read back in Step 3
    {
      "name": "spectre",
      "owner": {
        "name": "tweety53",
        "url": "https://github.com/tweety53"
      },
      "plugins": [
        {
          "name": "spectre",
          "source": "./",
          "description": "The /spectre and /spectre-new skills for the spectre CLI: spec-driven change tracking in markdown."
        }
      ]
    }
    ```

  - [x] **Step 3: Check both files parse.** Expected: no output at all. Any `INVALID:` line names a
    file that is not JSON.

    ```bash verified:run in this worktree
    for f in .claude-plugin/*.json; do
      python3 -c "import json,sys;json.load(open(sys.argv[1]))" "$f" || echo "INVALID: $f"
    done
    ```

  - [x] **Step 4: Install from this working tree and read the inventory back.** Expected: `details`
    prints `Skills (2)  spectre, spectre-new`. Run `claude plugin marketplace list` first — a
    marketplace named `spectre` must not already be configured, or `add` refuses.

    ```bash verified:run against this tree's manifests; reported both skills
    claude plugin marketplace add ./
    claude plugin install spectre@spectre
    claude plugin details spectre
    ```

  - [x] **Step 5: Remove the test install.** Not optional: leaving it configured points the
    operator's user settings at a relative local path instead of the GitHub repository. Expected:
    `list` no longer names `spectre`.

    ```bash verified:run after the cycle above; marketplace list returned to its prior contents
    claude plugin uninstall spectre@spectre
    claude plugin marketplace remove spectre
    claude plugin marketplace list
    ```

  - [x] **Step 6: Commit.**

    ```bash unverified:confirm these two paths are the only ones staged
    git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
    git commit -m "feat(plugin): add Claude Code plugin and marketplace manifests"
    ```

- [x] 2. Install the skills from Quick start

**Build:** green
**Files:** `README.md`
**Tests:** none — this task edits prose. The existing subtest `README file templates table` in
`internal/check/check_test.go` is the regression guard that the edit left the parsed part of the
file alone, and Step 3 runs it.
**Regression:** reverting this commit restores the sentence claiming `/spectre` and `/spectre-new`
are available to Claude Code users from `.claude/skills/`, which is false for anyone who installed
the binary with `go install` and has no checkout.
**Baseline:** before=197 after=197
**Commit:** `docs(readme): install the skills from Quick start`

The block goes immediately after Quick start's existing `go install` paragraph and before `Then,
one prompt per step:`, per decision `quick-start-placement`.

  - [x] **Step 1: Shrink the intro paragraph's claim to a pointer.** Replace these two lines, which
    are `README.md:11-12`:

    ```text verified:README.md lines 11-12 at 6181228
    Claude Code users can instead run `/spectre` and `/spectre-new`, shipped in this repository's
    [.claude/skills/](.claude/skills/). Prefer typing the commands yourself?
    ```

    with these:

    ```text unverified:confirm the whole paragraph still wraps within the file's existing width after the edit
    Claude Code users can instead run `/spectre` and `/spectre-new`, two skills that install
    separately — see below. Prefer typing the commands yourself?
    ```

  - [x] **Step 2: Add the install block.** Write this verbatim after the `go install` paragraph.
    The outer fence below is four backticks so the block's own three-backtick fences show; write
    them to `README.md` as ordinary three-backtick fences.

    ````markdown unverified:confirm every line wraps within the file's existing width
    The `/spectre` and `/spectre-new` skills install separately from the binary. In **Claude
    Code**, this repository is its own plugin marketplace:

    ```
    /plugin marketplace add tweety53/spectre
    /plugin install spectre@spectre
    ```

    `/plugin marketplace update spectre` picks up later changes to them. **With any other agent**,
    copy the two skill directories into whatever skills directory that agent reads — for Claude
    Code that directory is `~/.claude/skills/`, and others differ:

    ```bash
    git clone https://github.com/tweety53/spectre /tmp/spectre
    cp -r /tmp/spectre/.claude/skills/spectre /tmp/spectre/.claude/skills/spectre-new \
      ~/.claude/skills/
    ```

    Either way the skills only dispatch to the binary, so `spectre` still has to be on `PATH`.
    Inside a checkout of this repository no install is needed — Claude Code reads
    [.claude/skills/](.claude/skills/) directly.
    ````

  - [x] **Step 3: Check the width, then run the suite.** Expected: the width check prints only
    lines that were already over before this change, and none in the edited region; `gofmt` prints
    nothing; `go vet` and `go test` pass, `internal/check` included.

    ```bash verified:both commands run in this worktree
    python3 -c "
    for i, l in enumerate(open('README.md', encoding='utf-8'), 1):
        l = l.rstrip('\n')
        if len(l) > 99:
            print(i, len(l), l[:60])
    "
    gofmt -l . && go vet ./... && go test ./...
    ```

  - [x] **Step 4: Commit.**

    ```bash unverified:confirm README.md is the only path staged
    git add README.md
    git commit -m "docs(readme): install the skills from Quick start"
    ```

- [x] 3. Catch a manifest that disagrees with itself or with the README

**Build:** green
**Files:** `internal/check/plugin_manifest_test.go`
**Tests:** `TestPluginManifestsAgree`, with subtests `manifests parse`, `plugin name matches`,
`marketplace source resolves`, `skills path resolves`, `README install command matches`,
`README creates the destination before copying`.
**Regression:** reverting this commit restores the state the panel's Bugbot slot measured, in which
mutating the `skills` path, the marketplace `source`, the plugin `name` or the README's install
string one at a time each left the full suite green — all four ship silently.
**Baseline:** before=197 after=204
<!-- measured: go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN' @ branch spectre/kan-360-skills-installation-guide, before this task -->
<!-- predicted: the same command after this task — one test function plus six subtests -->
**Commit:** `test(check): assert the plugin manifests agree with the README`

This task exists because decision `manifest-validation-test` overturned the plan's original "no Go
code changes" constraint. It adds a test and no production code: the `spectre` binary's behaviour is
unchanged.

The test lives in `internal/check` because that package already holds the repository-consistency
tests that read `README.md` — see `docPaths` in `internal/check/check_test.go`, which resolves paths
from the test file's own location via `runtime.Caller` rather than from the working directory,
because `go test` may be invoked from anywhere. Follow that pattern rather than inventing a second
one.

**Assert nothing about the skills' own names.** They are `run` and `new` after decision
`skills-renamed-for-namespace`, and a test that hardcodes them fails the next time they change
without catching any of the four mutants this task targets. Assert that the directory the `skills`
field names exists and holds at least one `<dir>/SKILL.md`.

  - [x] **Step 1: Write the failing test.** Create `internal/check/plugin_manifest_test.go` in
    package `check`. Decode both manifests into minimal structs — only the fields asserted on — and
    resolve every path from the repository root the way `docPaths` does.

    ```go unverified:confirm the struct tags match the manifests' real field names before relying on the decode
    package check

    import (
        "encoding/json"
        "os"
        "path/filepath"
        "runtime"
        "strings"
        "testing"
    )

    // repoRoot resolves the repository root from this test file's own
    // location, for the same reason docPaths does: `go test` can be
    // invoked from anywhere, and internal/check sits two directories
    // below the root.
    func repoRoot(t *testing.T) string {
        t.Helper()
        _, thisFile, _, ok := runtime.Caller(0)
        if !ok {
            t.Fatal("could not determine this test file's own path")
        }
        return filepath.Join(filepath.Dir(thisFile), "..", "..")
    }

    type pluginManifest struct {
        Name   string `json:"name"`
        Skills string `json:"skills"`
    }

    type marketplaceManifest struct {
        Name    string `json:"name"`
        Plugins []struct {
            Name   string `json:"name"`
            Source string `json:"source"`
        } `json:"plugins"`
    }
    ```

  - [x] **Step 2: Add the five subtests.** Each one targets a mutant the panel proved survives
    today. Keep them as subtests of a single `TestPluginManifestsAgree` so the baseline count above
    holds.

    ```go unverified:confirm each failure message names the file and the value it read
    func TestPluginManifestsAgree(t *testing.T) {
        root := repoRoot(t)
        pluginPath := filepath.Join(root, ".claude-plugin", "plugin.json")
        marketPath := filepath.Join(root, ".claude-plugin", "marketplace.json")

        var plugin pluginManifest
        var market marketplaceManifest

        t.Run("manifests parse", func(t *testing.T) {
            for path, into := range map[string]any{pluginPath: &plugin, marketPath: &market} {
                raw, err := os.ReadFile(path)
                if err != nil {
                    t.Fatalf("reading %s: %v", path, err)
                }
                if err := json.Unmarshal(raw, into); err != nil {
                    t.Fatalf("%s is not valid JSON: %v", path, err)
                }
            }
        })

        t.Run("plugin name matches", func(t *testing.T) {
            if len(market.Plugins) != 1 {
                t.Fatalf("%s lists %d plugins, want exactly 1", marketPath, len(market.Plugins))
            }
            if market.Plugins[0].Name != plugin.Name {
                t.Errorf("marketplace names plugin %q, plugin.json names itself %q",
                    market.Plugins[0].Name, plugin.Name)
            }
        })

        t.Run("marketplace source resolves", func(t *testing.T) {
            src := filepath.Join(root, market.Plugins[0].Source, ".claude-plugin", "plugin.json")
            if _, err := os.Stat(src); err != nil {
                t.Errorf("marketplace source %q does not hold a plugin manifest: %v",
                    market.Plugins[0].Source, err)
            }
        })

        t.Run("skills path resolves", func(t *testing.T) {
            dir := filepath.Join(root, plugin.Skills)
            entries, err := os.ReadDir(dir)
            if err != nil {
                t.Fatalf("skills path %q does not resolve: %v", plugin.Skills, err)
            }
            var found int
            for _, e := range entries {
                if !e.IsDir() {
                    continue
                }
                if _, err := os.Stat(filepath.Join(dir, e.Name(), "SKILL.md")); err == nil {
                    found++
                }
            }
            if found == 0 {
                t.Errorf("skills path %q holds no <dir>/SKILL.md", plugin.Skills)
            }
        })

        t.Run("README install command matches", func(t *testing.T) {
            readme, err := os.ReadFile(filepath.Join(root, "README.md"))
            if err != nil {
                t.Fatalf("reading README.md: %v", err)
            }
            want := "/plugin install " + plugin.Name + "@" + market.Name
            if !strings.Contains(string(readme), want) {
                t.Errorf("README.md does not carry %q, which is what the manifests imply", want)
            }
        })
    }
    ```

  - [x] **Step 3: Prove each subtest catches its mutant.** For each of the four values below,
    change it, confirm the suite now fails and names the right thing, then revert. A subtest that
    stays green under its own mutation is not doing its job and must be fixed in this task.

    A sixth subtest guards the F3 fix, which the fix round's own mutation proof showed nothing
    catches: assert that in README.md the line `mkdir -p ~/.claude/skills` appears before the
    `cp -r` line in the manual-install block, so removing it fails the suite.

    ```bash verified:the five mutations below all survived the full suite before this task — four measured by the panel's Bugbot slot, the fifth by the fix round's own mutation proof
    # 1. plugin.json  "skills"            -> "./nonexistent-skills-dir/"
    # 2. marketplace.json plugins[0].source -> "./does-not-exist/"
    # 3. marketplace.json plugins[0].name   -> "spectre-typo"
    # 4. README.md    the install command   -> spectre@wrong
    # 5. README.md    delete the `mkdir -p ~/.claude/skills` line
    go test ./internal/check/ -run TestPluginManifestsAgree -v
    ```

  - [x] **Step 4: Run the suite and commit.**

    ```bash unverified:confirm the new test file is the only path staged
    gofmt -l . && go vet ./... && go test ./...
    git add internal/check/plugin_manifest_test.go
    git commit -m "test(check): assert the plugin manifests agree with the README"
    ```


- [x] 4. Merge the two skills into one

**Build:** green
**Files:** `.claude/skills/spectre/SKILL.md`, `.claude/skills/spectre-new/SKILL.md`,
`.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, `README.md`
**Tests:** none — this task is a file merge and prose. Task 3's `README install command matches`
subtest is what holds the README's install command to the manifests afterwards.
**Regression:** reverting this commit restores `.claude/skills/spectre-new/` and, with it, a README
and a pair of manifest descriptions that name a second skill the plugin route would invoke as
`/spectre:spectre-new`.
**Baseline:** before=197 after=197
<!-- measured: go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN' @ branch spectre/kan-360-skills-installation-guide, at commit b42693b -->
<!-- predicted: unchanged by this task, which touches no Go file -->
**Commit:** `refactor(skills): merge spectre-new into the spectre skill`

**This task was not planned. It came out of the review panel**, as the fix for finding F4: Claude
Code namespaces every plugin skill by its plugin name, so the README promised the plugin-install
audience commands they could not invoke. Its history is recorded rather than tidied away — the
operator first chose a rename to `run` and `new`, that dispatch was stopped mid-flight and its
staged work reverted before any commit, and decision `skills-renamed-for-namespace` is marked
superseded rather than deleted. See `.superpowers/sdd/final-review-panel.md` for the round's own log.

  - [x] **Step 1: Delete `.claude/skills/spectre-new/`** with `git rm -r`.

  - [x] **Step 2: Fold its behaviour into `.claude/skills/spectre/SKILL.md`** as a named exception to
    that skill's thin rule, keeping the directory name `spectre` so the bare command stays
    `/spectre`. The merged file states both routes: `/spectre <subcommand>` in a checkout or after a
    manual copy, `/spectre:spectre <subcommand>` after a plugin install.

  - [x] **Step 3: Rewrite both manifest descriptions** so neither names a second skill nor promises
    a command string that depends on install route.

  - [x] **Step 4: Rewrite the README's Quick start block** to state both routes, copy one skill
    directory instead of two, and `mkdir -p ~/.claude/skills` first.

  - [x] **Step 5: Commit** as `refactor(skills): merge spectre-new into the spectre skill`. Landed as
    `2d412a6`; the manifest and README halves were folded into `3ad7ef8` and `b42693b` as fixups.
    **That commit carries no `Task-Id` trailer**, alone among the eight: it was written during the
    review panel's first fix round, before this task existed to name it. It is left as it is rather
    than amended — amending would rewrite the five commits after it, and their shas are cited by this
    plan, the panel record and the session ledger.

- [x] 5. Harden the manifest test, and move it to the root

**Build:** green
**Files:** `plugin_manifest_test.go`, `internal/check/plugin_manifest_test.go`
**Tests:** `TestPluginManifestsAgree` moves to the root package and gains `skills path names the
real skill directory` and `README and skill agree on the directory name`.
**Regression:** reverting this commit restores a test that passes when `skills` is repointed at any
directory holding any `SKILL.md`, that panics rather than fails on an empty `plugins` array, and
that lets the skill directory be renamed while the README and the skill's own invocation line still
name the old one.
**Baseline:** before=204 after=206
<!-- measured: go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN' @ branch spectre/kan-360-skills-installation-guide, at commit 20dbe94 -->
<!-- predicted: the same command after this task — the move is count-neutral and two subtests are added -->
**Commit:** `test(spectre): harden the manifest test and move it to the root`

**This task was not planned.** It records the review panel's second pass: F7 and F8 from the
principles slot, and F9 through F12 from the Bugbot slot, whose mutation run found the first test
still let four mutants through. Decision `manifest-test-at-root` records the move and why.

  - [x] **Step 1: Move the file** to `plugin_manifest_test.go` at the repository root, package
    `spectre`, with `git mv`. The root helper becomes `filepath.Dir(thisFile)` — no `".."` segments,
    because the file now sits at the root it is resolving. That is what dissolves F7.

  - [x] **Step 2 (F11): guard the index.** `market.Plugins[0]` is read unguarded, so an empty
    `plugins` array panics and crashes the test binary instead of failing. Check the length in every
    subtest that indexes it, and fail with a message naming the file.

  - [x] **Step 3 (F9, F12): stop the shallow passes.** `skills path resolves` currently accepts any
    directory holding any `<dir>/SKILL.md`, so repointing `skills` at an unrelated directory with a
    stub still passes; `marketplace source resolves` accepts any directory holding a `plugin.json`.
    Assert instead that the resolved skills directory is the one the repository actually ships, and
    that the resolved source is this repository's own manifest — compare resolved absolute paths,
    not just existence.

  - [x] **Step 4 (F10): tie the directory name to what the docs claim.** Nothing currently connects
    the real skill directory name to the README's `cp -r ... /<name>` command or to the
    `/spectre:<name>` form the skill's own text claims, so renaming the directory breaks both with a
    green suite. Read the skill directory's real name and assert both documents carry it. Do not
    hardcode `spectre` — read it from disk, or the test asserts nothing when the name changes.

  - [x] **Step 5: Prove each new assertion.** For each of F9, F10, F11 and F12, apply the mutation,
    confirm the suite now fails and names the right thing, revert. A subtest that stays green under
    its own mutation is worthless and must be fixed in this task.

  - [x] **Step 6: Run the suite and commit.**

    ```bash unverified:confirm the moved file and its old path are the only paths staged
    gofmt -l . && go build ./... && go vet ./... && go test ./...
    git add plugin_manifest_test.go internal/check/plugin_manifest_test.go
    git commit -m "test(spectre): harden the manifest test and move it to the root"
    ```

- [x] 6. Anchor the name-agreement checks

**Build:** green
**Files:** `plugin_manifest_test.go`
**Tests:** `TestPluginManifestsAgree`'s `README and skill agree on the directory name` subtest,
tightened; no new subtest.
**Regression:** reverting this commit restores two unanchored `strings.Contains` calls over whole
files, under which a stale invocation string passes by prefix and a decoy mention anywhere in the
README satisfies a check about one specific command line.
**Baseline:** before=206 after=206
<!-- measured: go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN' @ branch spectre/kan-360-skills-installation-guide, at commit 1175529 -->
<!-- predicted: the same command after this task — the subtest is tightened, none added -->
**Commit:** `test(spectre): anchor the name-agreement checks to the lines they govern`

**This task was not planned.** It records the review panel's third pass, finding F13 from the Bugbot
slot, whose mutation run showed both checks added by task 5 can pass on a coincidence.

  - [x] **Step 1: Anchor the README check to the command line it governs.** The check currently
    matches the copy target anywhere in `README.md`, so an unrelated mention elsewhere satisfies it
    while the real `cp -r` line is broken. Find the manual-install command line itself and assert
    the target appears **on that line**, not in the file.

  - [x] **Step 2: Stop the invocation check passing by prefix.** `strings.Contains(own, invoke)` is
    satisfied by any longer string starting with `invoke`, so a stale `/spectre:spectre` passes a
    check for `/spectre:spec`. Require the match to end at a boundary — the invocation must not be
    followed by a character that could continue a directory name.

  - [x] **Step 3: Prove both.** Apply each mutation below, confirm the suite now fails and names the
    right thing, revert.

    ```bash verified:both mutations below were run by the panel's Bugbot slot against 1175529 and left the suite fully green
    # 1. rename the skill directory to a prefix of its old name, update README only,
    #    and leave SKILL.md's invocation string stale
    # 2. break the cp -r target, and add the correct string in an unrelated line elsewhere
    go test . -run TestPluginManifestsAgree -v
    ```

  - [x] **Step 4: Run the suite and commit.**

    ```bash unverified:confirm the test file is the only path staged
    gofmt -l . && go build ./... && go vet ./... && go test ./...
    git add plugin_manifest_test.go
    git commit -m "test(spectre): anchor the name-agreement checks to the lines they govern"
    ```

- [x] 7. Match tokens at their boundaries, not by substring

**Build:** green
**Files:** `plugin_manifest_test.go`
**Tests:** `TestPluginManifestsAgree`'s `README install command matches` and `README and skill agree
on the directory name` subtests, tightened; no new subtest.
**Regression:** reverting this commit restores three unanchored `strings.Contains` calls, under which
a wrong marketplace name, a wrong copy source and a stale invocation all pass whenever the correct
string is a prefix of the wrong one — and restores a `cp -r` line selection that fails on a correct
change if an unrelated `cp -r` example is ever added earlier in the README.
**Baseline:** before=206 after=206
<!-- measured: go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN' @ branch spectre/kan-360-skills-installation-guide, at commit dbe8f28 -->
<!-- predicted: the same command after this task — subtests are tightened, none added -->
**Commit:** `test(spectre): match documented tokens at their boundaries`

**This task was not planned.** It records the review panel's fourth pass: F14 from the principles
slot, F15 from Primary, and F16 and F17 from the Bugbot slot. F14 and F15 are the same underlying
defect seen from two directions — one says the boundary logic is hand-rolled where a `regexp` would
do, the other says it was applied to only one of the two checks that needed it.

  - [x] **Step 1: Replace the hand-rolled walk with one boundary matcher (F14).** The
    occurrence-walk and its `isDirNameChar` classifier reimplement what `regexp` states in two lines,
    and `regexp` is standard library, so it is allowed by the module's standard-library-only policy.
    Write one small helper that reports whether a token appears in some text followed by a
    directory-name boundary, and use it everywhere a documented token is checked.

  - [x] **Step 2: Apply it to the copy target (F15).** Round 3 anchored the invocation check and left
    `cpTarget` an unanchored `Contains`, so `.claude/skills/spectre-old` still satisfies a check for
    `.claude/skills/spectre`.

  - [x] **Step 3: Apply it to the install command (F16).** `/plugin install spectre@spectrecorp`
    satisfies an unanchored check for `/plugin install spectre@spectre` the same way.

  - [x] **Step 4: Stop selecting the copy line by document order (F17).** Taking the first line
    containing `cp -r` means an unrelated `cp -r` example added earlier in the README fails this
    subtest on a change that is entirely correct. Select the line by what it is for — the one that
    copies into the skills directory — and fail clearly if there is not exactly one such line.

  - [x] **Step 5: Prove all four.** Apply each mutation, confirm the suite now fails and names the
    right thing, revert.

    ```bash verified:all four mutations below were run against dbe8f28 and left the suite green, three by the panel and two re-confirmed by the parent on a clean tree
    # F15  README cp -r target      .claude/skills/spectre    -> .claude/skills/spectre-old
    # F16  README install command   spectre@spectre           -> spectre@spectrecorp
    # F17  insert an unrelated "cp -r" example ABOVE the Quick start section
    # F14  no mutation — a simplicity finding, proved by the diff being smaller
    go test . -run TestPluginManifestsAgree -v
    ```

  - [x] **Step 6: Run the suite and commit.**

    ```bash unverified:confirm the test file is the only path staged
    gofmt -l . && go build ./... && go vet ./... && go test ./...
    git add plugin_manifest_test.go
    git commit -m "test(spectre): match documented tokens at their boundaries"
    ```

- [x] 8. Drop the subsumed subtest

**Build:** green
**Files:** `plugin_manifest_test.go`
**Tests:** `TestPluginManifestsAgree` loses its `skills path resolves` subtest; no test is added.
**Regression:** reverting this commit restores a subtest whose two assertions are both already made
elsewhere — the path-identity assertion by `skills path names the real skill directory`, and the
holds-at-least-one-skill assertion by the `t.Fatalf` at the top of `README and skill agree on the
directory name`.
**Baseline:** before=206 after=205
<!-- measured: go test ./... -count=1 -v 2>&1 | grep -c '^=== RUN' @ branch spectre/kan-360-skills-installation-guide, at commit 278113c -->
<!-- predicted: the same command after this task — exactly one subtest is removed -->
**Commit:** `test(spectre): drop the subsumed skills-path subtest`

**This task was not planned.** It records finding F19 from the review panel's fifth pass, raised by
Primary and settled by the operator. The raising slot labelled it judgment rather than a defect; the
operator chose to act on it.

**The coverage argument was checked before acting, not assumed.** `skills path resolves` asserts two
things: that the directory `plugin.json`'s `skills` field names can be read, and that it holds at
least one `<dir>/SKILL.md`. `skills path names the real skill directory` asserts that same field
resolves to the repository's own `.claude/skills`, and `README and skill agree on the directory name`
opens by reading `.claude/skills` and calling `t.Fatalf` when it holds no `<dir>/SKILL.md`. Together
those two cover both assertions, so nothing is lost.

  - [x] **Step 1: Delete the `skills path resolves` subtest** and nothing else. Leave every other
    subtest, both helpers, and the shared `plugin`/`market` decoding exactly as they are.

  - [x] **Step 2: Confirm the coverage claim rather than trusting it.** Point `plugin.json`'s
    `skills` at a path that does not resolve and confirm the suite still fails; empty
    `.claude/skills/` of its skill directory and confirm the suite still fails. Revert both.

    ```bash verified:both assertions exist at 278113c in the two subtests named above
    go test . -run TestPluginManifestsAgree -v
    ```

  - [x] **Step 3: Run the suite and commit.**

    ```bash unverified:confirm the test file is the only path staged
    gofmt -l . && go build ./... && go vet ./... && go test ./...
    git add plugin_manifest_test.go
    git commit -m "test(spectre): drop the subsumed skills-path subtest"
    ```
