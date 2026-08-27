---
name: spectre
description: Dispatch to any spectre subcommand — init, new, list, validate, refs, archive — show its real output, and name what usually comes next. Use for /spectre.
allowed-tools: Bash(spectre:*)
license: MIT
compatibility: Requires the spectre CLI on PATH (or built locally as `spectre`).
metadata:
  author: tweety53
  version: "1.0"
---

`/spectre <subcommand> [args]` runs that spectre subcommand and shows its **real** output — never a
description of what it would print. It covers every subcommand spectre has by dispatching to the
binary, not by re-describing each one's behaviour here, which would drift from the binary the
moment either changes.

**Thin, plus next-step guidance — not automation.** Run the command, show what it actually printed,
then name what typically comes next. **Do not take that next step unasked** — if the likely next
command is `spectre validate` or `git add`, say so and stop; do not run it. (`/spectre-new` is a
separate, thicker command for exactly one case — see its own skill.)

## Workflow

1. Take the subcommand and any arguments from the user's `/spectre` invocation verbatim — no
   remapping, no added flags.
2. Run `spectre <subcommand> <args>` (add `--root <path>` only if the user supplied one; every
   spectre command accepts it, and it must come before any positional change id).
3. Print the command's real stdout/stderr and its exit code.
4. Name the likely next step, briefly, based on the exit code and subcommand — e.g. after `init`,
   that `new <change-id>` scaffolds a change; after `new`, that the three generated files need
   filling in (or `/spectre-new` for that); after `validate` reports findings, that they need
   fixing before archiving; after a clean `validate`, that the change is ready to work or archive.
   State it as a suggestion, not an action taken.

## Guardrails

- Never run a subcommand the user didn't ask for.
- Never invent output — only what the binary actually printed.
- Never invent a subcommand, flag, or behaviour spectre does not have; if unsure, run `spectre
  --help` or `spectre <subcommand> --help` and show that instead of guessing.
- Never commit, stage, or `git mv` on the user's behalf — `spectre archive` itself needs the
  change's files already `git add`ed; suggesting that is as far as this command goes.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run any spectre subcommand | `/spectre <subcommand> [args]` |
