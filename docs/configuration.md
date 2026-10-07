# Per-repository configuration

What `spectre/config.md` can override, and what stays fixed. The file is optional; absent, every
default below applies.

```markdown
# spectre config

## Rules
- shall-clause: error      # error | off, per rule
- placeholders: error
- headings: error
- malformed-bullet: error
- id-sequence: error
- task-sequence: error
- refs: error

## Vocabulary
- modal: SHALL             # the verb a requirement bullet must carry
- id-prefix: R              # replaces the accepted prefix entirely — "id-prefix: REQ-" makes
                             # REQ-1 legal and R1 no longer legal, not both at once

## Layout
- specs: specs              # relative to the tree root
- changes: changes
- extension: .md
```

- Keys are closed: an unknown key, rule name or value exits 2 rather than silently disabling a
  check.
- A key set twice — even within the same section — is an error naming both lines.
- Headings and bullets inside fenced code blocks are ignored, so an example in a ` ``` ` fence
  never changes a setting.
- Configuration is per tree: a peer's own `config.md` governs its files, so a `REQ-` tree can be
  cited from an `R` tree.
- What `new` scaffolds is not configurable; it is compiled into the binary.
