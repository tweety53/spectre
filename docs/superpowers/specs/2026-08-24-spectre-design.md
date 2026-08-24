# spectre — design

**Date:** 2026-08-24
**Status:** approved, pending implementation plan

## Purpose

spectre is a lightweight replacement for OpenSpec: a single Go binary operating statelessly
over a markdown tree that holds long-lived capability specs and per-change artifacts.

It exists because OpenSpec, at version 1.6.0, costs more than the workflow it serves. Four
specific costs drove this design:

- **CLI surface.** Twenty-plus commands (`schema`, `store`, `workset`, `doctor`, `context`,
  `completion`, …) of which a myflow run uses a handful. The rest is surface an agent has to
  reason about and a human has to remember.
- **Instruction bloat.** `openspec instructions`, resolved templates and injected AGENTS blocks
  put a large, tool-authored body of text into every run's context.
- **Format control.** The `### Requirement:` / `#### Scenario:` shape, the delta-spec syntax and
  the archive-time merge semantics are the tool's, not the author's.
- **Overhead per change.** The time between deciding to make a change and having artifacts that
  validate is larger than the change often warrants.

spectre answers those by shrinking the tool to five working commands plus an optional migration, keeping conventions out of the
binary, owning the artifact format, and persisting nothing but markdown.

## Scope

spectre stands alone. myflow continues to use OpenSpec until spectre has proven itself in a
project; migrating the pipeline is a separate, later change, and this design deliberately makes
no accommodation for it beyond not obstructing it.

## Architecture

### Stateless over the filesystem

The tree is the state. spectre never writes a file that describes another file, and every fact
it reports is recomputed from markdown on each invocation.

| Fact | Derived from |
|------|--------------|
| A change exists | a folder under `changes/` |
| Change id | the folder name |
| Open vs archived | path — `changes/<id>` vs `changes/archive/<id>` |
| Task progress (`3/7`) | count of `- [x]` vs `- [ ]` lines in `tasks.md` |
| Capability list | filenames in `specs/` |
| Requirement ids | `R<n>:` bullets parsed from a spec file |
| Which specs a change touches | `git diff --name-only <base>...HEAD -- spectre/specs/` |

That last row is the one a stored model would have held as an `## Impact` field the author has to
keep truthful. Git already knows which files a branch changed, so spectre asks git.

Consequences worth stating:

- No config to resolve, therefore no precedence rules and no per-project drift.
- No shared mutable state, therefore no locks and no transactions. Two agents in two worktrees
  cannot corrupt each other.
- `git mv`, `rm -rf` and hand-editing in an editor are all legal spectre operations. The tool is
  never the only way in, and so never needs a repair command.
- `validate` is a pure function of the tree: its tests are fixture directories, and a wrong answer
  always reproduces from the checkout.
- Anything not inferable from files does not exist — no start time, no assignee, no status beyond
  checkbox arithmetic. If such a fact is wanted later it is written into the markdown as visible
  text, not into a sidecar.
- `list` walks the tree on every call. That is the entire performance story; see **Measurements**.

### Writes go through parse → model → render

Every mutation parses the file into the Go model, changes the model, and renders the file from
the model. No command string-appends into a markdown file.

This makes malformed output unrepresentable, gives the renderer sole ownership of task numbering
and bullet shape, and yields `spectre fmt` — normalise a hand-edited file by round-tripping it —
for no extra machinery.

## Layout

```
spectre/
  peers                       # "<name> <relative-path>" lines, one per neighbour tree
  specs/<capability>.md       # flat file per capability; capability name = filename
  changes/<id>/proposal.md
             /tasks.md
             /design.md       # optional
  changes/archive/<id>/
```

Specs are flat files rather than `<capability>/spec.md` directories: one less level for the same
content.

`peers` is the only file spectre reads that is neither a spec nor a change. It carries the single
fact a tree cannot derive about itself — where its neighbours sit. Paths are relative, so the file
survives cloning, moving and git worktrees.

A name repeated in `peers` is an error naming both lines, on the same reasoning as `config.md`'s
closed keys: silently taking the last line would resolve citations against a tree the author did not
mean. Two different names resolving to one path are an alias, and legal.

