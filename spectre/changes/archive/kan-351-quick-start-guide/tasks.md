# kan-351-quick-start-guide

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Seven tasks. The binary changes come first, in dependency order — the heading checker gains
ordering before new files are handed to it, and `render.Tasks` changes before `new` calls it with a
new signature — then the docs, which describe the behaviour the earlier tasks produce. Every task
leaves a green build.

**Verification, every task:**

```bash verified:run at 6271d41 on the unmodified tree
gofmt -l .        # must print nothing
go vet ./...      # must print nothing
go test ./...     # must be all ok
```

Test counts are counts of `Test`/`Example` functions, as produced by:

```bash verified:run at 6271d41
git ls-files '*_test.go' | xargs grep -hcE '^func (Test|Example)' | paste -sd+ - | bc
```

The baseline is 131. <!-- measured: the command above @ 6271d41 -->

## Final Go verification — after task 7, before handoff

**Every Go file this change touches gets a dedicated Go-skills audit once all seven tasks are in.**
Per-task reviews look at one commit; this pass looks at the Go changes as a whole, which is where
cross-task problems live — a signature changed in one task and used in another, a constant hoisted
between packages, a rule split across two gates.

The Go files this change touches: `internal/check/check.go`, `internal/render/render.go`,
`internal/cmd/new.go`, `internal/tree/tree.go`, and their tests.

**Skills that apply, and must each be invoked and applied:** `go-code-review`,
`go-coding-standards`, `go-error-handling`, `go-documentation`, `go-test-quality`,
`go-test-table-driven`, `go-refactoring`, `go-architecture-review`, `go-data-structures`,
`go-interface-design`, `go-semantic-tools`, `go-security-audit`, `go-dependency-audit`,
`go-project-layout`, `go-cli`, `go-performance-review`, `go-troubleshooting`, `go-modernize` and
`modern-go-guidelines:use-modern-go`.

**Skills that do not apply, and why — stated rather than silently skipped:**
`go-concurrency-review`, `go-context` (no goroutines, no context in this codebase),
`go-api-design`, `go-grpc`, `go-database`, `go-observability`, `go-dependency-injection`,
`go-design-patterns` (no HTTP, RPC, database, logging or DI surface anywhere in the change),
`go-ci` (no CI configuration exists in this repository).

The pass reports per skill: applied with no findings, applied with findings, or not applicable with
the reason. A finding here is handled like any other — fixed before handoff, never waived.

**`validate` scans open changes only** — `tree.Changes()` calls `changes(false)`, which skips the
`archive` entry. No task below needs to repair an archived change. <!-- verified: read at 6271d41, internal/tree/read.go:51-73 -->

---

- [x] 1. Give the headings rule an ordering check

**Build:** green
**Files:** `internal/check/check.go`, `internal/check/check_test.go`
**Tests:** `TestHeadingOrder`
**Regression:** reverting this commit lets a `proposal.md` carrying `## What changes` before
`## Why` pass, which is the template drift task 5 exists to catch.
**Baseline:** before=131 after=132
**Commit:** `feat(check): report headings that appear out of template order`

`headingFindings` currently records presence in a `map[string]bool` and reports only what is
missing. Order is invisible to it.

  - [x] **Step 1: Write `TestHeadingOrder` first.** Table-driven — homogeneous cases, one input
    body and one expected finding set. Cases: correct order → no findings; two required headings
    swapped → one ordering finding naming both; a required heading missing → the existing missing
    finding and **not** an ordering finding on top of it; extra headings interleaved → no findings;
    required headings inside a ``` fence → still missing, not ordered. Run it; it fails.

  - [x] **Step 2: Record position, not just presence.** Keep the fence skipping and the
    trailing-whitespace trim exactly as they are — both are load-bearing and already tested. Record
    the line index of each wanted heading's **first** occurrence, then report a finding when the
    indices of the headings that are present are not ascending.

    **A missing heading must not also produce an ordering finding.** Two findings for one cause is
    noise; the missing-heading message already names the problem.

  - [x] **Step 3: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` all clean. The existing
    heading tests must pass untouched — this task adds a check, it does not change what "missing"
    means.

---

- [x] 2. Check the tasks.md and design.md templates

