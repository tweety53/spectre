# kan-350-spectre-updates

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Seven tasks, in dependency order: `RootPath` is extracted before `init` can call it, `init` exists
before the no-tree error can name it, and the deletions and renames follow. Every task leaves a
green build.

**Verification, every task:**

```bash verified:run at 3b16340 on the unmodified tree
gofmt -l .        # must print nothing
go vet ./...      # must print nothing
go test ./...     # must be all ok
```

Test counts below are counts of `Test`/`Example` functions, as produced by:

```bash verified:run at 3b16340
git ls-files '*_test.go' | xargs grep -hcE '^func (Test|Example)' | paste -sd+ - | bc
```

The baseline is 135. <!-- measured: the command above @ 3b16340 -->

---

- [x] 1. Extract `tree.RootPath` from `tree.Open`

**Build:** green
**Files:** `internal/tree/tree.go`, `internal/tree/tree_test.go`
**Tests:** `TestRootPath`
**Regression:** reverting this commit removes `RootPath`, and `internal/cmd/init.go` (task 2) no
longer compiles — the path rule would have to be restated there, which is the duplication this
extraction exists to prevent.
**Baseline:** before=135 after=136
**Commit:** `refactor(tree): extract RootPath from Open`

`Open` currently resolves a path and stats it in one function. `init` needs the resolution without
the stat, so the resolution becomes its own exported function and `Open` calls it.

  - [x] **Step 1: Write `TestRootPath` first.** Table-driven — the cases are homogeneous (one input
    path, one expected output path), which is exactly what a table suits. Cases: a relative path
    whose basename is not `spectre` gains `/spectre`; a path already ending in `spectre` is
    returned unchanged; `.` resolves to `<cwd>/spectre`; an absolute path is handled the same as a
    relative one. Assert against absolute expectations built with `filepath.Join(t.TempDir(), …)`,
    never string-compared relative paths. Run it; it fails to compile, which is the red state.

  - [x] **Step 2: Add `RootPath`.** In `internal/tree/tree.go`:

    ```go unverified:confirm the doc comment matches the final signature
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
    ```

  - [x] **Step 3: Rewrite `Open` to call it.** `Open` keeps its `Stat` and its `ErrNoRoot` wrap;
    only the two path lines move out. Nothing about `Open`'s behaviour changes, and its existing
    tests in `internal/tree/tree_test.go` must pass untouched — that is the proof this step is
    behaviour-preserving.

  - [x] **Step 4: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean.

---

- [x] 2. Add the `init` command

**Build:** green
**Files:** `internal/cmd/init.go`, `internal/cmd/init_test.go`, `cmd/spectre/main.go`
**Tests:** `TestInitCreatesTree`, `TestInitFillsMissingDir`, `TestInitPreservesExistingConfig`,
`TestInitRootResolution`, `TestInitIsIdempotent`, `TestInitUnwritableParent`,
`TestInitScaffoldParsesToDefaults`, `TestInitThenNewThenList`, `TestInitRejectsPositionalArgs`,
`TestInitRejectsFileWhereDirBelongs`
**Regression:** reverting this commit removes `spectre init`; `spectre new` in a directory with no
tree again fails with no way forward, and the scaffolded `config.md` — whose body
`TestInitScaffoldParsesToDefaults` pins to `config.Default()` — is gone with the test that pinned
it.
**Baseline:** before=136 after=146
**Commit:** `feat(cmd): add init command`