Root resolution walks up from the working directory until a `spectre/` directory appears. `--root`
overrides. With neither, spectre exits 2 naming the directory it searched from.

## Formats

### Spec file

```markdown
# <capability>

## Purpose
One paragraph stating what this capability is for.

## Requirements
- R1: The system SHALL refresh the session token before expiry.
- R2: The picker SHALL show only enabled plans (@gymie:plans#R7)
```

Requirement ids are `R<n>`, unique within a file and gap-free. An id is stable once written: it is
the anchor other trees cite, so renumbering is what makes citations dangle, and that breakage
surfacing is the point rather than a defect.

### Change files

`proposal.md` carries `## Why` and `## What changes`. `tasks.md` carries `- [ ] 1. <text>` lines,
numbered from 1 and ascending; the checkbox is the only progress signal spectre reads. `design.md`
is optional and unparsed.

### Cross-tree references

Three scopes, widening left to right:

| Form | Meaning |
|------|---------|
| `R4` | same file |
| `auth#R4` | same tree, capability `auth` |
| `gymie:auth#R4` | peer tree `gymie`, capability `auth` |

Resolution is four ordered checks, each with its own message: the peer is declared in `peers`; its
path exists; the capability file exists inside it; the id is present in that file. There is no
network access, no fetch and no cache — a neighbour that is not checked out produces a finding, not
a silent pass, and neither does one that is present but unreadable: a peer whose own files cannot be
read is reported against the citing reference rather than abandoning the run, since the citing tree
may be sound.

A citation naming a peer but no capability — `(@gymie:R4)` — is reported as malformed rather than
resolved or ignored. The parser admits the shape deliberately: a citation silently dropped at parse
time is the failure this tool exists to prevent, so it is read and then named.

## Commands

| Command | Behaviour |
|---------|-----------|
| `spectre new <id>` | Scaffolds `changes/<id>/` with proposal and tasks templates. Exits 1 if it already exists. |
| `spectre list` | Open changes with `3/7` progress. `--specs` lists capabilities with requirement counts. `--json` for machine consumption. |
| `spectre validate [id]` | No argument validates the whole tree. |
| `spectre archive <id>` | Asserts every task is checked, then `git mv`s the folder under `changes/archive/`. `--force` overrides the assertion. Does not commit. |
| `spectre refs <capability>#<id>` | Prints every citation of a requirement, scanning this tree and each declared peer, and prints which trees it scanned. |
| `spectre migrate <openspec-dir>` | Optional. Converts an existing OpenSpec tree into a spectre tree. Never required by any other command. |

A peer that cannot be read is skipped and marked in the `scanned:` line — `<name> (unreadable: <err>)`
beside `<name> (not present)` — and `refs` still exits 0. A neighbour's broken tree bounds the
answer's scope rather than failing the query, and the scanned line is where the user reads that
bound. Only a local failure — a bad argument, an unresolvable root, this tree's own unreadable files
— exits non-zero.

`refs` earns its place as the fifth command because without it a requirement's dependants only
surface when someone else runs `validate` — you would be deleting and renumbering requirements
blind. Its completeness depends on the neighbour declaring you in its own `peers` file; that is a
property of the data, and printing the scanned trees exposes the answer's scope rather than hiding
it.

## Migration from OpenSpec

`spectre migrate <openspec-dir>` exists so an OpenSpec tree can be adopted, but nothing in spectre
depends on it: a tree created by `spectre new` is indistinguishable from a migrated one, and the
command is never invoked by another command.

It is **non-destructive**. It writes a new `spectre/` tree and never modifies or deletes the
`openspec/` tree it reads. Re-running it on an existing target fails unless `--force` is given.

Conversions:

| OpenSpec input | spectre output |
|----------------|----------------|
| `specs/<cap>/spec.md` | `specs/<cap>.md` |
| `### Requirement: <text>` | `- R<n>: <text>`, numbered in document order |
| `#### Scenario:` blocks and their WHEN/THEN bullets | indented prose beneath their requirement bullet |
| `## Purpose` | `## Purpose`, verbatim |
| `changes/<id>/proposal.md`, `tasks.md`, `design.md` | copied, headings renamed to spectre's |
| `- [ ] 1.1 <text>` nested task numbering | `- [ ] <n>. <text>`, flattened and renumbered |
| `changes/archive/<id>/` | copied as-is |

