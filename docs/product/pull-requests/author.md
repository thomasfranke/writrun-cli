# `writrun author`

Ends the authoring flow: a rule was declared finished, and the derived
tasks and draft specs go up for review in one pull request.

- [The shape](shape.md) holds.
- The composed title and body are handed to the repository's own
  observance check before the push, so a title outside the declared
  style is refused here rather than on the forge.
- It carries a derivation already made; the command derives nothing
  itself ([rules](../rules.md)).
- The pull request opens ready for review — an authoring change
  announces no work.
