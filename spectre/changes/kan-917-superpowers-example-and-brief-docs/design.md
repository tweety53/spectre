## Context

Docs-only change plus one skill file; no Go code. Tests pin parts of the docs: `README.md`'s File
templates rows and install lines (`plugin_manifest_test.go`, `internal/check/check_test.go`), and
`docs/example.md`'s prompt markers and step-2 scaffold fences (`internal/check/check_test.go`,
`internal/cmd/new_test.go`). Every trim keeps those strings byte-identical.

## Decisions

### The superpowers walkthrough is its own page

**ID:** separate-superpowers-example
**Status:** active
**Chosen:** new `docs/superpowers-example.md` — `docs/example.md` stays the plain prompt path whose bodies tests pin.
**Considered:** a second part inside `docs/example.md` — one page twice as long, and the pinned step sections would sit beside unpinned look-alikes.

### The example shows full enriched bodies

**ID:** full-enriched-bodies
**Status:** active
**Chosen:** the enriched `proposal.md`, `design.md`, spec and `tasks.md` in full, each validated by a real `spectre` binary.
**Considered:** prompts plus a description of what the skills add — tells the reader the skills help without showing the difference.

### The skill auto-invokes the superpowers skills when present

**ID:** skill-auto-uses-superpowers
**Status:** superseded by bundle-superpowers-skills
**Chosen:** `new`'s fill step invokes `superpowers:brainstorming` then `superpowers:writing-plans` when available; otherwise recommends them once and fills as today — the same order `/flow` runs.
**Considered:** suggest only, auto-use in a later ticket — the operator asked for the skill to behave as `/flow` already does.
**Superseded because:** the operator asked mid-implementation for the two skills to ship inside the plugin, so they are always available rather than optional.

### Brevity scope is the user-facing docs

**ID:** brevity-scope
**Status:** active
**Chosen:** `README.md` and `docs/*.md` guides.
**Considered:** also `docs/superpowers/**` and `docs/self-review/**` — historical plan and review records, not reader docs; rewriting them falsifies the record.

### Validate accepts writing-plans' step lines

**ID:** steps-are-not-tasks
**Status:** active
**Chosen:** the example's `tasks.md` carries `  - [ ] **Step N: …**` lines two columns under each task, as writing-plans writes them.
**Considered:** flattening steps into prose — unnecessary: a spike (`spectre validate` on a tree with indented steps → `no findings`, `list` → `0/2`) shows spectre counts only column-0 task lines.

### The plugin bundles brainstorming and writing-plans

**ID:** bundle-superpowers-skills
**Status:** active
**Chosen:** copy `skills/brainstorming/` and `skills/writing-plans/` from superpowers v6.4.2 (commit `8ca22dba9a94f28898bbce59f2537ff4d87c747d`, the latest release) byte-identical into `.claude/skills/`, each with superpowers' `LICENSE` and an `UPSTREAM.md` naming the original repository, tag and commit; `new` invokes them unconditionally. `TestPluginManifestsAgree` exempts a skill directory carrying `UPSTREAM.md` from claiming `/spectre:<name>`, and the README's one `cp -r` line copies all three directories.
**Considered:** a reference snapshot under `third_party/` not loaded as skills — still needs a separate install; editing the copies to claim `/spectre:<name>` — they would no longer be the upstream version, and every refresh would re-apply the edit.

## Open questions
