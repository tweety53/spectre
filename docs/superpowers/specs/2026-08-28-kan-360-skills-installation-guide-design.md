# kan-360-skills-installation-guide — design

**Jira:** KAN-360 — Ship spectre's Claude Code skills as an installable plugin

## Problem

`README.md` tells the reader that Claude Code users can run `/spectre` and `/spectre-new`,
"shipped in this repository's `.claude/skills/`". That sentence is true only while the agent's
working directory is inside a spectre checkout, because `.claude/skills/` is project-scoped.
The documented way to get spectre is `go install`, which delivers the binary and nothing else, so
the reader who follows the README ends up with a promised command they cannot invoke.

## Approach

Make the repository its own Claude Code plugin marketplace, and split the Quick start's install
guidance in two: Claude Code through the marketplace, every other agent by hand.

### Considered and rejected

- **Move the skills to the plugin-default `skills/` directory and drop the manifest's `skills`
  field.** One less line of configuration, but it breaks the project-scoped path that works today
  for anyone with the repository checked out, and rewrites the README's existing link for no gain.
- **No plugin; document a manual copy for every agent including Claude Code.** Simplest to write,
  but it gives Claude Code users no update path, and a marketplace is what was asked for.
- **Have `spectre init` write the skills into the target repository.** Rejected: it couples a
  harness-agnostic CLI to one vendor's directory layout, and contradicts the principle that the
  tree on disk is the entire state.
- **Name `~/.cursor/skills/` and `~/.codex/skills/` explicitly in the manual path.** More useful
  if correct, but neither is confirmed as a documented, supported directory in either product, and
  a public README must not ship an instruction that may not work. The manual path names Claude
  Code's directory as the worked example and otherwise says "your agent's own skills directory".

## Design

### Manifests

Two files under `.claude-plugin/`:

- `plugin.json` — the plugin manifest. Its `skills` field points at the existing `.claude/skills/`
  directory, so no skill file moves and the in-checkout, project-scoped path keeps working exactly
  as it does today.
- `marketplace.json` — a single-entry catalogue whose one plugin's `source` is the repository
  itself.

An install and a checkout therefore serve the same two skill directories; there is no second copy
to drift.

### README

The Quick start's `go install` paragraph is followed by a skills-install block, because installing
the skills belongs with installing the binary rather than after the walkthrough. The block is split
exactly as the two audiences differ:

- **Claude Code** — `/plugin marketplace add tweety53/spectre`, then
  `/plugin install spectre@spectre`.
- **Any other agent** — copy the two skill directories into that agent's own skills directory,
  with Claude Code's `~/.claude/skills/` shown as the worked example and no claim made about any
  other product's path.

One line records that the skills only dispatch to the binary, so `spectre` must still be on `PATH`.

The intro paragraph's existing sentence shrinks from a claim that the skills are available to a
pointer at the block.

## Verification

- Both manifests parse as JSON.
- `claude plugin marketplace add ./`, then `claude plugin install spectre@spectre`, then
  `claude plugin details spectre` reports both skills; then uninstall and remove the marketplace so
  no local state is left behind.
- `gofmt -l .`, `go vet ./...` and `go test ./...` stay clean — no Go code is touched.
- README prose stays within the width the file already uses.

## Out of scope

`docs/example.md` and `docs/terminal.md` address "an AI coding agent" generically and mention no
harness or skill, so neither needs a corresponding edit.