`spectre init [--root <path>]`. Creates only what is absent, never overwrites, exits 0 whether it
created three things or none — and exits 1, creating nothing further, when a plain file sits where
`specs/` or `changes/` belongs.

  - [x] **Step 1: Write the tests first**, in `internal/cmd/init_test.go`, following the existing
    `internal/cmd/new_test.go` shape (call `cmd.Init(args, &stdout, &stderr)`, assert on the int
    return and the buffers — never on the real process streams).

    - `TestInitCreatesTree` — empty `t.TempDir()`, `--root <dir>`; asserts exit 0, and that
      `<dir>/spectre/specs`, `<dir>/spectre/changes` and `<dir>/spectre/config.md` all exist.
    - `TestInitFillsMissingDir` — a tree with `specs/` and `config.md` but no `changes/`; asserts
      exit 0 and that `changes/` now exists.
    - `TestInitPreservesExistingConfig` — writes a `config.md` with a sentinel body, runs `init`,
      asserts the file is byte-for-byte unchanged. This is the one destructive failure the command
      could have, so it is asserted on bytes, not on mtime or length.
    - `TestInitRootResolution` — table-driven over the same cases task 1 pinned on `RootPath`,
      asserted here through `Init`'s observable effect (which directory actually appeared).
    - `TestInitIsIdempotent` — runs `init` twice; asserts the second run exits 0, and that its
      stdout says nothing was created.
    - `TestInitUnwritableParent` — `os.Chmod(dir, 0o000)` on the parent, with a `t.Cleanup`
      restoring `0o755`, and — copied from `internal/cmd/refs_test.go`, which is the pattern to
      match — an `os.Geteuid() == 0` skip ahead of it. Without that guard the test fails under any
      root-run environment, a Docker CI image included, because root ignores the mode bits the test
      relies on. Asserts exit 2 and a non-empty stderr.
    - `TestInitRejectsFileWhereDirBelongs` — a plain file at `specs`, and separately at `changes`,
      makes `init` exit 1 and create nothing further.
    - `TestInitScaffoldParsesToDefaults` — strips the fence markers from `_configTemplate`,
      `config.Parse`s the un-fenced body, and asserts two things: that the result equals
      `config.Default()`, and that the template's `## Rules` lines equal `config.RuleNames` in
      order. Step 3 states why both assertions are needed. This test lives here, in `internal/cmd`,
      and **not** in `internal/config`: `internal/cmd` imports `internal/config`, so a test in
      package `config` reaching for `cmd.Init` would be an import cycle.
    - `TestInitRejectsPositionalArgs` — `init somewhere` exits `Usage` and creates nothing.
    - `TestInitThenNewThenList` — `init`, then `cmd.New`, then `cmd.List`; asserts the change shows
      up in `list`'s output. The scaffolded tree must be one the rest of the binary accepts.

  - [x] **Step 2: Write `internal/cmd/init.go`.** Match the file shape of `internal/cmd/new.go`:
    `func Init(args []string, stdout, stderr io.Writer) int`, using the shared `flagSet("init",
    stderr)` helper for `--root`, and returning `OK`/`Fail`/`Usage` from `internal/cmd/cmd.go`'s
    constants. It must **not** call `resolve` — that opens an existing tree, which is the thing
    `init` is creating. It calls `tree.RootPath(*root)` from task 1.

    Create `specs/`, `changes/` and `config.md` with `os.MkdirAll` / `os.WriteFile`, each guarded
    by an existence check so nothing is overwritten.

    **The existence check tests `fi.IsDir()`, not merely that `os.Stat` succeeded.** A plain file
    sitting where `specs/` or `changes/` belongs is not "already there" — treating it as such makes
    `init` report success while leaving a tree that no later command can use. It refuses instead,
    naming the offending path and exiting `Fail` (1), matching `new.go`'s `already exists` refusal:
    this is a content problem, not a usage or IO error. `config.md` existing as a *directory* is
    left to the skip path, where later commands fail with a clear error of their own. Report each created path to `stdout`; report
    "already complete" — or whatever wording reads plainly — when nothing was created. Every
    diagnostic goes to `stderr`, every result to `stdout`.

    Modes match `new.go`: `0o755` for directories, `0o644` for the file.

  - [x] **Step 3: Write the `config.md` template**, as a `const` in `init.go` beside
    `new.go`'s `_proposalTemplate`. Every default sits inside a fenced block so
    `internal/config`'s parser ignores it — the behaviour `internal/config/config_test.go` already
    covers — which is what makes the file documentation rather than configuration:

    ````markdown unverified:confirm against config.RuleNames order and config.Default() at implementation time
    # spectre config

    Every key below is a default, shown inside a fence so spectre ignores it.
    Move a line out of the fence to make it take effect.

    ## Rules

    ```
    - headings: error
    - placeholders: error
    - shall-clause: error
    - malformed-bullet: error
    - id-sequence: error
    - task-sequence: error
    - refs: error
    ```

    ## Vocabulary

    ```
    - modal: SHALL
    - id-prefix: R
    ```

    ## Layout

    ```
    - specs: specs
    - changes: changes
    - extension: .md
    ```
    ````

    The rule list must match `config.RuleNames` in order.

    **`TestInitScaffoldParsesToDefaults` must un-fence the template before parsing it.** Loading the
    scaffolded `config.md` as written proves nothing: `config.Parse` skips fenced content
    unconditionally, so `config.Load` returns `config.Default()` whatever the fence contains, and the
    test passes even against a template with a rule deleted or `id-prefix` changed. The test must
    strip the fence markers from `_configTemplate`, write the un-fenced body to a temp tree's
    `config.md`, `config.Load` that, and assert it equals `config.Default()`.

    **That assertion alone is still not enough, and the second one is the load-bearing half.**
    `config.Parse` starts from `config.Default()` and overrides only the keys it actually reads, so
    a rule line deleted or reordered — whose value already equals the default — parses to exactly
    the same `Config`. Parse-equality therefore catches a changed *value* (`id-prefix: R` → `X`) and
    nothing else. The test must **additionally** extract the `## Rules` section's literal lines from
    the template and assert they equal `config.RuleNames`, in order. Between the two assertions, a
    changed value, a deleted rule and a reordered pair all fail.

  - [x] **Step 4: Register the command** in `cmd/spectre/main.go` — a `case "init":` arm calling
    `cmd.Init`, and an `init [--root <path>]` line in the `_usage` block's command list, placed
    first, since it is the first command a new user runs.

    **The usage line states `--root`, never a bare positional.** `init` takes no positional
    argument, so a line reading `init <path>` documents an invocation that silently creates the
    tree in the working directory instead of at the named path — exit 0, no error, wrong location.
    The bracket form matches the convention `list [--specs] [--json]` already uses.

    **`Init` also rejects an unexpected positional argument** — `fs.NArg() != 0` exits `Usage`,
    matching `new`, `archive` and `refs`, which all validate their positional count. Documenting the
    flag is not enough on its own: a user who types `spectre init somewhere` should be told, not
    silently given a tree in the wrong directory.

  - [x] **Step 5: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean.

