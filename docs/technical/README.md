# Technical documentation

**How the system is built**, for developers and agents. It never restates
a product rule — it links to the rule and explains the machinery under
it.

[`architecture.md`](architecture.md) is the map — how `writrun` wraps
the adopted repository's scripts, and the contract it pins. Read it
first. Each subject then has its folder:

| | |
|---|---|
| [Runtime](runtime/README.md) | Language, platforms, and what `writrun` requires to run. |
| [Layout](layout/README.md) | Where the code lives and what is public. |
| [Engineering](engineering/README.md) | Style, boundaries, and the principles the code is held to. |
| [Versioning](versioning/README.md) | The computed version number and the release path. |
| [Testing](testing/README.md) | The suite, its tiers, and the two CI workflows. |
| [Distribution](distribution/README.md) | Homebrew, release binaries, `go install`. |
| [Decisions](decisions/README.md) | Dated, append-only, one numbered file each — what was decided and what was rejected. |
