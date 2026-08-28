# kan-360-skills-installation-guide

## Why

`README.md` tells the reader that Claude Code users can run `/spectre` and `/spectre-new`, "shipped
in this repository's `.claude/skills/`". That holds only while the agent's working directory is
inside a spectre checkout, because `.claude/skills/` is project-scoped. The documented way to get
spectre is `go install`, which delivers the binary and nothing else — so a reader who follows the
README is promised commands they have no way to invoke.

## What changes

- `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` make this repository its own
  Claude Code plugin marketplace. The manifest's `skills` field points at the existing
  `.claude/skills/` directory, so the in-checkout, project-scoped path keeps working unchanged.
- The two skills become one. `.claude/skills/spectre-new/` is deleted and its scaffold-then-fill
  behaviour folds into `.claude/skills/spectre/SKILL.md` as a named exception to that skill's thin
  rule, keeping the directory name so the bare command stays `/spectre`.
- `README.md`'s Quick start gains a skills-install block beside its `go install` paragraph, split
  by audience: Claude Code installs through the marketplace, every other agent copies the skill
  directory into whatever skills directory it reads. The block states both invocation forms, because
  they differ by route — `/spectre` in a checkout or after a manual copy, `/spectre:spectre` after a
  plugin install, since Claude Code namespaces every plugin skill by its plugin name. The intro
  paragraph's availability claim shrinks to a pointer at that block.
- `plugin_manifest_test.go` asserts the two manifests agree with each other and with the README's
  install command, and that the skill directory the docs name is the one that exists.

No new dependency, and no change to what the `spectre` binary does: the skill already existed and
already worked. What is added is its distribution, the documentation of it, and a test over both.

**Two of the bullets above were not in this proposal when it was written**, and are recorded here
rather than left to the plan alone. The skill merge and the two invocation forms answer a review
finding that the plugin route namespaces skills, so the original wording promised commands the
plugin audience could not invoke; the test answers a review finding that nothing caught four
separate ways the manifests and the README could silently disagree. `design.md` carries both
decisions, `single-spectre-skill` and `manifest-validation-test`, with the alternatives that were
rejected — including a rename that was chosen, begun, and reversed.
