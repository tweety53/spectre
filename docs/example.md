# Example: multi-channel-notifications

One change, end to end, as prompts for an AI coding agent, which runs the `spectre` commands
itself. Every body shown is real output against a built `spectre` binary in a scratch git repository. The subject — an
email-only backend adding SMS, messenger and push — is a language-free stand-in.

The recommended path fills the same change with superpowers' brainstorming and writing-plans:
[superpowers-example.md](superpowers-example.md).

## 1. Create the tree

Paste this into an AI coding agent:

> Set up a spectre tree in the current directory. If the directory is not already a git
> repository, run `git init` first. Then run `spectre init`.

Resulting layout:

```
spectre/
  config.md
  specs/
  changes/
```

- `spectre init` only fills in what's missing, so re-running it is safe.
- `specs/` holds capabilities, `changes/` changes; nothing needs editing yet.

## 2. Scaffold the change

Paste this into an AI coding agent:

> Run `spectre new multi-channel-notifications` to scaffold the change.

`new` writes three files, each a stub:

`spectre/changes/multi-channel-notifications/proposal.md`:

```markdown
# multi-channel-notifications

## Why

<!-- the problem this change solves -->

## What changes

<!-- the observable difference once this lands -->
```

`spectre/changes/multi-channel-notifications/tasks.md`:

```markdown
# multi-channel-notifications

```

`spectre/changes/multi-channel-notifications/design.md`:

```markdown
## Context

<!-- what constrains this change and why it is one change -->

## Decisions

<!-- what was chosen, what was considered, and why -->
```

Check by hand:

- Only titles and HTML comments — `new` scaffolds shape, not prose.
- `validate` here reports one finding, `no tasks`: the next thing to write. It clears once
  `tasks.md` has a task, after step 5.

## 3. Write the capability spec

The capability this change touches, `notifications`, doesn't exist yet, so it's written first.
Paste this into an AI coding agent:

> Write `spectre/specs/notifications.md`. It must have exactly these headings, in this order:
> `# notifications`, `## Purpose`, `## Requirements`. Under `## Requirements`, write numbered
> bullets in the form `- R<n>: <requirement text>`, ids gap-free starting at `R1`, and every
> bullet's text must contain the word `SHALL` stating what the system is required to do. Cover:
> selecting a delivery channel, falling back to email when no channel preference is set, and
> retrying on another channel when delivery fails.

Wrote `spectre/specs/notifications.md`.

```markdown
# notifications

## Purpose

Deliver events to a user through whichever channel — email, SMS, messenger or push — they are
actually likely to see, instead of always sending email regardless of whether it gets opened.

## Requirements

- R1: The system SHALL send every notification through at least one channel selected from the
  set the user has enabled: email, SMS, messenger, push.
- R2: When a user's preferred channel is unset for a given notification type, the system SHALL fall
  back to email.
- R3: When delivery on a channel fails, the system SHALL retry on the user's next-preferred
  enabled channel before marking the notification undelivered.
- R4: The system SHALL let a user set a distinct preferred channel per notification type, not
  only one preference for all notifications.
```

Check by hand:

- Ids run `R1`–`R4`, no skips or repeats, and every bullet contains `SHALL` — a dropped or reused
  id reads fine and still fails `validate`.

## 4. Fill in the proposal

Paste this into an AI coding agent:

> Write `spectre/changes/multi-channel-notifications/proposal.md`. It must have exactly these
> headings, in this order: `# multi-channel-notifications`, `## Why`, `## What changes`. `## Why`
> explains, as bullets, why email-only notifications hurt delivery and open rates. `## What
> changes` explains, as bullets, what the change adds: SMS, messenger and push channels, a
> per-notification-type channel preference, and fallback to the next-preferred channel on
> failure.

Wrote `spectre/changes/multi-channel-notifications/proposal.md`.

```markdown
# multi-channel-notifications

## Why

- The backend sends every notification by email today. Open rates on email lag every other
  channel the users we serve actually check first.
- A user with no email client open right now misses a time-sensitive notification entirely —
  there is no faster channel to fall back to.
- Support tickets already ask for SMS and push; messenger is the channel our newest users expect
  by default.

## What changes

- Notifications gain three additional channels — SMS, messenger and push — alongside the
  existing email channel.
- A user can set a preferred channel per notification type; email remains the fallback when no
  preference is set.
- When delivery on the preferred channel fails, the system retries on the user's next-preferred
  enabled channel before giving up.
```

Check by hand:

- The title is exactly the directory name, not `# Multi-Channel Notifications`, and `## Why`
  precedes `## What changes` — `validate` catches either drift.

## 5. Fill in the tasks

Paste this into an AI coding agent:

> Write `spectre/changes/multi-channel-notifications/tasks.md`. The first line must be exactly
> `# multi-channel-notifications`. Every task is a line of the form `- [ ] <n>. <short title>`,
> numbered from 1 with no gaps or repeats, followed by a short paragraph of what that task does.
> Order tasks by dependency: the channel abstraction before the channels that implement it, the
> channels before the preference and fallback logic that route through them. Write real,
> reviewable work — no placeholder tasks.

