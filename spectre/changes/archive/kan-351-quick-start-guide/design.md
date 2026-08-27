# kan-351-quick-start-guide — design

How each change in `proposal.md` works, and what was rejected on the way.

## Context

Four things move together, and they are one change because the docs are only honest if the binary
enforces what they show. A quick start that walks a reader through a template the tool does not
check teaches a shape the tool will happily let them break; and enforcing a template nobody has
seen filled in is a rule without an example. The worked example, the strict templates and the
scaffold that produces them are the same statement made in three places.

`validate` scans **open changes only** — `tree.Changes()` calls `changes(false)`, which skips the
`archive` entry — so no existing archived change is affected by any rule this change tightens. The
one archived change in this repository already titles its `tasks.md` `# kan-350-spectre-updates`,
which is exactly what the new template requires.

## 1. The README quick start

The first section after the one-line description, absorbing `## Install` and `## Getting started`,
which are deleted rather than left to say the same things twice.

It is a numbered walk, each step naming the command and what it produces:

1. install globally — `go install github.com/tweety53/spectre/cmd/spectre@latest`
2. `spectre init` in the project
3. `spectre new multi-channel-notifications`
4. fill in the generated files — linking to `docs/example.md` for the filled-in versions
5. write the capability spec
6. `spectre validate`
7. work the tasks, checking boxes
8. `spectre archive`

**Two facts move out of `## Getting started` and must survive**, because they are stated nowhere
else: `archive` needs the tree inside a git repository with the change's files already tracked, and
`--root` must precede any positional argument because Go's `flag` package stops parsing at the
first one.

**Every content-producing step carries the exact prompt to hand an AI agent**, verbatim, followed
by one line on what comes back. spectre's files are written far more often by an agent than by hand,
and a guide that shows only the commands leaves the actual work — producing a proposal, a spec, a
plan — as an exercise. The steps that are pure commands (`init`, `validate`, `archive`) show the
command and its real output instead; they need no agent.

**The guide is connected to the example rather than duplicating it.** Each step names the file it
produces and links to that file's full body in `docs/example.md`; the README shows commands and
outcomes, the example shows content. Nothing appears in full in both places.

## 2. `docs/example.md`

One change, end to end, with every generated file in full:

- `spectre/specs/notifications.md` — the capability, with numbered `SHALL` requirements
- `spectre/changes/multi-channel-notifications/proposal.md`
- `spectre/changes/multi-channel-notifications/tasks.md`
- `spectre/changes/multi-channel-notifications/design.md`

**This page is the end-to-end read, top to bottom.** It runs in the order a user hits the steps —
install, `init`, `new`, write the spec, fill in the files, `validate`, work the tasks, `archive` —
and each step is the same three beats: the prompt or command, the generated output verbatim, then
the resulting files in full. A reader starts at the top and reads straight down; nothing is deferred
to a link or a "see above".

Each file is shown three ways: **the exact prompt that produced it**, the file itself, and an
explanation of the result — what to look at, and what a reader must still verify by hand, because an
agent's output is a draft and not an authority. Prompts are agent-agnostic: no product name, no
slash command, no vendor syntax.

The subject is a backend that sends only email notifications and wants better delivery and open
rates, adding SMS, messenger and push channels. It is deliberately generic: no framework, no
vendor, no language, so the shape is what a reader takes away rather than the domain.

Every file body in that page must satisfy the rules **1** introduces — the example is the first
thing that would expose a template the rules cannot actually accept.

## 3. `spectre new` scaffolds `design.md`

A third file beside `proposal.md` and `tasks.md`, carrying `## Context` and `## Decisions`, in the
`<!-- -->` comment style `_proposalTemplate` already uses — never either of the two placeholder
tokens `_placeholders` in `internal/check/check.go` lists, which the `placeholders` rule reports on
sight. A stub using one would make every freshly scaffolded change fail `validate` immediately.

## 4. `render.Tasks` titles the file with the change id

`render.Tasks(ts []model.Task)` becomes `render.Tasks(id string, ts []model.Task)` and writes
`# <id>` instead of `# Tasks`. It has exactly one caller, `internal/cmd/new.go`, so the change is
contained.

