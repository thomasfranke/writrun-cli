# Language

- **Go**, one static binary per platform. Binary: `writrun`. Module
  path: `github.com/thomasfranke/writrun-cli`.
- Standard library for flags and dispatch — no CLI framework. The one
  interaction dependency is `github.com/charmbracelet/huh`, for the
  prompts [`product/rules.md`](../../product/rules.md) requires
  (decision
  [0009](../decisions/runtime/0009-interaction-via-charm-huh.md)).
