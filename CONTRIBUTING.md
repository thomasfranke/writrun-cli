# Contributing to writrun-cli

Thanks for being here. `writrun` is a command-line client: one binary
that wraps [WritRun](https://github.com/thomasfranke/writrun)'s own
scripts and files into human-shaped commands. It packages; it never
decides.

**Two repositories, two jobs.** The methodology — its rules, its skills,
its schemas — is developed at
[`thomasfranke/writrun`](https://github.com/thomasfranke/writrun), and a
change to what WritRun *means* belongs there. This repository is the
porcelain over it, and everything below is about this one.

Not sure whether an idea fits, or where to start? Open an issue and ask.
Asking early is usually faster than guessing — for both of us.

## Before you start

- **Read [`docs/about.md`](docs/about.md).** It carries what this project
  is reaching for and, more usefully, its non-goals. If one of them
  strikes you as wrong, that is worth hearing — open an issue and make
  the case.
- **[`docs/product/`](docs/product/README.md) is the source of truth**
  the implementation is checked against: every command, rule by rule, in
  non-technical language. If your change contradicts a rule there, that
  is a conversation for an issue, not a surprise in a pull request.
- **[`docs/technical/`](docs/technical/README.md)** covers how it is
  built, tested, versioned and distributed, and
  [`decisions/`](docs/technical/decisions/README.md) records the *why*
  and what was rejected.

**This project is docs-first.** `docs/product/` names every command
before any of them exists; the binary follows. A pull request that adds
behaviour no chapter describes is proposing a rule, and that is worth
saying out loud rather than reviewing as code.

## The four relationships that decide scope

A change that breaks one of these is out of scope no matter how
convenient. They live in [`docs/about.md`](docs/about.md); they are
repeated here because they are what a review looks for first.

- **Wraps, never reimplements.** Every check a command runs is the
  adopted repository's own `.writrun/` script, as a child process —
  nothing reimplemented in Go, nothing embedded in the binary. If the
  binary and the scripts disagree, the scripts are right.
- **Packages, never decides.** No command makes a call a human or an
  agent had not already made. There is no approve command, and nothing
  merges.
- **Launches agents, never is one.** `writrun work` starts one;
  `writrun` never reasons about a task's content.
- **Pins, never tracks.** Each release targets exactly one WritRun tag.
  Before changing anything that reads the adopted repository's layout,
  check that tag's own docs rather than the shape you remember — the
  methodology's contract is alpha and moves without notice.

## Development setup

**The binary is Go** — one static build, standard library for flags and
dispatch, `charmbracelet/huh` for the prompts. Entry point in
`cmd/writrun/`, everything else `internal/`; nothing is exported for
import, because the public contract is the command line. See
[`runtime/`](docs/technical/runtime/README.md) and
[`layout/`](docs/technical/layout/README.md).

**The release path is bash**, and it has a suite:

```bash
make tests               # everything — or: bash tests/run.sh
make test-integration    # one tier
make test-release        # one suite directory
```

Any case file also runs on its own. Each builds a throwaway repository
and asserts an exit code, on git, `bash` and POSIX `awk`/`sed` — the
same constraints as the scripts it exercises. Test on macOS's stock
`/usr/bin/awk`, not just whatever is on your `$PATH`. **If you change a
script, add the case that would have caught the change being wrong.**
[`testing/`](docs/technical/testing/README.md) has the layout.

## How contributions reach the project

Nobody needs write access. **Fork the repository, work on a branch in
your fork, and open a pull request against `main`.**

Merging is restricted to the maintainer. That is not a comment on trust:
a permanent doc never merges on single-reviewer approval alone, and
keeping that responsibility in one place is the simplest way to honour
it.

**Keep your fork in sync.** Branch from an up-to-date `main`, or your
pull request is reviewed against a moving target.

## Workflow

**Trunk-based.** `main` is the only long-lived branch and is always
green. Everything merges to `main` continuously.

1. Branch off the latest `main`, named `<type>/short-name` —
   `fix/broken-anchor`, `docs/release-chapter`. Rebase on `main` rather
   than merging it in; it keeps the squash clean.
2. **Commit subjects are [Conventional
   Commits](https://www.conventionalcommits.org/)**: `type(scope):
   imperative summary`. Types are `docs`, `feat`, `fix`, `refactor`,
   `chore`. The scope is optional and names the subsystem — `product`,
   `technical`, `release`, `tests`, `ci` — omitted when a change
   genuinely spans the repository.
3. Open the pull request against `main`. Say what changed and why, and
   name anything a reviewer should re-read by hand rather than trust the
   diff for.
4. **Merge is squash-only.** A messy branch history is fine; the commit
   landing on `main` is not.
5. **English everywhere:** prose, commits, documentation.

Update the docs in the same pull request that changes the behaviour they
describe — `docs/product/` when a rule moves, `docs/technical/` when the
machinery does, and a dated entry in
[`decisions/`](docs/technical/decisions/README.md) when a choice was made
that a future reader will ask about. Entries there are append-only:
never edit one, add the next.

## Releases

Tags on `main`, cut with `make release`. The number is computed from the
latest tag, never typed, and is SemVer — `patch` (the default), `minor`,
`major`. `CHANGELOG.md` is written by the cut and **never edited by
hand**. The whole path, and why it is shaped that way, is
[`versioning/`](docs/technical/versioning/README.md).

## Licence, and why there is no CLA

`writrun-cli` is [MIT-0](LICENSE) — MIT with no attribution requirement.
**There is no CLA to sign.** Inbound equals outbound: by opening a pull
request you offer your contribution under the same MIT-0 terms as the
rest of the project, and you keep the copyright to your own work.

That is deliberate, not an oversight. A CLA exists to let a project
relicense contributed code, usually for a commercial edition. This
project has none and no plan for one, so there is nothing a CLA would
secure and no reason to ask you for a signature.

## Code of conduct

Be decent. Assume good faith, critique the work and not the person, and
remember that most people here are doing this in their spare time.
