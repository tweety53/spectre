# kan-351-quick-start-guide

Jira: KAN-351 — "Quick start guide at the top of the README".

## Why

- A first-time reader meets `## Install` and `## Getting started` as separate sections and has to
  assemble the sequence themselves. There is no single path from "I have nothing" to "I have a
  validated change".
- The README shows command syntax but never a filled-in change, so nobody can see what a real
  `proposal.md`, `tasks.md` or capability spec looks like. The shape has to be invented.
- The generated files are only loosely enforced. `proposal.md` requires two headings but not its
  title or their order; `tasks.md` has no heading rule at all and passes with zero tasks;
  `design.md` is never read. A template nothing checks is a suggestion.
- `render.Tasks` titles `tasks.md` `# Tasks`, which repeats the filename and disagrees with
  `proposal.md`, whose title is the change id.

## What changes

- A step-by-step quick start becomes the first section of `README.md`, absorbing `## Install` and
  `## Getting started` — global `go install`, then `init`, `new`, fill in, `validate`, work, and
  `archive` — walking the reader through one real change and linking to it at each step.
- `docs/example.md` carries that change in full: every generated file, filled in, for a backend
  that sends only email notifications and is adding SMS, messenger and push.
- `spectre new` also scaffolds `design.md`.
- `render.Tasks` titles `tasks.md` with the change id instead of `# Tasks`.
- The `headings` rule gains ordering and covers `proposal.md`, `tasks.md` and `design.md`;
  `task-sequence` gains "a change has at least one task".

- `AGENTS.md` — this repository's first project instruction file — makes applying the Go skills a
  standing requirement for any agent working here, with a `CLAUDE.md` pointer at it. Added at the
  operator's request during implementation.

## Added during implementation

- The docs are **prompt-first**: the reader hands prompts to an AI coding agent, and the agent runs
  the `spectre` commands. The human is not walked through a terminal session. Requested by the
  operator after reading the first version, which led with commands and carried prompts alongside
  them.

- Two slash commands ship **with spectre**, in the repository's own `.claude/skills/`: `/spectre`,
  which dispatches to any subcommand, and `/spectre-new`, which scaffolds and then helps fill the
  files in. Requested by the operator during implementation, and explicitly part of spectre rather
  than of the flow tooling.

How each of these works, and what was rejected, is in `design.md`.
