# Links across repositories

How `link.md` records a change that spans more than one repository's spectre tree, what `spectre
link` writes, and what `validate` checks against it.

A link connects two trees the way `peers` connects two repositories, one level down: `peers`
declares that a neighbour tree exists, a link declares that *this change* is part of a specific
change over there. See [references across trees](references.md) for `peers` itself and the citation
forms built on it — this file does not restate either.

## Satellite and canonical

A change that spans repositories has exactly one **canonical** side — the repository holding the
real `proposal.md`, `tasks.md` and, optionally, `design.md` — and one or more **satellite** sides,
each a repository the change also touches with no plan of its own.

A change directory counts as a satellite when it holds `link.md` **and nothing else**: no
`proposal.md`, no `tasks.md`, no `design.md`. A directory carrying `link.md` next to a `design.md`
is not a satellite — `validate` runs every existing proposal, task and design check for it, in
addition to the link checks below. This is deliberate: a satellite is a pointer only because it
carries nothing else, and a `design.md` beside it means it isn't one.

`spectre validate` skips the "missing proposal.md" / "missing tasks.md" / "missing design.md"
findings entirely for a satellite — it does not expect a scaffold there, so it does not report one
as absent.

## The five sections

`link.md` has no free-text field. Each of its five sections has its own grammar, and every section
may appear at most once — a repeated heading is a parse error naming both lines, the same way a
repeated name in `peers` errors.

A file is one of two shapes, never a mixture:

- **Satellite:** `## Part of`, `## Branch`, and optionally `## Tasks here`. `## Parts` and
  `## Merge order` are rejected alongside `## Part of` — links are one level deep.
- **Canonical:** `## Parts`, `## Branch`, and optionally `## Merge order`. `## Tasks here` is
  rejected alongside `## Parts`.

`## Part of` and `## Parts` cannot both be present in the same file, for the same reason.

### `## Part of`

Exactly one line: a single `` `<peer>:<change-id>` `` code span, nothing else. Rejects: more than
one line, a line that isn't exactly one backtick-delimited span, a span with no `:`, an empty
`<peer>` or `<change-id>`.

### `## Parts`

One `` `<peer>:<change-id>` `` code span per line, at least one line. Same per-line grammar as
`## Part of`. Rejects an empty section.

`<peer>` in both sections must be non-empty and match the same character class a citation's peer
scope already uses (`internal/parse/ref.go`): lowercase letters, digits and hyphens, starting with
a letter or digit — `[a-z0-9][a-z0-9-]*`. `Agents`, `agents_repo` and a bare `:` after the peer are
all rejected. `<change-id>` must satisfy spectre's own change-id rule: non-empty, not `.` or `..`,
no `/` or `\`, not an absolute path, `filepath.Clean`-equal to itself, and free of control
characters.

### `## Branch`

Exactly one non-empty line, matching the git ref-name shape spectre already enforces elsewhere for
a base branch: the first character in `[A-Za-z0-9._]`, every character in `[A-Za-z0-9._/-]`. A
leading slash, an embedded space, or an empty section are all rejected. `feature/kan-363_x.v2` is
accepted; `/kan-363` and `kan 363` are not.

### `## Merge order`

A numbered list, at least one item, each line `N. `` `.` `` ` or `N. `` `<peer-name>` `` ` — a list
number, a period, whitespace, then a code span naming either the literal `.` (this tree) or a peer
name in the same character class `## Parts` uses. A line missing the code span, or naming an invalid
peer, is rejected.

Parsing only checks the *shape* of each item; whether the set of items agrees with `## Parts` is a
`validate` check, described below.

### `## Tasks here`

Exactly one line: comma-separated positive integers and ascending `lo-hi` ranges, e.g. `1,3,7-9` or
`4-12`. The whole list must be strictly ascending overall with no overlaps between items — `12-4`
(descending), `3,1` (out of order) and `2-5,4` (overlapping) are all rejected, each naming the
offending item. `0` is rejected: every task number is a positive integer.

## What `validate` checks

For every change carrying a `link.md`, `validate`:

- requires the peer named by `## Part of` or by each `## Parts` entry to be **declared** in
  `peers` — not declared is a finding;
- when the peer is declared and **present** on disk, requires the counterpart change to exist there
  (searched under `changes/<id>/`, then `changes/archive/<id>/`) and its own `link.md` to name this
  change back — a one-sided link is a finding naming both sides;
- reports **archive skew** — the counterpart resolved under `changes/archive/` while this side is
  still under `changes/`, or the reverse — as a finding naming which side is archived, never as a
  refusal; a link legitimately outlives either side's own archive timing;
- requires both sides' `## Branch` to be byte-identical once the counterpart resolves;
  disagreement is a finding quoting both;
- on the satellite side, requires every number in `## Tasks here` to match a task in the
  canonical's `tasks.md`; a number matching none is a finding naming it;
- on the canonical side, when `## Merge order` is present, requires it to name `.` exactly once and
  every `## Parts` entry exactly once, and nothing else — a missing entry, a duplicated entry and
  an unknown entry are three distinct findings, independent of any peer's resolution, since this
  check only compares the list against `link.md`'s own `## Parts`.

### What it deliberately does not check

A peer that is **declared but not present** on disk is reported as not checked, never as a finding.
`peers` paths are relative to the tree's parent directory (`agents ../agents`), and every `/flow`
worktree legitimately cannot see a sibling worktree of the other repository at that path — the
sibling lives in a different worktree root entirely. Treating that as a finding would fire on every
worktree of every linked change, so it is silence instead: the same peer-absence handling `refs`
already gives an unresolved citation.

