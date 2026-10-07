# Review panel — kan-917-superpowers-example-and-brief-docs

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | README.md:44 | README and superpowers-example.md name /flow, which no public doc defines or links |   |
| F2 | primary | Minor | .claude/skills/brainstorming/SKILL.md:3 | bundled brainstorming triggers in every project once installed and duplicates an existing superpowers install; README does not disclose it |   |
| F3 | primary | Minor | .claude-plugin/plugin.json:5 | version bump is in no task-6 step and has no Correction line |   |
| F4 | principles | Minor | .claude/skills/brainstorming/SKILL.md:3 | Least Astonishment: bundled skills apply plugin-wide and collide with a separate superpowers install, undisclosed |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-flow-undefined-1.sh
finding-reproducer: F2 none — a disclosure gap; no runnable behaviour to assert
finding-reproducer: F3 none — plan bookkeeping, no runtime behaviour
finding-reproducer: F4 none — a disclosure gap; no runnable behaviour to assert

## Pass log

### Round 0

- diff size: 3617 lines, under cap; docs-only: exit 1 (.claude-plugin/marketplace.json) — resolved roster unchanged; roster: compact — 72; dispatched primary+principles on opus/medium; no operator addition this round — the resolved list ran alone; standards: none declared; citation check: no project command declared
