# Review panel — kan-360-skills-installation-guide

## Pass 1 (round 0)

**Mode:** initial panel — the resolved roster, in full.
**Roster resolved from the settings store:** `primary`, `principles`, `code-review-low`, `bugbot`.
**Operator additions:** no addition this round — the resolved list ran alone.
**Diff read by every diff-reading slot:** `.superpowers/sdd/final-review.diff`, written from
`git diff 6181228` (merge base), 84 lines.
**Diff size guard:** `check-panel-diff-size.sh` measured 57 and reported `under cap`, exit 0. No
operator question was raised.
**Substituted slots:** `bugbot`. This harness offers no `bugbot` agent type — the Agent tool's own
listing of available types carries `claude`, `claude-code-guide`, `Explore`, `general-purpose`,
`Plan` and `statusline-setup` and no `bugbot`. It is therefore dispatched as a general-purpose
subagent carrying the Bugbot slot's mutation-testing brief, recorded under the slot it stood in
for, with the model actually given (`sonnet`) rather than `unknown (agent-defined)`.
**Standards files resolved for the Principles slot:** no `.flow/project.md` exists, so the
auto-detect path applied and resolved `CLAUDE.md` and `AGENTS.md` in the worktree.

## Fix round 1

**Findings carried in:** F3 (Minor, Code review (low)), F4 (Critical, Primary), F5 (Minor, Primary).
F1 was withdrawn by the operator as schema-forced duplication. F2 was not sent to the fix subagent:
the operator directed it be fixed as its own task, task 3, because its fix is a Go test and the
plan's original "no Go code changes" constraint had to be overturned to allow it.
**Reproducer guard:** `check-panel-reproducers.sh` exited 1. F1's and F3's reproducer shapes were
refused — shell metacharacters, and absolute paths — so both were recorded unverifiable and put to
the operator rather than silently rewritten. F2, F4 and F5 carry `none — <reason>` exemptions.
**First dispatch, abandoned:** the round was first dispatched to rename the skills to `run` and
`new`, per the operator's initial answer on F4. The operator questioned that mid-flight, the
dispatch was stopped, and its staged-but-uncommitted rename was reverted with `git reset --hard`.
Nothing from it reached a commit.
**Why it was abandoned:** the rename was chosen to make the plugin route read better, on the
premise that a better suffix was reachable. It is not. Claude Code namespaces every plugin skill by
its plugin name with no bare form and no documented exception, so any suffix still yields
`/spectre:<suffix>` — the rename could never produce the bare `/spectre` it was chosen for, and it
cost the bare names in a checkout. Decision `skills-renamed-for-namespace` is superseded by
`single-spectre-skill`.
**Second dispatch:** merge the two skills into one, keeping the name `spectre`.

### Fix round 1 — mutation proof

Every behaviour the fix changed was mutated in this worktree, the suite run, and the file restored.
The tree was confirmed clean afterwards and the suite green.

```text
fix-mutation: .claude-plugin/plugin.json — name -> spectre-typo — none, survived
fix-mutation: .claude-plugin/plugin.json — skills -> ./nope/ — none, survived
fix-mutation: .claude-plugin/marketplace.json — source -> ./nope/ — none, survived
fix-mutation: README.md — install string -> spectre@wrong — none, survived
fix-mutation: README.md — mkdir -p removed, undoing the F3 fix — none, survived
fix-mutation: .claude/skills/spectre/SKILL.md — thin rule inverted — none, survived
fix-mutations-total: 6
```

All six survived: nothing in the repository catches any of them today. That is finding F2, measured
a second time and now covering two behaviours this round itself introduced. Task 3 is the response,
and it is scoped to catch the first five; the sixth — the skill's own prose rule — is judgment a
test cannot assert, and is left to review.

## Pass 2 (round 1) — Full

