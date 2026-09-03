# Contributing to writrun-cli

Thanks for being here. `writrun` is the porcelain for
[WritRun](https://github.com/thomasfranke/writrun) — one binary that wraps
the methodology's own scripts and files into human-shaped commands. It
packages; it never decides.

Not sure whether an idea fits, or where to start? Open an issue and ask.
Asking early is usually faster than guessing — for both of us.

## Before you start

- **Read [`docs/about.md`](docs/about.md).** It carries what this project
  is reaching for and, more usefully, its non-goals. If one of them strikes
  you as wrong, that is worth hearing — open an issue and make the case.
- **Read the rules you are changing.**
  [`docs/product/`](docs/product/README.md) defines every command and is the
  source of truth the implementation is checked against.
  [`docs/technical/`](docs/technical/README.md) covers how it is built, and
  [`decisions.md`](docs/technical/decisions.md) records the *why* and what
  was rejected.
- **This repository runs on WritRun itself.** Its `work/` queue is where its
  own work is tracked, and [`AGENTS.md`](AGENTS.md) is the entry point for
  anyone — human or agent — arriving at a task.

## The four relationships that decide every change

A change that breaks one of these is out of scope no matter how convenient.
[`docs/about.md`](docs/about.md) is where they live; they are repeated here
because they are what a review looks for first.

- **Wraps, never reimplements.** Every check is the operated repository's
  own. If the binary and the scripts disagree, the scripts are right.
- **Packages, never decides.** No command makes a call a human or an agent
  had not already made. There is no approve command.
- **Launches agents, never is one.**
- **Pins, never tracks.** [`.writrun/VERSION`](.writrun/VERSION) records the
  one WritRun tag this client targets. Before changing anything that reads
  the operated repository's layout, check that tag's own docs rather than
  the shape you remember.

## The work is defined in tasks and specs

[`work/tasks/`](work/tasks/README.md) is **the queue** — what to do, when,
and what blocks it. [`work/specs/`](work/specs/README.md) is **the
detail** — scope, steps, criteria, edge cases.

**Nothing is "planned" in prose:** it is either a task in the queue or it
does not exist yet.

- **See what is available**, rather than browsing the folder:

  ```bash
  bash .writrun/skills/writrun-select-next-task/list_tasks.sh
  ```

  It prints what is eligible, what must be resumed first, and what is held
  back with the reason. **You may take any task it lists as available** —
  the order shown is a suggestion to you. An agent is bound by it, so that
  repeated sessions agree; you are not.
- **Take a task only when it is `ready`.** That status is written from the
  fact that every spec it references is `approved`. A `backlog` task has not
  passed the approval gate, so it is not authorized work.
- **Trivial work does not become a task** — a typo or a one-line fix goes
  straight to a commit.
- **Taking a task means opening its draft pull request, before the work
  starts.** A branch on your machine is invisible to everyone else, and the
  draft is the event the machinery answers by writing `in-progress`. It
  reserves nothing; it reports.
- **The task's status line is never yours to write.** The machinery owns
  it. What you write is the task's `completed` date, by hand, when the work
  is done.
- **Close the loop:** fill the spec's **Outcome** with what was actually
  built and anything that diverged, and why. Do not edit the
  Proposed-changes sections to match reality after the fact.
- **Update the permanent docs in the same change** — `docs/product/` if
  behaviour changed, `docs/technical/` if machinery did. A spec is a
  historical record, not where current behaviour is documented.

The full flow, the front-matter schemas, and the order the checks run in are
in [`AGENTS.md`](AGENTS.md) and the
[`.writrun/skills/`](.writrun/README.md) it points at. The generator script
scaffolds a new task or spec correctly rather than relying on getting the
schema right from memory.

## Development setup

**The binary is Go**, one static build, standard library first — see
[`runtime.md`](docs/technical/runtime.md) and
[`layout.md`](docs/technical/layout.md). No package manager beyond the Go
toolchain, and nothing is exported for import: the public contract is the
command line.

**The release path is bash**, and it has a suite:

```bash
make tests               # everything — or: bash tests/run.sh
make test-integration    # one tier
make test-release        # one suite directory
```

Any case file also runs on its own. Each builds a throwaway repository and
asserts an exit code, on git, `bash` and POSIX `awk`/`sed` — the same
constraints as the scripts it exercises. **If you change a script, add the
case that would have caught the change being wrong.**
[`testing.md`](docs/technical/testing.md) has the layout.

## How contributions reach the project

Nobody needs write access. **Fork the repository, work on a branch in your
fork, and open a pull request against `main`.**

Merging is restricted to the maintainer. That is not a comment on trust — a
permanent doc never merges on agent or single-reviewer approval alone, and
keeping that responsibility in one place is the simplest way to honour it.

**Keep your fork in sync.** Branch from an up-to-date `main`, or your pull
request is reviewed against a moving target.

## Workflow

**Trunk-based.** `main` is the only long-lived branch and is always green —
this repository's choice, per
[`.writrun/conventions/branches.md`](.writrun/conventions/branches.md).

1. Branch off the latest `main`, named per the conventions above. Rebase on
   `main` rather than merging it in — it keeps the squash clean.
2. Write commits, branch names and pull request titles per
   [`.writrun/conventions/`](.writrun/conventions/README.md): Conventional
   Commits with this repository's types and scopes, and squash-only merges.
   **That folder is this project's own convention, not the methodology's.**
3. Open the pull request against `main` and fill in
   [the template](.writrun/templates/pull_request_template.md).
4. **English everywhere:** prose, commits, documentation.

CI verifies the methodology, not the code: whether the change kept its
promises to the docs. Whether the code works is the suite's answer, and both
run on every pull request.

Releases are tags on `main`. Cut one with `make release` — the number is
computed from the latest tag, never typed. The vocabulary and what the cut
stamps are in [`versioning.md`](docs/technical/versioning.md).

**While alpha, any part of the contract may move without notice.** Each
release targets exactly one WritRun tag and nothing else.

## Licence, and why there is no CLA

`writrun-cli` is [MIT-0](LICENSE) — MIT with no attribution requirement.
**There is no CLA to sign.** Inbound equals outbound: by opening a pull
request you offer your contribution under the same MIT-0 terms as the rest
of the project, and you keep the copyright to your own work.

That is deliberate, not an oversight. A CLA exists to let a project
relicense contributed code, usually for a commercial edition. This project
has none and no plan for one, so there is nothing a CLA would secure and no
reason to ask you for a signature.

## Code of conduct

Be decent. Assume good faith, critique the work and not the person, and
remember that most people here are doing this in their spare time.
