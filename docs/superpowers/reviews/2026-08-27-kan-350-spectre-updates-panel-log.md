# Review panel — kan-350-spectre-updates

## Pass 1

**Mode:** initial panel (all three required slots).
**Diff:** `.superpowers/sdd/final-review.diff` — `git diff 3b16340` (merge base), 1795 changed
lines, under the guard's cap (`check-panel-diff-size.sh` exit 0).
**Slots dispatched:** 0 Primary, 2 Principles, 3 Code review (low) — each `general-purpose` on
`sonnet` (the settings store's `defaultModel`; no session override was given).

**Bugbot: not requested. Security: not requested.** Neither on-demand slot was named by the operator
at the start of this stage, and neither is ever added by a diff-size or touched-area trigger.

**Standards files:** none resolved — this project has no `.flow/project.md`, so `[STANDARDS_PATHS]`
was passed empty.

**Pass 1 findings:** F1 (Major, `internal/cmd/init.go:68`, Code review (low)), F2 (Major,
`internal/cmd/init_test.go:140`, Primary), F3 (Minor, `design.md`, Primary). Principles: CLEAN.

**Reproducers.** `check-panel-reproducers.sh` exit 0 (3 findings declared). F2's reproducer
(`go test ./internal/cmd/ -run TestInitUnwritableParent`) was refused by `run-reproducer.sh`, which
executes only repo-relative script paths — not a shape refusal, a runner limitation. The finding was
instead confirmed by direct inspection: `internal/cmd/refs_test.go` guards both of its chmod tests
with `os.Geteuid() == 0`; `internal/cmd/init_test.go` did not. F1 and F3 were recorded
`none — <reason>` and dispatched unverified, per the contract's exemption form.

## Fix round 1

**Findings dispatched:** F1, F2 to one fix subagent on `sonnet`. F3 was fixed by the parent: it is a
change artifact (`design.md`), which a fix subagent may never stage.

**Fix diff:** `internal/cmd/init.go`, `internal/cmd/init_test.go` only — inside the set the findings
named, so no escalation trigger fired and the re-run is **targeted**.

fix-mutation: internal/cmd/init.go — the `fi.IsDir()` check reverted — `TestInitRejectsFileWhereDirBelongs` failed on both subtests (`exit = 0, want 1`)
fix-mutation: internal/cmd/init_test.go — the `os.Geteuid() == 0` skip removed — none; this machine is uid 501, so chmod genuinely denies access here and the guarded failure cannot be demonstrated locally. It manifests only under euid 0, where chmod is a no-op. Reported honestly rather than claiming a run that did not happen.
fix-mutations-total: 2

**F3's fix is observational, not code:** `spectre init --root <worktree>` was run against this
repository's own half-made tree and created the `config.md` the hand-made bootstrap had not written
(exit 0). `design.md` section 5 now states what actually happened — a hand-made `specs/`+`changes/`
completed by `init`'s fill-in-what-is-missing path — instead of implying a greenfield run.

**Re-run mode:** targeted. Slot 0 Primary reads the rewritten whole diff; slot 3 Code review (low)
reads its delta from `d3b42de`. Principles raised nothing and is not re-run.

## Fix round 1 re-review (targeted)

**Slot 3 Code review (low)** — CLEAN. Confirmed F1 fixed at both shas, then hunted the fix itself:
symlink-to-directory, dangling symlink, tree root as a plain file, mixed `specs` dir + `changes`
file, and permission-denied on a parent. It correctly separated the two pre-existing `MkdirAll`
error paths from the new branch rather than reporting them as regressions introduced here, and
confirmed by mutation that `TestInitRejectsFileWhereDirBelongs` pins the mechanism.

**Slot 0 Primary** — confirmed F1, F2 and F3 fixed by independent reproduction, verified every
baseline at every commit boundary (135→136→146→147→131→131→131→131) and drove the documented journey
end to end. Raised one new Minor, **F4**.

## Fix round 2 (parent only — a change artifact)

**F4** — `design.md` section 1's "Exit codes" still said `init` has no exit 1, which the F1 fix made
false. Fixed by the parent, since a fix subagent may never stage `spectre/changes/`.

The sweep went wider than the finding: every exit-code claim in **both** artifacts was re-checked,
which turned up three more stale statements the finding did not name — `design.md`'s
"exits 0 whether it created three things or none", the `init-idempotent` decision's "Considered:
refuse with exit 1" (now ambiguous against the real exit 1 the fix added), and `tasks.md`'s
"returning `OK`/`Usage`". All four are corrected together, and the idempotence decision now states
explicitly what it did and did not reject.

fix-mutation: spectre/changes/kan-350-spectre-updates/design.md — none — a documentation sentence, not an executable behaviour; nothing to mutate and no test to catch it
fix-mutation: spectre/changes/kan-350-spectre-updates/tasks.md — none — same
fix-mutations-total: 2

**No code changed in this round**, so no slot's diff moved and no code-reading slot has anything new
to read. Slot 0 is re-run once more as the integration check, because it raised F4 and its subject —
the change artifacts — is the one thing that did change.