**Mode:** Full, escalated automatically and not asked. Two triggers fired: the fix touched files
outside the set the findings named — `.claude/skills/spectre/SKILL.md` was rewritten and
`.claude/skills/spectre-new/SKILL.md` deleted, where the findings named `README.md` and
`.claude-plugin/plugin.json` — and the round added a new file, `internal/check/plugin_manifest_test.go`.
**Roster:** the resolved list again — `primary`, `principles`, `code-review-low`, `bugbot`. Every one
of them raised a finding in pass 1, so targeted and Full name the same four slots here; Full changes
what they read, not who runs.
**Operator additions:** no addition this round — the resolved list ran alone.
**Diffs:** Primary reads the rewritten `.superpowers/sdd/final-review.diff` (whole branch, 353
lines). Principles and Code review (low) read their delta,
`.superpowers/sdd/slot-delta-1-shared.diff` — `git diff 844ff20..HEAD`, 343 lines — since 844ff20 is
the HEAD each last reviewed. Bugbot reads no diff file.
**Diff size guard:** measured 277, `under cap`, exit 0. No operator question raised.
**Substituted slots:** `bugbot` again, as general-purpose with the mutation-testing brief, recorded
under the Bugbot slot with the model actually given.
**Carried into this pass:** F1 withdrawn by the operator; F2, F3, F4 and F5 fixed and mutation-proved.

### Fix round 2 — mutation proof

Re-proved independently by the parent, not taken from the fix subagent's report. Each mutation was
applied in this worktree, the suite run, and the file restored; the tree was clean afterwards.

```text
fix-mutation: .claude/skills/spectre — directory renamed — README and skill agree on the directory name
fix-mutation: .claude-plugin/plugin.json — skills repointed at a decoy directory holding a stub SKILL.md — skills path names the real skill directory
fix-mutation: .claude-plugin/marketplace.json — plugins emptied — plugin name matches, and marketplace source resolves, both failing rather than panicking
fix-mutations-total: 3
```

All three are now caught. The empty-plugins case fails cleanly with no panic, which was F11.

## Pass 3 (round 2) — Targeted

**Mode:** Targeted, the default. No escalation trigger fired: the fix touched only the one file the
findings named, git recorded it as a rename rather than a new file, the diff is 83 lines, no new
Critical was raised, and this is the second fix round rather than the fourth.
**Who re-runs:** Primary as the integration check, plus Principles and Bugbot — the two slots that
raised findings in pass 2. **Code review (low) is not re-dispatched**: it returned clean in pass 2
and raised nothing for this round to have addressed.
**Operator additions:** no addition this round — the resolved list ran alone.
**Diff they read:** `.superpowers/sdd/fix-round-2.diff`, `git diff 20dbe94..HEAD`.
**Substituted slots:** `bugbot`, as general-purpose with the mutation-testing brief.

### Fix round 3 — mutation proof

Re-proved independently by the parent, not taken from the fix subagent's report.

```text
fix-mutation: .claude/skills/spectre — renamed to a prefix of its own name, README updated, SKILL.md left stale — README and skill agree on the directory name
fix-mutation: README.md — cp -r target broken, the correct string planted in an unrelated line — README and skill agree on the directory name
fix-mutations-total: 2
```

Both are now caught. The tree was clean afterwards and the suite green.

## Pass 4 (round 3) — Full

**Mode:** Full, escalated automatically and not asked, on the stated trigger that three fix rounds
have already run.
**Roster:** the resolved list — `primary`, `principles`, `code-review-low`, `bugbot`. Code review
(low) returns this pass despite being clean in pass 2 and not re-run in pass 3, because Full covers
the whole resolved roster.
**Operator additions:** no addition this round — the resolved list ran alone.
**Diffs:** Primary reads the rewritten whole-branch `final-review.diff`. Principles reads
`slot-delta-3-principles.diff` (`git diff 1175529..HEAD`), the HEAD it last reviewed. Code review
(low) reads `slot-delta-3-codereview.diff` (`git diff 20dbe94..HEAD`), the HEAD it last reviewed.
Bugbot reads no diff file.
**Substituted slots:** `bugbot`, as general-purpose with the mutation-testing brief.