Wrote `spectre/changes/multi-channel-notifications/tasks.md`.

```markdown
# multi-channel-notifications

- [ ] 1. Add a channel abstraction behind the existing email sender

Introduce a `Channel` interface with one method, send a notification and report success or
failure, and make the current email code its first implementation. No behaviour changes yet —
this only gives the later channels a seam to plug into.

- [ ] 2. Add the SMS, messenger and push channel implementations

Each new channel implements the interface from task 1 against its own provider. Every channel is
tested in isolation with a fake transport, so a provider outage never reaches a real test run.

- [ ] 3. Add per-notification-type channel preference

Store one preferred channel per user per notification type; read it when a notification is sent.
When no preference is stored, the send path uses email, matching the current behaviour for every
existing user.

- [ ] 4. Add fallback on delivery failure

When the preferred channel reports failure, retry once on the user's next-preferred enabled
channel, in the order the user configured, before marking the notification undelivered.

- [ ] 5. Backfill channel preference for existing users

Every current user gets an explicit email preference row for every notification type they
already receive, so the fallback in task 4 has a defined starting point instead of relying on the
implicit default.
```

Check by hand:

- Numbers run 1–5 with no gaps or repeats.
- Dependency order holds: task 2 needs task 1's interface; tasks 3 and 4 need task 2's channels.

## 6. Fill in the design

Paste this into an AI coding agent:

> Write `spectre/changes/multi-channel-notifications/design.md`. It must have exactly these
> headings, in this order: `## Context`, `## Decisions`. `## Context` states what constrains this
> change in a short paragraph. `## Decisions` records each real design choice: what was chosen,
> what alternative was considered, and the concrete reason the alternative was rejected. Do not
> use the words `TODO` or `TBD` anywhere in the file.

Wrote `spectre/changes/multi-channel-notifications/design.md`.

```markdown
## Context

The current sender talks directly to the email provider from the code path that decides a
notification should go out — there is no seam between "decide to notify" and "deliver it". Adding
a channel without that seam would mean a fourth copy of the same decision logic, one per channel,
which is why the channel abstraction (task 1) is the first task rather than the SMS channel
itself.

## Decisions

### One `Channel` interface, not a per-channel notification path

**Chosen:** a single interface every channel implements, so the send path is written once against
the interface and never learns which concrete channel it is talking to.
**Considered:** a separate send function per channel, called from a switch on notification type —
rejected because the fallback logic in task 4 would then need a case per channel pair, growing
quadratically as channels are added.

### Fallback is one retry on the next-preferred channel, not a broadcast

**Chosen:** on failure, retry once on the user's next-preferred enabled channel, then stop.
**Considered:** sending on every enabled channel simultaneously — rejected: it guarantees delivery
but also guarantees a user who prefers push still gets an email every time push is briefly down,
which is the exact overuse of email this change exists to reduce.

### Preference is per notification type, not one setting per user

**Chosen:** a user can prefer push for time-sensitive alerts and email for weekly summaries.
**Considered:** one preferred channel per user, applied to every notification type — rejected as
simpler but wrong for the stated goal: a single global preference cannot express "urgent things on
push, digests by email," which is the case that motivated the request.

### Existing users get an explicit email row, not an implicit default

**Chosen:** a task writes an explicit email preference for every existing user and notification
type at migration time.
**Considered:** leaving unset preferences to fall through to the code default forever — rejected:
it works today but means "email" is encoded in two places, the fallback rule and every existing
user's absence of a row, and the two would need to be changed together if the default ever
changed.
```

Check by hand:

- `## Context` precedes `## Decisions`.
- Every decision names a real alternative and a concrete reason it lost — not "chosen for
  simplicity".

## 7. Validate the finished change

Paste this into an AI coding agent:

> Run `spectre validate` and confirm it reports no findings.

`validate` reports no findings and exits 0; step 2's `no tasks` finding is gone.

## 8. Work the tasks

Paste this into an AI coding agent:

> Implement task 1 of `spectre/changes/multi-channel-notifications/tasks.md` — the `Channel`
> abstraction — then run the project's own build and tests, then check its box in `tasks.md`, then
> commit. Repeat for tasks 2 through 5, in order.

Per task: write the code, then check its box:

```diff
-- [ ] 1. Add a channel abstraction behind the existing email sender
+- [x] 1. Add a channel abstraction behind the existing email sender
```

- `validate` ignores box state; a task-sequence finding fires only on a missing, duplicate or
  malformed task number. The box exists for `archive`.
- This loop, not any `spectre` command, is the bulk of a real change.

## 9. Archive the finished change

Paste this into an AI coding agent:

> Stage the change with `git add spectre`, then run `spectre archive multi-channel-notifications`.
> If it refuses because a task is unchecked, go back and check the remaining boxes in `tasks.md`,
> then run `archive` again.

- `archive` uses `git mv`, so the change's files must be tracked — hence the `git add`.
- It refuses while any task is unchecked.
- The change moves to `spectre/changes/archive/multi-channel-notifications/`, staged, not
  committed — committing is the last step.
