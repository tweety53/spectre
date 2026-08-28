---
name: spectre
description: Dispatch to any spectre subcommand — init, new, list, validate, refs, archive — show
  its real output, name what typically comes next, and for `new`, help fill in the three files it
  scaffolds.
allowed-tools: Bash(spectre:*), Read, Edit
license: MIT
compatibility: Requires the spectre CLI on PATH (or built locally as `spectre`).
metadata:
  author: tweety53
  version: "1.0"
---

`/spectre <subcommand> [args]` in a checkout or after a manual copy, or `/spectre:spectre
<subcommand> [args]` after a `/plugin install`, runs that spectre subcommand and shows its **real**
output — never a description of what it would print. It covers every subcommand spectre has by
dispatching to the binary, not by re-describing each one's behaviour here, which would drift from
the binary the moment either changes.

**Thin, plus next-step guidance — not automation.** Run the command, show what it actually printed,
then name what typically comes next. **Do not take that next step unasked** — if the likely next
command is `spectre validate` or `git add`, say so and stop; do not run it. `new` is the one named
exception — see below.

## Workflow

1. Take the subcommand and any arguments from the user's invocation verbatim — no remapping, no
   added flags.
2. Run `spectre <subcommand> <args>` (add `--root <path>` only if the user supplied one; every
   spectre command accepts it, and it must come before any positional change id).
3. Print the command's real stdout/stderr and its exit code.
4. Name the likely next step, briefly, based on the exit code and subcommand — e.g. after `init`,
   that `new <change-id>` scaffolds a change; after `validate` reports findings, that they need
   fixing before archiving; after a clean `validate`, that the change is ready to work or archive.
   State it as a suggestion, not an action taken — except for `new`, below.

### `new` — the one exception

`new` alone leaves three stubs and a change that reports `no tasks` until a plan exists, so
scaffold-then-fill is the useful unit, and it is the one case where this skill goes past showing
output:

1. Run `spectre new <change-id>` (with `--root <path>` if the user supplied one). Print its real
   output and exit code. If it exits non-zero (e.g. the change already exists, or no spectre tree
   was found), stop and show that — do not proceed to filling anything in.
2. Read the three files it wrote — `proposal.md`, `tasks.md`, `design.md` under
   `changes/<change-id>/`. Their required headings are the **`## File templates`** table in
   [README.md](../../../README.md); do not restate that order here.
3. Help the user write each file's content, following the prompts and worked example in
   [docs/example.md](../../../docs/example.md) — link to it and to the README rather than repeating
   what those documents already state.
4. Once filled in, name the next step — running `spectre validate` — without running it unasked.

## Guardrails

- Never run a subcommand the user didn't ask for.
- Never invent output — only what the binary actually printed.
- Never invent a subcommand, flag, or behaviour spectre does not have; if unsure, run `spectre
  --help` or `spectre <subcommand> --help` and show that instead of guessing.
- Never commit, stage, or `git mv` on the user's behalf — `spectre archive` itself needs the
  change's files already `git add`ed; suggesting that is as far as this command goes.
- For `new`: never invent the required headings or their order — they live only in the README's
  `## File templates` table and must be read from there, not repeated from memory.
- For `new`: never run `spectre validate`, `git add`, or any other follow-on command unasked.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run any spectre subcommand | `/spectre <subcommand> [args]` (or `/spectre:spectre` after a plugin install) |
| Scaffold and fill in a new change | `/spectre new <change-id>` (or `/spectre:spectre new <change-id>`) |
