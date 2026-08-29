# kan-364-spectre-archive-refuses-a-link-only-satellite

KAN-364 · `spectre archive` refuses a link-only satellite change.

## Why

KAN-363 introduced satellite change directories: in a repository that is not the canonical one, a
change carries `spectre/changes/<id>/link.md` and nothing else. Its tasks live in the canonical
repository's `tasks.md`, by design — that is the whole point of the pointer tree.

`spectre archive` reads a missing `tasks.md` as a content refusal:

```text verified:run against a scratch fixture holding only a valid link.md, at 09912dd
$ spectre archive sat-demo
sat-demo: no tasks.md (use --force to archive anyway)
```

`/flow` never passes `--force`, so a cross-repo change reaches its archive phase and stops. The
canonical repository archives normally and the satellite cannot; cleanup verification then reports
`LEFTOVER` for that repository and `FINISHED` is never written.

This is not hypothetical: KAN-363 is parked at `IN_PROGRESS` on exactly this refusal, with both its
pull requests already merged.

## What changes

- **`check.IsSatellite(dir)`**, exported — the predicate `Structural` currently inlines: a change
  directory holding `link.md` and no `proposal.md`, `tasks.md` or `design.md`. `Structural` calls it
  instead of carrying its own copy, so the two callers cannot drift.
- **`archive` skips the three plan-related refusals for a satellite** — missing `tasks.md`,
  `tasks.md` with no tasks, and unchecked tasks. A satellite has no plan of its own to be missing.
- **Nothing else moves.** A canonical change missing its `tasks.md` refuses exactly as before, the
  destination-exists refusal stays un-overridable, and `--force` keeps the same scope it has today.

`archive` deliberately does **not** check whether the canonical side is archived first — KAN-363's
`independent-archive` decision settles that each tree archives on its own timeline, and skew is a
`validate` finding rather than an archive refusal.
