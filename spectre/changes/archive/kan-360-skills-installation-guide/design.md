## Context

The narrative design — the problem, the approaches weighed, and the tradeoff that ruled each
rejected one out — is `docs/superpowers/specs/2026-08-28-kan-360-skills-installation-guide-design.md`.
This file records the decisions under stable ids and the question left open.

The plugin mechanism was exercised against this working tree before planning: adding the tree as a
local marketplace, installing from it, and reading the inventory back reported both skills present.

```text verified:ran against this working tree, then uninstalled and removed the marketplace
$ claude plugin details spectre
spectre 1.0.0
  Component inventory
    Skills (2)  spectre, spectre-new
```

## Decisions

### Where the plugin's skills live

**ID:** skills-stay-in-claude-skills
**Status:** active
**Chosen:** keep the skills at `.claude/skills/` and point the manifest's `skills` field at that
directory — an install and a checkout then serve the same files, with no second copy to drift.
**Considered:** moving them to the plugin-default `skills/` and dropping the manifest field — one
less line of configuration, but it breaks the project-scoped path that works today for anyone with
the repository checked out, and rewrites the README's existing link for no gain.

### How the skills are distributed

**ID:** repo-is-own-marketplace
**Status:** active
**Chosen:** this repository is its own marketplace, so one `/plugin` pair installs the skills and a
later `/plugin marketplace update` refreshes them.
**Considered:** documenting a manual copy for every agent including Claude Code — simplest to
write, but it leaves Claude Code users no update path. Having `spectre init` write the skills into
the target repository — rejected because it couples a harness-agnostic CLI to one vendor's
directory layout, and contradicts the principle that the tree on disk is the entire state.

### What the manual path names

**ID:** generic-non-claude-path
**Status:** active
**Chosen:** name Claude Code's `~/.claude/skills/` as the worked example and otherwise say "your
agent's own skills directory", claiming nothing about any other product's path.
**Considered:** naming `~/.cursor/skills/` and `~/.codex/skills/` explicitly — more useful if
correct, but neither is confirmed as a documented, supported directory in either product, and a
public README must not ship an instruction that may not work.

### Where the block sits in the README

**ID:** quick-start-placement
**Status:** active
**Chosen:** immediately after Quick start's `go install` paragraph — installing the skills is part
of getting set up, so it belongs beside installing the binary rather than after the walkthrough.
**Considered:** a subsection at the end of Quick start, which keeps the fast path fast but
separates the two halves of one install; and expanding the intro paragraph in place, which is the
smallest diff but puts a multi-line command block inside a prose paragraph.

### What the skills are called, now that the plugin route namespaces them

**ID:** skills-renamed-for-namespace
**Status:** superseded by single-spectre-skill
**Chosen:** rename the skill directories to `run` and `new`, so the plugin route reads
`/spectre:run` and `/spectre:new`. Claude Code namespaces every plugin skill by its plugin name —
`https://code.claude.com/docs/en/discover-plugins` states it verbatim — so the pre-rename names
would have produced `/spectre:spectre` and `/spectre:spectre-new`.
**Considered:** shortening only `spectre-new`, which keeps `/spectre:spectre` stuttering as the
namesake entry point; and renaming the plugin to `spec` instead, which leaves the bare names
untouched but still repeats the word and makes the install line `spec@spectre`. The operator chose
the reading of the plugin form over the reading of the bare form, having been shown that a skill's
directory name is both — the bare command in a checkout and the suffix under the plugin — so every
rename that improves one changes the other by the same string.
**Consequence, accepted knowingly:** in a checkout the commands become `/run` and `/new`, which are
generic enough to collide with another installed skill and say nothing about spectre to a reader.

### Whether anything checks the manifests and the README agree

**ID:** manifest-validation-test
**Status:** active
**Chosen:** add a Go test asserting that `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`
and the README's install command agree, and that the `skills` path resolves to real skill
directories.
**Considered:** leaving it, on the grounds that this is a documentation change. Rejected by the
operator after the panel's Bugbot slot mutated the `skills` path, the marketplace `source`, the
plugin `name` and the README's install string one at a time and found all four survived the full
suite — nothing in the repository catches any of them.

### One skill, not two, and the bare name kept

**ID:** single-spectre-skill
**Status:** active
**Chosen:** merge `spectre-new` into `spectre`, leaving one skill. Invoked `/spectre <whatever>` in
a checkout or after a manual copy, and `/spectre:spectre` after a plugin install.
**Considered:** the rename `skills-renamed-for-namespace` chose, which this supersedes. It was
selected to make the plugin route read better, on the understanding that a better suffix was
available. It is not: namespacing applies to the suffix whatever the suffix is, so `run`/`new` still
yields `/spectre:run` and never the bare `/spectre` it was chosen for, while costing the bare names
in every context where the bare form does work. Also considered: merging but keeping the thin rule
strict, so `new` would print its stubs and stop — rejected because scaffold-then-fill is the useful
unit, which is the whole reason the second skill existed.
**Why one skill is possible at all:** `/spectre` already dispatches to every subcommand, `new`
included, so the second skill was never carrying a subcommand the first could not reach — only a
thicker behaviour for one of them. That behaviour survives the merge as a named exception to the
thin rule rather than as a separate skill.
**Cost, accepted knowingly:** the two files split thin-dispatch from scaffold-then-fill explicitly,
and each cites the other for it. One file now carries both, so its thin rule needs a stated
exception for `new` instead of a sibling to point at.

### Where the manifest test lives

**ID:** manifest-test-at-root
**Status:** active
**Chosen:** a package-less test at the repository root, `package spectre` in `plugin_manifest_test.go`.
Probed before committing to it: `go build ./...`, `go vet ./...`, `go test .` and `gofmt -l .` are
all clean with a root package holding only a test file.
**Considered:** leaving it in `internal/check`, whose doc comment scopes the package to spectre's
validation rules — the panel's principles slot showed that package's existing README test binds back
to `check`'s own heading rules while this one calls none of its logic; and a new `internal/plugincheck`
package, which restores cohesion but invents a package holding one test file and no production code.
**A second finding dissolves with this one:** the panel also raised that the test's `repoRoot` helper
re-derived how deep `internal/check` sits below the repository root, a fact `docPaths` already
encoded. At the root there is no depth to encode — the helper is `filepath.Dir(thisFile)` — so the
duplicated fact stops existing rather than being deduplicated.

## Open questions

### Do Cursor and Codex read a skills directory of their own?

**ID:** other-harness-skill-paths
**Status:** open
**Why it is open:** deferred by the operator in favour of wording that is correct either way.
Neither product is confirmed to document a supported skills directory, and this change had no
reason to find out.
**What it affects:** whether the manual path can later name concrete per-agent directories instead
of the generic phrasing `generic-non-claude-path` chose. An answer changes wording only; nothing
in the manifests depends on it.