Scenarios are preserved rather than discarded. spectre's format has no scenario concept and its
parser ignores lines that are not requirement bullets, so the block survives as indented prose that
validates cleanly. Discarding it would delete acceptance criteria a human wrote, which migration
has no mandate to do.

Delta specs — `changes/<id>/specs/<cap>/spec.md` carrying `## ADDED Requirements` — are the one
input with no faithful target, because spectre has no delta concept and applying a delta is archive
semantics rather than migration. `migrate` copies each delta file verbatim into the change folder
as `deltas-<cap>.md`, reports one warning per file naming it, and exits 1 so the run is visibly
incomplete. Resolving those files is a human edit against the migrated spec.

Every run prints a report: files written, requirements converted, tasks renumbered, and each
warning with its source path.

## Validation rules and exit codes

Findings print as `file:line: message`. Exit 0 clean, 1 findings, 2 usage or IO error.

- Spec files: `# <name>`, `## Purpose` and `## Requirements` headings present.
- No `TBD` or `TODO` anywhere in a spec or proposal.
- Requirement bullets match `- R<n>: … SHALL …`; a bullet under `## Requirements` that does not is
  reported as malformed rather than ignored.
- Requirement ids unique within a file and gap-free.
- Task lines match `- [ ] <n>. <text>` or `- [x] <n>. <text>`, numbers unique and ascending; a line
  opening `- [` that does not match is reported as malformed rather than ignored.
- `proposal.md` has `## Why` and `## What changes`; a change missing either `proposal.md` or
  `tasks.md` is reported, since `archive` refuses only on unchecked tasks and a change with no task
  file would otherwise pass unremarked.
- Headings and requirement bullets are matched outside fenced code blocks only, so an example inside
  a fence neither satisfies a heading rule nor trips the malformed-bullet rule.
- Every reference resolves by the four checks above.

## Implementation

Module `github.com/tweety53/spectre`, `go 1.26`, standard library only — no CLI framework, so
there is no dependency to track and no generated command scaffolding.

```
cmd/spectre/main.go   subcommand dispatch and exit codes
internal/tree/        root resolution, peers, reading changes and specs
internal/parse/       markdown to model
internal/render/      model to markdown
internal/check/       validation rules, pure functions over the model
internal/cmd/         new.go list.go validate.go archive.go refs.go migrate.go
internal/openspec/    the OpenSpec reader, used only by migrate
```

## Testing

Table-driven tests per command over fixture trees copied into `t.TempDir()`, with golden files for
`list`, `validate` and `refs` output. Cross-tree cases build two sibling trees in the temporary
directory with a real relative `peers` file — the filesystem is not faked. `archive` shells out to
`git mv`, so its test runs in a real temporary repository. `migrate` is tested against a fixture
copied from a real OpenSpec tree, with golden output covering a converted spec, a preserved
scenario block, flattened task numbering, and the delta-spec warning path with its exit 1.

## Measurements

Measured 2026-08-24 on the corpus shape of `agents/openspec` (35 spec files, 558 KB, 255
requirements), plus synthetic 10× and 100× trees. Go 1.26.5, darwin/arm64, warm page cache, best
of five.

| Tree | Files | Requirements | Parse whole tree | Validate | `refs` query |
|------|-------|--------------|------------------|----------|--------------|
| today | 35 | 560 | 3.4 ms | 35 µs | 1 µs |
| 10× | 350 | 5 600 | 28 ms | 255 µs | 6 µs |
| 100× | 3 500 | 56 000 | 267 ms | 3.7 ms | 64 µs |

SQLite over the same 56 000 rows: indexed lookup 9 µs, connection open 80 µs, unindexed `LIKE`
scan over references 2.9 ms, and a full rebuild from markdown 605 ms — twice the cost of simply
parsing the markdown.