## 5. Strict templates

The `headings` rule gains **ordering** — required headings must appear in the order the template
states, not merely be present — and covers two more files:

| File | Required, in order |
|---|---|
| `proposal.md` | `# <change-id>`, `## Why`, `## What changes` |
| `tasks.md` | `# <change-id>` |
| `design.md` | `## Context`, `## Decisions` — **only when the file exists** |
| `<capability>.md` | `# <capability>`, `## Purpose`, `## Requirements` — unchanged, now also ordered |

**Extra headings remain legal.** A change is free to add `## Open questions`, `## Testing`, or
numbered sections; the rule states what must be there and in what relative order, not what may not.

`task-sequence` gains one check: a `tasks.md` with no tasks at all is a finding. That rule already
owns malformed task lines and numbering, so the "there are no tasks" case belongs with them.

`design.md` is validated **only when present** — it stays optional. The README describes it as
"optional, unparsed"; that wording is now false and is updated with the rule.

**Only the README, not `docs/spec-format.md`.** This paragraph originally named both, on the
assumption that the file documenting the spec format would also describe the change files. It never
mentioned `design.md` at all — `git show 6271d41:docs/spec-format.md | grep -c design` returns `0`.
Task 7 found this while implementing and its file list was corrected; this sentence was not, and is
corrected here.

## 6. The project instruction file

`AGENTS.md` is canonical and `CLAUDE.md` is a one-line pointer at it. `AGENTS.md` because it is the
cross-harness convention rather than one vendor's; the pointer because Claude Code reads
`CLAUDE.md`, and two files carrying the same rules is how they come to disagree.

It states which Go skills apply to work in this repository, **and which do not, with the reason** —
a skipped skill should be a stated judgement rather than an omission a reader has to re-derive — plus
the project facts an agent would otherwise guess: the check commands, the standard-library-only
constraint, and the exit-code contract.

## 7. Prompt-first, not a terminal walk

**The operator's direction, after reading the first version:** spectre is driven through an AI
coding agent, so the docs teach prompts and let the agent run the commands. A reader is never told
to type `spectre init`.

What this changes, and what it deliberately does not:

- **`README.md`'s quick start** becomes a short sequence of prompts. The agent runs `init`, `new`,
  `validate` and `archive` itself; the human supplies intent. The per-step command blocks and their
  captured output go.
- **`docs/example.md`** keeps being the end-to-end read, and keeps **every generated file body in
  full** — those are the output the operator asked to see. Its `spectre` terminal transcripts go.
- **A small command reference table stays** in the README. The commands are still the tool's real
  surface: an agent has to know they exist, and a reader who wants them should not have to run
  `--help` to find out. Dropping it entirely was considered and rejected.

**The `scaffold-reports-no-tasks` decision survives this reorientation**, but its *presentation*
changes: the walkthrough no longer shows a `validate` transcript reporting `no tasks`, so wherever
that behaviour still matters to a reader it is stated in prose rather than shown as output.

## 8. The commands between spectre commands

spectre commands are the punctuation, not the work. A guide showing only them implies the gaps
between are empty, and **most of what fills them is another prompt**: *"implement task 1"*, the
project's own build and tests, *"tick task 1's box"*, commit, repeated per task — which is the bulk
of a real change. The one genuine shell requirement is `git add spectre` before `archive`, which
moves the change with `git mv` and refuses on untracked files.

**Short, and in the prompt where it belongs.** One or two lines per gap, not a git tutorial; and
where the command is the agent's to run, it goes in that step's prompt like any other, because the
docs are prompt-first and the reader types nothing.

**End to end means the whole cycle, not the application code.** The walkthrough runs from an empty
directory through tree, change, spec, plan, validation, the implement/test/tick/commit loop, and
archive — every stage present, none glossed. It shows **no application code**: the implement loop is
described as prompts, because the example is deliberately language- and vendor-neutral and showing
real channel code would force it to pick a framework. Confirmed with the operator rather than
assumed.

**Both documents say the same thing.** The README and the walkthrough each carry it, and they must
agree — two documents disagreeing about what to run between steps is worse than neither mentioning
it.