**Build:** green
**Files:** `internal/check/check.go`, `internal/check/check_test.go`
**Allowed-collateral:** `internal/check/config_test.go`, `internal/cmd/archive_test.go`, `internal/cmd/list_test.go`
**Tests:** `TestTasksHeading`, `TestDesignFindings`, `TestEmptyTasksReported`,
`TestStructuralDesignAndTasksTemplates`, `TestStructuralDesignUnreadablePropagatesError`
**Regression:** reverting this commit stops `validate` reading `design.md` at all and lets a
`tasks.md` with no title and no tasks pass.
**Baseline:** before=132 after=137
**Commit:** `feat(check): check the tasks.md and design.md templates`

  - [x] **Step 1: Write the three tests first.**
    - `TestTasksHeading` — a `tasks.md` whose title is not `# <change-id>` is a finding; the correct
      title is not. The change id has to reach the checker; see step 3.
    - `TestDesignFindings` — a `design.md` missing `## Decisions` is a finding; one carrying both
      headings out of order is an ordering finding (task 1's check, reused); one carrying both in
      order plus extra headings is clean.
    - `TestEmptyTasksReported` — a `tasks.md` parsing to zero tasks is a finding under
      `task-sequence`; with the rule off, it is not.

  - [x] **Step 2: Add `DesignFindings`**, mirroring `ProposalFindings`: gate on `headings`, want
    `## Context` then `## Decisions`; gate placeholders on `placeholders`.

  - [x] **Step 3: Thread the change id into the change-level checks.** `ProposalFindings` and the
    tasks check both need `# <change-id>` as a wanted heading, and neither takes the id today. Add
    it as a parameter rather than re-deriving it inside the checker — `Structural` already has
    `c.ID` in hand at the call site, and a checker that re-derives an id from a path is a second
    source of truth for what a change is called.

  - [x] **Step 4: Read `design.md` in `Structural`, only when it exists.** `os.ReadFile` returning
    `os.ErrNotExist` is **not** a finding — the file is optional. Any other read error propagates,
    exactly as the `proposal.md` and `tasks.md` branches already do. Do not add a "missing
    design.md" finding.

    **The unreadable case needs a test of its own, or nothing pins it.** The whole three-armed
    switch can be replaced by `if err == nil { … }` and every other test still passes — measured.
    Assert that an unreadable `design.md` makes `Structural` return an error, guarding the test with
    `os.Geteuid() == 0` exactly as `internal/cmd/refs_test.go` does, since root ignores the mode
    bits the test relies on. The identical gap on the `proposal.md` and `tasks.md` branches predates
    this change and is deliberately left alone.

  - [x] **Step 5: Report an empty tasks.md** inside `TaskFindings`, under the existing
    `task-sequence` gate, so turning that rule off turns this off with it.

  - [x] **Step 6: Existing fixtures will fail, and are updated rather than accommodated.** Requiring
    `# <change-id>` makes every fixture whose `tasks.md` is titled `# Tasks` fail — in
    `internal/cmd/list_test.go`, `internal/cmd/archive_test.go` and `internal/check/config_test.go`,
    which is what `**Allowed-collateral:**` covers. Bring each fixture's title in line with its own
    change id. **Never weaken the new rule to let an old fixture pass.**

  - [x] **Step 7: Verify.** All three commands clean.

---

- [x] 3. Title tasks.md with the change id

**Build:** green
**Files:** `internal/render/render.go`, `internal/render/tasks_test.go`, `internal/cmd/new.go`
**Allowed-collateral:** `internal/cmd/new_test.go`
**Tests:** `TestTasksTitle`
**Regression:** reverting this commit returns the title to `# Tasks`, which task 2's
`TestTasksHeading` then rejects — the scaffold would produce a change that does not validate.
**Baseline:** before=137 after=138
**Commit:** `refactor(render): title tasks.md with the change id`

  - [x] **Step 1: Write `TestTasksTitle`** — `Tasks("my-change", nil)` starts `# my-change`, and the
    existing task-rendering behaviour is unchanged.

  - [x] **Step 2: Change the signature** to `Tasks(id string, ts []model.Task) []byte` and write
    `# <id>`. Update the doc comment.

  - [x] **Step 3: Update the one caller**, `internal/cmd/new.go:70`, to pass the change id. It is
    the only caller — `grep -rn 'render\.' --include='*.go' internal cmd` returns that line alone.
    <!-- verified: grep run at 6271d41 -->

  - [x] **Step 4: Fix the existing render tests** that assert `# Tasks`. Update them to the new
    title; do not weaken an assertion to accept either.

  - [x] **Step 5: Verify.** All three commands clean.

---

- [x] 4. Scaffold design.md

**Build:** green
**Files:** `internal/cmd/new.go`, `internal/cmd/new_test.go`, `internal/tree/tree.go`,
`internal/check/check.go`
**Tests:** `TestNewWritesDesign`, `TestNewScaffoldValidates`
**Regression:** reverting this commit stops `new` writing `design.md`; `TestNewScaffoldValidates`
still passes, since design.md is optional — `TestNewWritesDesign` is what fails.
**Baseline:** before=138 after=140
**Commit:** `feat(cmd): scaffold design.md alongside the proposal`

  - [x] **Step 1: Write both tests first.**
    - `TestNewWritesDesign` — `new` writes `design.md` carrying `## Context` and `## Decisions`.
    - `TestNewScaffoldValidates` — scaffold a change with `cmd.New`, then run the checkers over the
      tree and assert the findings are **exactly one**, the `no tasks` finding on `tasks.md`, and
      nothing else. This is the test that stops the scaffold and the rules drifting apart; without
      it, tasks 1–2 can tighten a rule the scaffold then violates and nothing notices until a user
      hits it.

      **It asserts one finding rather than none, and that is deliberate** — see the
      `scaffold-reports-no-tasks` decision in `design.md`. A freshly scaffolded change genuinely has
      no tasks yet, and task 2 made that a finding. Asserting zero findings here would be asserting
      something false; the test's real job is that the scaffold produces *no other* finding, which
      is what pins scaffold and rules together. **Do not** relax the empty-tasks rule, and **do
      not** have `new` write a placeholder task to dodge this — a scaffolded fake task is worse than
      an honest finding, because `archive` would happily archive it unchecked.

  - [x] **Step 2: Hoist the filename to `tree.DesignFile` first.** Task 2 put `design.md`'s name in
    an unexported `designFile` const inside `internal/check`, correctly, because `internal/check`
    was then its only caller. This task is the second caller, so the name moves beside
    `tree.ProposalFile` and `tree.TasksFile` — whose own doc comment calls them "the one source of
    truth every package that creates, reads, or validates a change agrees on". Add
    `tree.DesignFile`, point `internal/check` at it, and delete the local const. Two packages each
    holding their own spelling of the same filename is exactly the drift that comment exists to
    prevent.

  - [x] **Step 3: Add `_designTemplate`** beside `_proposalTemplate`, same `<!-- -->` comment style.
    **Never the literal `TODO` or `TBD`** — the `placeholders` rule reports both, so a stub using
    them would make every freshly scaffolded change fail `validate` immediately.

  - [x] **Step 4: Write it in `New`**, in the same `files` map, so the existing failure path — which
    removes the whole directory when any write fails — covers it unchanged.

  - [x] **Step 5: Verify.** All three commands clean, and confirm by hand:
    `cd $(mktemp -d) && git init -q && spectre init && spectre new demo && spectre validate`
    reports exactly `changes/demo/tasks.md:1: no tasks` and exits 1. Then add one task to
    `tasks.md` and re-run `validate`: it must now report **no findings**. Both halves are the check
    — the first proves the scaffold trips only the rule it should, the second proves it trips
    nothing else.

---

- [x] 5. Write docs/example.md — the end-to-end walkthrough

**Build:** green
**Files:** `docs/example.md`
**Tests:** none added — this task writes no Go. Conformance of the published file bodies is proved
by step 4 instead: the fences are extracted from the page, written into a scratch tree, and run
through spectre validate. That proof is a step in this task rather than a test in the suite, which
is exactly the gap the open question example-bodies-untested in design.md records.
**Regression:** reverting this commit removes the worked example the README's quick start links to
at every step, leaving those links dangling.
**Baseline:** before=140 after=140
**Commit:** `docs(example): add the worked multi-channel notifications change`

**This page is the end-to-end read.** It runs top to bottom, in the order a real user hits the
steps, from an empty directory to an archived change. Every step is the same three beats, in this
order:

1. **the prompt or command** — verbatim and copy-pasteable;
2. **the generated output** — exactly what came back, whether that is the binary's own stdout or the
   agent's reply;
3. **the resulting files** — in full, followed by a short explanation of what to look at and what to
   verify by hand.

A reader must be able to start at the top, read straight down, and understand how to use spectre end
to end without following a single link or jumping back. Nothing is deferred to "see above" or "see
the README".

  - [x] **Step 1: Lay out the steps in the order a user hits them**, and keep that order for the
    whole page: install → `init` → `new` → write the spec → fill in the change's files → `validate`
    → work the tasks → `archive`. The page's structure is the walk; there is no separate reference
    section and no appendix.

  - [x] **Step 2: Write the capability spec** — `spectre/specs/notifications.md`, shown in full:
    `# notifications`, `## Purpose`, `## Requirements` with numbered `SHALL` bullets covering
    channel selection, fallback when a channel fails, and per-user channel preference. Requirement
    ids `R1`, `R2`, … gap-free.

  - [x] **Step 3: Write the change's three files** in full, for change id
    `multi-channel-notifications`: `proposal.md` (why email-only limits delivery and open rates,
    what changes), `tasks.md` (numbered tasks that read like real work), `design.md` (`## Context`,
    `## Decisions`, recording why the chosen channels were chosen and what was rejected).

  - [x] **Step 4: Prove every body against the rules mechanically, not by reading.** Each file shown
    must satisfy tasks 1–4: titles, heading order, at least one task, `SHALL` on every requirement
    bullet, gap-free ids.

    **Extract the fences from the page programmatically** — do not retype them into the scratch
    tree. Retyping proves that something you typed validates, not that the page's own content does,
    and the two drift the moment the page is edited. Write the extracted bodies into a scratch tree
    and run `spectre validate` there.

  - [x] **Step 5: Pair every file with the exact agent prompt that produced it.** For each of the
    four files, the page carries, in this order: the **verbatim prompt** a reader can paste into an
    AI coding agent, the **file the agent produced**, and a short **explanation of the result** —
    what to look at in it, and why it is shaped that way.

    The prompts are literal and copy-pasteable, not paraphrases like "ask the agent to write a
    proposal". Each names the file to write, the headings the template requires, and the constraint
    that matters — for the spec, that every requirement bullet carries `SHALL` and ids are gap-free;
    for `tasks.md`, that tasks are `- [ ] <n>. …` numbered from 1.

    **Agent-agnostic.** No product name, no slash command, no vendor-specific syntax — a prompt, not
    a tool invocation, so it works in whatever agent the reader uses.

  - [x] **Step 6: The explanations say what the agent got right and what a reader must check.** An
    agent's output is a draft, not an authority. Each explanation names at least one thing to verify
    by hand — that requirement ids are gap-free, that the tasks are genuinely ordered by dependency,
    that the design records what was rejected and not only what was chosen.

  - [x] **Step 7: Show the real command output, not an idealised one.** Every `spectre` invocation
    the page shows must be run and its actual stdout pasted — including the `validate` that reports
    `no tasks` on a freshly scaffolded change, which is the tool telling the reader what to do next
    (`scaffold-reports-no-tasks` in `design.md`). A walkthrough showing output nobody will see is
    the failure this step exists to prevent.

  - [x] **Step 8: Keep it vendor-neutral.** No framework, no product name, no language-specific API.
    The reader is meant to take the shape, not the domain.

  - [x] **Step 9: Read it top to bottom as a stranger would.** Nothing may depend on knowing the
    README, the other docs pages, or spectre's internals. Every term is introduced before it is used.
    If a step only makes sense after reading a later one, reorder.

---

- [x] 6. Rewrite the README's opening as a quick start

**Build:** green
**Files:** `README.md`
**Tests:** none — documentation only.
**Regression:** reverting this commit restores the separate `## Install` and `## Getting started`
sections and removes the single path a first-time reader follows.
**Baseline:** before=140 after=140
**Commit:** `docs(readme): open with a step-by-step quick start`

  - [x] **Step 1: Replace `## Install` and `## Getting started`** with one `## Quick start`,
    immediately after the one-line description. Both absorbed sections are deleted, not left in
    place — stating installation twice is the duplication this README was just cut to remove.

  - [x] **Step 2: Write the numbered walk** — install globally, `init`, `new`, fill in the files,
    write the spec, `validate`, work the tasks, `archive`. Each step names its command and what it
    produces, and links to that file's full body in `docs/example.md`.

    **Every step that produces content carries the exact prompt to hand an AI agent**, verbatim and
    copy-pasteable, followed by one line on what comes back. The steps that are pure commands
    (`init`, `validate`, `archive`) carry the command and its real output instead — they need no
    agent.

    **The README is the short path; `docs/example.md` is the full read.** The quick start gets a
    reader from nothing to a validated change in one screen and points at the walkthrough for the
    complete prompt-output-files sequence. It does not retell the walkthrough. Prompt text is the
    one thing deliberately duplicated between them, since a prompt you have to follow a link to copy
    is a prompt the reader will not use.

  - [x] **Step 3: Carry both surviving facts across, verbatim in substance.** `archive` needs the
    tree inside a git repository with the change's files already tracked; `--root` must precede any
    positional argument, because Go's `flag` package stops parsing flags at the first one. Neither
    is stated anywhere else in the repository. Losing either is the failure mode of an absorb.

  - [x] **Step 4: The walk must not claim `validate` is clean right after `new`.** It is not: a
    freshly scaffolded change reports `no tasks` until you write the plan. Order the steps so
    `validate` comes *after* filling in `tasks.md`, or show the `no tasks` finding as the expected
    prompt to write it. A quick start whose very first `validate` contradicts what the reader sees
    is worse than no quick start.

  - [x] **Step 5: Do not duplicate the example.** The README shows commands and outcomes;
    `docs/example.md` shows file contents. No file body appears in full in both.

  - [x] **Step 6: Run every command the new section shows**, in a scratch directory, and confirm
    each behaves as written — including the `--root`-after-positional line, which must still fail.

---

- [x] 7. Update the docs that describe the old behaviour

**Build:** green
**Files:** `README.md`
**Tests:** none — documentation only. Correctness is shown by triggering each documented rule
against the real binary, in step 5.

`docs/spec-format.md` was in this field when the plan was written, on the assumption it also called
`design.md` unparsed. It never mentioned `design.md` at all — checked, not assumed — so it needed no
edit and was dropped from the list rather than touched to match a guess.
**Regression:** reverting this commit restores prose stating that `design.md` is unparsed and that
the tree carries two generated files, both false once tasks 2 and 4 land.
**Baseline:** before=140 after=140
**Commit:** `docs: describe the enforced templates and the third generated file`

  - [x] **Step 1: Fix every "optional, unparsed" claim about `design.md`.** It is now validated when
    present. Grep both files for `unparsed` and `optional` and correct each hit rather than the one
    you remember.

  - [x] **Step 2: Update the tree diagram** in the README to show `design.md` as a file `new`
    creates, alongside `proposal.md` and `tasks.md`.

  - [x] **Step 3: State the templates** — which headings each generated file must carry, in order,
    and that extra headings are allowed. Put this where the reader meets the file shapes, not in the
    quick start, which stays a walk rather than a reference.

  - [x] **Step 4: Sweep for anything else the change falsified.** Task 3 changed the `tasks.md`
    title; task 2 added an empty-tasks finding. Grep for `# Tasks` and for any prose describing what
    `validate` checks, and correct what is now wrong.

  - [x] **Step 5: Verify.** All three commands clean, and `spectre validate` in this repository
    reports no findings.

---

- [x] 8. Add the project instruction file

**Build:** green
**Files:** `AGENTS.md`, `CLAUDE.md`
**Tests:** none — project configuration.
**Regression:** reverting this commit removes the standing instruction, and an agent working in this
repository is again told nothing about the Go skills it is expected to apply.
**Baseline:** before=140 after=140
**Commit:** `docs(agents): require the Go skills on every change`

Requested by the operator during this change's implementation, and recorded in the Jira issue's
description rather than left only in this plan. This repository currently has **no** instruction
file of any kind — `AGENTS.md`, `CLAUDE.md`, `.cursorrules` and `.codex` are all absent — so this
creates the first one.

  - [x] **Step 1: Write `AGENTS.md`** as the canonical file, because it is the cross-harness
    convention rather than one vendor's. It states, as a standing instruction for any agent working
    in this repository: every change touching Go code applies the applicable Go skills, and names
    them — `go-code-review`, `go-coding-standards`, `go-error-handling`, `go-documentation`,
    `go-test-quality`, `go-test-table-driven`, `go-refactoring`, `go-architecture-review`,
    `go-data-structures`, `go-interface-design`, `go-semantic-tools`, `go-security-audit`,
    `go-dependency-audit`, `go-project-layout`, `go-cli`, `go-performance-review`,
    `go-troubleshooting`, `go-modernize`, `modern-go-guidelines:use-modern-go`.

  - [x] **Step 2: Say which do not apply here, and why**, so a reader does not have to re-derive it
    every time: `go-concurrency-review` and `go-context` (this codebase has no goroutines and no
    `context.Context`); `go-api-design`, `go-grpc`, `go-database`, `go-observability`,
    `go-dependency-injection`, `go-design-patterns` (no HTTP, RPC, database, logging or DI surface);
    `go-ci` (no CI configuration exists). **A skill is skipped by a stated judgement, never
    silently.** Where the codebase later grows one of those surfaces, the corresponding line is
    deleted rather than the rule quietly ignored.

  - [x] **Step 3: State the project's own facts an agent needs** and would otherwise guess: the
    module targets the Go version in `go.mod`; the checks are `gofmt -l .`, `go vet ./...` and
    `go test ./...`, all of which must be clean; the standard library only, no external
    dependencies; exit codes are a contract (`0` success, `1` findings or a content refusal, `2`
    usage or IO error).

  - [x] **Step 4: Add `CLAUDE.md` as a one-line pointer** at `AGENTS.md`, so Claude Code picks the
    instruction up without a second copy to drift. **It carries no rules of its own** — a duplicated
    instruction file is how the two disagree six months from now.

  - [x] **Step 5: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; test count 140,
    unchanged. Confirm `CLAUDE.md` restates nothing `AGENTS.md` says.

---

- [x] 9. Print created paths relative to the working directory

**Build:** green
**Files:** `internal/cmd/cmd.go`, `internal/cmd/init.go`, `internal/cmd/new.go`,
`internal/cmd/cmd_test.go`, `internal/cmd/init_test.go`, `internal/cmd/new_test.go`, `README.md`,
`docs/example.md`
**Tests:** `TestDisplayPath`
**Regression:** reverting this commit makes `init` and `new` print absolute paths again, so every
command output in `README.md` and `docs/example.md` becomes false — the exact mismatch that produced
this task.
**Baseline:** before=140 after=141
**Commit:** `fix(cmd): print created paths relative to the working directory`

`init` and `new` print `created <path>` using the tree's own absolute root, so a reader sees
`created /var/folders/.../spectre/specs`. The README has claimed the relative form
(`created spectre/changes/my-change`) since before this change — **the docs and the binary have
disagreed all along**, and writing a walkthrough of real output is what exposed it. The operator
chose to fix the binary.

  - [x] **Step 1: Write `TestDisplayPath` first**, in `internal/cmd/cmd_test.go` beside the other
    tests for that file's shared helpers, table-driven: a path under the working directory
    renders relative (`spectre/specs`); a path outside it renders absolute, unchanged; the working
    directory itself renders as `.`; a path reached through a symlink is not silently rewritten into
    something that resolves elsewhere.

  - [x] **Step 2: Add the helper to `internal/cmd/cmd.go`**, beside `flagSet` and `resolve` — the
    file that already holds what the subcommands share. `displayPath(p string) string` returns
    `filepath.Rel(wd, p)` when that succeeds and the result does not escape upward, and `p`
    otherwise. **Two callers exist the moment it is written** (`init` and `new`), so it is an earned
    helper rather than a speculative one.

  - [x] **Step 3: Never let a failure change what is created.** `os.Getwd` can fail, and
    `filepath.Rel` can too. Either one returns the absolute path unchanged — this is a **display**
    concern, and a display concern must never turn into an error path or alter behaviour. Do not
    add an error return.

  - [x] **Step 4: A path outside the working directory stays absolute.** Use `filepath.IsLocal`
    rather than a hand-rolled `..` prefix test: it is stdlib (Go 1.20+, and this module targets
    1.26.5) and it drops the `strings` import. Verified before adopting: `IsLocal(".")` is `true`,
    so the working-directory-itself case still renders as `.` rather than falling back to the
    absolute path.

    **Its extra strictness is unreachable here, and that is stated rather than claimed as a
    benefit.** `IsLocal` also rejects an absolute or empty string, which a `..` prefix test lets
    through — but `filepath.Rel` returns a non-nil error for every input that would produce either,
    so `displayPath` has already returned by then. Measured, not assumed. The swap is therefore
    worth making for clarity and for the dropped import, **not** for behaviour: no test can
    distinguish the two guards through `displayPath`, and none should be written pretending to.

  - [x] **Step 4b: The guard's original hand-rolled form.** `--root ../elsewhere`
    would otherwise render as `../../elsewhere/spectre/specs`, which is harder to read than the
    absolute path and easy to misread as a path under the current directory. Relative only when the
    result does not start with `..`.

  - [x] **Step 5: Use it at both print sites** — `init`'s `created <path>` lines and `new`'s single
    one — and update the existing assertions in `internal/cmd/init_test.go` and
    `internal/cmd/new_test.go` that expect absolute paths. **Do not weaken an assertion into a
    substring match** to accept either form.

    **If a print site turns out to have no exact assertion today, add one** rather than concluding
    there is nothing to update: an output format this task exists to fix should not be able to
    regress silently afterwards.

  - [x] **Step 6: Verify against the docs this task exists to make true.** Run, from inside a
    scratch git repository:

    ```bash unverified:confirm the exact strings once the helper lands
    spectre init          # must print: created spectre/specs, created spectre/changes, created spectre/config.md
    spectre new demo      # must print: created spectre/changes/demo
    ```

    Then re-read `docs/example.md`'s and `README.md`'s `init`/`new` output blocks and make them
    match byte for byte. **Both need editing, in opposite directions**, and both are in this task's
    `**Files:**` for that reason: `docs/example.md` already shows the relative form this task makes
    true, so it needs checking rather than changing; `README.md` shows the real absolute paths its
    author captured (`created /tmp/myapp/spectre/specs`), which this task makes false, so it must be
    rewritten to the relative form. Re-run both pages' commands after the helper lands and paste
    what you actually saw.

---

- [x] 10. Fix what the whole-change Go audit found

**Build:** green
**Files:** `internal/cmd/new.go`, `internal/cmd/new_test.go`, `internal/tree/tree.go`,
`internal/check/check_test.go`
**Tests:** `TestNewRejectsControlCharacterID`
**Regression:** reverting this commit lets `spectre new` accept an id containing a newline, which
makes that change's own required title unsatisfiable and breaks `validate`'s `file:line: message`
output contract; and it removes the case that pins the missing-heading guarantee.
**Baseline:** before=141 after=142
**Commit:** `fix(cmd): reject control characters in a change id`

The three-pass Go audit over the whole change raised four findings. Fixing them together, because
they were found together and none is worth its own commit.

  - [x] **Step 1: `isValidID` accepts control characters — Major.** `spectre new $'weird\nid'`
    succeeds. Because task 2 made the change id a required heading, the resulting change can never
    validate: `headingFindings` scans line by line, so a wanted heading containing a newline matches
    nothing, and `missing "# weird\nid"` is unsatisfiable by any file content. Worse, the finding
    itself spans two lines, breaking the `file:line: message` contract that `validate`'s output
    holds and that anything parsing it relies on.

    **This interaction is new in this change.** `isValidID`'s gap is older, but it was inert until
    the id started reaching `headingFindings`. Reject control characters in `isValidID` — an id that
    cannot be a safe directory name, a heading, *and* a well-formed finding path is not a valid id.
    Write `TestNewRejectsControlCharacterID` first; cover at least a newline, a carriage return and
    a tab.

  - [x] **Step 2: The missing-heading guarantee is not pinned — the one surviving mutant.** Deleting
    the `continue` after `headingFindings` appends its missing-heading finding lets execution fall
    through into the ordering check, and the whole suite still passes.

    It survives for a precise reason, and the fix follows from it: `TestHeadingOrder`'s
    missing-heading case lists the missing heading **first** in `want`, so `prevLine` is still `-1`
    and the comparison behaves identically either way. Add a case with the missing heading **second
    or later** — `want = []string{"## Present", "## Missing"}`, only `## Present` in the file — where
    the fallen-through zero `line` is less than `prevLine` and a spurious ordering finding appears.
    Add it as a **row in the existing table**, not a new test function, so the count moves by one
    only for step 1's test.

    **Prove it:** delete the `continue`, run the suite, watch your new case fail; restore, watch it
    pass.

  - [x] **Step 3: `tree.go`'s constant doc comment contradicts itself — Minor.** It says
    `ProposalFile, TasksFile and DesignFile are the fixed filenames every change carries`, then four
    lines later that `a change need not carry a DesignFile`. The first sentence was not adjusted when
    `DesignFile` joined the group. Rewrite so the group's shared purpose and `DesignFile`'s
    optionality are stated once, without contradiction.

  - [x] **Step 4: `New`'s doc comment is stale — Minor.** It reads `New scaffolds changes/<id>/ with
    proposal and task templates`; `New` has written three files since task 4.

  - [x] **Step 5: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; count 142. Then
    re-run the audit's own reproducer and confirm it is closed:

    ```bash unverified:confirm the exact refusal wording once step 1 lands
    spectre new "$(printf 'weird\nid')"   # must be refused, exit 2, and create nothing
    ```

---

- [x] 11. Pin the walkthrough's scaffold output to the real templates

**Build:** green
**Files:** `internal/cmd/new_test.go`, `docs/example.md`
**Tests:** `TestExampleDocMatchesScaffold`
**Regression:** reverting this commit lets `_designTemplate` or `_proposalTemplate` be edited while
`docs/example.md` keeps showing the old body, with every check still green — measured, not assumed.
**Baseline:** before=142 after=143
**Commit:** `test(cmd): pin the example walkthrough to the real scaffold templates`

The review panel's principles slot raised this (F1): the `design.md` scaffold body exists twice —
compiled into `internal/cmd/new.go` as `_designTemplate`, and quoted in `docs/example.md` — with
nothing binding the copies. The reviewer demonstrated it by editing the template and watching
`gofmt`, `go vet` and `go test` all stay green while the page went stale.

  - [x] **Step 1: Write `TestExampleDocMatchesScaffold` first**, and watch it fail before writing
    any extraction helper — mutate the template first so you are certain the test can fail for the
    right reason.

  - [x] **Step 2: Read the page from disk, at a path relative to the test.** `docs/example.md` sits
    two directories above `internal/cmd`. Resolve it explicitly rather than with a working-directory
    assumption, and **fail loudly if the file is missing** — a test that silently skips when it
    cannot find its subject is worse than no test, because it reports success forever.

  - [x] **Step 3: Extract the scaffolded bodies the page shows**, for `proposal.md`, `tasks.md` and
    `design.md`, and compare each against what `cmd.New` actually writes — call `New` into a
    `t.TempDir()` and read the three files back, rather than comparing against the template
    constants directly. Comparing to the constants would pass even if `New` stopped using them.

  - [x] **Step 4: Compare on exact bytes**, not a substring or a normalised form. A loose comparison
    is what let this drift in the first place.

  - [x] **Step 5: Prove it both ways.** Change one character of `_designTemplate` → the test FAILS;
    restore → it PASSES. Paste both.

  - [x] **Step 6: Expect the test to find real drift, and fix the page rather than the test.**
    Writing it surfaced one: `docs/example.md` showed the `tasks.md` scaffold without the trailing
    blank line `render.Tasks` actually emits. `docs/example.md` is in this task's `**Files:**` for
    that reason — the fix belongs in the same commit as the test that caught it, or the commit ships
    a test that fails against its own tree.

  - [x] **Step 7: Say what this does and does not close.** It pins the *scaffold* section of
    `docs/example.md`. The page's *filled-in* bodies — the spec, and the completed proposal, tasks
    and design — are still unpinned; the open question `example-bodies-untested` stays open and
    must be narrowed rather than deleted, to say that the scaffold half is now covered.

---

- [x] 12. Make the walkthrough prompt-first

**Build:** green
**Files:** `docs/example.md`, `internal/cmd/new_test.go`
**Tests:** none added — this task writes no Go. The existing scaffold guard added by task 11 must
keep passing, and must still fail under a substring mutation of the design template; step 6 proves
both.
**Regression:** reverting this commit restores the terminal transcripts the operator asked to
remove, and re-couples the page to a command-driven reading it is meant to drop.
**Baseline:** before=143 after=143
**Commit:** `docs(example): drive the walkthrough with prompts, not commands`

Per `docs-are-prompt-first` in `design.md`. The page stays the end-to-end read and keeps **every
generated file body in full** — those are the output the operator asked for. What goes is the
`spectre` terminal transcripts.

  - [x] **Step 1: Each step becomes prompt → resulting files.** The prompt tells the agent what to
    do, including running the `spectre` command where one is needed; the files it produced follow, in
    full, then the short explanation of what to check by hand. The reader types nothing.

  - [x] **Step 2: The prompts must now instruct the agent to run the commands**, because nobody else
    will. `spectre init`, `spectre new <id>`, `spectre validate`, `spectre archive <id>` each move
    into the prompt of the step that needs them. A prompt that produces a file but never tells the
    agent to create the change is a prompt that cannot work.

  - [x] **Step 3: `TestExampleDocMatchesScaffold` hardcodes the section heading
    `## 2. Scaffold the change`** (`internal/cmd/new_test.go:282`). Restructuring the page **will**
    break it — loudly, which is the design. Either keep that exact heading, or update the test's
    constant to the new one. **Do not weaken the test to tolerate either heading**, and do not
    delete it: it is the guard task 11 added for panel finding F1.

  - [x] **Step 4: Behaviour that was shown as output is now stated in prose where it still matters.**
    Most concretely, a freshly scaffolded change reports `no tasks` until the plan is written
    (`scaffold-reports-no-tasks`). With no `validate` transcript on the page, say it in a sentence
    rather than dropping it — it is the tool telling the reader what to do next.

  - [x] **Step 5: Do not invent output.** Where the page still needs to state what a command does,
    state it; do not write a fabricated transcript. Every remaining factual claim about behaviour
    must be one you verified against the built binary.

  - [x] **Step 6: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; count 143.
    `TestExampleDocMatchesScaffold` must still pass, and must still **fail** when
    `_designTemplate` is mutated by substring — prove both.

---

- [x] 13. Make the README quick start prompt-first

**Build:** green
**Files:** `README.md`
**Tests:** none — documentation only.
**Regression:** reverting this commit restores the numbered terminal walk the operator rejected.
**Baseline:** before=143 after=143
**Commit:** `docs(readme): drive the quick start with prompts, not commands`

  - [x] **Step 1: Replace the numbered command walk with a short prompt sequence.** The reader hands
    each prompt to an agent; the agent runs `spectre` itself. No "run this command" steps, and no
    captured command output.

  - [x] **Step 2: Keep the command reference table.** `docs-are-prompt-first` states why: the
    commands are the tool's real surface, an agent has to know they exist, and dropping them
    entirely was considered and rejected. It is a reference table, not a walk.

  - [x] **Step 3: The two absorbed facts must still survive** — `archive` needs the tree inside a
    git repository with the change's files tracked, and `--root` must precede any positional
    argument. They are stated nowhere else in the repository. Losing them here is the same failure
    task 6 was warned about; carry them as prose or fold them into the prompts that need them.

  - [x] **Step 4: Show, briefly, the ordinary commands that go between spectre executions.**
    spectre commands are the punctuation, not the work. A guide showing only them implies the gaps
    between are empty. They are not, and **most of what fills them is another prompt**:

    - after `validate` passes — *"implement task 1"*, then the project's own build and tests, then
      *"tick task 1's box"*, repeated per task. This is the bulk of a real change.
    - alongside that — committing as the work proceeds.
    - before `archive` — `git add spectre`, because `archive` moves the change with `git mv` and
      refuses on untracked files.

    **Short is the requirement, not a nicety** — a line or two per gap, not a git tutorial and not a
    development-workflow essay. And since the docs are prompt-first, an in-between step that is the
    agent's to do is written as a prompt like any other, not as a command addressed to the reader.

  - [x] **Step 5: The quick start is SHORT and defers to the walkthrough.** This is the operator's
    explicit instruction and it overrides any inclination to make the section self-sufficient: keep
    it to the shortest prompt sequence that gets a reader from nothing to a validated, archived
    change, and link to `docs/example.md` for the full example — the complete prompts, every
    generated file body, the whole cycle.

    **Do not retell the walkthrough.** Where a step needs a long prompt, give the short form and
    link; the walkthrough is the long read. The prior `quick-start-length-accepted` decision is
    **superseded** by `quick-start-is-short` in `design.md` — 123 lines is no longer the target to
    beat, it is the thing being removed. If you land near it, you have not done this step.

  - [x] **Step 6: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean (no Go touched);
    count 143. Confirm nothing in the README now contradicts `docs/example.md` or the three
    `docs/` reference pages, and that every link still resolves.

---

- [x] 14. Show the in-between commands in the walkthrough too

**Build:** green
**Files:** `docs/example.md`
**Tests:** none added — documentation only. The existing scaffold guard must keep passing.
**Regression:** reverting this commit leaves the walkthrough implying that nothing happens between
spectre commands, and in particular drops the `git add spectre` that `archive` requires.
**Baseline:** before=143 after=143
**Commit:** `docs(example): show the ordinary commands between spectre steps`

Task 13 does this for the README; this does it for the walkthrough, so the two agree.

**Most of it is already there — this task is small, and inflating it would be wrong.** Step 8
already carries the implement-then-tick loop and the checkbox diff; step 9's prompt already tells
the agent to `git add spectre` before archiving. Checked at `d4dc331`, not assumed. What is missing
is only the two beats between implementing and ticking: running the project's own build and tests,
and committing as the work proceeds.

  - [x] **Step 1: Add the two missing beats to step 8's loop.** Its prompt currently says implement
    task 1 then check its box. Between those, a real change runs the project's own build and tests,
    and commits. Fold both into that step's prompt and say in one line that this loop — not any
    single `spectre` command — is the bulk of the work.

    **Do not restructure step 8, and do not touch steps 1–7 or 9.** They were reviewed and passed at
    `3456625`; step 9's prompt already stages with `git add spectre`. This is an addition of two
    beats, not a rewrite.

  - [x] **Step 2: Write them as prompts, because the page is prompt-first.** The reader types
    nothing. *"Implement task 1, then run the tests, then tick its box"* is a prompt; `git add
    spectre` belongs inside the archive step's prompt. Neither is a code block addressed to the
    reader.

  - [x] **Step 3: End to end means every stage present, and no application code.** The page must
    run empty directory → tree → change → spec → plan → validate → implement/test/tick/commit →
    archive, with no stage glossed. It shows **no channel implementation code**: the loop is prompts
    only, because the example is deliberately language- and vendor-neutral and real code would force
    it to pick a framework. Confirmed with the operator, not assumed.

  - [x] **Step 4: Keep it short.** One or two lines per gap. This is orientation, not a git guide,
    and the page is already 306 lines.

  - [x] **Step 5: Do not reintroduce terminal transcripts.** No captured output, no `$` prompts. That
    is what task 12 removed at the operator's request.

  - [x] **Step 6: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; count 143;
    `TestExampleDocMatchesScaffold` still passing. Confirm the walkthrough and the README describe
    the same in-between commands — two documents disagreeing about what to run is worse than
    neither mentioning it.

---

- [x] 15. Restore the two facts the cut lost, and make step 3 a real prompt

**Build:** green
**Files:** `README.md`
**Tests:** none — documentation only.
**Regression:** reverting this commit loses two facts that then exist nowhere in the documentation
— only in a runtime error message — and restores a numbered "prompt" a reader cannot paste.
**Baseline:** before=143 after=143
**Commit:** `docs(readme): restore the facts lost in the cut`

Task 13's review found two genuine losses in going 213 → 123 lines, and one usability wrinkle it
judged borderline. All three are fixed here. **The quick start stays short** — each fix is a clause
or a line, not a section. If this pushes it past ~40 lines, something is being restated rather than
restored.

  - [x] **Step 1: `spectre new` refuses to run until the tree exists.** It will not create one for
    you; it only scaffolds changes inside one. This is in no documentation file at all now — it
    survives only as the binary's own error. Restore it where step 2 mentions `new`, in a clause.

  - [x] **Step 2: Every command searches upwards for a directory *literally named* `spectre`.** The
    README currently says `--root` "names the tree explicitly instead of searching for it" without
    ever saying what is searched for, which makes `--root`'s own paragraph incomplete. Add the
    missing half there — it is one clause and it belongs exactly where the sentence already gestures
    at searching.

  - [x] **Step 3: Make step 3 a prompt a reader can actually paste.** The section says "one prompt
    per step", and steps 1, 2, 4, 5 and 6 each are one; step 3 is a meta-pointer telling the reader
    to go read four walkthrough sections and copy four prompts. **A reader following the section
    literally has nothing to hand their agent at step 3.**

    Write a single prompt that names the four files, says each must follow the template its heading
    row in `## File templates` gives, and links to the walkthrough for the worked bodies. It does
    not need to reproduce the four detailed prompts — it needs to be one thing a reader can paste
    and get a sensible result from.

  - [x] **Step 4: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; count 143. Report
    the quick start's new line count. Confirm nothing added contradicts `docs/example.md`, and that
    the two restored facts are true of the binary — run them, do not assume.

---

- [x] 16. Bind the documented heading order to the checker

**Build:** green
**Files:** `internal/check/check.go`, `internal/check/check_test.go`
**Allowed-collateral:** `README.md`, `docs/example.md`
**Tests:** `TestDocumentedHeadingsMatchChecker`
**Regression:** reverting this commit lets a `want` literal in `internal/check` change while the
README's File templates table and the walkthrough's prompts keep stating the old order, with every
check green — the exact drift task 11's guard prevents for the scaffold bodies.
**Baseline:** before=143 after=144
**Commit:** `test(check): bind the documented heading order to the checker`

Panel finding F4. The required-heading-order knowledge is stated three ways with nothing binding
it: the `want` literals in `SpecFindings`, `ProposalFindings`, `TaskFindings` and `DesignFindings`;
the README's `## File templates` table; and `docs/example.md`'s prompts ("It must have exactly these
headings, in this order: …"). Change a literal and both documents silently go false.

  - [x] **Step 1: Make the heading lists readable from outside the functions.** They are currently
    inline `want := []string{…}` literals inside four functions, so no test can compare against them
    without duplicating them a fourth time. Lift each into a package-level function or var in
    `internal/check` — parameterised where it must be (`# <change-id>`, `# <capability>`) — and have
    the four `*Findings` functions use it. **This is the point of the task**: it creates the single
    source the docs can be checked against. A test that restates the lists has changed nothing.

  - [x] **Step 2: Write `TestDocumentedHeadingsMatchChecker` first**, and watch it fail before
    step 1 lands — mutate a `want` literal and confirm your test would have caught it.

  - [x] **Step 3: Read both documents from disk** and assert each states exactly what the checker
    requires, in order. Resolve the paths explicitly (`runtime.Caller`, as
    `internal/cmd/new_test.go` already does) and **fail loudly if either file is missing** — a test
    that skips when it cannot find its subject reports success forever.

  - [x] **Step 4: You may adjust the documents' wording so it is machine-checkable**, and they are
    in `**Allowed-collateral:**` for that reason — but **only the wording, never the facts**. If a
    document currently states the order correctly, the test adapts to it, not the reverse. Say
    exactly what you changed and why.

  - [x] **Step 5: Prove it.** Change one `want` entry — reorder two headings, and separately add a
    fourth — and confirm the test fails each time, naming which document disagrees. Restore after
    each. Paste all of it.

  - [x] **Step 6: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; count 144;
    `TestExampleDocMatchesScaffold` still passing — this task must not disturb task 11's guard.

---

- [x] 17. Ship `/spectre` and `/spectre-new` with the repository

**Build:** green
**Files:** `.claude/skills/spectre/SKILL.md`, `.claude/skills/spectre-new/SKILL.md`, `README.md`
**Tests:** none — skill definitions are markdown, not Go. Correctness is shown by invoking each
command against the real binary, in step 5.
**Regression:** reverting this commit removes both slash commands, and anyone cloning spectre loses
them — which is the point of shipping them here rather than in the agent tooling.
**Baseline:** before=144 after=144
**Commit:** `feat(skills): ship /spectre and /spectre-new with the repository`

Per `slash-commands-ship-with-spectre` and `slash-commands-are-thin` in `design.md`. `.claude/` is
not in `.gitignore` — checked, not assumed — so these commit normally.

  - [x] **Step 1: Write `.claude/skills/spectre/SKILL.md`.** It dispatches: `/spectre <subcommand>
    [args]` runs that spectre subcommand, shows its **real** output, and names what usually comes
    next. Cover every subcommand the binary has — `init`, `new`, `list`, `validate`, `refs`,
    `archive` — by dispatching, not by enumerating behaviour it would then have to keep in step.

    Follow the frontmatter shape the existing skills use (`name`, `description` ending in "Use for
    /spectre.", `allowed-tools` scoped to what it actually needs, `license`, `compatibility`,
    `metadata`). Read `~/Projects/agents/skills/flow-status/SKILL.md` for the house style.

  - [x] **Step 2: Write `.claude/skills/spectre-new/SKILL.md`.** It scaffolds with `spectre new
    <change-id>` and then helps fill the three generated files, following the templates
    `## File templates` in the README states and the prompts `docs/example.md` carries. It links to
    those rather than restating them — a third copy of the heading order is exactly what task 16's
    guard exists to prevent.

  - [x] **Step 3: Thin, plus guidance — not automation.** Both commands run what was asked, show the
    real output, and name the likely next step. **Neither takes that next step unasked.**
    `/spectre-new` is the single exception, because `new` alone leaves three stubs and a change that
    reports `no tasks` until a plan exists.

  - [x] **Step 4: Mention them in the README**, briefly, where a reader would look for them — one or
    two lines. The quick start is 35 lines and deliberately short; do not restate what the skills
    do, and do not lengthen it materially.

  - [x] **Step 5: Invoke both against the real binary and paste what happened.** Not a description of
    what they would do — run them. At minimum: `/spectre init` and `/spectre validate` in a scratch
    tree, and `/spectre-new` end to end producing a change that validates clean. If a command cannot
    be invoked in this environment, say so plainly rather than claiming it works.

  - [x] **Step 6: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean and count still 144 —
    no Go is touched, so any movement means something unintended happened. Both guards still pass.
    Confirm `git status` shows `.claude/` as a normal untracked addition, not ignored.

---

- [x] 18. Add a terminal guide beside the prompt-first one

**Build:** green
**Files:** `docs/terminal.md`, `README.md`
**Tests:** none added — documentation only. Correctness is shown by running every command the guide
shows and pasting the real output, in step 5.
**Regression:** reverting this commit removes the only route for a reader who does not drive spectre
through an agent — the position the docs were in before this task, and the reason it exists.
**Baseline:** before=144 after=144
**Commit:** `docs(terminal): add a command-driven guide beside the prompt-first one`

Per `terminal-guide-alongside` in `design.md`. It does **not** supersede `docs-are-prompt-first`:
the README still leads with the prompt-first path, and `docs/example.md` is unchanged.

  - [x] **Step 1: Mirror the walkthrough's nine steps, in the same order**, with the same subject —
    the notifications change, change id `multi-channel-notifications`. A reader moving between the
    two documents should recognise the same journey.

  - [x] **Step 2: Commands and their real output, in place of prompts.** Every command shown must be
    one you ran, with the output you actually saw. This is the document the prompt-first rewrite
    removed; it is being restored deliberately, so it has to be accurate.

  - [x] **Step 3: Do NOT repeat the generated file bodies.** They live in `docs/example.md` and are
    pinned to `cmd.New`'s real output by `TestExampleDocMatchesScaffold`. A second copy would sit
    **outside** that guard and drift silently — the exact defect the panel raised twice in this
    change. Name what each step produces and link to the walkthrough for the contents.

  - [x] **Step 4: Carry the facts a terminal reader needs and an agent reader did not.** The
    `--root`-before-positional rule; that `archive` needs the tree in a git repository with the
    change's files tracked; that a freshly scaffolded change reports `no tasks` until the plan
    exists; and the implement/build-and-test/tick/commit loop between `new` and `archive`.

  - [x] **Step 5: Run every command the guide shows**, in a scratch git repository, against a binary
    built from this worktree, and paste the transcript. Include the two refusals — `archive` with
    unchecked tasks, and `archive` before `git add` — since a terminal reader will hit both.

  - [x] **Step 6: Point the README at both guides**, in one or two lines. It still leads with the
    prompt-first path; the terminal guide is the alternative, not the default. The quick start is 35
    lines and must not grow materially.

  - [x] **Step 7: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` clean; count 144; both
    guards still pass. Confirm `docs/example.md` is untouched, and that no file body appears in both
    guides.

