# kan-917-superpowers-example-and-brief-docs

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — the doc trims move and merge existing prose between sections and files.

**Goal:** name superpowers brainstorming + writing-plans as the recommended way to fill spectre's
artifacts, show it in a second worked example, make the `spectre` skill use them, and make every
user-facing doc brief.

## Global Constraints

- Docs plus `.claude/skills/`; the only Go change is task 6's exemption in `plugin_manifest_test.go`.
- Never lose a fact: every rule, exit code, refusal, flag, grammar line and worked example in a
  doc survives the trim (reworded at most once, never duplicated across files).
- Test-pinned strings stay byte-identical: `README.md`'s `| \`proposal.md\` |`, `| \`tasks.md\` |`,
  `| \`design.md\` |`, `| \`specs/<capability>.md\` |` rows' heading columns; `/plugin install
  spectre@spectre`; `mkdir -p ~/.claude/skills` before exactly one `cp -r` line into
  `~/.claude/skills` naming `.claude/skills/spectre`; `docs/example.md`'s `## 2. Scaffold the
  change` section with its three labelled ```` ```markdown ```` fences, and its four
  "It must have exactly these headings, in this order:" / "The first line must be exactly `"
  prompt markers.
- Brief: bullets and tables over prose, no preamble, no recap, each thing stated once with a link
  elsewhere.
- Checks, all clean: `gofmt -l .`, `go vet ./...`, `go test ./...`.

## Review Focus

- A reader following `docs/superpowers-example.md` without superpowers installed — the page says
  what to do instead (the plain path in `docs/example.md`).
- A trimmed doc that drops a refusal, exit code or grammar rule — compare against the base
  section by section.
- A broken relative link or anchor after headings change — every `](…#…)` resolves.
- The skill auto-invoking superpowers on a subcommand other than `new` — it must not.
- A claimed-validated body in the new example that `spectre validate` actually rejects.

- [x] 1. Make the `spectre` skill use superpowers when filling `new`'s stubs

**Build:** green
**After:** none
**Files:** `.claude/skills/spectre/SKILL.md`
**Tests:** none — skill prose; `TestPluginManifestsAgree` already pins the skill's invocation line.
**Regression:** none
**Baseline:** before=176 after=176
<!-- measured: grep -rh '^func Test' --include='*_test.go' . | wc -l @ merge-base 6802cb0 -->
**Commit:** docs(skill): fill new's stubs with superpowers brainstorming and writing-plans

