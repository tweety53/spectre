# Example: multi-channel-notifications, with superpowers

The same change as [example.md](example.md), filled by the `brainstorming` and `writing-plans`
skills instead of hand-written prompts — the recommended path. Both ship with spectre, unmodified
from [superpowers](https://github.com/obra/superpowers) (see the [README](../README.md#bundled-brainstorming-and-writing-plans)).
Every body below was validated clean by a real `spectre` binary.

The `spectre` skill (`/spectre new <change-id>`) runs steps 2–3 for you. Filling by hand instead:
[example.md](example.md).

## 1. Create the tree and scaffold the change

Paste this into an AI coding agent:

> Set up a spectre tree in the current directory (run `git init` first if it isn't a git
> repository, then `spectre init`), then run `spectre new multi-channel-notifications`.

Result: the tree and three stubs — see [example.md steps 1–2](example.md#1-create-the-tree).

## 2. Brainstorm the design

Paste this into an AI coding agent:

> Use the brainstorming skill to design `multi-channel-notifications`: a backend that sends
> every notification by email should add SMS, messenger and push. Once I approve the design, write
> it into spectre's artifacts instead of a separate design doc: `spectre/specs/notifications.md`
> (`# notifications`, `## Purpose`, `## Requirements`, bullets `- R<n>: …` with `SHALL`),
> `spectre/changes/multi-channel-notifications/proposal.md` (`# multi-channel-notifications`,
> `## Why`, `## What changes`) and `design.md` (`## Context`, `## Decisions` — each decision with
> `**ID:**`, `**Status:**`, `**Chosen:**`, `**Considered:**` — then `## Open questions`).

Brainstorming asks one question at a time before writing anything, for example:

```text
Who picks the channel — the user per notification type, one global setting, or the sender?
When the preferred channel fails, retry one other channel or send on all of them?
Is a delivery receipt (read/seen) in scope, or only "the provider accepted it"?
Do existing users keep email-only behaviour until they choose otherwise?
```

`spectre/specs/notifications.md`:

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
- R3: When delivery on a channel fails, the system SHALL retry once on the user's next-preferred
  enabled channel before marking the notification undelivered.
- R4: The system SHALL let a user set a distinct preferred channel per notification type, not
  only one preference for all notifications.
- R5: The system SHALL treat delivery as succeeded when the channel's provider accepts the
  message; read and seen receipts are out of scope.
- R6: The system SHALL record, per notification, the channel it was delivered on or that it went
  undelivered.
```

What the skill added over example.md's spec:

- R3 states "once" — the retry count the design settled, not left open.
- R5 pins what "delivered" means, answering a question the hand-written spec never asked.
- R6 makes the outcome observable, so the fallback in R3 is testable.

`spectre/changes/multi-channel-notifications/proposal.md`:

```markdown
# multi-channel-notifications

## Why

- The backend sends every notification by email today; open rates on email lag every channel
  our users check first.
- A user with no email client open misses a time-sensitive notification entirely — there is no
  faster channel to fall back to.
- Support tickets already ask for SMS and push; messenger is the channel our newest users expect
  by default.

## What changes

- Notifications gain SMS, messenger and push channels beside email (R1).
- A user sets a preferred channel per notification type (R4); email stays the fallback when
  none is set (R2).
- A failed delivery retries once on the next-preferred enabled channel (R3), and every
  notification records where it landed (R6).
- Existing users get an explicit email preference, so their behaviour does not change.

## Out of scope

- Read and seen receipts (R5).
- Quiet hours, batching and digests.
```

What the skill added:

- Each change names the requirement it satisfies.
- `## Out of scope` records what brainstorming ruled out — an extra heading `validate` allows.

`spectre/changes/multi-channel-notifications/design.md`:

```markdown
## Context

- The sender calls the email provider directly from the code that decides a notification should
  go out; there is no seam between "decide to notify" and "deliver it".
- Adding channels without that seam means one copy of the decision logic per channel, so the
  channel abstraction comes first.

## Decisions

### One channel interface

**ID:** channel-interface
**Status:** active
**Chosen:** one `Channel` interface every channel implements; the send path is written once
against it.
**Considered:** a send function per channel behind a switch — the fallback would need a case per
channel pair, growing quadratically as channels are added.

### Fallback is one retry, not a broadcast

**ID:** single-retry-fallback
**Status:** active
**Chosen:** on failure, retry once on the next-preferred enabled channel, then mark undelivered.
**Considered:** send on every enabled channel at once — guarantees delivery, but a user who
prefers push gets an email every time push is briefly down, the overuse this change exists to
reduce.

### Preference is per notification type

**ID:** per-type-preference
**Status:** active
**Chosen:** one preferred channel per user per notification type.
**Considered:** one preference per user — cannot express "urgent on push, digests by email", the
case that motivated the request.

### Existing users get an explicit email row

**ID:** explicit-email-backfill
**Status:** active
**Chosen:** a backfill writes an email preference for every existing user and notification type.
**Considered:** leave them unset and rely on the fallback — "email" would then live in two places,
the fallback rule and every user's missing row, which must change together.

## Open questions

- Which SMS provider, and its per-message cost ceiling — decided before task 2's SMS channel ships.
```

What the skill added:

- Every decision has an `**ID:**` and `**Status:**`, so a later change can supersede one by name.
- `## Open questions` keeps what brainstorming left undecided visible instead of silently guessed.

## 3. Write the plan

Paste this into an AI coding agent:

> Use the writing-plans skill to plan `multi-channel-notifications` from its proposal, design
> and spec, written into `spectre/changes/multi-channel-notifications/tasks.md`. Keep spectre's
> task shape: first line `# multi-channel-notifications`; each task a column-0
> `- [ ] <n>. <title>` line, numbered from 1 with no gaps; its steps indented two columns as
> `  - [ ] **Step N: …**`. Give each task its files, its test, a verify command and a commit.

`spectre/changes/multi-channel-notifications/tasks.md`:

```markdown
# multi-channel-notifications

**Goal:** deliver notifications on email, SMS, messenger or push per user preference, with one
fallback retry.

- [ ] 1. Add a channel abstraction behind the existing email sender

**Files:** `notify/channel`, `notify/email`, `notify/send`
**Test:** `notify/send_test` — the send path delivers through a fake channel.

  - [ ] **Step 1: Write the failing test.** The send path delivers one notification through a fake
    `Channel` and gets its success back.
  - [ ] **Step 2: Run it.** `make test TEST=notify/send` — fails: no `Channel` exists.
  - [ ] **Step 3: Add `Channel`** (send one notification, report success or failure) and make the
    email sender its first implementation; route the send path through it.
  - [ ] **Step 4: Verify.** `make test` — all pass, email behaviour unchanged.
  - [ ] **Step 5: Commit** `notify: send through a Channel interface`.

- [ ] 2. Add the SMS, messenger and push channel implementations

**Files:** `notify/sms`, `notify/messenger`, `notify/push`
**Test:** one test per channel against a fake transport (R1).

  - [ ] **Step 1: Write the failing tests.** Each channel sends through a fake transport and
    reports a transport error as failure.
  - [ ] **Step 2: Run them.** `make test TEST=notify/sms,notify/messenger,notify/push` — fail.
  - [ ] **Step 3: Implement each channel** against its provider client.
  - [ ] **Step 4: Verify.** `make test` — all pass; no test reaches a real provider.
  - [ ] **Step 5: Commit** `notify: add SMS, messenger and push channels`.

- [ ] 3. Add per-notification-type channel preference

**Files:** `notify/preference`, `notify/send`
**Test:** `notify/preference_test` (R4, R2).

  - [ ] **Step 1: Write the failing tests.** A stored push preference for `alert` sends `alert` on
    push and `digest` on email; no stored preference sends on email.
  - [ ] **Step 2: Run them.** `make test TEST=notify/preference` — fail.
  - [ ] **Step 3: Store one channel per user per notification type** and read it in the send path.
  - [ ] **Step 4: Verify.** `make test` — all pass.
  - [ ] **Step 5: Commit** `notify: choose the channel by per-type preference`.

- [ ] 4. Add fallback on delivery failure

**Files:** `notify/send`, `notify/outcome`
**Test:** `notify/fallback_test` (R3, R6).

  - [ ] **Step 1: Write the failing tests.** Preferred channel fails → one retry on the
    next-preferred enabled channel, outcome records it; both fail → outcome `undelivered`, no third
    attempt.
  - [ ] **Step 2: Run them.** `make test TEST=notify/fallback` — fail.
  - [ ] **Step 3: Retry once** in the user's configured order, then record the outcome.
  - [ ] **Step 4: Verify.** `make test` — all pass.
  - [ ] **Step 5: Commit** `notify: retry once on the next-preferred channel`.

- [ ] 5. Backfill channel preference for existing users

**Files:** `migrations/backfill_email_preference`
**Test:** `migrations/backfill_email_preference_test`.

  - [ ] **Step 1: Write the failing test.** After the backfill, every existing user has an email
    row for every notification type they receive; a second run changes nothing.
  - [ ] **Step 2: Run it.** `make test TEST=migrations/backfill_email_preference` — fails.
  - [ ] **Step 3: Write the backfill**, idempotent.
  - [ ] **Step 4: Verify.** `make test` — all pass.
  - [ ] **Step 5: Commit** `migrations: backfill email preference for existing users`.
```

What the skill added over example.md's plan:

- Every task names its files and its test; tasks 2–4 name the requirements they deliver.
- Every task is test-first: a failing test, the command that shows it failing, the fix, a verify
  command and a commit.
- Steps are indented `- [ ]` lines; `spectre list` still counts only the five column-0 tasks.

## 4. Validate

Paste this into an AI coding agent:

> Run `spectre validate` and `spectre list`.

```text
$ spectre validate
no findings
$ spectre list
multi-channel-notifications  0/5
```

## 5. Work and archive

Work and archive the change exactly as in [example.md steps 8–9](example.md#8-work-the-tasks):
implement each task, tick its column-0 box, commit, then `git add spectre` and `spectre archive
multi-channel-notifications`. Step boxes are optional progress marks; only task boxes gate
`archive`.
