# References across trees

How a citation names a requirement in this file, this tree, or a peer tree — and how `validate`
and `refs` use that.

Three scopes, widening left to right:

| Form | Meaning |
|------|---------|
| `(@R4)` | same file |
| `(@plans#R4)` | same tree, capability `plans` |
| `(@web:plans#R4)` | peer tree `web`, capability `plans` |

A citation naming a peer accepts any id shape, since the peer's own id prefix is that peer's own
business. A citation naming no peer must match this tree's own configured id prefix, or it is not
read as a reference at all — this stops `(@v2)` in ordinary prose from becoming a phantom citation.

`validate` resolves each reference it does read: peer declared in `peers`, its tree present on
disk, the capability file found inside it, the id found in that file — and reports the first
failing check as `file:line: message`. `refs` answers the reverse question — every place a
requirement is cited — before you renumber or delete it.

`peers` is the only file spectre reads besides specs, changes and `config.md`:

```
web ../web
web-frontend ../web-frontend
```

One name per line, paths relative to the tree's parent directory. A name repeated in `peers` is an
error naming both lines. Two different names resolving to the same path is legal — an alias.

A path may name either the peer repository or its tree directly — `web ../web` and
`web ../web/spectre` both resolve to `../web/spectre`. Which one it already is is decided by
probing for `changes/` inside it, not by comparing the path's basename to `spectre`: a repository
whose own directory happens to be named `spectre` is not mistaken for the tree it contains.
