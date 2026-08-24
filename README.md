# spectre

Spec-driven change tracking in markdown. One binary, no dependencies, no database: the tree on
disk is the entire state.

## Install

```bash
go install github.com/tweety53/spectre/cmd/spectre@latest
```

## Getting started

A spectre tree lives inside a git repository: `archive` moves a finished change with `git mv`, so
both the repository and the change's files have to exist in git before you can archive anything.

There is no `init` command. A tree is a plain directory named `spectre` holding `specs/` and
`changes/`, and you make it by hand:

```bash
git init                              # if this isn't already a git repository
mkdir -p spectre/specs spectre/changes
```

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
| `spectre new <change-id>` | `--root` | scaffolds `changes/<id>/`; exits 1 if it already exists |
| `spectre list` | `--root`, `--specs`, `--json` | open changes with `3/7` progress; `--specs` lists capabilities instead |
| `spectre validate [change-id]` | `--root` | structural checks and reference resolution, whole tree or one change |
| `spectre refs <capability>#<id>` | `--root` | every citation of a requirement, in this tree and each declared peer |
| `spectre archive <change-id>` | `--root`, `--force` | `git mv`s a finished change into `changes/archive/` |
| `spectre migrate <openspec-dir>` | `--out`, `--force` | converts an OpenSpec tree into a new spectre tree |

`migrate` is the one exception to `--root`: it takes a source path and writes to `--out` instead
of resolving an existing spectre tree.

Exit codes are a contract every command holds to: `0` success, `1` findings or a content refusal,
`2` a usage or IO error.

## Spec format

```markdown
# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the session token before expiry.
  Indented prose beneath a bullet is free-form and preserved.
- R2: The picker SHALL show only enabled plans (@R1)
```

Requirement ids are `R<n>` by default (configurable, see below), unique within the file and
gap-free. They are stable once written: other trees cite them, and renumbering makes those
citations dangle. `(@R1)` above cites a requirement in this same file — see the next section for
the other two citation forms, which cite a capability or a peer tree and only resolve once that
capability, or a `peers` entry for that peer, actually exists.

## References across trees

Three scopes, widening left to right:

| Form | Meaning |
|------|---------|
| `(@R4)` | same file |
| `(@plans#R4)` | same tree, capability `plans` |
| `(@gymie:plans#R4)` | peer tree `gymie`, capability `plans` |

A citation naming a peer accepts any id shape, since the peer's own id prefix is that peer's own
business. A citation naming no peer must match this tree's own configured id prefix, or it is not
read as a reference at all — this stops `(@v2)` in ordinary prose from becoming a phantom citation.

`validate` resolves each reference it does read: peer declared in `peers`, its tree present on
disk, the capability file found inside it, the id found in that file — and reports the first
failing check as `file:line: message`. `refs` answers the reverse question — every place a
requirement is cited — before you renumber or delete it.

`peers` is the only file spectre reads besides specs, changes and `config.md`:

```
gymie ../gymie
gymie-frontend ../gymie-frontend
```

One name per line, paths relative to the tree's parent directory. A name repeated in `peers` is an
error naming both lines. Two different names resolving to the same path is legal — an alias.

## Per-repository configuration

`spectre/config.md` is optional; absent, every default below applies.

```markdown
# spectre config

## Rules
- shall-clause: error      # error | off, per rule
- placeholders: error
- headings: error
- malformed-bullet: error
- id-sequence: error
- task-sequence: error
- refs: error

## Vocabulary
- modal: SHALL             # the verb a requirement bullet must carry
- id-prefix: R              # replaces the accepted prefix entirely — "id-prefix: REQ-" makes
                             # REQ-1 legal and R1 no longer legal, not both at once

## Layout
- specs: specs              # relative to the tree root
- changes: changes
- extension: .md
```

Keys are closed: an unknown key, an unknown rule name, or a value outside the ones shown above
exits 2 rather than silently disabling a check. A key set twice — even across the same section —
is also an error, naming both lines. Headings and bullets inside fenced code blocks are ignored, so
an example inside a ` ``` ` fence never changes a real setting.

Configuration is per tree — a peer's own `config.md` governs how that peer's files are read (and
how `migrate` writes into it), so a tree using `REQ-` ids can be cited from a tree using `R` ids.
What `new` scaffolds is not configurable; it is compiled into the binary.

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

## `migrate`

```bash
spectre migrate [--out <dir>] [--force] <openspec-dir>
```

Converts an existing OpenSpec tree into a new spectre tree, written to `--out` (default: `./spectre`
in the working directory). `--out` is resolved the same way every other command resolves `--root`:
if its basename isn't literally `spectre`, spectre writes to `<out>/spectre` instead of `<out>`
directly, and the report line names the path actually written — so passing that same `--out` value
as `--root` to any other command always finds the tree `migrate` produced. It is
**non-destructive**: it never modifies or deletes the OpenSpec tree it reads. Re-running it against
an existing target fails unless `--force` is given, which clears that target's `specs/` and
`changes/` before writing — nothing else in the target is touched.

Conversions:

| OpenSpec input | spectre output |
|---|---|
| `specs/<cap>/spec.md` | `specs/<cap>.md` |
| `### Requirement: <text>` | `- R<n>: <text>`, numbered in document order |
| `#### Scenario:` blocks and their WHEN/THEN bullets | indented prose beneath the requirement bullet |
| `changes/<id>/proposal.md`, `tasks.md`, `design.md` | copied, headings renamed to spectre's |
| `- [ ] 1.1 <text>` nested task numbering | `- [ ] <n>. <text>`, flattened and renumbered |
| `changes/archive/<id>/` | copied as-is |
| `changes/<id>/specs/<cap>/spec.md` (delta spec) | copied verbatim as `deltas-<cap>.md`, with a warning — spectre has no delta concept |

After writing, `migrate` runs the same checks `spectre validate` runs over the produced tree and
reports every finding as a warning, so exit 0 means the migrated tree has zero warnings —
including delta-spec warnings, not only structural ones.

**A migrated tree almost always needs a per-requirement hand pass before it validates.** OpenSpec
states a requirement's normative sentence in the requirement body; spectre's `shall-clause` rule
requires it on the bullet itself. A straightforward conversion therefore fails `shall-clause` for
nearly every requirement — 249 of 278 warnings on one real corpus were exactly this conversion
artefact. `migrate` does not paper over this: when it emits more than one such warning, it adds one
line naming the count and the reason, so the fix is to move each requirement's modal sentence onto
its own bullet, not to disable the rule.

## What spectre does not do

No delta specs, no merge-back at archive time: a change edits spec files directly and git computes
the diff. No status field, assignee or timestamps. No network access, daemon or cache. Anything not
derivable from the markdown does not exist.
