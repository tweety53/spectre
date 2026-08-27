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

Requirement ids are `R<n>` by default (configurable, see [per-repository configuration](configuration.md)), unique within the file and
gap-free. They are stable once written: other trees cite them, and renumbering makes those
citations dangle. `(@R1)` above cites a requirement in this same file — see
[references across trees](references.md) for the other two citation forms, which cite a
capability or a peer tree and only resolve once that capability, or a `peers` entry for that
peer, actually exists.
