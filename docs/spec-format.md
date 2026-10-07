# Spec format

How a capability spec file is written, and what makes a requirement id valid.

```markdown
# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the session token before expiry.
  Indented prose beneath a bullet is free-form and preserved.
- R2: The picker SHALL show only enabled plans (@R1)
```

- Ids are `R<n>` by default ([configurable](configuration.md)), unique within the file and
  gap-free.
- Ids are stable once written: other trees cite them, and renumbering leaves those citations
  dangling.
- `(@R1)` cites this same file; the capability and peer forms
  ([references across trees](references.md)) resolve only once that capability, or that peer's
  `peers` entry, exists.
