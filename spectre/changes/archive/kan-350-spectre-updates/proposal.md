# kan-350-spectre-updates

Jira: KAN-350 — "Spectre updates".

## Why

- `spectre` cannot create its own tree. The README's getting-started section is a `mkdir` recipe
  the tool could run itself, and the error every command emits when no tree exists names no way
  forward.
- `migrate` is a one-corpus OpenSpec converter carrying ~960 lines, and it is the binary's only
  exception to `--root`.
- The README carries reference material at a length that buries what a first reader needs.
- The peer-tree fixtures name a real project, in the README and in ~85 test-file occurrences.

## What changes

- `spectre init [--root <path>]` scaffolds `specs/`, `changes/` and a `config.md` whose defaults
  are present but inert. Idempotent; never overwrites.
- `spectre migrate`, `internal/cmd/migrate.go` and `internal/openspec/` are deleted outright, with
  their tests and README section. `--root` becomes uniform across every command.
- The README keeps install, the tree, commands, exit codes and `archive`; its three reference
  sections move to `docs/spec-format.md`, `docs/configuration.md` and `docs/references.md`.
- `gymie` / `gymie-frontend` become `web` / `web-frontend` throughout code and README.
- This repository gains its own `spectre/` tree, and `.gitignore` stops ignoring the path that
  tree occupies.

How each of these works, and what was rejected, is in `design.md`.