---

- [x] 3. Name `spectre init` in the no-tree error

**Build:** green
**Files:** `internal/tree/tree.go`, `internal/tree/tree_test.go`
**Tests:** `TestErrNoRootNamesInit`
**Regression:** reverting this commit returns the error to a message that reports the problem and
names no fix; `TestErrNoRootNamesInit` fails.
**Baseline:** before=146 after=147
**Commit:** `fix(tree): name spectre init in the no-tree error`

  - [x] **Step 1: Write `TestErrNoRootNamesInit`** — `tree.Find` in an empty `t.TempDir()`, assert
    the returned error's message contains `spectre init`. Assert `errors.Is(err, tree.ErrNoRoot)`
    still holds, so the wrap is not accidentally broken.

  - [x] **Step 2: Change both messages.** `Find`'s wrap and `Open`'s wrap both gain the pointer,
    e.g. `no spectre/ directory found (searched upwards from %s); run "spectre init" to create
    one`. `ErrNoRoot` itself — the sentinel `errors.Is` matches on — is left alone.

  - [x] **Step 3: Fix the callers' expectations.** Grep the test tree for assertions on the old
    message text and update each. Do not weaken an exact assertion into a substring match to avoid
    the edit.

  - [x] **Step 4: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean.

---

- [x] 4. Remove the `migrate` command

**Build:** green
**Files:** `internal/cmd/migrate.go`, `internal/cmd/migrate_test.go`, `internal/openspec/read.go`,
`internal/openspec/read_test.go`, `cmd/spectre/main.go`, `README.md`
**Allowed-collateral:** `internal/tree/tree.go`
**Tests:** none added — this task deletes 16 test functions and adds none. A green build plus the
surviving suite is the verification.
**Regression:** reverting this commit restores ~960 lines of OpenSpec conversion and reintroduces
the binary's only `--root` exception.
**Baseline:** before=147 after=131
**Commit:** `feat(cmd)!: remove the migrate command`

Deleted outright — no dormant package, no second binary.

  - [x] **Step 1: Delete the four files** — `internal/cmd/migrate.go`,
    `internal/cmd/migrate_test.go`, `internal/openspec/read.go`, `internal/openspec/read_test.go`.
    The `internal/openspec/` directory goes with them.

  - [x] **Step 2: Unregister it** in `cmd/spectre/main.go`: remove the `case "migrate":` arm and
    the `migrate <openspec-dir>` usage line. Rewrite `_usage`'s closing sentence — currently
    `every command except migrate accepts --root <path> to name the tree explicitly; migrate takes
    a source path and --out instead` — as `every command accepts --root <path> to name the tree
    explicitly`.

  - [x] **Step 3: Remove it from the README** — the `## migrate` section, the Commands-table row,
    and the paragraph beginning `migrate is the one exception to --root`. (Task 6 rewrites the rest
    of the README; this task removes only what has no subject left.)

  - [x] **Step 4: Confirm nothing else references it, and fix what does.** Two references to
    `migrate` live outside the files listed above and dangle once it is deleted: `tree.At`'s doc
    comment names `cmd.Migrate` as its example caller, and the README's per-repository-configuration
    section carries the parenthetical `(and how migrate writes into it)`. Both are rewritten here —
    that is what `**Allowed-collateral:**` covers. Do not leave a comment naming a function that no
    longer exists.

  - [x] **Step 4b: Run the grep.** `grep -rn 'openspec\|migrate' --include='*.go'
    --include='*.md' .` should return only hits inside `spectre/changes/` and `docs/superpowers/`,
    which are this change's own artefacts and a historical record.

  - [x] **Step 5: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean.

