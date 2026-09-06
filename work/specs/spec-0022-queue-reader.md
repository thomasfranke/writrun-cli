---
id: spec-0022
task_ref: task-0023
status: approved
created: 2026-09-06T04:28:24Z
---

# spec-0022 — Read the queue in one place

**References:** [task-0023](../tasks/task-0023-queue-reader.md)

- **Goal:** the four packages that read the queue's front matter read it through one package, and every point where their copies disagree is resolved against the kit's own reader rather than against each other.

## Scope

In: the front-matter reader, the field reader and writer, the
inline-list reader, the heading reader and the id-to-file resolution in
`internal/command/amendcmd/`, `finishcmd/`, `authorcmd/` and
`statuscmd/` — and, per disagreement, which behaviour is correct, named
against `queue_lib.sh`'s `ql_fm_field`, `ql_set_field` and
`ql_task_num`, and against `check_front_matter.sh`'s `fm_block`.

Out: `split`, which stands in four copies whose flag sets are their own
and each of which already records why it is kept — a duplication
cheaper than the coupling. Out: `.writrun/` itself. Out: the words each
command gives a refusal, a finding or a status line; the reading is
shared, the wording is not.

**This is not a behaviour-preserving refactor, and must not be scoped as
one.** The copies disagree on thirteen points. Unifying them changes
what three of the four commands do. Every such change is in scope only
where this spec names it, says which copy was right, and shows the kit
agreeing; any output change this spec does not name is out.

## Steps

1. Name the disagreements. Enumerate every point on which the four copies differ, with the kit's answer beside each. The enumeration is a deliverable, not working notes — it is what makes the behaviour change reviewable instead of smuggled.
2. Write `internal/queue`: a leaf package reading through the caller's `vfs.FS`, with no port and no package-level variable another package writes (`technical/engineering/boundaries.md`). `boundaries.md` authorises exactly this vocabulary — task, spec, report, stage — and no more.
3. Give the id parser the kind it is resolving, so it strips only that prefix as `ql_task_num` does, and the cross-kind refusal falls out of the parser rather than sitting beside it.
4. Move the four callers onto it. What stays with each command is every word: `statuscmd`'s "no status", `amendcmd`'s refusal sentence, `finishcmd`'s "resolves to no file".
5. Prove the change against the matrix in Tests required, and record the result as the enumeration from step 1 with each cell's verdict.

## Acceptance criteria (EARS)

- When any of the four commands reads a queue file, the system shall read it through one package.
- When a queue file's fence is not exactly `---` at column 0, the system shall report no front matter, in every command.
- When a front-matter block is never closed, the system shall report no fields, in every command.
- When a file carries two `status:` lines, the system shall read the first, in every command.
- When an id declares a kind, the system shall resolve only that kind, in every command.
- When this spec does not name an output change, the system's output shall be unchanged.

## Edge cases

- A queue file over 1 MiB on one line: `statuscmd` alone caps its scanner today and silently drops every field after the cap. The unified reader has one answer, and this spec names it.
- `spec_ref: [null]`: three copies read it as empty, one as a single entry named `null`. The kit's reader decides.
- A file with no front matter at all, such as the `README.md` beside the queue: it is read by the walk and must not become an error.
- A CRLF field line under an LF fence: `setField` preserves the ending it found in one copy and drops it in another.

## Tests required

The proof is not "every byte matched" — two of the four commands write,
and the copies disagree. The matrix: eleven file states (canonical LF;
CRLF throughout; a fence with trailing whitespace; unclosed front
matter; no front matter; a duplicate `status:` key; `status :
approved`; a CRLF field line under an LF fence; a line over 1 MiB;
missing `id`; missing `status`) by nine id forms (`task-0012`,
`spec-0012`, `0012`, `12`, `task/0012-x`, `task-0000`, `task-abc-0012`,
`report-0020`, an id no file holds) by five commands (`amend`,
`finish`, `author`, `status`, and `writrun` with no command) by four
comparisons: stdout, stderr, exit code, and the bytes of every file
under `work/` afterwards, plus whether anything reached the fake forge.

A wrong resolution that reached a write is a different failure from one
that reached a push, and streams alone cannot tell them apart.

## Definition of Done

- [ ] One package reads the queue, and the four callers hold no reader of their own.
- [ ] Every disagreement is named, resolved against the kit, and recorded in the Outcome.
- [ ] The matrix is run, and every cell either matches or is a named, justified change.

## Proposed product changes

- none — no command's documented behaviour changes except where this spec names it, and none of those is stated in `docs/product/`.

## Proposed technical changes

- `technical/layout/tree.md` — a row for the extracted package.

## Outcome

_(fill after execution)_
