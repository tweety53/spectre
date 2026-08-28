# Review panel — kan-360-skills-installation-guide

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | .claude-plugin/marketplace.json:12 | The plugin description string is duplicated verbatim between plugin.json and marketplace.json — Single Source of Truth, stated twice. |
| F2 | Bugbot | Minor | .claude-plugin/plugin.json:1 | Four surviving mutants: nothing validates that plugin.json, marketplace.json and the README install command agree, so a rename or typo in any of them ships silently. |
| F3 | Code review (low) | Minor | README.md:29 | The manual-install cp command assumes ~/.claude/skills/ already exists; with two sources cp requires an existing destination directory, so it fails outright for a reader who has never installed a skill. |
| F4 | Primary | Critical | README.md:18 | Plugin skills are namespaced, so after /plugin install spectre@spectre the commands are /spectre:spectre and /spectre:spectre-new, not /spectre and /spectre-new as the README and its intro sentence both claim. |
| F5 | Primary | Minor | .claude-plugin/plugin.json:4 | Both manifest description fields name the skills as /spectre and /spectre-new, carrying the unnamespaced form into the plugin metadata. |
| F6 | Primary | Important | spectre/changes/kan-360-skills-installation-guide/tasks.md:33 | tasks.md step bodies still show the pre-merge two-skill text, and no task records commit 2d412a6, so the plan misdescribes what the branch actually did. |
| F7 | Principles | Important | internal/check/plugin_manifest_test.go:16 | repoRoot re-derives how deep internal/check sits below the repository root, a fact docPaths already encodes; a package move would fix only one of them. |
| F8 | Principles | Important | internal/check/plugin_manifest_test.go:1 | The test calls none of package check logic and asserts on repository metadata, where the packages existing README test binds back to checks own heading rules — cohesion. |
| F9 | Bugbot | Important | internal/check/plugin_manifest_test.go:88 | skills path resolves only requires some directory holding some SKILL.md, so repointing skills at an unrelated directory with a stub still passes. |
| F10 | Bugbot | Important | internal/check/plugin_manifest_test.go:88 | Nothing ties the real skill directory name to the README manual-install cp command or to the /spectre:spectre form the skill claims, so renaming it breaks both user-facing instructions with a green suite. |
| F11 | Bugbot | Minor | internal/check/plugin_manifest_test.go:69 | marketplace source resolves indexes Plugins[0] unguarded, so an empty plugins array panics and crashes the test binary instead of failing. |
| F12 | Bugbot | Minor | internal/check/plugin_manifest_test.go:69 | marketplace source resolves only requires the target to contain a plugin.json, not that it is this repository own manifest. |
| F13 | Bugbot | Important | plugin_manifest_test.go:159 | Both name-agreement checks are unanchored substring matches over whole files: a stale /spectre:spectre satisfies a check for /spectre:spec by prefix, and a decoy mention anywhere in README satisfies the cp -r check while the real command is broken. |
| F14 | Principles | Minor | plugin_manifest_test.go:187 | KISS: the hand-rolled occurrence walk and isDirNameChar classifier reimplement what a two-line stdlib regexp match expresses, and regexp is allowed under the standard-library-only policy. |
| F15 | Primary | Important | plugin_manifest_test.go:196 | Round 3 applied the directory-name boundary check to the invoke match but not to cpTarget, so a corrupted source path that keeps the real directory name as a prefix still passes. |
| F16 | Bugbot | Important | plugin_manifest_test.go:114 | The install-command check is an unanchored Contains, so /plugin install spectre@spectrecorp satisfies a check for /plugin install spectre@spectre by prefix. |
| F17 | Bugbot | Minor | plugin_manifest_test.go:170 | The cp line is selected as the first line containing cp -r in document order, so an unrelated cp -r example added earlier fails the subtest on a correct change. |
| F18 | Primary | Minor | spectre/changes/kan-360-skills-installation-guide/proposal.md:19 | proposal.md still says every other agent copies the skill directories plural, and that the change touches no Go code — both untrue after the skill merge and the added test, and unlike tasks.md and design.md it carries no reconciliation. |
| F19 | Primary | Minor | plugin_manifest_test.go:1 | skills path resolves is close to subsumed by skills path names the real skill directory plus the nonempty check in the name-agreement subtest; it could be dropped without losing coverage. |

findings-total: 19
finding-status: F1 withdrawn schema-forced duplication: plugin.json and marketplace.json are read by different consumers and each mandates its own description field, so the text cannot be stated once
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed

reproducers-total: 19
finding-reproducer: F1 grep -n "spec-driven change tracking in markdown" .claude-plugin/plugin.json .claude-plugin/marketplace.json
finding-reproducer: F2 none — mutating the skills path, the marketplace source, the plugin name and the README install string one at a time each left gofmt, go vet and go test green; the repo has no CI workflows and no check covers these files
finding-reproducer: F3 rm -rf /tmp/dest && mkdir -p /tmp/src/a /tmp/src/b && cp -r /tmp/src/a /tmp/src/b /tmp/dest/
finding-reproducer: F4 none — verified against https://code.claude.com/docs/en/discover-plugins which states verbatim that plugin skills are namespaced by the plugin name; no local command reproduces it without mutating the operator user settings
finding-reproducer: F5 none — same root cause as F4; the description text is shown by claude plugin details and carries the unnamespaced names
finding-reproducer: F6 none — a plan-provenance inconsistency in a planning artifact, not a code defect; no runnable command demonstrates it
finding-reproducer: F7 grep -n runtime.Caller internal/check/plugin_manifest_test.go internal/check/check_test.go
finding-reproducer: F8 sed -n 1,3p internal/check/check.go
finding-reproducer: F9 none — requires repointing the skills field and planting a stub SKILL.md before running go test; no single bare command demonstrates it
finding-reproducer: F10 none — requires renaming the skill directory before running go test; no single bare command demonstrates it
finding-reproducer: F11 none — requires emptying the marketplace plugins array before running go test; no single bare command demonstrates it
finding-reproducer: F12 none — informational, the same shallow-field pattern as F9
finding-reproducer: F13 none — requires renaming the skill directory and editing README.md before running go test; no single bare command demonstrates it
finding-reproducer: F14 go test ./... -run TestPluginManifestsAgree
finding-reproducer: F15 none — requires editing the cp -r line in README.md before running go test; pending re-verification against a clean tree, the first attempt was contaminated by a concurrent agent mutation
finding-reproducer: F16 none — requires editing the install command in README.md before running go test
finding-reproducer: F17 none — requires inserting an unrelated cp -r section earlier in README.md before running go test
finding-reproducer: F18 none — prose staleness in a planning file that is not part of the seven-commit diff
finding-reproducer: F19 none — a proportionality judgment the raising slot explicitly stated is not a defect