### Pass 4 — dedupe and a contaminated verification

**Deduped:** Bugbot's first finding and Primary's have the same defect identity — `plugin_manifest_test.go`
at the `cpTarget` comparison, theme *unanchored substring match on the copy target* — and are carried
as F15 alone.

**A verification this parent ran was contaminated, and is recorded rather than quietly redone.** The
first attempt to confirm F15 snapshotted `README.md` while the Bugbot slot was concurrently mutating
it, ran a mutation on top of that, and restored the slot's version. The net edit was zero and the
slot reverted its own work, but the result proved nothing: the subtest selects the first `cp -r`
line, and the slot had injected an earlier one. F15 and F16 were both re-confirmed afterwards
against a tree verified clean, with `HEAD`'s README carrying exactly one `cp -r` line. The
order-dependence that caused it is itself F17.

### Fix round 4 — mutation proof

Re-proved independently by the parent on a tree verified clean first, after pass 4's contamination.

```text
fix-mutation: README.md — cp -r target changed to a longer name keeping the real one as a prefix — README and skill agree on the directory name
fix-mutation: README.md — install command marketplace name extended to spectrecorp — README install command matches
fix-mutation: README.md — an unrelated cp -r example inserted above Quick start — no subtest fails, which is the correct direction for F17
fix-mutations-total: 3
```

The round fixed the defect class with two shared helpers rather than four separate patches, and the
file shrank by 4 lines. The fix subagent also reported extending the fix to a third subtest carrying
the same document-order defect, which the plan's Tests field had not named — disclosed rather than
folded in silently.

## Pass 5 (round 4) — Full

**Mode:** Full, escalated automatically on the same standing trigger as pass 4: more than three fix
rounds have run.
**Roster:** the resolved list — `primary`, `principles`, `code-review-low`, `bugbot`.
**Operator additions:** no addition this round — the resolved list ran alone.
**Diffs:** Primary reads the rewritten whole-branch `final-review.diff`; every other diff-reading
slot reads `slot-delta-4-shared.diff` (`git diff dbe8f28..HEAD`), the single commit since each last
read. Bugbot reads no diff file.
**Substituted slots:** `bugbot`, as general-purpose with the mutation-testing brief.
**Convergence note carried to every slot:** passes 2, 3 and 4 all found variants of one defect — an
unanchored substring match — entirely within the test file, while the change's actual deliverable has
been clean since pass 2. This round fixed the class rather than the instances.

### Fix round 5 — coverage proof

The round was a pure deletion, so the proof runs the other way: rather than showing a mutant is
caught, it shows the two assertions the deleted subtest made are still made. Re-proved independently
by the parent on a clean tree.

```text
fix-mutation: .claude-plugin/plugin.json — skills pointed at a path that does not resolve — skills path names the real skill directory
fix-mutation: .claude/skills/spectre — moved out, leaving the skills directory with no SKILL.md — README and skill agree on the directory name
fix-mutations-total: 2
```

Both still fail, so nothing the deleted subtest asserted is now unasserted. The file lost 20 lines
and one subtest; the run count went 206 to 205, as the plan predicted.

## Pass 6 (round 5) — Full

**Mode:** Full, on the standing trigger that more than three fix rounds have run. The diff under
review is a 20-line deletion in one file, so Full here costs breadth rather than depth.
**Roster:** the resolved list — `primary`, `principles`, `code-review-low`, `bugbot`.
**Operator additions:** no addition this round — the resolved list ran alone.
**Diffs:** Primary reads the whole-branch `final-review.diff`; every other diff-reading slot reads
`slot-delta-5-shared.diff` (`git diff 278113c..HEAD`). Bugbot reads no diff file.
**Substituted slots:** `bugbot`, as general-purpose with the mutation-testing brief.
**Standing at the start of this pass:** nineteen findings raised, one withdrawn and eighteen fixed.
Pass 5 returned clean on Principles, Code review (low) and Bugbot, and Primary's verdict was that the
branch is ready to hand to a human.