## 9. Slash commands, shipped with spectre

Two skills in the repository's own `.claude/skills/`, so they travel with the project: anyone who
clones spectre gets them, and they are versioned alongside the code they drive.

**Not in the `agents` repository.** Every other skill on this machine is authored there and
symlinked into `~/.claude/skills/`. These are deliberately different: the operator's instruction was
that they are *part of spectre, not flow*. A skill describing spectre's own commands belongs beside
spectre, where it cannot drift from the binary it documents and where it reaches anyone using the
project — not only this machine.

| Command | Does |
|---|---|
| `/spectre <subcommand> [args]` | dispatches to any spectre subcommand, shows its real output, then names what usually comes next |
| `/spectre-new <change-id>` | scaffolds the change, then helps fill in the generated files |

**Thin, plus next-step guidance — not workflow automation.** They run the command, show what it
really printed, and say what typically follows; they do not go on to do that next thing unasked.
`/spectre-new` is the one exception, and only because scaffolding is inert on its own: `new` leaves
three stubs and a change that reports `no tasks` until someone writes the plan, so the useful unit
is scaffold-then-fill.

**Why two commands and not six.** A dispatcher mirrors the binary's own surface exactly, so a
subcommand added to spectre later needs no new skill. `new` earns a dedicated command because it is
the only one with real work around it.

## 10. A second guide, for the terminal

`docs/terminal.md` covers the same ground as `docs/example.md` for a reader who would rather type
commands than hand prompts to an agent. Same subject, same nine steps, same order — commands and
their real output in place of prompts.

**It does not repeat the generated file bodies.** They are shown in full in `docs/example.md` and
pinned to `cmd.New`'s real output by `TestExampleDocMatchesScaffold`; a second copy would be a third
statement of the same content, outside that guard, and free to drift. The terminal guide names what
each step produces and links to the walkthrough for the contents. **A reader following it therefore
follows one link for file content** — accepted deliberately, because the alternative is a copy no
test protects.

**This reverses part of the prompt-first decision, and does not supersede it.** `docs-are-prompt-first`
stands: the README's quick start and `docs/example.md` remain prompt-first, and that is the path the
README leads with. What changed is that a terminal reader is no longer left without a route —
previously the terminal steps were deleted outright rather than relocated.

**Some prose is deliberately duplicated between the guides, and this is the reasoning.** The
`--root`-before-positional rule appears in both the README and `docs/terminal.md`, and the `archive`
prerequisites appear in the README, the walkthrough and the terminal guide — three independent
statements of one behaviour, bound by no test. This is the same trade already taken for the agent
prompts: a rule a reader needs *at the moment they are typing* is worse behind a link than
duplicated. **The cost is stated rather than hidden: a change to either rule requires three manual
edits to stay honest, and nothing will catch a missed one.** A prose-diffing guard was considered
and rejected as brittle — it would fail on every legitimate reword, which is how guards get deleted.
The file bodies, which are bulk content rather than a rule needed mid-keystroke, are **not**
duplicated and remain guarded.

**Two guides is the cost.** They cover the same nine steps and can drift apart in structure even
though the file bodies cannot. That is the accepted price of serving both audiences; the mitigation
is that neither restates the other's substance — prompts live in one, command output in the other,
file bodies in one place only.

## Testing

- `internal/check`: ordering violations per file (present but out of order); `design.md` absent →
  no findings; `design.md` present and short a heading → finding; `tasks.md` with zero tasks →
  finding; extra headings → no finding.
- `internal/render`: `Tasks` renders `# <id>`; the round trip through `parse` still holds.
- `internal/cmd`: `new` writes three files; the scaffolded `design.md` carries both headings; a
  freshly scaffolded change passes `validate` — the test that stops the scaffold and the rules
  drifting apart, which is exactly the failure KAN-350's `config.md` pairing was added to prevent.
- A test asserting `docs/example.md`'s embedded file bodies satisfy the rules would be ideal but is
  deferred; see **Open questions**.

## Decisions

### Ordering and the two new files extend the existing `headings` rule

