# Spec format

How a capability spec file is written, and what makes a requirement id valid.

```markdown
# auth

## Purpose
Sessions and tokens.

## Requirements
- R1: The system SHALL refresh the session token before expiry.
  Indented prose beneath a bullet is free-form and preserved.
- R2: The picker SHALL show only enabled plans.
```

- Ids are `R<n>` by default ([configurable](configuration.md)), unique within the file and
  gap-free.
- Ids are stable once written.
