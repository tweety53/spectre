# spectre

Spec-driven change tracking in markdown. One binary, no dependencies, no database: the tree on
disk is the entire state.

## Install

```bash
go install github.com/tweety53/spectre/cmd/spectre@latest
```

Or build locally: `go build -o bin/spectre ./cmd/spectre`.

## Getting started

A spectre tree lives inside a git repository: `archive` moves a finished change with `git mv`, so
both the repository and the change's files have to exist in git before you can archive anything.

```bash
git init                              # if this isn't already a git repository
spectre init
```

`spectre init` creates a plain directory named `spectre` holding `specs/`, `changes/` and a
`config.md` of commented defaults; it fills in whatever's missing and never overwrites what's
already there.

`spectre new` refuses to run until that directory exists — it will not create the tree for you,
only scaffold changes inside one:

```bash
spectre new my-change
# created spectre/changes/my-change
```

Stage what `new` wrote — and every file you add or edit inside the tree afterwards — so `archive`
has something to move later:

```bash
git add spectre
```

Every command searches upwards from the working directory for a directory literally named
`spectre`. `--root <path>` names one explicitly instead of searching, and — because Go's `flag`
package stops parsing flags at the first positional argument — `--root` must come **before** any
positional argument:

```bash
spectre validate --root . my-change    # works
spectre validate my-change --root .    # fails: --root is read as a second positional argument
```

## The tree

```
spectre/
  peers                       # "<name> <relative-path>" per neighbour tree
  config.md                   # optional; rules, vocabulary and layout for this tree
  specs/<capability>.md       # long-lived capability specs
  changes/<id>/proposal.md    # why, and what changes
             /tasks.md        # "- [ ] 1. ..." — the only progress signal
             /design.md       # optional, unparsed
  changes/archive/<id>/
```

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
