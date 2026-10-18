## Context

`internal/cmd/archive.go` applies four refusals in order: no `tasks.md`, a `tasks.md` carrying no
tasks, unchecked tasks, and a destination that already exists. The first three are content refusals
`--force` overrides; the fourth is not.

The satellite predicate already exists, but as a local expression inside `internal/check`'s
`Structural` — the thing that lets `validate` skip proposal/tasks/design heading checks for a
pointer tree. It is not reachable from `internal/cmd`.

`internal/cmd` already imports `internal/check` (`validate.go`), so consulting it from `archive.go`
adds no dependency edge and no import cycle.

## Decisions

### Extract the predicate rather than copy it

**ID:** extract-is-satellite
**Status:** active
**Chosen:** Export `check.IsSatellite(dir string) bool` carrying exactly the expression `Structural`
inlines today, and have `Structural` call it. One definition, two callers.
**Considered:** Re-deriving the same condition inside `archive.go`. Rejected because the two would
drift the first time the definition of "and nothing else" changed — and it has already changed once,
when review widened it from proposal/tasks to include `design.md`. A second copy would have kept the
old meaning silently.

### A satellite skips the plan refusals, and only those

**ID:** satellite-skips-plan-refusals
**Status:** active
**Chosen:** For a satellite, `archive` skips the missing-`tasks.md`, no-tasks and unchecked-tasks
refusals. The destination-exists refusal, the git-tracking requirement and every exit code stay
exactly as they are.
**Considered:** Skipping every refusal for a satellite, which would have been simpler to write.
Rejected because the destination-exists refusal protects against overwriting an already-archived
change, and that hazard is identical for a satellite — it has nothing to do with whether the change
carries a plan.

### `--force` keeps its current scope

**ID:** force-unchanged
**Status:** active
**Chosen:** `--force` still overrides the same three content refusals for a canonical change, and
gains no new power. A satellite simply stops needing it.
**Considered:** Making `/flow` pass `--force` for a satellite instead of changing `archive`.
Rejected as the wrong shape: `--force` also suppresses genuine refusals such as unchecked tasks, so
using it to route around a false positive would hide real problems in every future run. The operator
rejected it explicitly when the blocker was first raised.

### Archive does not police link skew

**ID:** no-skew-check-in-archive
**Status:** active
**Chosen:** `archive` does not check whether the counterpart change is archived. It archives what it
was asked to archive.
**Considered:** Refusing a satellite whose canonical is still open, so the two stay in step.
Rejected because KAN-363's `independent-archive` decision already settles the opposite — each tree
archives on its own timeline, links resolve into `changes/archive/`, and skew is reported by
`validate` as a finding rather than blocking anything.

## Open questions

None.
