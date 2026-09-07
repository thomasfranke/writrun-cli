---
id: spec-0026
task_ref: task-0027
status: approved
created: 2026-09-06T21:22:47Z
---

# spec-0026 — Each kit path declared once

**References:** [task-0027](../tasks/task-0027-kit-paths-once.md)

- **Goal:** every path into the adopted repository is declared in exactly one package, so a path the kit moves is one edit.

## Scope

In: `internal/kit`, `internal/queue`, and every package holding a
duplicate — `listcmd`, `takecmd`, `workcmd`, `authorcmd`, `amendcmd`,
`finishcmd`, `statuscmd`, `doctorcmd`, `initcmd`, `internal/hook`,
`cmd/writrun`.

Out: which paths may exist at all, which is
[coupling](../../docs/technical/engineering/coupling.md) rule 1 and
spec-0025's work. Out: what any command does with a script's output.
Out: test fixtures naming a queue file by id — those are data, not the
binary's knowledge of the kit.

## What is wrong

One path, several names, several packages:

| Path | Packages declaring it |
|---|---|
| `.writrun/skills/writrun-select-next-task/list_tasks.sh` | `listcmd`, `takecmd`, `workcmd`, `cmd/writrun` |
| `work/tasks` | `authorcmd`, `amendcmd`, `finishcmd`, `statuscmd`, `doctorcmd`, `initcmd` |
| `work/specs` | `authorcmd`, `amendcmd`, `finishcmd`, `statuscmd`, `doctorcmd` |
| `.writrun/scripts/stage-2-pull-requests/read_setting.sh` | `authorcmd`, `amendcmd`, `doctorcmd` |
| `.writrun/scripts/stage-2-pull-requests/check_observance.sh` | `authorcmd`, `amendcmd`, `internal/hook` |
| `.writrun/skills/writrun-check-front-matter/check_front_matter.sh` | `authorcmd`, `doctorcmd` |
| `.writrun/scripts/stage-1-tasks-and-specs/preflight.sh` | `finishcmd`, `statuscmd` |
| `.writrun/templates/pull_request_template.md` | `authorcmd`, `amendcmd` |

`list_tasks.sh` carries four different constant names for one file. A
rename that updates three of the four compiles, and the fourth command
fails at run time against a repository that is correct.

`internal/queue` centralised the queue's *parsing* (task-0023) and left
its *paths* out: `Resolve` takes the directory as a parameter, so every
caller supplies its own copy.

## Steps

1. Give `internal/kit` a constant per kit path it runs or reads, named after the act rather than the file — the package already owns running them, and the path is the same fact as the runner.
2. Give `internal/queue` `TasksDir`, `SpecsDir` and `ReportsDir`, and default `Resolve`'s directory from the kind rather than from the caller.
3. Replace every duplicate declaration with a reference. No package outside those two, `internal/kittag` and `internal/pointer` declares a path into the adopted repository.
4. Add the check that keeps it: a test walking the packages' source for a string literal beginning `.writrun/` or `work/` outside the four that may hold one.

## Acceptance criteria (EARS)

- When a kit path is searched for across the production packages, the system shall hold exactly one declaration of it.
- When `internal/kit`'s constant for a script changes, every command running that script shall follow with no further edit.
- When a package outside `internal/kit`, `internal/queue`, `internal/kittag` and `internal/pointer` declares a literal path into the adopted repository, the suite shall fail naming the file.
- When every command runs after the change, each shall invoke the same script path it invoked before.

## Edge cases

- A test fixture writing `work/tasks/task-0011-finish-command.md`: it names a document, not a directory, and the check reads declarations rather than every occurrence.
- `internal/hook` runs `check_observance.sh` from a git hook rather than a command: it references the same constant, since the path is the same fact wherever it is called from.
- A path used exactly once today: it is still declared in the owning package — the rule is about where a path lives, not how many callers it has.
- `cmd/writrun` wiring a script into the screen's frame: production wiring is where the composition happens, and it references rather than declares.

## Tests required

The source-walking check of step 4, asserting it fails on a planted
duplicate and passes on the tree.

The existing per-command tests assert which script each command runs;
they are the regression net for step 3 and are expected to pass
unchanged. Any that assert a local constant's name rather than its value
are rewritten to assert the value; the Outcome names each.

## Definition of Done

- [x] Every path in the table above has one declaration.
- [x] The source-walking check is in the suite and green.
- [x] No command's behaviour changed — the existing suites pass unchanged.

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- `technical/layout/tree.md` — `internal/kit/` and `internal/queue/` name the paths as well as act on them.

## Outcome

Every path in the table has one declaration.

**`internal/kit` names what it runs.** Thirteen scripts and three kit
files, each after the act rather than the file — `Preflight`,
`CheckObservance`, `TakeTask`, `PullRequestTemplate`. Running a script
is that package's act, so the paths sit beside the runner. The tag
stayed `internal/kittag`'s, and `.writrun/AGENTS.md` stayed
`internal/pointer`'s: each is already the package that owns its act.

**`internal/queue` names the queue's folders**, and `Resolve` lost its
`dir` parameter. Every one of its five callers passed the folder
matching the kind it asked for, so the parameter was the whole reason
five copies of `work/tasks` and `work/specs` existed; `Dir(kind)`
answers it now. `Root` carries the tree for the one caller that tests a
path against `work/` without caring which kind it is.

**`internal/kitpaths` references rather than repeats.** `Untouchable`
and `Seeded` name `kit.Settings` and `kit.Gates`, and what stays a
literal there is what nothing calls: `.writrun/conventions`, `docs`,
`work`, `AGENTS.md`, `CLAUDE.md`.

`internal/kit/paths_test.go` is what keeps it collapsed: it parses
every production `.go` file under `internal/` and `cmd/` and reports a
string literal opening `.writrun/` or `work/` outside the five
packages that may hold one. Run against the tree as it stood before
this change it names all ten duplicates;
`TestTheCheckSeesAPlantedPath` proves it can fail at all.

**Two refinements to the spec's own words, both found by running the
check.**

The Scope named four declaring packages and the check allows five. The
fifth is `internal/kitpaths`, whose subject is exactly the paths the
binary never calls — an inventory that had to reference every path it
governs would have nothing left to say. `Dir(kind)` covers the queue's
folders, and the check's comment states the split.

The check reads a *declaration*, not a sentence. Seventeen literals
matched the prefixes on the first run, and seven were messages opening
with a path — `".writrun/VERSION records no tag"` is a command's own
words about a file, and rule 1 governs what the binary calls, not what
it prints. Whitespace is the whole test, since a path has none;
`TestASentenceIsNotADeclaration` holds the line.

The other ten were the duplication this task is about, and three of
them were not in the spec's table: the `[]string{"work/tasks",
"work/specs", "work/reports"}` lists in `doctorcmd/files.go` and
`initcmd/checks.go`, and `authorcmd`'s `queueTree`. The check found
them; the table, written by hand, had missed them.

No command's behaviour changed. Every tier passes unchanged — unit,
all sixteen integration domains and the adopt e2e, run under
`HOME=$(mktemp -d)` so no fixture borrows the machine's git identity.
Coverage 97.1% over `internal/`, with `internal/kit` at 100% and
`internal/queue` at 98.6%.
