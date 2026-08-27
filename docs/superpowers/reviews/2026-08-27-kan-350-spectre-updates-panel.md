# Review panel — kan-350-spectre-updates

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Major | internal/cmd/init.go:68 | init checks only err == nil from os.Stat and never fi.IsDir(), so a plain file sitting where specs/ or changes/ should be is treated as already present: init reports success, exits 0, creates the rest of the tree, and leaves a broken half-initialized tree with no error or warning. Observed: after touching spectre/specs as a 0-byte file, init printed only "created .../changes" and "created .../config.md" and exited 0, with specs still a file. |
| F2 | Primary | Major | internal/cmd/init_test.go:140 | TestInitUnwritableParent chmods a directory to 0o000 and expects Init to fail with Usage, but omits the os.Geteuid() == 0 skip guard that every other chmod-based test in this codebase carries (internal/cmd/refs_test.go). Running as root — a Docker CI image, for instance — chmod does not deny access, Init succeeds, and the test fails with exit = 0, want 2. design.md claims this test does the chmod exactly as refs_test.go does; it does the chmod the same way but omits the root-skip that makes the pattern safe. |
| F3 | Primary | Minor | spectre/changes/kan-350-spectre-updates/design.md:119 | design.md section 5 states that spectre init runs in this repository and calls it the first use of the command being added, implying the tree on disk is what init produced. The tree has specs/ and changes/ but no config.md, which init always writes when absent — so init did not produce it. The directories were created by hand during pipeline bootstrap, before init existed to run. |
| F4 | Primary | Minor | spectre/changes/kan-350-spectre-updates/design.md:66 | design.md section 1 Exit codes still states there is no exit 1 because init inspects no content and has no content refusal to make. The F1 fix added exactly that content refusal: a file blocking specs/ or changes/ exits Fail (1). Section 5 was updated during the fix round but this sentence in section 1 was missed — the same stale-artifact class of defect F3 caught. |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 none — the defect needs a multi-step filesystem setup (mkdir the tree, touch a plain file named specs, run init) which the reproducer shape rules reject; the steps are in this note and a regression test is required as part of the fix
finding-reproducer: F2 go test ./internal/cmd/ -run TestInitUnwritableParent
finding-reproducer: F3 none — the finding is an untrue sentence in a change artifact, not a runtime behaviour; it is confirmed by the absence of spectre/config.md in this worktree
finding-reproducer: F4 none — the finding is an untrue sentence in a change artifact; it is confirmed by the F1 reproducer showing init now exits 1
