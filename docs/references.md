# References across trees

How a citation names a requirement in this file, this tree or a peer tree, and how `validate`
and `refs` use it.

| Form | Meaning |
|------|---------|
| `(@R4)` | same file |
| `(@plans#R4)` | same tree, capability `plans` |
| `(@web:plans#R4)` | peer tree `web`, capability `plans` |

- A citation naming a peer accepts any id shape — the peer's prefix is its own business.
- A citation naming no peer must match this tree's configured id prefix, or it is not read as a
  reference — so `(@v2)` in prose is no phantom citation.
- `validate` resolves each reference — peer declared in `peers`, its tree on disk, the capability
  file in it, the id in that file — and reports the first failing check as `file:line: message`.
- `refs` answers the reverse — every place a requirement is cited — before you renumber or delete
  it.

`peers` is the only file spectre reads besides specs, changes and `config.md`:

```
web ../web
web-frontend ../web-frontend
```

- One name per line; paths relative to the tree's parent directory.
- A repeated name is an error naming both lines; two names resolving to one path is legal — an
  alias.
- A path names the peer repository or its tree: `web ../web` and `web ../web/spectre` both resolve
  to `../web/spectre`. Which it is comes from probing for `changes/` inside it, not from the
  basename — a repository whose own directory is named `spectre` is not mistaken for its tree.
