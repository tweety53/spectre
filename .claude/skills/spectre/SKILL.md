---
name: spectre
description: Dispatch to any spectre subcommand — init, new, list, validate, refs, archive — show
  its real output, name what typically comes next, and for `new`, fill in the three files it
  scaffolds, through the bundled brainstorming and writing-plans skills.
allowed-tools: Bash(spectre:*), Read, Edit, Write, Skill
license: MIT
compatibility: Requires the spectre CLI on PATH (or built locally as `spectre`).
metadata:
  author: tweety53
  version: "1.0"
---

`/spectre <subcommand> [args]` in a checkout or after a manual copy, or `/spectre:spectre
<subcommand> [args]` after a `/plugin install`, runs that spectre subcommand and shows its **real**
output. It dispatches to the binary rather than describing each subcommand here, which would drift
from the binary.

**Thin, plus next-step guidance — not automation.** Run the command, show its output, name what
typically comes next, and **do not take that next step unasked**. `new` is the one exception.

## Workflow

1. Take the subcommand and arguments from the invocation verbatim — no remapping, no added flags.
2. Run `spectre <subcommand> <args>` (`--root <path>` only if the user supplied one; it must come
   before any positional change id).
3. Print the real stdout/stderr and the exit code.
4. Suggest the likely next step from the exit code and subcommand — e.g. after `init`, `new
   <change-id>`; after `validate` with findings, fix them before archiving; after a clean
   `validate`, the change is ready to work or archive.

### `new` — the one exception

`new` leaves three stubs and a change reporting `no tasks`, so scaffold-then-fill is the unit:

1. Run `spectre new <change-id>` (with `--root <path>` if supplied); print its output and exit
   code. On a non-zero exit (the change exists, no spectre tree found), stop there.
2. Read the three files under `changes/<change-id>/`. Their required headings are the
   **`## File templates`** table in [README.md](../../../README.md) — read them from there.
3. Fill them with the `brainstorming` and `writing-plans` skills bundled beside this one
   (`spectre:brainstorming` / `spectre:writing-plans` after a plugin install; a separately
   installed `superpowers:` copy works the same), as
   [docs/superpowers-example.md](../../../docs/superpowers-example.md) shows:
   - Invoke `brainstorming` to settle the design with the user, then write `proposal.md`,
     `design.md` and any capability spec from the approved design — into these files, not a
     separate design doc.
   - Then invoke `writing-plans` to write `tasks.md` in spectre's task shape: column-0
     `- [ ] <n>. <title>` task lines, flat integer ids, and `  - [ ] **Step N: …**` steps indented
     two columns beneath their task. Stop once it is written — executing the plan is not part of
     `new`.
   - Skip the skills' own commit steps: the guardrails below still hold.
   - If neither skill is available, fill the files with the user by hand, following
     [docs/example.md](../../../docs/example.md).
4. Name `spectre validate` as the next step — do not run it.

Only `new` invokes the bundled skills; every other subcommand stays dispatch-only.

## Guardrails

- Never run a subcommand the user didn't ask for.
- Never invent output — only what the binary printed.
- Never invent a subcommand, flag or behaviour; if unsure, run `spectre --help` or `spectre
  <subcommand> --help` and show that.
- Never commit, stage or `git mv` for the user — `spectre archive` needs the change's files
  already `git add`ed; suggesting that is as far as this skill goes.
- For `new`: never invent the required headings or their order — read them from the README's
  `## File templates` table.
- For `new`: never run `spectre validate`, `git add` or any other follow-on command unasked.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run any spectre subcommand | `/spectre <subcommand> [args]` (or `/spectre:spectre` after a plugin install) |
| Scaffold and fill in a new change | `/spectre new <change-id>` (or `/spectre:spectre new <change-id>`) |
