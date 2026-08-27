# kan-350-spectre-updates — design

How each change in `proposal.md` works, and what was rejected on the way.

## 1. `spectre init`

```
spectre init [--root <path>]
```

`--root` carries the same meaning it carries in every other command: a path whose basename is not
literally `spectre` gets `spectre` appended, so `--root .` creates `./spectre` and `--root
./spectre` uses that directory. Bare `spectre init` creates `./spectre` in the working directory.

That rule lives today inside `tree.Open`, bound to a `Stat` requiring the directory to already
exist — the one thing `init` cannot assume. It is extracted to an exported
`tree.RootPath(root string) (string, error)`, which both `Open` and `Init` call. One statement of
the rule, two callers; `Open` keeps its `Stat` on the result.

### What it writes

| Path | Written |
|---|---|
| `<tree>/specs/` | when absent |
| `<tree>/changes/` | when absent |
| `<tree>/config.md` | when absent — **never overwritten** |

`init` is idempotent: it creates only what is missing, never overwrites, and exits 0 whether it
created three things or none — unless it refuses, per **Exit codes** below. It reports each path it created, and reports plainly when there was
nothing to do. A tree holding `specs/` but no `changes/` — a half-made tree, or one whose empty
`changes/` git declined to track — is repaired by re-running.

There is no `--force`. A command whose entire job is creating what is not there has nothing to
force, and the only file it could destructively rewrite is the one file a tree owner edits by hand.

### The scaffolded `config.md`

Written with every default present but **inert**, inside a fenced block. `internal/config`'s
parser already ignores headings and bullets inside fences — a behaviour `config_test.go` exercises
today, not one this change introduces. So the file documents every knob without changing any, and
enabling a key means moving its line out of the fence.

Its content mirrors `config.Default()` exactly: the seven rules of `config.RuleNames` at `error`,
`modal: SHALL`, `id-prefix: R`, `specs: specs`, `changes: changes`, `extension: .md`.

### Empty directories and git

`specs/` and `changes/` are created empty, and git tracks files rather than directories, so
neither is staged by a `git add spectre` that follows. This is accepted rather than worked around:
`config.md` makes `spectre/` itself tracked, and `spectre new` writes files under `changes/`
immediately. Two `.gitkeep` placeholders would buy a tracked empty directory nobody needs and add
two files every tree carries forever.

### Exit codes

`0` — created, or already complete. `1` — a content refusal: a plain file sits where `specs/` or
`changes/` belongs, so the path cannot be the directory a tree needs. `2` — usage error, or an IO
failure creating a directory or file.

The existence check therefore tests `fi.IsDir()`, not merely that `os.Stat` succeeded. Treating a
blocking file as "already there" would make `init` report success and leave a tree no later command
can use. `1` rather than `2` because this is a refusal about what is on disk, matching `new.go`'s
own `already exists` refusal, not a usage or IO error.

### The error that points at it

`tree.ErrNoRoot`'s message gains a pointer:

```
no spectre/ directory found (searched upwards from <start>); run "spectre init" to create one
```

The error every command emits when no tree exists should name the command that makes one.

## 2. Removing `migrate`

Deleted:

- `internal/cmd/migrate.go` (282 lines), `internal/cmd/migrate_test.go` (411 lines)
- `internal/openspec/read.go` (158 lines), `internal/openspec/read_test.go` (111 lines)
- `cmd/spectre/main.go`'s `case "migrate"` arm and its usage line
- the usage block's closing sentence, which becomes
  `every command accepts --root <path> to name the tree explicitly`
- README's `## migrate` section, its Commands-table row, and the paragraph beginning
  `migrate is the one exception to --root`

The capability does not survive anywhere. A dormant `internal/openspec/` would be code no caller
reaches, and a separate `cmd/spectre-migrate` binary would be a second thing to maintain for a
conversion that has run against one corpus.

Removing it makes `--root` uniform across every command for the first time, which is what makes
`init` reusing `--root` (§1) the natural choice rather than a second exception to a rule with one
exception already.

## 3. README, and three `docs/` pages

README keeps: title, install, getting started, the tree diagram, the command table, the exit-code
contract, `archive`, and "What spectre does not do" — and links the three pages below. Its
getting-started section shrinks most: the `mkdir -p` / `git init` recipe collapses to one
`spectre init`, and the paragraph explaining that no `init` command exists is deleted along with
the fact it stated.

