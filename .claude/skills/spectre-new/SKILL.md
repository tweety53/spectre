---
name: spectre-new
description: Scaffold a change with `spectre new <change-id>`, then help fill in the three generated files. Use for /spectre-new.
allowed-tools: Bash(spectre new *), Read, Edit
license: MIT
compatibility: Requires the spectre CLI on PATH (or built locally as `spectre`).
metadata:
  author: tweety53
  version: "1.0"
---

`/spectre-new <change-id>` runs `spectre new <change-id>`, shows its **real** output, then helps
fill in the three files it just scaffolded. This is the one command that goes past showing output
and naming a next step: `new` alone leaves three stubs and a change that reports `no tasks` until a
plan exists, so scaffold-then-fill is the useful unit — see `spectre-new` skill's sibling `/spectre`
for the thin, stop-after-showing-output behaviour that otherwise applies everywhere.

## Workflow

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

- Never invent the required headings or their order — they live only in the README's `## File
  templates` table and must be read from there, not repeated from memory.
- Never run `spectre validate`, `git add`, or any other follow-on command unasked.
- Never invent spectre behaviour the binary does not have.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Scaffold and fill in a new change | `/spectre-new <change-id>` |
