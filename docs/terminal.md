# Terminal guide: multi-channel-notifications

The same journey as [docs/example.md](example.md), for a reader who would rather type commands
than hand prompts to an agent. Same subject — a backend that sends every notification by email and
wants to add SMS, messenger and push — same change id, `multi-channel-notifications`, same nine
steps in the same order. Each step here shows the command and the real output it produced, not a
prompt; file bodies are not repeated here — they are shown in full, and pinned to `cmd.New`'s real
output by a test, in `docs/example.md`. Where a step writes a file, this guide names what the file
contains and links to the matching step there.

Every transcript below is a real run, in a scratch git repository, of a `spectre` binary built from
this repository: `go build -o bin/spectre ./cmd/spectre`. The commands below assume that binary is
on `PATH` as `spectre`; substitute `./bin/spectre` if it isn't.

**One flag-order rule applies to every command below:** `--root <path>` must come *before* any
positional argument, e.g. `spectre validate --root . my-change`, not after — Go's `flag` package
stops parsing flags at the first positional argument.

## 1. Create the tree

```
$ git init
$ spectre init
created spectre/specs
created spectre/changes
created spectre/config.md
```

Same result as [step 1](example.md#1-create-the-tree): `spectre/config.md`, `spectre/specs/` and
`spectre/changes/`. `spectre init` never overwrites what already exists, so re-running it is safe.

## 2. Scaffold the change

```
$ spectre new multi-channel-notifications
created spectre/changes/multi-channel-notifications
```

This writes `proposal.md`, `tasks.md` and `design.md` under
`spectre/changes/multi-channel-notifications/`, each a stub — every heading present, no content yet.
See [step 2](example.md#2-scaffold-the-change) for the three bodies verbatim.

A `spectre validate` run here (see step 7 below) reports one finding, `no tasks`, until `tasks.md`
has at least one task — expected, not a fault; it's the tool naming what to do next.

## 3. Write the capability spec

Write `spectre/specs/notifications.md` by hand, or however you write files, following the required
headings row for `specs/<capability>.md` in the [README's File templates](../README.md#file-templates)
table. The full body this guide's run produced is in
[step 3](example.md#3-write-the-capability-spec).

## 4. Fill in the proposal

Write `spectre/changes/multi-channel-notifications/proposal.md`, following the `proposal.md` row in
[File templates](../README.md#file-templates). Body: [step 4](example.md#4-fill-in-the-proposal).

## 5. Fill in the tasks

Write `spectre/changes/multi-channel-notifications/tasks.md`, following the `tasks.md` row in
[File templates](../README.md#file-templates) — numbered, gap-free, dependency-ordered tasks. Body:
[step 5](example.md#5-fill-in-the-tasks).

## 6. Fill in the design

Write `spectre/changes/multi-channel-notifications/design.md`, following the `design.md` row in
[File templates](../README.md#file-templates). Body: [step 6](example.md#6-fill-in-the-design).

## 7. Validate the finished change

```
$ spectre validate
no findings
```

Exit code 0. With every file above satisfying its template, `validate` reports nothing — the
`no tasks` finding from step 2 is gone now that `tasks.md` has tasks.

## 8. Work the tasks

Between `new` and `archive` sits the real work, repeated per task: implement what the task
describes, run the project's own build and tests, tick the task's box, commit.

```diff
-- [ ] 1. Add a channel abstraction behind the existing email sender
+- [x] 1. Add a channel abstraction behind the existing email sender
```

`validate` doesn't care whether a box is checked — a task-sequence finding only fires on a missing,
duplicate or malformed task number. The checkbox exists for `archive`, next.

## 9. Archive the finished change

`archive` moves a change with `git mv`, so it needs the tree inside a git repository with the
change's files already `git add`ed. It also refuses while any task is unchecked. Both refusals,
hit in order:

```
$ spectre archive multi-channel-notifications
multi-channel-notifications: 5 of 5 tasks are unchecked (use --force to archive anyway)
```

Exit code 1. After checking every task's box:

```
$ spectre archive multi-channel-notifications
archive requires the tree to be inside a git repository with multi-channel-notifications's files already tracked (git add); git mv failed: exit status 128
fatal: source directory is empty, source=spectre/changes/multi-channel-notifications, destination=spectre/changes/archive/multi-channel-notifications
```

Exit code 2. Once the files are staged, `archive` succeeds:

```
$ git add spectre
$ spectre archive multi-channel-notifications
archived multi-channel-notifications
```

Exit code 0. `spectre/changes/multi-channel-notifications/` becomes
`spectre/changes/archive/multi-channel-notifications/`, moved with `git mv`, staged but not
committed — committing it is the last step, same as any other change to the tree.
