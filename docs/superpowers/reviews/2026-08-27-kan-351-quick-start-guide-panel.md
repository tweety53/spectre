# Review panel — kan-351-quick-start-guide

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | internal/cmd/new.go:25 | The design.md scaffold body exists twice with nothing binding the copies: as the compiled Go string _designTemplate in internal/cmd/new.go, and as quoted output in docs/example.md. Editing the template leaves the docs page silently stale and every check passes. The same shape pre-exists for _proposalTemplate, but this change doubles it and adds no guard. Single Source of Truth. |
| F2 | Primary | Minor | spectre/changes/kan-351-quick-start-guide/design.md:100 | design.md was corrupted by the parent orchestrator, not by any task. Section 5 contains the sentence "A change is free to add `## Open questions`, `## Testing`, or numbered sections" — two later scripted edits anchored on those inline-code mentions rather than the real headings, splicing the placeholders open-question entry and the whole of section 6 into the middle of that sentence. Raised by the Primary panel slot as an awareness note because the file is untracked; recorded as a finding regardless, since the integrate phase commits spectre/changes/ and the artifact ships. |
| F3 | Primary | Minor | spectre/changes/kan-351-quick-start-guide/design.md:101 | design.md corrupted a SECOND time by the parent orchestrator, by the same mechanism as F2 and after F2 was fixed. Section 5 contains the sentence "a change is free to add ## Open questions, ## Testing, or numbered sections"; later scripted edits anchored on the bare string ## Testing, whose first occurrence is inside that sentence, splicing sections 7 and 8 into it and displacing section 6. Raised independently by both the Primary and Principles panel slots. |
| F4 | Principles | Major | internal/check/check.go:159 | The required-heading-order knowledge is duplicated three ways with nothing binding it: the want literals in ProposalFindings/TaskFindings/DesignFindings/SpecFindings, the README File templates table, and docs/example.md prompt text stating exactly these headings, in this order. Task 11 guarded the scaffold BODIES against exactly this drift; the headings fact was left unguarded. Change a want literal and both documents silently go false with every check green. |
| F5 | Principles | Minor | internal/check/check.go:133 | designHeadings is the only unexported package-level var in the repository not prefixed with an underscore. Its two siblings in the same file, _placeholders and _wellFormedTask, both carry it, as does _taskRe in internal/parse. Introduced by task 16 when the inline heading literals were lifted out. Verified independently by the parent across the whole repository, not only in the reviewed file. |
| F6 | Primary | Minor | spectre/changes/kan-351-quick-start-guide/design.md:107 | The design-optional-but-checked decision claims the README AND docs/spec-format.md described design.md as optional, unparsed. Only the README ever did; docs/spec-format.md has never mentioned design.md at all. Task 7 discovered this during implementation and the task Files list was corrected, but the decision text that carried the wrong assumption was not. A decision record stating a false fact about the codebase is the same defect class as a stale doc. |
| F7 | Primary | Major | .claude/skills/spectre/SKILL.md:8 | Both new skills carry metadata author: gymie, copied from the house style of the agents repository where that name is legitimate. KAN-350 deliberately removed every occurrence of gymie from this repository as one of its four stated goals; this reintroduces it in two shipped, committed files. Found by the parent while reading the skills, not by a review slot. |
| F8 | Principles | Minor | .claude/skills/spectre-new/SKILL.md:4 | allowed-tools grants Write alongside Read and Edit, but spectre new pre-creates proposal.md, tasks.md and design.md, so the skill only ever edits existing content. Write additionally permits creating arbitrary new files, which the documented workflow never does. Narrow to Bash(spectre:*), Bash(git:*), Read, Edit. |
| F9 | Primary | Minor | .claude/skills/spectre/SKILL.md:4 | Both skills grant Bash(git:*) in allowed-tools, but neither workflow runs any git command. Both guardrails explicitly forbid staging, committing or git mv on the user behalf, so the grant permits precisely what the skills promise not to do. Raised as an observation by the task-17 reviewer rather than as a finding; recorded as one here because it is the same least-privilege class as F8 and should be settled with it. |
| F10 | Principles | Minor | .claude/skills/spectre-new/SKILL.md:4 | allowed-tools grants Bash(spectre:*) though the documented workflow only ever runs spectre new. Narrowing to Bash(spectre new:*) would cover every documented invocation including --root. Least privilege. Flagged by the reviewer as pre-existing rather than introduced by the permission fix, and non-blocking. |

findings-total: 10
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed

reproducers-total: 10
finding-reproducer: F1 none — demonstrated by mutation rather than a single command: edit the placeholder text inside _designTemplate, then run gofmt -l . and go vet ./... and go test ./... — all stay green while docs example.md still shows the old body
finding-reproducer: F2 none — a prose corruption in a change artifact, confirmed by reading the raw file: a stray heading line reading "## Testing`, or" and an open-question entry sitting inside section 5
finding-reproducer: F3 none — a prose corruption in an untracked change artifact; confirmed by grep -n on the file heading order showing sections out of sequence and a malformed heading
finding-reproducer: F4 grep -n "exactly these headings" docs/example.md README.md — three independent statements of the required heading order, none checked against check.go
finding-reproducer: F5 grep -rn "^var [a-z]" internal/*/*.go cmd/*/*.go | grep -v _test — designHeadings is the only hit; grep -rn "^var _" internal/*/*.go shows three siblings following the convention, two in the same file
finding-reproducer: F6 git show 6271d41:docs/spec-format.md | grep -c design — returns 0, so the file never mentioned design.md either before or after the change
finding-reproducer: F7 grep -rn gymie .claude/ — two hits, in files this change adds and commits
finding-reproducer: F8 none — a static tool-permission scope, verified by reading .claude/skills/spectre-new/SKILL.md steps 2 to 4 against internal/cmd/new.go lines 76 to 88, where spectre new already writes all three files before the skill acts
finding-reproducer: F9 grep -n "git" .claude/skills/spectre/SKILL.md .claude/skills/spectre-new/SKILL.md — every mention is a guardrail forbidding git actions or prose naming git add as a next step, never a workflow step that runs one
finding-reproducer: F10 grep -n "spectre " .claude/skills/spectre-new/SKILL.md — only spectre new appears as an invoked command, yet the grant covers every subcommand