**ID:** headings-rule-extended
**Status:** active
**Chosen:** extend `headings` — it is literally the headings rule doing more.
**Considered:** a new rule name (`heading-order`, `templates`) — rejected because
`config.RuleNames` is a closed set and `init`'s scaffolded `config.md` is pinned to it by a test, so
a new name means config-surface churn and a scaffold change for no new concept; leaving ordering
unchecked — rejected, since a reordered template is exactly the drift the requirement names.

### "No tasks at all" belongs to `task-sequence`

**ID:** empty-tasks-under-task-sequence
**Status:** active
**Chosen:** `task-sequence`, which already owns malformed task lines and numbering.
**Considered:** `headings` — rejected: the absence of tasks is not a heading fact.

### `init` and `new` print paths relative to the working directory

**ID:** created-paths-relative
**Status:** active
**Chosen:** relativise the printed path against the working directory, falling back to the absolute
path when the tree sits outside it or when the relativisation cannot be computed.
**Considered:** printing the absolute path in the docs instead — rejected because every example
would then carry a long path no reader will actually see, working directly against the
readable-top-to-bottom requirement; collapsing `init`'s three lines into one summary line —
rejected as the largest behaviour change of the three, discarding per-path information a scripted
caller may parse.

**This was a pre-existing mismatch, not one this change introduced.** The README has shown
`created spectre/changes/my-change` since before this work, while `new` has always printed the
absolute path. Writing a walkthrough whose every output is real is what surfaced it.

### The docs are prompt-first; the agent runs the commands

**ID:** docs-are-prompt-first
**Status:** active
**Chosen:** the reader hands prompts to an agent, which runs the `spectre` commands. No terminal
walk for the human, in either document; a command reference table stays in the README.
**Narrowed by `terminal-guide-alongside`:** the "no terminal walk in either document" clause held
for the two documents that existed when it was written — the README and `docs/example.md` — and
still does. A third document, `docs/terminal.md`, now carries one. The alternative rejected below
was a fallback *inside the README*; a separate document is a different shape and does not carry the
cost that rejection named.
**Considered:** dropping command documentation entirely — rejected because the commands are the
tool's real surface and an agent needs to know they exist; keeping the terminal walk lower down as a
fallback for agent-less use — rejected because it leaves the README long (213 lines) and hedges on
which path is the real one, which is the thing the operator objected to.

**This supersedes the presentation half of `guide-carries-agent-prompts`, not its substance.** That
decision said every content-producing step carries a verbatim prompt, and command-only steps show
the command and its real output instead. The first half stands. The second is now wrong: there are
no command-only steps for the human, because the human runs no commands.

### The guide shows agent prompts, not just commands

**ID:** guide-carries-agent-prompts
**Status:** superseded by docs-are-prompt-first
**Chosen:** every content-producing step carries a verbatim, copy-pasteable prompt plus an
explanation of the result; command-only steps show the command and its real output.
**The README is the short path and the example is the long one**, which is why the split survives
the "readable end to end" requirement: the walkthrough is the thing that reads end to end, and the
quick start is a one-screen route into it rather than a second telling of the same story.
**Considered:** commands alone — rejected, because the hard part of using spectre is producing the
content, not running the binary, and that is the part a reader most needs shown; prompts only in
`docs/example.md` — rejected, since a prompt behind a link is a prompt nobody copies, so prompt text
is the one thing deliberately allowed in both places.

### The quick start is short and defers to the walkthrough

**ID:** quick-start-is-short
**Status:** active
**Chosen:** the README's quick start is deliberately brief — the shortest prompt sequence that gets
a reader from nothing to a validated, archived change — and points at `docs/example.md` for the full
thing. The walkthrough is the long read; the README is the door.
**Considered:** keeping the quick start long enough to stand alone — rejected by the operator, who
asked for it short with a reference to the full example. A README that retells the walkthrough gives
the reader two documents to keep in step and no reason to prefer either.

**This supersedes `quick-start-length-accepted`.** That decision accepted ~123 lines because four
verbatim prompts forced it. The premise no longer holds: with the walkthrough carrying the full
prompts and every file body, the README does not need to.

