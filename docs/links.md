# Links across repositories

How `link.md` records a change spanning several repositories' spectre trees, what `spectre link`
writes, and what `validate` checks. A link is `peers` one level down: `peers` declares a neighbour
tree; a link declares that *this change* is part of a specific change there. `peers` and citations:
[references across trees](references.md).

## Satellite and canonical

- A spanning change has exactly one **canonical** side — the repository holding the real
  `proposal.md`, `tasks.md` and optional `design.md` — and one or more **satellite** sides with no
  plan of their own.
- A change directory is a satellite when it holds `link.md` **and nothing else**. `link.md` beside
  any of `proposal.md`, `tasks.md` or `design.md` is not a satellite: every proposal, task and
  design check runs, plus the link checks.
- For a satellite, `validate` skips the "missing proposal.md" / "missing tasks.md" / "missing
  design.md" findings.

## The five sections

`link.md` has no free-text field. Each section has its own grammar and appears at most once — a
repeated heading is a parse error naming both lines, as in `peers`. A file is one of two shapes:

- **Satellite:** `## Part of`, `## Branch`, optional `## Tasks here`. `## Parts` and `## Merge order`
  are rejected beside `## Part of` — links are one level deep.
- **Canonical:** `## Parts`, `## Branch`, optional `## Merge order`. `## Tasks here` is rejected
  beside `## Parts`.
- `## Part of` and `## Parts` never appear together.

### `## Part of`

- Exactly one line: a single `` `<peer>:<change-id>` `` code span, nothing else.
- Rejects: more than one line, a line that isn't exactly one backtick-delimited span, a span with
  no `:`, an empty `<peer>` or `<change-id>`.

### `## Parts`

- One `` `<peer>:<change-id>` `` span per line, at least one line; per-line grammar as `## Part of`.
  An empty section is rejected.
- `<peer>` (both sections): non-empty, in the citation peer class (`internal/parse/ref.go`)
  `[a-z0-9][a-z0-9-]*`. `Agents`, `agents_repo` and a bare `:` after the peer are rejected.
- `<change-id>`: non-empty, not `.` or `..`, no `/` or `\`, not absolute, `filepath.Clean`-equal to
  itself, no control characters.

### `## Branch`

- Exactly one non-empty line in the git ref-name shape spectre enforces for a base branch: first
  character `[A-Za-z0-9._]`, every character `[A-Za-z0-9._/-]`.
- `feature/kan-363_x.v2` is accepted; a leading slash (`/kan-363`), a space (`kan 363`) or an
  empty section are rejected.

### `## Merge order`

- A numbered list, at least one item, each `N. `` `.` `` ` or `N. `` `<peer-name>` `` ` — `.` is
  this tree; a peer name uses the `## Parts` class.
- A line without the code span, or with an invalid peer, is rejected.
- Parsing checks shape only; agreement with `## Parts` is a `validate` check, below.

### `## Tasks here`

- Exactly one line of comma-separated positive integers and ascending `lo-hi` ranges: `1,3,7-9`,
  `4-12`.
- Strictly ascending overall, no overlaps: `12-4`, `3,1` and `2-5,4` are rejected, each naming the
  offending item. `0` is rejected.

## What `validate` checks

For every change carrying `link.md`:

- the peer named by `## Part of` or each `## Parts` entry must be **declared** in `peers`;
- for a declared, **present** peer, the counterpart must exist (`changes/<id>/`, then
  `changes/archive/<id>/`) and its `link.md` must name this change back — a one-sided link is a
  finding naming both sides;
- **archive skew** — one side archived, the other not — is a finding naming the archived side,
  never a refusal: a link outlives either side's archive timing;
- both sides' `## Branch` must be byte-identical once the counterpart resolves; a mismatch quotes
  both;
- satellite: every `## Tasks here` number must match a task in the canonical `tasks.md`; each
  unmatched number is a finding naming it;
- canonical, with `## Merge order`: it names `.` exactly once and every `## Parts` entry exactly
  once, nothing else — missing, duplicated and unknown entries are three distinct findings,
  compared against this `link.md` alone, independent of peer resolution.

### What it deliberately does not check

- A peer **declared but not present** is reported as not checked, never a finding. `peers` paths
  are relative to the tree's parent (`agents ../agents`), and a `/flow` worktree cannot see the
  other repository's sibling worktree there; a finding would fire on every worktree of every
  linked change. `refs` treats an unresolved peer citation the same way.
- A peer declared and **unreadable** — the path exists but the tree or its specs cannot be read —
  IS a finding: a broken peer, not the ordinary worktree case.

## `spectre link`

```
spectre link [--root <path>] [--force] <peer>:<canonical-id>
```

Run in the satellite tree. One command writes both sides:

- The peer's canonical `link.md` gains `## Parts` and `## Merge order`; the local tree gains
  `changes/<canonical-id>/` holding only `link.md`, with `## Part of` naming the peer. **The
  satellite's change id is always the canonical id** — no separate argument.
- The peer side is written first: a failed local write leaves only the peer side, never the
  reverse. Retrying is refused as "already exists", never appending a duplicate part entry.
- `## Merge order` is only **appended** to: a first link seeds `` `.` `` then the peer; a later link
  appends its peer. Reordering is a hand edit.
- `## Tasks here` is never written; add it by hand.

### The six refusals

Each is exit 1 naming the failed check. A malformed invocation — missing argument, no `:`, invalid
change id, unresolvable `--root` — is exit 2.

1. the peer is not declared in `peers`, or is declared and not present;
2. the canonical change id does not exist in the peer tree;
3. the canonical change is itself a satellite (it carries `## Part of`) — links are one level deep;
4. this link already exists, on either side;
5. the peer's own change directory has uncommitted modifications;
6. the peer's `peers` file has no name for *this* tree — the new `## Parts` entry would be
   unresolvable the moment it was written.

`--force` overrides refusal 5 alone. Refusal 5 checks only the peer's change directory
(`changes/<canonical-id>/`); uncommitted work elsewhere in the peer repository does not refuse.

## Worked example

`gymie-backend` (canonical, holds the plan for `route-c-checkout`) and `gymie-frontend`
(satellite, the UI half), each declaring the other in `spectre/peers`:

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

Frontend `spectre list` shows the satellite with the canonical's progress:

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

The canonical side shows plain progress, and `"parts"` in `--json`:

```
$ spectre list --root .
route-c-checkout  0/3
```

Peer declared but not checked out: `validate` stays quiet and `list` prints `-` for progress:

```
$ spectre list --root .
route-c-checkout  ->  gymie-backend  -
```

and `--json` marks the row `"progressUnknown": true`, so `0`/`0` is not mistaken for zero tasks:

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

A resolved row carries no `"progressUnknown"` key (`omitempty`).

## `list` degrades on a malformed link.md

- A `link.md` that fails to parse renders as a plain row (no `## Part of` / `## Parts` read);
  every other change still lists.
- A malformed `peers` file: every row renders, no link's peer resolved.

```
$ spectre list --root .
route-c-checkout  0/0
route-d-refund  1/1
$ spectre validate --root .
changes/route-c-checkout/link.md:1: link.md:3: "not-a-code-span" must be a single "<peer>:<change-id>" code span
1 finding(s)
```

`list` never exits `2` over an unparseable `link.md` or `peers`; that is a content problem
`validate` reports with exit `1`.