**Decision:** skill-auto-uses-superpowers

  - [x] **Step 1: Rewrite `### \`new\` — the one exception` steps 2–3.** After reading the three
    stubs: if `superpowers:brainstorming` is available, invoke it to settle the design with the
    user and write `proposal.md`, `design.md` and any capability spec from the approved design;
    then, if `superpowers:writing-plans` is available, invoke it to enrich `tasks.md` in spectre's
    task shape (column-0 `- [ ] <n>. <title>` task lines, `  - [ ] **Step N: …**` steps indented
    two columns beneath, flat integer ids). If either skill is missing, say once that installing
    superpowers is the recommended path, link `docs/superpowers-example.md`, and fill that file by
    hand as today, following `docs/example.md`. Keep step 4 (name `spectre validate`, don't run it).
  - [x] **Step 2: Scope it.** State that only `new` invokes these skills; every other subcommand
    stays dispatch-only. Update the frontmatter `description` to mention it in one clause.
  - [x] **Step 3: Trim the rest of the file** to the same brevity rule, keeping every guardrail
    and the `/spectre:spectre` invocation line.
  - [x] **Step 4: Verify.** `gofmt -l .` (expect no output), `go test -run 'TestPluginManifestsAgree' .`
    (expect `ok`).
  - [x] **Step 5: Commit** with the subject above.

Correction (2026-10-07): the plan left `allowed-tools` unchanged; shipped `Bash(spectre:*), Read, Edit, Write, Skill` — `Skill` to invoke the two skills, `Write` because a capability spec is a new file `Edit` cannot create.

- [x] 2. Add the superpowers-enriched example

**Build:** green
**After:** none
**Files:** `docs/superpowers-example.md`
**Tests:** none — prose; bodies validated by hand in step 3.
**Regression:** none
**Baseline:** before=176 after=176
<!-- measured: grep -rh '^func Test' --include='*_test.go' . | wc -l @ merge-base 6802cb0 -->
**Commit:** docs(example): add the superpowers-enriched multi-channel-notifications walkthrough

**Decision:** separate-superpowers-example
**Decision:** full-enriched-bodies
**Decision:** steps-are-not-tasks

  - [x] **Step 1: Write the page.** Title `# Example: multi-channel-notifications, with superpowers`.
    One-paragraph intro: same change as `docs/example.md`, filled by `superpowers:brainstorming`
    and `superpowers:writing-plans` instead of hand-written prompts; without superpowers, use
    `docs/example.md`. Sections: `## 1. Create the tree and scaffold the change` (prompt + link to
    example.md steps 1–2); `## 2. Brainstorm the design` (the prompt invoking brainstorming, a
    3–5 line excerpt of the questions it asks, then the full enriched `specs/notifications.md`,
    `proposal.md` and `design.md` — design carrying `## Decisions` entries with `**ID:**`,
    `**Status:**`, `**Chosen:**`, `**Considered:**` and an `## Open questions` section);
    `## 3. Write the plan` (the prompt invoking writing-plans, naming spectre's task shape, then
    the full enriched `tasks.md`: 5 tasks matching example.md's, each with `**Files:**`, a test
    line, and `  - [ ] **Step N: …**` steps ending in a verify command and a commit);
    `## 4. Validate`; `## 5. Work and archive` (link to example.md steps 8–9). A short
    "What the skills added" bullet list after each body comparing it with example.md's version.
  - [x] **Step 2: Note `/flow` and the `spectre` skill** do steps 2–3 automatically when the
    skills are installed (one line).
  - [x] **Step 3: Validate every body for real.** In a scratch git repo: `go build -o <scratch>/spectre
    ./cmd/spectre`, `spectre init`, `spectre new multi-channel-notifications`, write each fenced
    body to its path, `spectre validate` (expect `no findings`, exit 0) and `spectre list`
    (expect `multi-channel-notifications  0/5`). Fix any body that fails, then re-run.
  - [x] **Step 4: Verify.** `gofmt -l .` (expect no output).
  - [x] **Step 5: Commit** with the subject above.

Correction (2026-10-07): the change files cite requirements as plain `(R3)`, not `(@notifications#R3)` — spectre reads citations only in spec files, so `@` forms in a proposal or plan are neither validated nor found by `refs`.

- [x] 3. Rewrite the README short, with the superpowers recommendation

**Build:** green
**After:** Task 2
**Files:** `README.md`
**Tests:** none — `TestPluginManifestsAgree` and `TestDocumentedHeadingsMatchChecker` already pin it.
**Regression:** none
**Baseline:** before=176 after=176
<!-- measured: grep -rh '^func Test' --include='*_test.go' . | wc -l @ merge-base 6802cb0 -->
**Commit:** docs(readme): lead with the superpowers pairing and cut restatement

**Decision:** brevity-scope

  - [x] **Step 1: Restructure** to: one-line pitch → `## Install` (binary; skill via plugin
    marketplace or manual copy — pinned lines kept) → `## Recommended: pair with superpowers`
    (2–4 bullets: brainstorming fills proposal/design/spec, writing-plans enriches tasks.md;
    `/flow` and the `spectre` skill invoke them automatically when installed; link
    `docs/superpowers-example.md`) → `## Quick start` (the six steps as a short list linking
    example.md anchors; the `--root`-before-positional rule once) → `## The tree` → `## File
    templates` (table rows unchanged) → `## Commands` (table + exit codes) → `## archive`
    refusals → `## Further reading` (one bullet per doc) → `## What spectre does not do`.
  - [x] **Step 2: Keep every fact** from the base README (diff it section by section).
  - [x] **Step 3: Verify.** `gofmt -l .` (expect no output), `go test -run
    'TestPluginManifestsAgree|TestDocumentedHeadingsMatchChecker' ./...` (expect `ok` for `.` and
    `internal/check`).
  - [x] **Step 4: Commit** with the subject above.

- [x] 4. Trim the plain walkthrough and the terminal guide

**Build:** green
**After:** Task 3
**Files:** `docs/example.md`, `docs/terminal.md`
**Tests:** none — `TestExampleDocMatchesScaffold` and `TestDocumentedHeadingsMatchChecker` already pin example.md.
**Regression:** none
**Baseline:** before=176 after=176
<!-- measured: grep -rh '^func Test' --include='*_test.go' . | wc -l @ merge-base 6802cb0 -->
**Commit:** docs(example): cut restatement from the walkthrough and terminal guide

**Decision:** brevity-scope

  - [x] **Step 1: example.md.** Shorten the intro to 2–3 lines plus a pointer to
    `docs/superpowers-example.md` as the recommended path; collapse each "Check by hand"
    paragraph to one or two bullets; keep every prompt, every fenced body and every `## N.`
    heading verbatim (anchors are linked from README and terminal.md).
  - [x] **Step 2: terminal.md.** Shorten the intro; drop restated rules that README states
    (link instead); keep every transcript, exit code and the archive refusal sequence.
  - [x] **Step 3: Verify.** `gofmt -l .` (expect no output), `go test -run
    'TestExampleDocMatchesScaffold|TestDocumentedHeadingsMatchChecker' ./...` (expect `ok`).
  - [x] **Step 4: Commit** with the subject above.

- [x] 5. Trim the reference docs

**Build:** green
**After:** none
**Files:** `docs/links.md`, `docs/references.md`, `docs/configuration.md`, `docs/spec-format.md`
**Tests:** none — prose only.
**Regression:** none
**Baseline:** before=176 after=176
<!-- measured: grep -rh '^func Test' --include='*_test.go' . | wc -l @ merge-base 6802cb0 -->
**Commit:** docs(reference): make the links, references, config and spec-format pages brief

**Decision:** brevity-scope

  - [x] **Step 1: links.md.** Turn grammar prose into per-section bullet rules, the validate
    checks and six refusals into lists; keep the worked example's commands and output, cut its
    repeated explanation; keep every rejection rule and the `--force` scope.
  - [x] **Step 2: references.md, configuration.md, spec-format.md.** Cut restatement; keep every
    rule (closed keys, duplicate-key error, fenced-block ignore, id-prefix replacement, peer path
    resolution, alias legality).
  - [x] **Step 3: Check links.** Every relative `](…)` link and `#anchor` across `README.md` and
    `docs/*.md` resolves to an existing file and heading.
  - [x] **Step 4: Run the full checks.** `gofmt -l .`, `go vet ./...`, `go test ./...` — all clean.
  - [x] **Step 5: Commit** with the subject above.

- [x] 6. Bundle superpowers' brainstorming and writing-plans skills in the plugin

**Build:** green
**After:** Task 1, 2, 3, 5
**Files:** `.claude/skills/brainstorming/LICENSE`, `.claude/skills/brainstorming/SKILL.md`, `.claude/skills/brainstorming/UPSTREAM.md`, `.claude/skills/brainstorming/scripts/frame-template.html`, `.claude/skills/brainstorming/scripts/helper.js`, `.claude/skills/brainstorming/scripts/server.cjs`, `.claude/skills/brainstorming/scripts/start-server.sh`, `.claude/skills/brainstorming/scripts/stop-server.sh`, `.claude/skills/brainstorming/spec-document-reviewer-prompt.md`, `.claude/skills/brainstorming/visual-companion.md`, `.claude/skills/writing-plans/LICENSE`, `.claude/skills/writing-plans/SKILL.md`, `.claude/skills/writing-plans/UPSTREAM.md`, `plugin_manifest_test.go`, `README.md`, `.claude/skills/spectre/SKILL.md`, `docs/superpowers-example.md`, `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`
**Tests:** `TestPluginManifestsAgree` — the vendored directories, exempted by `UPSTREAM.md`, and the README `cp -r` line naming all three.
**Regression:** `TestPluginManifestsAgree` fails with the vendored directories present and the exemption absent.
**Baseline:** before=176 after=176
<!-- measured: grep -rh '^func Test' --include='*_test.go' . | wc -l @ merge-base 6802cb0 -->
**Commit:** feat(skills): bundle superpowers brainstorming and writing-plans

**Decision:** bundle-superpowers-skills

  - [x] **Step 1: Copy upstream.** From the v6.4.2 tarball, `skills/brainstorming/` (with
    `scripts/`) and `skills/writing-plans/` into `.claude/skills/`, byte-identical; add superpowers'
    `LICENSE` and an `UPSTREAM.md` (repository URL, tag, commit, "unmodified; refresh by
    re-copying") to each.
  - [x] **Step 2: RED.** `go test -run TestPluginManifestsAgree .` — fails: the `cp -r` line and the
    `/spectre:<name>` claim.
  - [x] **Step 3: GREEN.** Exempt a skill directory holding `UPSTREAM.md` from the invocation claim
    (never from the `cp -r` check); widen the README's one `cp -r` line to all three directories.
  - [x] **Step 4: Rewire the docs.** README's superpowers section says the plugin bundles both
    skills with attribution; the `spectre` skill's `new` invokes them without the missing-skill
    branch beyond a one-line hand-fill fallback; `docs/superpowers-example.md`'s prompts name the
    bundled skills; the plugin descriptions mention them.
  - [x] **Step 5: Verify.** `gofmt -l .`, `go vet ./...`, `go test ./...` — all clean.
  - [x] **Step 6: Commit** with the subject above.

Correction (2026-10-07): review added a step the plan lacked — `.claude-plugin/plugin.json`'s `version` bumped `1.0.0` → `1.1.0`, since an unchanged version keeps existing plugin installs on the cached copy without the bundled skills. Panel review also dropped the `/flow` mentions task 2 step 2 and task 3 step 1 asked for: `/flow` is defined nowhere a public reader can follow.
