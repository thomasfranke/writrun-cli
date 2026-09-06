---
id: spec-0022
task_ref: task-0023
status: implemented
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

`internal/queue` reads the queue for `amendcmd`, `finishcmd`,
`authorcmd` and `statuscmd`, and none of the four holds a reader of its
own. It is a leaf package: it reads through the caller's `vfs.FS`, owns
no port, and holds no package-level variable another package writes.
Its vocabulary is `task` and `spec` and stops there.

### The disagreements, and the kit's answer

| # | The point | Where the four copies stood | The kit's answer | Who moved |
|---|---|---|---|---|
| 1 | The opening fence | `amend`, `finish` matched `---` exactly; `author`, `status` trimmed it | `ql_fm_field`'s `NR == 1 { if ($0 != "---") exit }` matches exactly, and a CRLF file opens with `---\r` | `author`, `status` |
| 2 | The closing fence | the same split | `/^---$/`, exactly | `author`, `status` |
| 3 | A block never closed | `amend`, `finish`, `author` read no fields; `status` read every one, on to the end of the file | `fm_block` exits 1 — no fields | `status` |
| 4 | Two lines carrying one key | `amend`, `finish`, `author` read the first; `status` read the last | `ql_fm_field` exits on the first, and `get` is `head -n1` | `status` |
| 5 | `status : approved` | `amend`, `finish`, `author` read no field; `status` read one | `sub("^" f ": *", "")` wants the colon against the name — no field | `status` |
| 6 | A line over 1 MiB | `amend`, `finish`, `author` read past it; `status` stopped its scanner and dropped every field after it in silence | awk and sed read a line of any length | `status` |
| 7 | `spec_ref: [null]` | `amend`, `finish`, `author` read no entries; `status` read one named `null` | `ql_resting` deletes the brackets, makes the commas spaces, and keeps the word | `amend`, `finish`, `author` |
| 8 | A list separated by spaces | all four split on commas alone | `tr ',' ' '` and word splitting take either | all four |
| 9 | A field line's carriage return under an LF fence | `amend` kept it; `finish` dropped it | `ql_set_field` prints `field ": " value` and nothing else — dropped | `amend` |
| 10 | A queue directory that cannot be walked | `amend`, `finish` returned the error; `status` discarded it | `find … 2>/dev/null` cannot tell an unreadable directory from an empty one, so its caller names the file it did not find | nobody — the reader returns the error, and each command keeps the answer it already gave |
| 11 | Which prefix an id parser strips | `amend`, `finish` stripped `task-`, `spec-`, `task/` and `spec/` alike; `status` stripped `task-` or `spec-` and never the slash | `ql_task_num` strips the kind's own two and nothing else | `amend`, `finish`, `status` |
| 12 | Where an id's digits are | `amend`, `finish` took digits from anywhere; `status` took leading digits | `s/[^0-9].*$//` keeps the leading run | `amend`, `finish` |
| 13 | What the number is | `amend`, `finish` kept the digits as text; `status` made an `int`, which reads `task-0000` as task 0 | text: `s/^0+//` leaves `task-0000` naming no number | `status` |
| 14 | The file's title | `author` took the sentence after the id's em dash; `status` took the whole line | neither — the kit has no title reader | nobody |
| 15 | What a resolution returns | `amend`, `finish` returned a repository-relative path; `status` returned a joined one | neither | nobody |

**The triage counted thirteen; enumerated one point per behaviour, there
are fifteen.** Points 8, 14 and 15 are the three it did not name. Point
14 is resolved by keeping both: `queue.Heading` gives the whole line and
`authorcmd`'s own `subject` drops the id, because the wording of a
column is that command's. Point 15 is resolved to the relative path, so
every caller joins its own root.

**The cross-kind refusal is the parser's.** `Num` takes the kind it is
resolving, so `Num(Spec, "task-0012")` is empty and there is no second
guard beside it; `Declares` only chooses which sentence explains the
empty answer. `amend`'s refusal sentence, `finish`'s "resolves to no
file" and `status`'s "no status" are unchanged, word for word.

### What the matrix says

Twelve file states — the eleven this spec names, plus `spec_ref:
[null]`, which its Edge cases name and its eleven states do not reach —
by nine id forms by five commands: 348 cells, each compared on the exit
code, the bytes of every file under `work/`, whether anything reached
the fake forge or the bare origin, the Derived-work table `author`
composed out of what it read, and what the run said.

`author` and `writrun` with no command take no id, so they are one cell
per state; what varies for `author` is the state of the queue files the
change adds. `check_front_matter.sh` refuses ten of the twelve states
at `author`'s door, so `author` reaches its own reading in two — where
it composes the table, pushes and opens the pull request. The table is
in the cell because nothing else in it would show a reading that
changed: the run writes nothing under `work/`, and the exit code is the
same on any table at all.

**242 cells matched and 106 moved, and every moved cell is one of the
fifteen points above.** No cell gained a write, a forge call or a pushed
ref. Two lost all three.

| What moved | Cells | The point |
|---|---|---|
| `amend report-0020` and `finish report-0020` refuse the id instead of resolving `spec-20` / `task-20` | 24 | 11 |
| `amend task-abc-0012` refuses the id instead of naming a near miss | 12 | 12 |
| `finish task-abc-0012` refuses the id | 12 | 12 |
| `status` counts the open reports the way the kit reads their status | 48 | 1, 2, 3, 4, 5, 6 |
| `status` reads the branch's task and its spec the way the kit reads them | 6 | 1, 2, 3, 4, 5, 6 |
| `finish` on a task whose `spec_ref` is `[null]` hands `null` to `check_deltas.sh`, which refuses it (exit 3) | 4 | 7 |

**The two cells that lost a write are the reason this was not a
behaviour-preserving refactor.** On a canonical queue, `finish
task-abc-0012` wrote `status: implemented` onto spec-0012, stamped
task-0012's `completed` date and marked the pull request ready — on an
id that names no task. It now refuses before anything is read.

**`writrun` with no command moved in no cell.** It reads the lister's
output and no front matter, and it still does.

### Named, and not resolved

- `ql_set_field` rewrites **every** line in the front matter carrying
  the field, where `queue.Set` rewrites the first. Both Go copies wrote
  the first, so this is not a disagreement between them, and this spec
  does not name it. A file with two `status:` lines is a file
  `check_front_matter.sh` refuses either way.
- `ql_set_field` also writes into a block `fm_block` refuses for never
  closing. Both Go copies refuse it, and this spec's acceptance criteria
  say an unclosed block holds no fields.
- `ql_spec_file` resolves a spec by its literal id — `find work/specs
  -iname "$1.md" -o -iname "$1-*.md"` — so the kit does not resolve
  `amend 11` to `spec-0011` and this binary does. Both copies already
  did, so it is not a disagreement between them; the reading kept is the
  one both comments state, that `ql_task_num`'s rule applies to either
  vocabulary.