## Non-goals

No delta specs and no archive-time merge-back. No configuration file beyond `peers`. No status
field, assignee or timestamps. No network access, daemon or server. No `instructions` command:
conventions live in whatever skill drives spectre, not in the binary.

## Rejected alternatives

**Changes without a long-lived spec base.** Rejected because the accumulated capability specs are
the artefact with the longest useful life; git history and code do not answer "what is this system
required to do" the way a specs tree does.

**Keeping OpenSpec's delta specs.** Rejected because a change branch can edit spec files directly
and let git compute the diff. Delta blocks are a second encoding of a diff that git already
represents, and their merge-back at archive time is the step most likely to lose an edit.

**`### Requirement:` headings with `#### Scenario:` WHEN/THEN blocks.** Rejected in favour of flat
`- R<n>:` bullets. The heading form costs several lines of ceremony per requirement, and the
scenario blocks duplicate what tests state executably. The cost is weaker structural validation
and no per-requirement acceptance criteria in the spec — accepted deliberately.

**Convention plus skills with no binary at all.** Rejected because validation and archiving become
agent-performed, and therefore non-deterministic; a binary makes them checkable in CI.

**A per-project `spectre.yaml`.** Rejected under YAGNI — no second caller needs different paths,
templates or strictness, and it reintroduces the configuration surface this tool exists to remove.

**A per-change state file with an explicit status.** Rejected because it is a second source of
truth that drifts from the checkboxes, and myflow already owns pipeline state separately.

**A shared spec tree with many consumer repos.** Rejected because it splits a single change across
two repositories — artefacts in the specs repo, code in the consumer — requiring two branches and
two pull requests for one unit of work.

**One change spanning several repos, with tasks tagged by repo.** Rejected as unneeded: the
coupling between repositories is at the level of requirements, not tasks, and cross-tree references
express that coupling without spectre having to know about multiple checkouts at once.

**A machine-level registry of trees (`~/.spectre/trees`).** Rejected because it is per-machine
state that is not in git: CI, a fresh clone and each new worktree would have to rebuild it, and a
tree would no longer describe itself completely.

**Peer resolution by sibling-directory convention.** Rejected because `gymie:auth#R4` would
silently resolve to whatever directory happens to sit beside the tree, making correctness depend on
the checkout layout of the machine.

**Peers as git URLs with a local cache.** Rejected because it adds fetching, staleness and a
network dependency to `validate`.

**A store — SQLite or JSON — as the source of truth, with markdown rendered from it.** Rejected on
both measurements above and error analysis. The error classes a store closes (malformed output,
duplicate ids, numbering drift) are closed already by rendering from the model. The one class it
cannot close is the dangling cross-tree reference, because peers are separate repositories and no
foreign key spans them — and that is the whole multi-repo feature. Against that it introduces
store-versus-markdown drift after any human edit, PR review or merge; conflicts in a committed
binary or JSON artefact that cannot be resolved by hand; and a fresh clone, CI run or worktree with
no data at all.

**A derived index cache in v1.** Deferred, not rejected. At 3.4 ms for a full parse the cache
would buy nothing and cost an invalidation path. If the tree ever passes roughly 1 000 spec files,
the addition is safe and additive: `spectre/.cache/index.json`, keyed by file mtime, gitignored,
rebuildable in under a second, never authoritative, and deleting it can only cost time.

**Dropping scenarios during migration.** Rejected because a `#### Scenario:` block is
human-authored acceptance criteria, and a conversion tool deleting it silently is the worst failure
this design can produce. Preserving it as indented prose costs nothing: the parser ignores those
lines and validation passes.

**Applying delta specs during migration.** Rejected because applying a delta is archive semantics,
not conversion — it decides which requirements land and in what order, which is a judgement the
tool cannot make for an in-flight change. `migrate` copies deltas verbatim, warns per file, and
exits 1.

**Making migration a prerequisite or a first-class workflow step.** Rejected: spectre stands alone,
and a migrated tree has no properties a freshly created one lacks. The command is optional in the
strict sense that no other command calls it and no feature assumes it has run.