---

- [x] 5. Rename the example peer tree from `gymie` to `web`

**Build:** green
**Files:** `README.md`, `internal/tree/tree_test.go`, `internal/tree/peer_test.go`,
`internal/cmd/refs_test.go`, `internal/cmd/validate_test.go`, `internal/check/refs_test.go`,
`internal/check/citations_test.go`, `internal/parse/spec_test.go`, `internal/parse/ref_test.go`,
`internal/render/render_test.go`
**Allowed-collateral:** `internal/model/model.go`
**Tests:** none added — the existing tests are the subject, and must pass unchanged in behaviour
after the rename.
**Regression:** reverting this commit puts a real project's name back into the README's citation
example and ~85 fixture occurrences.
**Baseline:** before=131 after=131
**Commit:** `refactor: rename the example peer tree to web`

  - [x] **Step 1: Rename.** `gymie-frontend` → `web-frontend` **first**, then `gymie` → `web` —
    in that order, so the longer name is not half-rewritten by the shorter one's pass. Covers
    fixture directory names, `peers` file bodies, expected-output strings, and citation literals
    such as `(@gymie:plans#R4)`.

  - [x] **Step 2: One occurrence sits outside the test tree.** `internal/model/model.go`'s `Ref`
    struct documents its `Raw` field with the literal `// "@gymie:plans#R7"`. It is a doc comment,
    not a fixture, which is why it is listed under `**Allowed-collateral:**` rather than `**Files:**`
    — but it is a real occurrence and step 4's grep fails without it.

  - [x] **Step 2b: Leave the neighbours alone.** The capability fixtures `auth`, `billing`, `plans`
    and `app` are already neutral. `github.com/tweety53/spectre` is this module's real import path,
    not an example.

  - [x] **Step 3: Confirm.** `grep -rn gymie --include='*.go' --include='*.md' .` returns hits only
    under `docs/superpowers/`, which is a finished change's planning record and deliberately
    untouched.

  - [x] **Step 4: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean. The suite must
    be green with no test-count change: a rename that alters a count changed behaviour.

---

- [x] 6. Split the README's reference sections into `docs/`

**Build:** green
**Files:** `README.md`, `docs/spec-format.md`, `docs/configuration.md`, `docs/references.md`
**Tests:** none — documentation only.
**Regression:** reverting this commit restores the long README and deletes the three pages; no
behaviour changes either way.
**Baseline:** before=131 after=131
**Commit:** `docs(readme): split the reference sections into docs/`

  - [x] **Step 1: Move three sections out**, each into its own page, in full — worked examples
    included. `## Spec format` → `docs/spec-format.md`; `## Per-repository configuration` →
    `docs/configuration.md`; `## References across trees` → `docs/references.md`. Give each page a
    title heading and a one-line opener; nothing else is added, and nothing is summarised away.

  - [x] **Step 2: Rewrite getting started.** The `mkdir -p spectre/specs spectre/changes` recipe,
    the `git init` line, and the paragraph stating that no `init` command exists all collapse into
    `spectre init`. Keep the sentence explaining that `archive` needs git — that is still true and
    is stated nowhere else.

  - [x] **Step 3: Add the `init` row** to the Commands table: `spectre init`, flag `--root`, "creates
    `specs/`, `changes/` and a `config.md` of commented defaults; fills in what's missing".

  - [x] **Step 4: Link the three pages** from the README, each where its section used to sit — one
    line naming what the page covers, so a reader who wanted that section knows where it went.

  - [x] **Step 5: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean (unchanged —
    this task touches no Go). Re-read the README end to end: every fact it used to state is either
    still there or on one of the three pages.

---

- [x] 7. Stop ignoring the tree path; build to `bin/`

**Build:** green
**Files:** `.gitignore`, `README.md`
**Tests:** none — repository configuration.
**Regression:** reverting this commit makes `.gitignore` hide `spectre/` again, so this
repository's own tree — and every change in it — silently disappears from `git status`.
**Baseline:** before=131 after=131
**Commit:** `chore: ignore /bin/ instead of the tree path`

  - [x] **Step 1: Edit `.gitignore`** — remove `/spectre`, add `/bin/`.

  - [x] **Step 2: Document the build** in the README's install section: `go install` stays the way
    to get the binary; add one line for a local build, `go build -o bin/spectre ./cmd/spectre`.

  - [x] **Step 3: Confirm the tree is visible.** `git status --short` must now show
    `spectre/changes/kan-350-spectre-updates/` as untracked rather than hiding it, and
    `git check-ignore -v spectre` must report no match.

  - [x] **Step 4: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean.