### The quick start runs to ~123 lines, and that is accepted

**ID:** quick-start-length-accepted
**Status:** superseded by quick-start-is-short
**Chosen:** keep it. The plan asked for roughly one screen and the result is about three; the
overshoot is four verbatim agent prompts and eight command-plus-real-output pairs, and no file body
is retold from `docs/example.md`.
**Considered:** cutting the prompts to links — rejected, because the prompts are the requested
feature and `guide-carries-agent-prompts` already records why a prompt behind a link is a prompt
nobody copies; cutting the shown output — rejected, since a step that shows a command without its
result is the thing the walkthrough exists to stop.

**Recorded rather than silently ignored:** the review raised the length as a Minor finding, and this
is the answer to it, not a dismissal of it. The one-screen target was a guess written before the
prompts existed; the prompts are what made it wrong.

### A freshly scaffolded change reports `no tasks`, and that is correct

**ID:** scaffold-reports-no-tasks
**Status:** active
**Chosen:** keep the empty-tasks finding, and accept that `spectre new` followed immediately by
`spectre validate` reports exactly one finding. It is the tool saying what to do next — write the
plan — and it is consistent with `spectre archive`, which already refuses a `tasks.md` with no
tasks (`internal/cmd/archive.go`). The scaffold is not a finished change and should not claim to be.
**Considered:** dropping the empty-tasks check so a fresh scaffold validates clean — rejected,
since "a change with no tasks" is precisely the loose case the strict-template requirement names;
having `spectre new` write a placeholder task — rejected as worse than the finding, because a
scaffolded fake task is a real task to every other command, and `archive` would archive it
unchecked once someone ticked its box.

**This was found by running the binary, not by reading the plan.** The plan as first written had
task 2 make zero tasks a finding and task 4 assert a fresh scaffold validates clean — two
requirements that cannot both hold. Task 3's end-to-end run surfaced the contradiction.

### `design.md` is validated only when it exists

**ID:** design-optional-but-checked
**Status:** active
**Chosen:** check the template when the file is there; a change without one is fine.
**Considered:** requiring `design.md` on every change — rejected as forcing a design document onto
trivial changes; leaving it unparsed — rejected, since the scaffold now creates it and an unchecked
scaffold is the loose template this change exists to tighten.

### The quick start absorbs Install and Getting started

**ID:** absorb-install-and-getting-started
**Status:** active
**Chosen:** fold both into the quick start and delete them.
**Considered:** adding the quick start above them — rejected as stating installation twice, the
duplication KAN-350 just removed from this README; putting the quick start above the description
line — rejected because a reader then meets commands before learning what spectre is.

### The example lives in its own file

**ID:** example-in-its-own-file
**Status:** active
**Chosen:** `docs/example.md`, linked from each step of the quick start.
**Considered:** full bodies inline in the README — rejected: it would add roughly as many lines as
KAN-350 removed, to a section whose whole point is being quick; abbreviating the bodies — rejected,
since "all the generated files" is the requirement and a trimmed file does not show the shape.

### `tasks.md` is titled with the change id

**ID:** tasks-titled-by-change-id
**Status:** active
**Chosen:** `# <change-id>`, matching `proposal.md`.
**Considered:** keeping `# Tasks` and fixing the archived file — rejected: the title then repeats
the filename and carries nothing, and the two generated files disagree; accepting any level-1
heading — rejected as enforcing shape without enforcing the template.

### The slash commands ship with spectre, not with the agent tooling

**ID:** slash-commands-ship-with-spectre
**Status:** active
**Chosen:** author them in this repository's `.claude/skills/`, versioned with the code they drive.
**Considered:** the `agents` repository, where every other skill on this machine lives and from
which `setup.sh global` symlinks them into `~/.claude/skills/` — rejected on the operator's explicit
instruction that these are part of spectre rather than of flow, and because a skill describing
spectre's commands drifts from the binary the moment it lives in a different repository; six skills,
one per subcommand — rejected as six files to keep in step where a dispatcher mirrors the binary's
surface for free.

### Slash commands are thin, plus next-step guidance

