---
id: spec-0001
task_ref: task-0001
status: implemented
created: 2026-09-03T22:30:34Z
---

# spec-0001 — Dispatch, repo detection, version and help

**References:** [task-0001](../tasks/task-0001-command-frame.md)

- **Goal:** one Go skeleton every command plugs into, implementing `docs/product/rules.md` once.

## Scope

In: module layout (`cmd/writrun/`, `internal/`), subcommand dispatch
(standard library), adopted-repository detection, the interaction
helpers — arrow-key selection, confirmation, `--yes` — plain-text
reporting, exit-status discipline, `--version` and `--help`. Out: the
behaviour of any real command.

## Steps

1. `go.mod` — module `github.com/thomasfranke/writrun-cli`; `cmd/writrun/main.go` holds `main` and production wiring only.
2. Packages per `technical/layout/tree.md`: `internal/command/` (one use case per command, depending on ports only), `internal/kit/` (exec port), `internal/forge/` (gh port), `internal/term/` (terminal port), `internal/wrepo/` (detection, settings). Each port is an interface defined in its consuming package, with a fake beside it (`technical/engineering/boundaries.md`).
3. Dispatch: subcommand table; `--help`/`--version` answer anywhere — `--version` names the client's version and the pinned WritRun tag; `--help` is one line per command plus the docs' address. Detection: walk to the git toplevel, require `.writrun/` per the command's declared need (adopted / not adopted / any).
4. Interaction helpers on `github.com/charmbracelet/huh` behind the term port: a select (arrow keys move, Enter confirms), a confirm, and a spinner for waits; each renders only where stdin is a terminal, each has a flag equivalent (`--yes`, the command's own flags), and each aborts — never hangs — when no terminal and no flag answers.
5. Confirmation flow: print what will be done, ask; `--yes` skips; decline changes nothing.
6. Reporting: plain text, no pager; color only where stdout is a terminal, off under `NO_COLOR` or `--no-color`; a failure names the failing check; exit 0 only when the command did what it said.

## Acceptance criteria (EARS)

- When invoked with `--version` or `--help`, in any directory, the system shall answer and exit 0.
- When `--version` answers, it shall name the client's version and the pinned WritRun tag.
- When stdout is not a terminal, or `NO_COLOR` is set, or `--no-color` is given, output shall carry no color.
- When a command requiring an adopted repository runs outside one, the system shall abort naming the cause, change nothing, and exit non-zero.
- When a mutating command runs without `--yes`, the system shall show its plan and ask before acting.
- When the user declines the prompt, the system shall exit non-zero having changed nothing.
- When stdin is not a terminal and a question has no flag answer, the system shall abort naming the missing flag.

## Edge cases

- Not a git repository at all.
- `.writrun/` present in a parent directory (command run from a subdirectory).
- Unknown subcommand or flag: usage on stderr, exit non-zero.

## Tests required

The three tiers of `technical/testing/tiers.md`: table-driven unit tests with
every port faked; `tests/integration/cli/` driving the compiled binary
against fixture repositories on the release suite's harness; the CI
order (`gofmt` → `go vet` → unit with the 85% coverage gate →
integration → e2e) wired into `tests.yml` by this task.

## Definition of Done

- [ ] `go build ./...` cross-compiles for macOS, Linux, Windows without cgo.
- [ ] `gofmt -l` clean; `go vet` clean; coverage over `internal/` ≥ 85%, gated in CI.
- [ ] Every acceptance criterion has a test; all tiers green in the pipeline.

## Proposed product changes

- none — the behaviour implemented is already stated in `product/rules.md`.

## Proposed technical changes

- none — `technical/engineering/`, `technical/layout/` and `technical/testing/` already state the machinery this task builds.

## Outcome

Shipped as specified. The frame lives in `internal/command/` —
dispatch, the three needs, the interaction helpers, the color rule,
exit discipline — with `wrepo.Find` for detection, the term port on
`huh` (arrow keys exercised headless in tests), and the `kit`/`forge`
adapters. `--version` names the client and the pinned tag. 55 unit
tests and 4 CLI integration cases; coverage over `internal/` at 95%
against the 85% gate, wired into `tests.yml` in the tier order;
cross-compiles for macOS, Linux and Windows without cgo, gated in
`tests.yml`. One reading
of scope: `Frame.Commands` ships empty — the table exists, and each
command joins it with its own task.
