# spectre

Spec-driven change tracking in markdown. One binary, no dependencies, no database: the tree on
disk is the entire state.

## Quick start

spectre is meant to be driven by an AI coding agent, not typed at by hand: hand it each prompt
below and it runs the `spectre` commands itself. This is the short path — the full walkthrough,
with every prompt and every generated file body in full, is [docs/example.md](docs/example.md).
Claude Code users can instead run `/spectre` and `/spectre-new`, shipped in this repository's
[.claude/skills/](.claude/skills/).

Install first: `go install github.com/tweety53/spectre/cmd/spectre@latest`, or build locally with
`go build -o bin/spectre ./cmd/spectre`.

Then, one prompt per step:

1. *Set up a spectre tree in the current directory* (run `git init` first if it isn't already a
   git repository, then `spectre init`) — see [step 1](docs/example.md#1-create-the-tree).
2. *Scaffold the change* with `spectre new <change-id>` (`new` refuses to run until the tree from
   step 1 exists — it does not create one for you) — see
   [step 2](docs/example.md#2-scaffold-the-change).
3. *Write the capability spec, then `proposal.md`, `tasks.md` and `design.md`, each following the
   required headings its row gives in [File templates](#file-templates)* — see
   [steps 3–6](docs/example.md#3-write-the-capability-spec) for the worked bodies.
4. *Run `spectre validate` and confirm no findings* — see
   [step 7](docs/example.md#7-validate-the-finished-change).
5. Implement the change, one task at a time: *"implement task 1"*, then the project's own build
   and tests, then *"tick task 1's box"*, then commit — repeated per task. This loop, not any
   single `spectre` command, is the bulk of a real change; see
   [step 8](docs/example.md#8-work-the-tasks).
6. Stage the change (`git add spectre` — `archive` moves it with `git mv`, which needs the tree
   inside a git repository with the change's files already tracked) and *archive it* with `spectre
   archive <change-id>` — see [step 9](docs/example.md#9-archive-the-finished-change).

One flag-order rule applies to every command: without it, every command searches upwards from the
working directory for a directory literally named `spectre`; `--root <path>` names the tree
explicitly instead of searching for it. And because Go's `flag` package stops parsing flags at the
first positional argument, `--root` must come **before** the change id, e.g. `spectre validate
--root . my-change`, not after it.

## The tree

```
spectre/
  peers                       # "<name> <relative-path>" per neighbour tree
  config.md                   # optional; rules, vocabulary and layout for this tree
  specs/<capability>.md       # long-lived capability specs
  changes/<id>/proposal.md    # why, and what changes
             /tasks.md        # "- [ ] 1. ..." — the only progress signal
             /design.md       # context and decisions; validated when present
  changes/archive/<id>/
```

`new` scaffolds all three files under `changes/<id>/` — see [file templates](#file-templates) below
for the headings each one must carry.

## File templates

`validate` checks that each generated file carries its required headings, in order; a file is free
to carry extra headings beyond these.

| File | Required headings, in order |
|---|---|
| `proposal.md` | `# <change-id>`, `## Why`, `## What changes` |
| `tasks.md` | `# <change-id>` |
| `design.md` | `## Context`, `## Decisions` — checked only when the file exists; it stays optional |
| `specs/<capability>.md` | `# <capability>`, `## Purpose`, `## Requirements` — see [spec format](docs/spec-format.md) |

`tasks.md` is also checked for having at least one task: a freshly scaffolded change reports `no
tasks` under `task-sequence` until a task is added.

## Commands

| Command | Flags | What it does |
|---------|-------|--------------|
| `spectre init` | `--root` | creates `specs/`, `changes/` and a `config.md` of commented defaults; fills in what's missing |
| `spectre new <change-id>` | `--root` | scaffolds `changes/<id>/`; exits 1 if it already exists |
| `spectre list` | `--root`, `--specs`, `--json` | open changes with `3/7` progress; `--specs` lists capabilities instead |
| `spectre validate [change-id]` | `--root` | structural checks and reference resolution, whole tree or one change |
| `spectre refs <capability>#<id>` | `--root` | every citation of a requirement, in this tree and each declared peer |
| `spectre archive <change-id>` | `--root`, `--force` | `git mv`s a finished change into `changes/archive/` |

Exit codes are a contract every command holds to: `0` success, `1` findings or a content refusal,
`2` a usage or IO error.

## Spec format

How a capability spec file is written and how requirement ids work — see
[docs/spec-format.md](docs/spec-format.md).

## References across trees

The three citation forms — same file, same tree, peer tree — and how `validate` and `refs` use
them — see [docs/references.md](docs/references.md).

## Per-repository configuration

What `spectre/config.md` can override, and what stays fixed — see
[docs/configuration.md](docs/configuration.md).

## `archive`

```bash
spectre archive <change-id>
```

`git mv`s a finished change into `changes/archive/`. It does not commit. This means the tree has to
be inside a git repository with the change's files already tracked (`git add`) — if either isn't
true, `git mv`'s own error is prefixed with that requirement and the command exits 2, since this is
an environment problem rather than one of the content refusals below.

It refuses three content problems, each overridable with `--force`:

- no `tasks.md`
- `tasks.md` with no tasks at all
- one or more unchecked tasks

A fourth refusal — the destination already exists under `changes/archive/` — is **not** overridden
by `--force`; the collision has to be resolved by hand.

## What spectre does not do

No delta specs, no merge-back at archive time: a change edits spec files directly and git computes
the diff. No status field, assignee or timestamps. No network access, daemon or cache. Anything not
derivable from the markdown does not exist.