**ID:** slash-commands-are-thin
**Status:** active
**Chosen:** run the command, show its real output, name what usually comes next — and stop.
**Considered:** workflow-aware commands that do the following step automatically — rejected because
a command that acts beyond what was asked is the opposite of convenient when it guesses wrong;
purely thin wrappers with no guidance — rejected as saving almost nothing over typing a short
command against an already-global binary.

**`/spectre-new` is the deliberate exception**, and the reason is stated rather than assumed: `new`
alone leaves three stubs and a change that reports `no tasks` until a plan exists, so scaffolding
without filling is not a useful unit of work.

### A terminal guide is added beside the prompt-first one, not instead of it

**ID:** terminal-guide-alongside
**Status:** active
**Chosen:** a second document, `docs/terminal.md`, mirroring the walkthrough with commands and real
output; the README points at both and still leads with the prompt-first path.
**Considered:** a terminal appendix inside `docs/example.md` — rejected because that page is already
308 lines and an appendix would push it past 400, burying the prompt-first reading the operator
asked for; a dual-column format showing prompt and command per step — rejected because it undoes
prompt-first for every reader in order to serve some of them.

**The file bodies are not repeated**, by decision: they are guarded in `docs/example.md` by
`TestExampleDocMatchesScaffold`, and a second copy would sit outside that guard. The terminal guide
links for content. This costs its reader one link and buys the guarantee that the bodies cannot
drift.

**This narrows `docs-are-prompt-first` rather than merely coexisting with it**, and saying "it does
not supersede it" would have been too glib. That decision states "No terminal walk for the human, in
either document", and rejects "keeping the terminal walk lower down as a fallback for agent-less
use" on the grounds that it leaves the README long and hedges about which path is real. Both
objections were about a fallback **inside the README**. A separate document costs the README two
lines, and the README still leads unambiguously with the prompt-first path — so the reasons for the
rejection do not reach this shape. What is now false in the earlier decision is its blanket "either
document" clause, and that clause has been marked narrowed where it is written rather than left to
be discovered.

**What is unchanged:** the README's quick start and `docs/example.md` remain prompt-first, and the
README leads with that path.

## Open questions

### The `placeholders` rule makes its own tokens undocumentable in a checked file

**ID:** placeholders-rule-blocks-its-own-documentation
**Status:** open
**Why it is open:** found by dogfooding, not by design. Task 2 made `design.md` a checked file, and
this very document then failed `validate` twice — because it named the two tokens the
`placeholders` rule forbids, in prose explaining that rule. The rule matches a bare substring
anywhere in the file, with no exemption for a fenced block, an inline code span, or a sentence that
is plainly describing rather than deferring work. The wording here was changed to refer to the
tokens indirectly, which is a workaround and not an answer.
**What it affects:** any change whose subject *is* the placeholders rule cannot describe it in a
checked file. `tasks.md` escapes only because it is not placeholder-checked, which is an accident of
where `placeholderFindings` is called rather than a decision. The fix, if it is wanted, is to exempt
fenced and inline-code spans the way `headings` and `malformed-bullet` already skip fences — but
that widens this change well past its subject, so it is recorded rather than taken.

### Should `docs/example.md`'s embedded file bodies be checked by a test?

**ID:** example-bodies-untested
**Status:** open, narrowed
**Why it is open:** task 11 added `TestExampleDocMatchesScaffold`, which extracts the "## 2.
Scaffold the change" section's three fenced bodies from `docs/example.md` and compares them, byte
for byte, against what `cmd.New` actually writes — pinning the *scaffold* half of the page. The
*filled-in* bodies — the capability spec, and the completed `proposal.md`, `tasks.md` and
`design.md` shown from "## 4. Write the capability spec" onward — are still prose with no test
extracting and validating them, and remain deferred for the same reason: doing so means extracting
them from the markdown and running the checkers over them, which is scope beyond this change.
**What it affects:** whether `docs/example.md`'s filled-in bodies can silently drift out of
conformance with the rules they demonstrate. The scaffold bodies can no longer drift silently; the
filled-in bodies' conformance still rests on review alone.
