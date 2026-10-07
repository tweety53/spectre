# kan-917-superpowers-example-and-brief-docs

## Why

- spectre's artifacts are only as good as what fills them; the docs never say that
  `superpowers:brainstorming` and `superpowers:writing-plans` are the best way to fill them, although
  `/flow` already works that way.
- The standalone `spectre` skill fills `new`'s stubs by hand even when those skills are installed,
  and a reader without superpowers must find and install it separately.
- The README and docs restate themselves and each other; readers wade through prose to find a rule.

## What changes

- README names the superpowers pairing as the recommended way to use spectre, and is rewritten short.
- New `docs/superpowers-example.md`: the `multi-channel-notifications` change, filled by
  brainstorming and writing-plans, every body validated clean.
- The `spectre` skill's `new` step invokes both skills, falling back to hand-filling only when
  neither is available.
- The plugin bundles `brainstorming` and `writing-plans`, unmodified from superpowers v6.4.2 with
  its MIT licence and a pointer to the original repository, so one install brings everything.
- Every user-facing doc under `docs/` is trimmed to say each thing once, losing no fact.