| Page | Absorbs |
|---|---|
| `docs/spec-format.md` | `## Spec format` — requirement bullets, id rules, id stability |
| `docs/configuration.md` | `## Per-repository configuration` — the key table, the closed-key rules, per-tree scoping |
| `docs/references.md` | `## References across trees` — the three citation scopes, `peers`, what `validate` and `refs` each resolve |

Nothing is lost in the move: each page receives its section's full content, including the worked
examples. Only `migrate`'s prose is deleted, and only because its subject is.

## 4. Anonymization

`gymie` becomes `web`, `gymie-frontend` becomes `web-frontend`, across README (6 occurrences) and
the test fixtures in `internal/tree/tree_test.go`, `internal/tree/peer_test.go`,
`internal/cmd/refs_test.go`, `internal/cmd/validate_test.go`, `internal/check/refs_test.go`,
`internal/check/citations_test.go`, `internal/parse/spec_test.go`, `internal/parse/ref_test.go`
and `internal/render/render_test.go` — roughly 85 occurrences.

Mechanical throughout: fixture directory names, `peers` file bodies, expected-output strings, and
citation literals such as `(@gymie:plans#R4)`. The capability fixtures already in use — `auth`,
`billing`, `plans`, `app` — are neutral and are left alone. `github.com/tweety53/spectre` is this
module's real import path rather than an example, and stays.

## 5. This repository's own tree

`.gitignore` drops `/spectre` and gains `/bin/`. Local builds go to
`go build -o bin/spectre ./cmd/spectre`; `go install` remains the documented way to get the binary.

The `/spectre` entry existed to hide a root-level build output, and in doing so occupied — and
silently ignored — the exact path a spectre tree occupies. After the change a stray `./spectre`
binary appears in `git status` rather than hiding, which is how this went unnoticed.

This change's own `proposal.md`, `design.md` and `tasks.md` live in
`spectre/changes/kan-350-spectre-updates/`.

**The tree's `specs/` and `changes/` directories were made by hand, not by `init`** — the pipeline
needed somewhere to write those artifacts before task 2 existed to build the command, so the
bootstrap used the `mkdir -p` recipe the README documented at the time. `spectre init` was then run
against that half-made tree once it existed, and filled in the `config.md` the bootstrap had not
written. That is the idempotent fill-in-what-is-missing path (`init-idempotent`) exercised against a
real half-made tree, which is a better first use of the command than a greenfield one would have
been — but it is not the greenfield use, and saying otherwise would misdescribe what is on disk.

## Testing

**`internal/cmd/init_test.go`** (new):

- creates a complete tree from nothing;
- fills in a missing `changes/` in a tree that already has `specs/`;
- leaves an existing `config.md` byte-for-byte unchanged;
- `--root` whose basename is not `spectre` appends `spectre`;
- `--root` naming a `spectre` directory uses it directly;
- a re-run against a complete tree exits 0 and creates nothing;
- an unwritable parent exits 2;
- a plain file at `specs`, and separately at `changes`, makes `init` exit 1 and create nothing
  further;
- a round trip — `init`, then `new`, then `list` — proving the scaffolded tree is one the rest of
  the binary accepts.

**`internal/tree/tree_test.go`**: `RootPath` covered directly. `Open`'s existing tests reach the
path rule only through a `Stat` that now sits above it, so the extracted function needs its own
cases.

**`internal/cmd/init_test.go`**, one further case: the scaffolded `config.md`, **un-fenced**,
parses to `config.Default()`. It lives in `internal/cmd` rather than `internal/config` because
`internal/cmd` imports `internal/config`, so the reverse would be an import cycle.

**Un-fencing is what makes the test mean anything.** `config.Parse` skips fenced content
unconditionally — that is exactly the property that makes the scaffolded defaults inert — so loading
the file as written returns `config.Default()` regardless of what the fence contains, and the test
would pass against a template with a rule deleted. The test strips the fence markers first, parses
the body, and asserts equality; it fails when a rule is reordered, removed, or given a different
value.

