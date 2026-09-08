---
id: spec-0032
task_ref: task-0029
status: draft
created: 2026-09-08T15:04:04Z
---

# spec-0032 — Settle the documents the screens contradict

**References:** [task-0029](../tasks/task-0029-ui-screens-refactor.md)

- **Goal:** Three contradictions that predate the screens, and that a
  screen naming a stage or a section cannot straddle, are settled.

## Scope

In: the stage vocabulary, the queue's sections, and the grouping of the
flow table.

Out: anything the other five specs change. This spec touches no
behaviour except where a filter is added for a section that has none.

## Steps

1. **The stage vocabulary.** `init.md` names stages `files`,
   `pull requests`, `GitHub issues`; `doctor.md`, in the folder beside
   it, names them `files`, `the forge`, `Issues`. The code carries both
   lists — `internal/command/initcmd/plan.go` and
   `internal/command/doctorcmd/report.go`. Settle it as two concepts
   rather than one word forced to serve twice: a stage has a **name**,
   which is `init`'s list, and `doctor` has a **subject** it examines at
   that stage, which is its own. Say so once, and let both files point
   at the saying.
2. **The queue's sections.** `list.md` describes three — available, held
   back, reports — and the lister writes five, adding *In progress* and
   *In flight*. `screens/README.md` says the screen shows the sections
   `list` shows, which binds three documents to a count that does not
   hold. Name all five.
3. **The filters.** `--available`, `--held` and `--reports` cover three
   of the five. Add the two missing, so the flag set and the section set
   are the same set.
4. **The grouping.** `docs/product/README.md` groups by flow; the entry
   screen groups by what a person acts on. Regroup the table to match
   the screen, and keep
   [`shape.md`](../../docs/product/pull-requests/shape.md) reachable
   from each of the four commands that share its shape, since it is no
   longer a section heading.

## Acceptance criteria (EARS)

- When a document names a stage, the system shall use the stage names
  `init` offers, and no other list.
- When `doctor` heads a stage group, the system shall name the subject
  it examined, which is not the stage's name and is stated to be so.
- When `list.md` enumerates the queue's sections, the system shall
  enumerate all five the lister writes.
- When a section filter is given, the system shall print that section,
  for each of the five.
- When no filter is given, the system shall print what it prints today.
- When the flow table is read, the system shall group the commands as
  the entry screen groups them.
- When `shape.md` is reached, the system shall be reachable from each of
  `take`, `author`, `finish` and `amend`.

## Edge cases

- **A section the queue has nothing for.** The existing behaviour holds:
  a section with no rows is not printed, and a filter naming it says so
  rather than printing an empty heading.
- **Two filters at once.** Already supported; the new two join the same
  set without a new rule.
- **`doctor`'s stage 0.** It has a subject — the environment — and no
  stage name, because adoption never offers it. The saying in step 1
  covers that: subjects run 0–3, names run 1–3.

## Tests required

- Unit, `internal/command/listcmd`: each of the five filters prints its
  section; the unfiltered output is unchanged.
- A docs check is not automated here; the three contradictions are
  settled by review, which is the gate `gates.md` already puts on
  `docs/`.

## Definition of Done

- [ ] One stage vocabulary, stated once and pointed at twice.
- [ ] `list.md` names five sections.
- [ ] Five sections, five filters.
- [ ] The flow table and the entry screen group one set one way.

## Proposed product changes

- `product/adoption/init.md` — the stage names, pointing at the one
  statement of them.
- `product/adoption/doctor.md` — the subjects, named as subjects.
- `product/queue/list.md` — all five sections, and the five
  filters.
- `product/README.md` — the regrouped flow table.
- `product/pull-requests/shape.md` — reached from the four
  commands' pages rather than from a section heading.

## Proposed technical changes

- none — no machinery change.

## Outcome

_(fill after execution)_
