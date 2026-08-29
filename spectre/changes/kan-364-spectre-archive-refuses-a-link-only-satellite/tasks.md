# kan-364-spectre-archive-refuses-a-link-only-satellite

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Two tasks, both in `spectre`. Task 1 extracts the satellite predicate with no behaviour change;
task 2 is the behaviour change that consumes it. They are split because a reviewer can reject either
one while approving the other: the first is a pure refactor whose whole claim is that nothing moved,
the second changes what `archive` refuses.

`design.md` is canonical for the decisions; `proposal.md` for the refusal this change removes.

**Baseline, measured before any edit:**

- 169 top-level `func Test…` across the module.
  <!-- measured: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l @ 09912dd -->
- `internal/cmd/archive_test.go` carries 11 top-level test functions.
  <!-- measured: grep -c '^func Test' internal/cmd/archive_test.go @ 09912dd -->
- `go test ./...` clean at `09912dd`.
  <!-- measured: go test ./... -count=1 @ 09912dd -->
- `internal/cmd` already imports `internal/check`, so consulting it from `archive.go` adds no
  dependency edge.
  <!-- measured: grep -n 'internal/check' internal/cmd/validate.go @ 09912dd -->

**Global constraints:**

- Standard library only — `go.mod` carries no `require` block and this change adds none.
- Exit codes are a contract: `0` success, `1` findings or a content refusal, `2` usage or IO error.
  This change moves no exit code.
- `gofmt -l .`, `go vet ./...` and `go test ./...` must all be clean before a task is done.
- Per KAN-197, every assertion added here carries a mutation test: prove it fails when the behaviour
  it checks is broken.

---

- [x] 1. `check.IsSatellite` — extract the predicate, change nothing

**Repository:** `/Users/tweety53/Projects/spectre`

`internal/check/check.go`'s `Structural` inlines the satellite test:

```go verified:read from internal/check/check.go at 09912dd
satellite := hasLink && !fileExists(filepath.Join(c.Dir, tree.ProposalFile)) &&
    !fileExists(filepath.Join(c.Dir, tree.TasksFile)) &&
    !fileExists(filepath.Join(c.Dir, tree.DesignFile))
```

Move it to an exported `IsSatellite(dir string) bool` in `internal/check/link.go`, beside the
`LinkFile` constant it depends on, and have `Structural` call it. **Behaviour must not move**:
`IsSatellite` returns true for exactly the directories the expression above returns true for, and
the existing `check` suite must pass untouched — that is this task's whole claim.

Its doc comment states *why* the predicate is what it is, in this package's established style: a
satellite is a pointer tree, "and nothing else" includes `design.md` because a directory carrying
one keeps every check a full scaffold gets, and the predicate is exported so `archive` reads the
same definition rather than a copy that would drift.

Write `TestIsSatellite` **first, RED before GREEN**, table-driven: `link.md` alone → true; `link.md`
plus each of `proposal.md`, `tasks.md`, `design.md` in turn → false; no `link.md` at all → false; an
empty directory → false; a directory that does not exist → false.

**Files:** `internal/check/link.go`, `internal/check/check.go`, `internal/check/link_test.go`
**Tests:** `TestIsSatellite`
**Regression:** reverting this commit returns the predicate to a local expression `internal/cmd`
cannot reach, so task 2 would have to re-derive it and the two copies would drift the first time
"and nothing else" changed — as it already has once, when review widened it to include `design.md`.
**Baseline:** before=169 after=170 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 1 -->
**Commit:** `refactor(check): export the satellite predicate`
**Build:** green

- [x] 2. `archive` skips the plan refusals for a satellite

**Repository:** `/Users/tweety53/Projects/spectre`

`internal/cmd/archive.go` applies four refusals in order. For a change `check.IsSatellite` reports
true, skip the first three — missing `tasks.md`, a `tasks.md` carrying no tasks, and unchecked
tasks. A satellite has no plan of its own to be missing; its tasks live in the canonical repository.

**Skip those three and nothing else**, per `design.md`'s `satellite-skips-plan-refusals`:

| Refusal | Canonical | Satellite |
|---------|-----------|-----------|
| no `tasks.md` | refuses, exit 1 | skipped |
| `tasks.md` with no tasks | refuses, exit 1 | skipped — unreachable, there is no file |
| unchecked tasks | refuses, exit 1 | skipped |
| destination already exists | refuses, exit 1, never `--force`-able | refuses, unchanged |
| tree not a git repo, or change untracked | exit 2 | exit 2, unchanged |

`--force` keeps exactly the scope it has (`force-unchanged`): it still overrides the same three
refusals for a canonical change, and a satellite stops needing it.

Write the cases **first, RED before GREEN**, in `internal/cmd/archive_test.go`: a link-only satellite
archives at exit 0 and lands under `changes/archive/<id>/`; a satellite whose destination already
exists still refuses at exit 1; and — the case that must **not** regress — a canonical change with no
`tasks.md` still refuses at exit 1 with its existing message.

Mutation-prove each: make `IsSatellite` return false unconditionally and confirm the satellite case
fails; make it return true unconditionally and confirm the canonical no-`tasks.md` case fails. A test
that only passes is not evidence — three reviews on KAN-363 caught cases that passed for reasons
unrelated to what they named.

**Files:** `internal/cmd/archive.go`, `internal/cmd/archive_test.go`
**Tests:** `TestArchiveSatellite`, `TestArchiveSatelliteDestinationExists`
**Regression:** reverting this commit restores the refusal that parks every cross-repo change at its
archive phase — `/flow` never passes `--force`, so a satellite could not be archived at all and
`FINISHED` would never be written for a change spanning repositories.
**Baseline:** before=170 after=172 top-level `func Test…`
<!-- predicted: grep -rhoE '^func Test[A-Za-z0-9_]*' internal cmd *_test.go | wc -l after task 2 -->
**Commit:** `fix(cmd): archive a link-only satellite without --force`
**Build:** green