**Two assertions, because one does not suffice.** `config.Parse` starts from `config.Default()` and
overrides only the keys it reads, so a deleted or reordered rule whose value already equals the
default parses to an identical `Config`. Parse-equality alone therefore catches a changed value and
nothing else. The test also extracts the template's `## Rules` lines and asserts they equal
`config.RuleNames` in order. Between the two, a changed value, a deleted rule and a reordered pair
all fail — that is the pair of assertions that stops the scaffold and the defaults drifting apart.

`Init` additionally rejects an unexpected positional argument, matching `new`, `archive` and `refs`;
`TestInitRejectsPositionalArgs` covers it.

Deleted with their subjects: `internal/cmd/migrate_test.go`, `internal/openspec/read_test.go`.

## Decisions

### Where `init` creates the tree

**ID:** init-target-root
**Status:** active
**Chosen:** `--root`, the same flag every other command takes — one argument shape across the whole
binary, and `tree.Open`'s existing basename rule is reused rather than restated.
**Considered:** a positional path (`spectre init [dir]`, reads like `git init <dir>`) — rejected
because it would make `init` the second command with its own argument shape at the same moment
`migrate`, the first, is being deleted to remove that exception; no argument at all (always
`./spectre`) — rejected as needlessly narrow when the flag already exists and costs nothing.

### `init` re-run against an existing tree

**ID:** init-idempotent
**Status:** active
**Chosen:** fill in what is missing and exit 0, never overwriting `config.md` — safe to re-run, and
repairs a half-made tree.
**Considered:** refuse whenever any part of the tree already exists, matching `spectre new`'s
blanket `already exists` refusal — rejected because a tree missing only
`changes/` would then be unfixable by the tool. **This is narrower than it sounds:** `init` does
refuse, with that same exit 1, when an existing path is a plain *file* rather than a directory (see
**Exit codes**). What was rejected is refusing on a healthy partial tree, not refusing at all;
refuse with a `--force` override — rejected as a
destructive flag on a command that creates what is absent, whose only overwritable file is the one
the tree owner hand-edits.

### How much of `migrate` goes

**ID:** migrate-delete-all
**Status:** active
**Chosen:** delete the command, `internal/cmd/migrate.go`, `internal/openspec/`, both test files
and the README section — ~960 lines, and the binary's only `--root` exception with them.
**Considered:** keep `internal/openspec/` for a possible future migration — rejected as dead code
no caller reaches; move the conversion to a separate `cmd/spectre-migrate` binary — rejected as a
second binary to maintain for a converter that has run against one corpus.

### README shape

**ID:** readme-plus-docs
**Status:** active
**Chosen:** a brief README linking three `docs/` pages, each receiving one of the relocated
reference sections in full.
**Considered:** one README with the reference material cut rather than moved — rejected because the
config key table and the citation scopes are facts a user needs somewhere, and `--help` does not
carry them; compressing every section in place — rejected as too small a reduction for a ticket
asking for "as brief as you can".

### Fixture naming

**ID:** peer-fixture-web
**Status:** active
**Chosen:** `web` / `web-frontend` — reads as a real neighbouring service, and preserves the
backend/frontend pair the original fixture illustrated.
**Considered:** `core` / `core-frontend` — equally workable, no deciding difference; `peer` /
`peer-frontend` — rejected because a fixture literally named `peer` in a `peers` file reads as a
placeholder and makes the README example less concrete; `upstream` / `downstream` — rejected as
awkward in `peers` lines and fixture directory names.

### Build output for this repository

**ID:** build-to-bin
**Status:** active
**Chosen:** `.gitignore` ignores `/bin/` and not `/spectre`; local builds use
`go build -o bin/spectre ./cmd/spectre`.
**Considered:** ignoring `/spectre.bin` and continuing to build at the root — rejected because
nothing stops the next `go build .` producing `./spectre` again; leaving `/spectre` ignored and
placing this repository's tree elsewhere — rejected because it breaks the upwards-search
convention every command relies on.

### Empty scaffold directories and git

**ID:** no-gitkeep
**Status:** active
**Chosen:** create `specs/` and `changes/` empty and accept that git does not track them.
**Considered:** writing `.gitkeep` into each — rejected because `config.md` already makes
`spectre/` tracked and `spectre new` populates `changes/` immediately, so the placeholders would be
two files every tree carries forever to solve nothing.

## Open questions

None.
