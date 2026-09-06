---
id: spec-0023
task_ref: task-0024
status: approved
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

- `product/pull-requests/shape.md`, `amend.md`, `author.md` — only if step 4 decides a sentence is needed; the Outcome records the decision either way.

## Proposed technical changes

- none — the seam is internal and adds no package.

## Outcome

_(fill after execution)_