A peer that is declared and **unreadable** — the path exists but the tree cannot be opened, or its
specs cannot be read — is not the same case, and IS a finding. Absent is the ordinary worktree case
above; unreadable means a peer tree that is there and broken, which a validator says out loud rather
than swallowing alongside the ordinary case.

## `spectre link`

```
spectre link [--root <path>] [--force] <peer>:<canonical-id>
```

Run in the satellite tree. It writes both sides of the link in one command: the peer's canonical
`link.md` gains `## Parts` and `## Merge order`; the local tree gains a new `changes/<canonical-id>/`
directory holding only `link.md`, with `## Part of` naming the peer. **The satellite's change id is
always the canonical id** — there is no separate argument for it.

The peer's side is written first, the local side second: a failure writing the local side leaves
only the peer side landed, never the reverse. Retrying the same command after such a partial write
is refused as "already exists" rather than appending a second, duplicate part entry.

`## Merge order` is only ever **appended** to, never reordered: a first link seeds it as
`` `.` ``, then the new peer name; a later link to a second peer appends that peer's name after the
existing entries. Changing the landing order is a hand edit to `link.md` afterwards — the command
does not do it.

`spectre link` never writes `## Tasks here`; that section, when wanted, is added by hand afterward.

### The six refusals

Each refusal is exit 1 and names the check that failed. A malformed invocation — a missing
argument, an argument with no `:`, an invalid change id, an unresolvable `--root` — is exit 2.

1. the peer is not declared in `peers`, or is declared and not present;
2. the canonical change id does not exist in the peer tree;
3. the canonical change is itself a satellite (it carries `## Part of`) — links are one level deep;
4. this link already exists, on either side;
5. the peer's own change directory has uncommitted modifications;
6. the peer has no declared name for *this* tree in its own `peers` file — without one there is no
   name to write into the peer's new `## Parts` entry, so the entry would be unresolvable the
   moment it was written.

`--force` overrides refusal 5 alone, and no other.

Refusal 5's dirty-tree check is scoped to the peer's own **change directory**
(`changes/<canonical-id>/`), not the whole peer repository: uncommitted work anywhere else in the
peer repository — a different change in progress, an unrelated edit — does not refuse the link.

## Worked example

Two sibling repositories, `gymie-backend` (canonical: it holds the real plan for
`route-c-checkout`) and `gymie-frontend` (satellite: it only needs the UI half). Each declares the
other in its own `spectre/peers`:

```
# gymie-backend/spectre/peers
gymie-frontend ../gymie-frontend

# gymie-frontend/spectre/peers
gymie-backend ../gymie-backend
```

`gymie-backend`'s `changes/route-c-checkout/tasks.md` has three tasks. From the frontend repo, on
branch `route-c-checkout`:

```
$ spectre link gymie-backend:route-c-checkout
linked spectre/changes/route-c-checkout to gymie-backend:route-c-checkout
```

This writes the frontend's `changes/route-c-checkout/link.md`:

```
## Part of

`gymie-backend:route-c-checkout`

## Branch

route-c-checkout
```

and the backend's `changes/route-c-checkout/link.md`:

```
## Parts

`gymie-frontend:route-c-checkout`

## Branch

route-c-checkout

## Merge order

1. `.`
2. `gymie-frontend`
```

Adding `## Tasks here` to the frontend side by hand, then `spectre validate` in both trees:

```
$ spectre validate --root .
no findings
```

`spectre list` on the frontend renders the satellite as the change it's part of, with the
canonical's progress:

```
$ spectre list --root .
route-c-checkout  ->  gymie-backend  0/3
```

and, with `--json`, carries `"partOf"`:

```json
{
  "changes": [
    {
      "id": "route-c-checkout",
      "done": 0,
      "total": 3,
      "partOf": {
        "peer": "gymie-backend",
        "changeId": "route-c-checkout"
      }
    }
  ]
}
```

The canonical side's `spectre list` shows plain progress in text form and carries `"parts"` in
`--json` — `list`'s special rendering is for a satellite's row, not a canonical's:

```
$ spectre list --root .
route-c-checkout  0/3
```

If the peer becomes unreachable — declared in `peers` but not checked out — `validate` stays quiet
and `list` prints `-` for the unresolved progress column instead of refusing:

```
$ spectre list --root .
route-c-checkout  ->  gymie-backend  -
```

and, with `--json`, `"progressUnknown": true` marks the row explicitly rather than leaving `done`
and `total` at `0` indistinguishable from a canonical change with genuinely zero tasks:

```json
{
  "changes": [
    {
      "id": "route-c-checkout",
      "done": 0,
      "total": 0,
      "partOf": {
        "peer": "gymie-backend",
        "changeId": "route-c-checkout"
      },
      "progressUnknown": true
    }
  ]
}
```

A row whose progress resolved carries no `"progressUnknown"` key at all — the field only ever
appears, `omitempty`, on a row it applies to.

## `list` degrades on a malformed link.md

A `link.md` that fails to parse — the same content problem `validate` reports as a finding — does
not stop `list` from rendering every other change. The broken change's own row falls back to a
plain one, with no `"## Part of"` or `"## Parts"` read from it, rather than the whole listing
aborting over one bad file. A malformed `peers` file degrades the same way: every row still
renders, with no peer resolved for any link.

```
$ spectre list --root .
route-c-checkout  0/0
route-d-refund  1/1
$ spectre validate --root .
changes/route-c-checkout/link.md:1: link.md:3: "not-a-code-span" must be a single "<peer>:<change-id>" code span
1 finding(s)
```

`list` never exits `2` over a `link.md` or `peers` file it cannot parse — that is a content
problem for `validate` to report as exit `1`, not an IO or usage error.
