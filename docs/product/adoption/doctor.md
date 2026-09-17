# `writrun doctor`

Reports whether the repository still satisfies what the methodology
assumes. Reports; it never repairs.

## The report's shape

- **Every requirement is a row**, met or not. A requirement that holds
  prints its row, so a repository that satisfies nine checks reads
  differently from one this binary never examined.
- **A stage is a heading that counts its rows** — `Stage 1 — files: 8
  of 9 met.`
- **Four marks, four states.** The mark carries the state and the
  colour repeats it ([rules](../rules.md)).

| Mark | State | Reaches the exit status |
|---|---|---|
| `✓` | met | no |
| `✗` | breaks a flow the methodology runs | yes |
| `!` | the methodology recommends it, and nothing breaks without it | no |
| `?` | the check could not be made at all | no |

- **A row that is not met names the file or the setting** and what is
  expected of it; a wrapped script's own words print under it, indented
  and unedited.
- **The report ends with two lines**: what the declared stage answers
  for, and whether the rung above it is within reach.

## Checks are grouped by stage

- **Checks run from stage 0 up to the declared one** — a project is
  never judged against machinery it did not enable.
- **A group is named for what it examines, not for the stage.** The
  stages' own names are [`init`](init.md)'s; the subjects below are this
  report's, and they run 0–3 where the names run 1–3, because adoption
  never offers stage 0.
- **Stage 0 — environment:** the wrapped scripts' own requirements
  present on the `PATH`
  ([requirements](../../technical/runtime/requirements.md)).
- **Stage 1 — files:** an About file, a real product chapter, a real
  technical chapter, the `docs/` / `work/` split, `AGENTS.md`, the
  gates answered, the kit's version recorded, the queue readable, the
  settings canonical.
- **A gate the kit answers is answered.** The gates are read at the
  project's address, and a project that has not written its own is
  answered by the kit's default — so deferring is a complete answer,
  not a missing one. The gates are the rows that file states — a gate a
  newer tag adds is named without this binary knowing it
  ([coupling](../../technical/engineering/coupling.md)).
- **Stage 2 — the forge:** `gh` present and authenticated; squash
  merging on; the recording push able to reach `main` — Actions
  workflow permissions read-and-write, or read where every workflow
  that pushes to `main` raises `contents: write` of its own; `main`
  governed by a ruleset; and no rule over `main` refusing the recording
  push — the four that do are restrict updates, require signed commits,
  require status checks to pass, and require a pull request before
  merging, and each is named where the Actions bot has no way past it:
  always on a user-owned repository, and on an organization-owned one
  where the ruleset names no bypass actor.
- **Stage 3 — Issues:** enabled, so the mirror has somewhere to land.

## The rung above is previewed

- **The stage above the declared one is previewed**, through the code
  path `--at` already uses. No check is written twice for it.
- **A previewed row never reaches the exit status.** A stage nobody
  declared cannot fail a build.
- **A read the preview needs and cannot make is `?`.** At stage 1 the
  preview is the first forge read the run makes, and a forge that does
  not answer leaves the requirements unknown, never unmet.
- **Stage 3 previews nothing** and says so: there is no rung above it.
- **`--at <stage>` examines that stage instead of the declared one**, so
  a project can ask what a stage it has not taken on would require. The
  report names both numbers, and the exit status still answers for the
  declared stage alone.
- `--at` outside 1–3 is refused, and nothing is examined.

## The screen

- **On a terminal the report opens as a screen**
  ([adoption/doctor.excalidraw](../screens/adoption/doctor.excalidraw)):
  the same rows, a cursor, and the selected requirement explained under
  them. Where there is no terminal it prints the report instead, so a
  script reading this command keeps reading it.
- **Every requirement row is selectable**, the met ones included. The
  headings, the counts and the blank lines are shown and skipped.
- **The footer explains the selected requirement** in the words the
  table below carries, and says whether this repository meets it.
- **A requirement the table does not name is explained by the check's
  own sentence**, never by a guess — a check a newer kit adds is
  reported without this binary knowing it.
- `r` re-runs every check and keeps the cursor where it was. `esc`
  leaves, `q` quits.
- **The screen writes nothing.** `doctor` repairs nothing on a terminal
  either.

## The requirements

One row per requirement. The binary reads this table and looks a row up
by the requirement's name, so the screen and this document cannot
disagree ([coupling](../../technical/engineering/coupling.md)).

| Requirement | What it is | Why the stage needs it | What clears it |
|---|---|---|---|
| `git` | the version control the wrapped scripts run | every flow reads and writes history through it | installing git and putting it on the `PATH` |
| `bash` | the shell every kit script is written for | the scripts are the execution authority, and they are bash | installing bash and putting it on the `PATH` |
| `awk` | the text processor the kit's readers use | the front-matter and settings readers are awk | installing awk and putting it on the `PATH` |
| `sed` | the stream editor the kit's writers use | the queue's edits are sed | installing sed and putting it on the `PATH` |
| `docs/about.md` | the About file every project owes | a reader arriving at the repository is told what it is before anything else | writing `docs/about.md` |
| `docs/product/` | the product chapter, beyond its README | the product half is where rules a task may derive from live | writing one real product doc beside the README |
| `docs/technical/` | the technical chapter, beyond its README | the technical half is where the machinery is stated | writing one real technical doc beside the README |
| `docs/, work/tasks/, work/specs/, work/reports/` | the docs and work split | documents and queue are two halves, and every flow addresses them by these names | creating the folder that is missing |
| `AGENTS.md` | the agents' entry point at the repository root | an agent reads this file first, and the flow is reached from it | writing `AGENTS.md`, and deleting a `writrun:begin`/`writrun:end` section a kit before v0.0.04 left |
| `writrun/gates.md` | who operates each gate the methodology names | the rows are that file's own, so a gate a newer kit adds is named without this binary knowing it | answering the rows still carrying the kit's placeholder |
| `.writrun/VERSION` | the kit tag this repository has installed | no refresh can tell what is installed without it | running `writrun update`, which records the tag it installs |
| `writrun/settings.json` | the adopter's own answers, the declared stage first | every flow reads the stage and the conduct flags from it | writing the file, or running `writrun config` to correct the value it refuses |
| `check_front_matter.sh` | the kit's own sweep over the queue's front matter | the queue is read by line-based readers that need canonical fields | fixing each fault the script named under the row |
| `check_settings.sh` | the kit's own check on the settings file's shape | the same readers read the settings | fixing each fault the script named under the row |
| `gh on the PATH` | the forge client the flows ask the forge through | from stage 2 every forge read and the pull request go through it | installing the GitHub CLI |
| `gh authenticated` | a forge client with credentials | an unauthenticated client answers no read | running `gh auth login` |
| `squash merging is on` | the repository's merge setting | the methodology lands every pull request as one commit | turning squash merging on in the repository settings |
| `the recording push can write to main` | from stage 2 the workflows record the queue's state by pushing to main | a push with no right to write leaves the queue's state unrecorded | setting the Actions workflow permissions to read-and-write, or raising `contents: write` in that file |
| `main is governed by a ruleset` | branch protection over the branch every flow lands on | the methodology recommends protecting it | adding a ruleset that targets `main` |
| `no rule over main refuses the recording push` | whether any rule over main stops the Actions bot | a rule the bot cannot get past stops the recording push | taking the rule off `main`, or putting the Actions bot on the ruleset's bypass list |
| `Issues are enabled` | somewhere for the upstream mirror to land | from stage 3 the flows open issues on this repository | enabling Issues in the repository settings |
