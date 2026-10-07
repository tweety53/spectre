# spectre

Spec-driven change tracking in markdown. One binary, no dependencies, no database: the tree on
disk is the entire state.

## Install

- Binary: `go install github.com/tweety53/spectre/cmd/spectre@latest`, or build locally with
  `go build -o bin/spectre ./cmd/spectre`.
- The skills — `spectre`, plus the bundled `brainstorming` and `writing-plans` — install
  separately from the binary. In **Claude Code**, this repository is its own plugin
  marketplace:

  ```
  /plugin marketplace add tweety53/spectre
  /plugin install spectre@spectre
  ```

  A plugin install namespaces the skills: the command becomes `/spectre:spectre <subcommand>
  [args]`. `/plugin marketplace update spectre` picks up later changes.
- **With any other agent**, copy the skill directories into the skills directory that agent reads
  (for Claude Code, `~/.claude/skills/`):

  ```bash
  git clone https://github.com/tweety53/spectre /tmp/spectre
  mkdir -p ~/.claude/skills
  cp -r /tmp/spectre/.claude/skills/spectre /tmp/spectre/.claude/skills/brainstorming /tmp/spectre/.claude/skills/writing-plans ~/.claude/skills/
  ```

- The `spectre` skill dispatches to the binary, so `spectre` must be on `PATH`. Inside a checkout of
  this repository no install is needed — Claude Code reads [.claude/skills/](.claude/skills/)
  directly, as `/spectre` — as does a manual copy.

## Bundled: brainstorming and writing-plans

spectre checks the shape of a change; what fills it decides its quality. Two skills from
[superpowers](https://github.com/obra/superpowers) fill it best, and ship with this plugin —
nothing else to install:

- `brainstorming` settles the design with you, then writes `proposal.md`, `design.md` and the
  capability spec.
- `writing-plans` turns them into a test-first `tasks.md` — files, tests, verify commands and
  commits per task, steps indented under spectre's column-0 tasks.
- The `spectre` skill's `new` runs both.
- They are unmodified superpowers v6.4.2, MIT, © Jesse Vincent; each directory's `UPSTREAM.md`
  names the source commit. A separately installed superpowers works the same.
- Installed, they are ordinary skills available in every project, not only inside `spectre new`;
  beside a separate superpowers install you have two copies of each, and the manual `cp -r` merges
  into an existing `~/.claude/skills/brainstorming` or `writing-plans` rather than replacing it.
- Worked example: [docs/superpowers-example.md](docs/superpowers-example.md). Filling by hand
  instead: [docs/example.md](docs/example.md).

## Quick start

spectre is driven by an AI coding agent: hand it one prompt per step and it runs the commands —
or run the `spectre` skill instead.
Every prompt and file body is in [docs/example.md](docs/example.md); the same journey as typed
commands and real output is [docs/terminal.md](docs/terminal.md).

1. *Set up a spectre tree* (`git init` first if needed, then `spectre init`) —
   [step 1](docs/example.md#1-create-the-tree).
2. *Scaffold the change* with `spectre new <change-id>`; `new` refuses until the tree exists and
   never creates one — [step 2](docs/example.md#2-scaffold-the-change).
3. *Write the capability spec, then `proposal.md`, `tasks.md` and `design.md`*, each with the
   headings [File templates](#file-templates) requires —
   [steps 3–6](docs/example.md#3-write-the-capability-spec).
4. *Run `spectre validate` and confirm no findings* —
   [step 7](docs/example.md#7-validate-the-finished-change).
5. Per task: implement it, run the project's build and tests, tick its box, commit. This loop, not
   any `spectre` command, is the bulk of a change — [step 8](docs/example.md#8-work-the-tasks).
6. `git add spectre`, then `spectre archive <change-id>` —
   [step 9](docs/example.md#9-archive-the-finished-change).

Every command searches upwards from the working directory for a directory literally named
`spectre`; `--root <path>` names the tree instead. Go's `flag` package stops at the first
positional argument, so `--root` must come **before** the change id: `spectre validate --root .
my-change`, not after it.

## The tree

```
spectre/
  peers                       # "<name> <relative-path>" per neighbour tree
  config.md                   # optional; rules, vocabulary and layout for this tree
  specs/<capability>.md       # long-lived capability specs
  changes/<id>/proposal.md    # why, and what changes
             /tasks.md        # "- [ ] 1. ..." — the only progress signal
             /design.md       # context and decisions; validated when present
             /link.md         # a change spanning repositories; alone, marks a satellite
  changes/archive/<id>/
```

`new` scaffolds `proposal.md`, `tasks.md` and `design.md` under `changes/<id>/`.

## File templates

`validate` checks that each file carries its required headings, in order; extra headings are
allowed.

| File | Required headings, in order |
|---|---|
| `proposal.md` | `# <change-id>`, `## Why`, `## What changes` |
| `tasks.md` | `# <change-id>` |
| `design.md` | `## Context`, `## Decisions` — checked only when the file exists; it stays optional |
| `link.md` | no fixed headings; `## Part of`/`## Parts`/`## Branch`/`## Merge order`/`## Tasks here` grammar — see [links across repositories](docs/links.md) |
| `specs/<capability>.md` | `# <capability>`, `## Purpose`, `## Requirements` — see [spec format](docs/spec-format.md) |

`tasks.md` must also hold at least one task: a fresh scaffold reports `no tasks` under
`task-sequence` until one is added.

## Commands

| Command | Flags | What it does |
|---------|-------|--------------|
| `spectre init` | `--root` | creates `specs/`, `changes/` and a `config.md` of commented defaults; fills in what's missing |
| `spectre new <change-id>` | `--root` | scaffolds `changes/<id>/`; exits 1 if it already exists |
| `spectre list` | `--root`, `--specs`, `--json` | open changes with `3/7` progress; `--specs` lists capabilities instead |
| `spectre validate [change-id]` | `--root` | structural checks and reference resolution, whole tree or one change |
| `spectre refs <capability>#<id>` | `--root` | every citation of a requirement, in this tree and each declared peer |
| `spectre archive <change-id>` | `--root`, `--force` | `git mv`s a finished change into `changes/archive/` |
| `spectre link <peer>:<canonical-id>` | `--root`, `--force` | links this tree to a canonical change in a peer tree |

Exit codes, for every command: `0` success, `1` findings or a content refusal, `2` a usage or IO
error.

## `archive`

`spectre archive <change-id>` `git mv`s a finished change into `changes/archive/` and does not
commit.

- The tree must be inside a git repository with the change's files tracked (`git add`); otherwise
  `git mv`'s own error is prefixed with that requirement and the command exits 2 — an environment
  problem, not a content refusal.
- Content refusals, each overridable with `--force`: no `tasks.md`; `tasks.md` with no tasks;
  one or more unchecked tasks.
- The destination already existing under `changes/archive/` is refused and **not** overridden by
  `--force`; resolve the collision by hand.

## Further reading

- [Spec format](docs/spec-format.md) — how a capability spec is written and how requirement ids
  work.
- [References across trees](docs/references.md) — the three citation forms and how `validate`
  and `refs` use them.
- [Links across repositories](docs/links.md) — a change spanning several trees, `spectre link`'s
  guarded two-sided write, and what `validate` checks and deliberately does not.
- [Per-repository configuration](docs/configuration.md) — what `spectre/config.md` can override,
  and what stays fixed.

## What spectre does not do

No delta specs, no merge-back at archive time: a change edits spec files directly and git computes
the diff. No status field, assignee or timestamps. No network access, daemon or cache. Anything not
derivable from the markdown does not exist.
