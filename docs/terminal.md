# Terminal guide: multi-channel-notifications

[example.md](example.md)'s nine steps as typed commands and their real output, same change. File
bodies live there, pinned by a test to `cmd.New`'s output; each step here links to its body.

- Transcripts are real runs, in a scratch git repository, of `go build -o bin/spectre ./cmd/spectre`, on `PATH` as `spectre`
  (else `./bin/spectre`).
- `--root` goes before any positional argument — see the [README](../README.md#quick-start).

## 1. Create the tree

```
$ git init
$ spectre init
created spectre/specs
created spectre/changes
created spectre/config.md
```

Same result as [step 1](example.md#1-create-the-tree); re-running `init` is safe.

## 2. Scaffold the change

```
$ spectre new multi-channel-notifications
created spectre/changes/multi-channel-notifications
```

- Three stubs, headings only — bodies in [step 2](example.md#2-scaffold-the-change).
- `spectre validate` here reports `no tasks` until `tasks.md` has a task — expected.

## 3. Write the capability spec

Write `spectre/specs/notifications.md` per the `specs/<capability>.md` row of
[File templates](../README.md#file-templates). Body: [step 3](example.md#3-write-the-capability-spec).

## 4. Fill in the proposal

Write `spectre/changes/multi-channel-notifications/proposal.md` per its
[File templates](../README.md#file-templates) row. Body: [step 4](example.md#4-fill-in-the-proposal).

## 5. Fill in the tasks

Write `spectre/changes/multi-channel-notifications/tasks.md` per its
[File templates](../README.md#file-templates) row — numbered, gap-free, dependency-ordered. Body:
[step 5](example.md#5-fill-in-the-tasks).

## 6. Fill in the design

Write `spectre/changes/multi-channel-notifications/design.md` per its
[File templates](../README.md#file-templates) row. Body: [step 6](example.md#6-fill-in-the-design).

## 7. Validate the finished change

```
$ spectre validate
no findings
```

Exit code 0; step 2's `no tasks` finding is gone.

## 8. Work the tasks

Per task: implement it, run the project's build and tests, tick its box, commit.

```diff
-- [ ] 1. Add a channel abstraction behind the existing email sender
+- [x] 1. Add a channel abstraction behind the existing email sender
```

`validate` ignores box state ([example.md step 8](example.md#8-work-the-tasks)); the box exists for
`archive`.

## 9. Archive the finished change

`archive` refuses while any task is unchecked, and needs the change's files `git add`ed for its
`git mv`. Both refusals, hit in order:

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

Exit code 0. The change moves to `spectre/changes/archive/multi-channel-notifications/`, staged,
not committed — committing is the last step.
