---
id: spec-0023
task_ref: task-0024
status: implemented
created: 2026-09-06T04:28:28Z
---

# spec-0023 — Judge the title at composition

**References:** [task-0024](../tasks/task-0024-observance-at-composition.md)

- **Goal:** `writrun amend` and `writrun author` hand `check_observance.sh` the title and body they composed, before either writes anything, so a composition the door would reject is refused locally instead of becoming a pushed branch and an open pull request with a red check.

## Scope

In: `internal/kit` — an environment seam on the script runner, so
`PR_TITLE` and `PR_BODY` reach a script through the environment as
`check_observance.sh` requires and never through argv, which the script
forbids by name. Whether that changes `kit.Runner`'s signature across
its nine call sites or adds a sibling entry point is this spec's to
decide.

In: `internal/command/amendcmd` — the check invoked after the title and
body are composed and before the confirmation, with the range
`HEAD...HEAD`, which names no commits and is correct because at that
instant the amendment's commit does not exist. `amend`'s first write is
`git switch -c`, so a refusal there leaves nothing behind.

In: `internal/command/authorcmd` — the same check added to its ordered
list, after the title is asked for and before the push, with the real
range in hand so both halves of the script are live. `author`'s first
write is the push, so a refusal there also leaves nothing behind.

In: `docs/product/pull-requests/` — whether `shape.md`'s "checks first,
in their load-bearing order" needs a sentence for a check that can only
run once the composition exists, and whether `amend.md` and `author.md`
say the title is judged before the forge sees it.

**The vocabulary stays the kit's.** Nothing in Go holds a copy of
`TYPES` or `SCOPES`; the verdict is the script's.

Out: the commit subject and the branch name. The subject is already
judged by the commit-msg hook `init` installs, which reads the same
field; the branch prefix is judged by nothing anywhere, and giving it a
judge is a separate finding. Out: `take` and `finish`. Out: any change
to `check_observance.sh` or anything under `.writrun/`.

## Steps

1. Give the script runner a way to pass environment, and say at the seam why argv is not an option.
2. Invoke the check in `amend` after composition and before the confirmation, passing the script's verdict up unchanged the way the other commands pass a script's exit code.
3. Invoke it in `author` in its ordered check list, with the range it already resolved.
4. Decide the doc question in Scope, and either write the sentences or record why none is needed.

## Acceptance criteria (EARS)

- When `amend` composes a title outside the declared vocabulary, the system shall refuse before creating a branch, and the working tree shall be unchanged.
- When `author` composes such a title, the system shall refuse before pushing.
- When the composed title observes the declared style, the system shall proceed exactly as it does today.
- When the check refuses, the system shall carry the script's own exit code and its own words.
- The system shall hold no copy of the type or scope vocabulary in Go.

## Edge cases

- A repository whose `check_observance.sh` is missing or unreadable: the command says so rather than passing silently, and rather than falling back to a judgement of its own.
- `--yes`: the refusal is a check, not a question, so it fires on the unattended path too.
- A title that is valid but whose body carries a credit line while `agent_coauthor` is false: the same call catches it, which is a second thing this buys.

## Tests required

Integration over the fixture: `amend --type wibble` refused with a clean
tree, no local branch and nothing on the stub forge; the same for
`author`; the valid path unchanged. One case asserting the vocabulary
lives only in the kit — grep the Go tree for the type list and find
nothing.

## Definition of Done

- [ ] `writrun amend --type wibble` is refused locally, and `git status --porcelain` is empty afterwards.
- [ ] `writrun author` with a title outside the style is refused before the push.
- [ ] No Go file contains the type or scope vocabulary.

## Proposed product changes

- `product/pull-requests/shape.md` — the order bullet, for a check whose input is the composition.
- `product/pull-requests/amend.md` — the check named as amend's, before the branch is cut.
- `product/pull-requests/author.md` — the check named as author's, before the push.

Step 4 decided all three are needed; the Outcome records why. One path
per bullet, each written from `docs/`: `check_deltas.sh` reads the first
backticked span of a bullet line and prefixes `docs/`, so three paths on
one line declare only the first, and a bare `amend.md` declares
`docs/amend.md`.

## Proposed technical changes

- none — the seam is internal and adds no package.

## Outcome

`amend` and `author` hand the composed title and body to
`check_observance.sh` and carry its verdict up unchanged. `amend` asks
after the composition and before the confirmation, over `HEAD...HEAD`;
`author` asks after its three diff checks and before the push, over the
range it already resolved. Neither holds a copy of the vocabulary.

**The seam is `kit.Runner`'s own signature, not a sibling type.** It
now takes `env []string` between the streams and the script, and `Run`
layers those entries on `os.Environ()` rather than replacing it. A
sibling `EnvRunner` would have been the smaller diff — thirteen call
sites and nine fakes pass `nil` — and it was rejected because the
narrow type is the shape that invites the mistake: a consumer holding a
runner without an environment has argv as its only way to hand a script
a string, and argv is what `check_observance.sh` forbids by name. One
signature also keeps one recorder per package in the suite, which is
what `author`'s check-order assertions read.

**Step 4 decided that a sentence is needed, in all three files.**
`shape.md`'s "Checks first, in their load-bearing order" became false
the moment a check read the composition, so it is now two bullets: a
non-zero check stops the command there, and a check runs where its
input exists. The same page's opening recital was replaced under
report-0026, triaged `fixed` on this branch — only one of its five
clauses held for all four commands. `amend.md` and `author.md` each
name the observance check as theirs, and both dropped the
"checks first" recital that `shape.md` no longer carries.

Tests: `internal/kit` proves the environment reaches the child, that it
is layered rather than substituted, and that a caller's entry beats an
inherited one; a walk over the shipped Go tree proves no file carries
the kit's `TYPES` or `SCOPES` list. `amendcmd` and `authorcmd` each
gained an `observance_test.go` — the refusal before the first write,
the text arriving through the environment and not argv, and a runner
failure named rather than passed off as a verdict. Two integration
cases run the compiled binary: `amend --type wibble` and `author` with
a title outside the style, each refused with a clean tree, no branch
and nothing on the stub forge.

Left standing, out of scope: the branch name is judged by nothing
anywhere, `shape.md`'s "Nothing reaches the forge without confirmation"
is loose (`amend` lists pull requests and `finish` views one before
their questions), and `take.md` and `finish.md` still recite a phrase
`shape.md` dropped.
