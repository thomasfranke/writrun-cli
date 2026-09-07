#!/usr/bin/env bash
# queue_lib.sh — the fixture behind the queue-reader integration case
# (tests/integration/queue/): one adopted repository rebuilt at each
# front-matter state, and every command that reads the queue run against
# every id form over it.
#
# The four packages that read the queue's front matter used to carry
# four copies of the reader, and the copies had drifted apart. Unifying
# them changed what three of the four commands do, so the proof is not
# "every byte matched": it is the whole matrix, cell by cell, with each
# difference named (spec-0022, tests required).
#
# The matrix is states × ids × commands, and each cell is answered in
# five ways: the exit code, what the run left under work/, whether
# anything reached the fake forge or the bare origin, the Derived-work
# table `author` composed out of what it read, and what the run said.
# The five are one line of `matrix.txt`, which the case diffs against
# `matrix.golden`.
#
# The suite never reaches a real forge: every `gh` invocation is the
# stub's, and GH_LOG records them.

. "$(dirname "${BASH_SOURCE[0]}")/cli_lib.sh"

git_q() { git -c user.name=suite -c user.email=suite@test -c commit.gpgsign=false "$@"; }

TARGET="$WORK/target"
ORIGIN="$WORK/origin.git"
PRISTINE="$WORK/pristine"
GH_LOG="$WORK/gh.log"
GH_BODY="$WORK/gh.body"
export GH_LOG GH_BODY

# The forge, stubbed. Two questions are asked of it: whether it was
# called at all, and what body `author` handed it — the Derived-work
# table is the only place a command shows what it read out of a queue
# file it did not write to.
make_gh() {
  mkdir -p "$WORK/bin"
  : > "$GH_LOG"
  : > "$GH_BODY"
  cat > "$WORK/bin/gh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_LOG"
prev=""
for a in "$@"; do
  [ "$prev" = "--body" ] && printf '%s\n' "$a" > "$GH_BODY"
  prev="$a"
done
case "$1 $2" in
  "auth status") exit 0 ;;
  "pr list") exit 0 ;;
  "pr view") printf '%s\n' '{"number":7,"title":"[TASK-0012] A thing","state":"OPEN","isDraft":true}'; exit 0 ;;
  "pr create") echo "https://forge.test/pull/9"; exit 0 ;;
  "pr ready") exit 0 ;;
esac
exit 0
STUB
  chmod +x "$WORK/bin/gh"
  export PATH="$WORK/bin:$PATH"
}

# LONG holds the line the `long-line` state carries: over the 1 MiB a
# capped scanner stopped at, so a reader that caps drops every field
# after it and says nothing. It is a file because it is longer than an
# argument may be.
LONG="$WORK/long-line.txt"
awk 'BEGIN { s = "x"; while (length(s) < 1200000) s = s s; printf "%s", substr(s, 1, 1200000) }' > "$LONG"

# canonical_task <dir> <number> <slug> <spec-ref> <status> — a task in
# the shape check_front_matter.sh accepts, which every state is a
# departure from.
canonical_task() {
  cat > "$1/work/tasks/task-$2-$3.md" <<EOF
---
id: task-$2
status: $5
blocked_reason: null
taken_by: null
spec_ref: [$4]
doc_ref: null
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-01-01T00:00:00Z
queued: 2026-01-01T00:00:00Z
completed: null
merged: null
provenance: []
---

# Read the queue the way the kit reads it

One paragraph of brief.
EOF
}

# canonical_spec <dir> <number> <slug> <task-ref> <status>
canonical_spec() {
  cat > "$1/work/specs/spec-$2-$3.md" <<EOF
---
id: spec-$2
task_ref: task-$4
status: $5
created: 2026-01-01T00:00:00Z
---

# spec-$2 — the contract

**References:** [task-$4](../tasks/task-$4-$3.md)

- **Goal:** something the task implements.

## Scope

In: the thing. Out: everything else.

## Steps

1. Do the thing.

## Acceptance criteria (EARS)

- When the thing is done, the system shall say so.

## Edge cases

- none

## Tests required

Integration over the fixture.

## Definition of Done

- [ ] Suite green.

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- none — no machinery change

## Outcome

What was built, and nothing diverged.
EOF
}

# canonical_report <dir> <number> <slug>
canonical_report() {
  cat > "$1/work/reports/report-$2-$3.md" <<EOF
---
id: report-$2
status: open
task_ref: []
doc_ref: null
created: 2026-01-01T00:00:00Z
triaged: null
---

# Something was noticed

One paragraph of finding.
EOF
}

# shape <state> <file> — the state, written over a canonical file. Each
# one is a departure the four copies of the reader answered differently
# (spec-0022, tests required).
shape() {
  local state="$1" f="$2" tmp="$2.shaping"
  case "$state" in
    canonical) return 0 ;;
    crlf)              awk '{ printf "%s\r\n", $0 }' "$f" > "$tmp" ;;
    fence-space)       awk 'NR == 1 { print "--- "; next } { print }' "$f" > "$tmp" ;;
    unclosed)          awk 'NR > 1 && $0 == "---" && !done { done = 1; next } { print }' "$f" > "$tmp" ;;
    no-front-matter)   awk 'NR == 1 { next } $0 == "---" { stop = 1; next } stop { print }' "$f" > "$tmp" ;;
    duplicate-status)  awk '/^status: / && !done { print; print "status: dropped"; done = 1; next } { print }' "$f" > "$tmp" ;;
    space-colon)       sed 's/^status: /status : /' "$f" > "$tmp" ;;
    mixed-endings)     awk 'NR == 1 { print; fm = 1; next } fm && $0 == "---" { print; fm = 0; next } fm { printf "%s\r\n", $0; next } { print }' "$f" > "$tmp" ;;
    long-line)         { head -n 2 "$f"; printf 'note: '; cat "$LONG"; printf '\n'; tail -n +3 "$f"; } > "$tmp" ;;
    missing-id)        grep -v '^id: ' "$f" > "$tmp" ;;
    missing-status)    grep -v '^status: ' "$f" > "$tmp" ;;
    null-list)         sed 's/^spec_ref: \[.*\]$/spec_ref: [null]/' "$f" > "$tmp" ;;
    *) printf 'no such state: %s\n' "$state" >&2; return 1 ;;
  esac
  mv "$tmp" "$f"
}

# STATES is the eleven the spec names, plus `[null]` — the list edge
# case its Edge cases name and its eleven do not reach.
STATES="canonical crlf fence-space unclosed no-front-matter duplicate-status space-colon mixed-endings long-line missing-id missing-status null-list"

# IDS is the nine forms a person types or a branch spells, ending in one
# no file holds.
IDS="task-0012 spec-0012 0012 12 task/0012-x task-0000 task-abc-0012 report-0020 task-0099"

# COMMANDS is every command that reads the queue's front matter, plus
# `writrun` with no command, which reads none and must stay that way.
COMMANDS="amend finish author status screen"

# build_state <state> — the pristine repository at one state: main
# carrying the kit, the queue and the docs; a task branch for `finish`;
# an authoring branch for `author`; and a bare origin holding all three.
build_state() {
  rm -rf "$PRISTINE"
  mkdir -p "$PRISTINE/target/.writrun" \
           "$PRISTINE/target/work/tasks" "$PRISTINE/target/work/specs" \
           "$PRISTINE/target/work/reports" "$PRISTINE/target/docs/product"
  local t="$PRISTINE/target"
  cp -R "$REPO_ROOT/.writrun/scripts"     "$t/.writrun/scripts"
  cp -R "$REPO_ROOT/.writrun/skills"      "$t/.writrun/skills"
  cp -R "$REPO_ROOT/.writrun/templates"   "$t/.writrun/templates"
  cp -R "$REPO_ROOT/.writrun/conventions" "$t/.writrun/conventions"
  cp "$REPO_ROOT/.writrun/VERSION"        "$t/.writrun/VERSION"
  cat > "$t/.writrun/settings.json" <<'EOF'
{
  "stage": 2,
  "stage_1": {
    "decisions_style": "chronological",
    "product_layout": "by-feature",
    "provenance_ledger": false,
    "spec_required": "when-warranted"
  },
  "stage_2": {
    "agent_coauthor": false,
    "auto_commit": false,
    "auto_pr": false,
    "auto_push": true,
    "pr_title_style": "bracketed"
  }
}
EOF
  printf '# A project\n'  > "$t/README.md"
  printf '# Tasks\n'      > "$t/work/tasks/README.md"
  printf '# Specs\n'      > "$t/work/specs/README.md"
  printf '# Reports\n'    > "$t/work/reports/README.md"
  printf '# Product\n'    > "$t/docs/product/README.md"

  # The pair `amend`, `finish` and `status` work: a task in flight and
  # the approved spec it implements.
  canonical_task   "$t" 0012 queue-reader spec-0012 in-progress
  canonical_spec   "$t" 0012 queue-reader 0012 approved
  canonical_report "$t" 0020 a-finding
  shape "$1" "$t/work/tasks/task-0012-queue-reader.md"
  shape "$1" "$t/work/specs/spec-0012-queue-reader.md"
  shape "$1" "$t/work/reports/report-0020-a-finding.md"

  (
    cd "$t" || exit 1
    git_q init -q
    git_q symbolic-ref HEAD refs/heads/main

    # The identity and the line endings are written into the repository,
    # not passed per invocation: `amend` commits through the binary's own
    # git, which reads this config and never sees git_q's `-c` flags. A
    # machine whose ~/.gitconfig names a committer hides that; a runner
    # with none fails the commit and the cell reads as a forge that was
    # never called. autocrlf is pinned for this fixture in particular —
    # its states are line endings, and a checkout that rewrote them would
    # be testing the wrong bytes.
    git_q config user.name suite
    git_q config user.email suite@test
    git_q config commit.gpgsign false
    git_q config core.autocrlf false
    git_q config core.safecrlf false

    git_q add -A
    git_q commit -q -m "the kit and the queue"

    # The branch `finish` is run on: one commit touching no permanent
    # doc, so a spec promising none has its delta contract honoured.
    git_q checkout -q -b task/0012-queue-reader
    printf 'the work\n' >> README.md
    git_q commit -q -am "the work"

    # The branch `author` is run on: a rule written, and the work it
    # derives added beside it. Derived work enters the tree at
    # `backlog` behind a `draft` spec — check_task_state.sh calls any
    # other entry FORBIDDEN, and a fixture it refuses never reaches the
    # command's own reading.
    git_q checkout -q main
    git_q checkout -q -b docs/the-rule
    printf '# Rules\n\nA rule.\n' > docs/product/rules.md
    canonical_task   . 0013 derived spec-0013 backlog
    canonical_spec   . 0013 derived 0013 draft
    shape "$1" work/tasks/task-0013-derived.md
    shape "$1" work/specs/spec-0013-derived.md
    git_q add -A
    git_q commit -q -m "the rule, and the work it derives"
    git_q checkout -q main
  )
  git_q init --bare -q "$PRISTINE/origin.git"
  git_q -C "$t" remote add origin "$PRISTINE/origin.git"
  # The authoring branch is not pushed: `author` refuses a branch the
  # forge already has, and authoring starting locally is its own rule.
  git_q -C "$t" push -q -u origin main task/0012-queue-reader
}

# reset_repo — the repository as build_state left it. A cell may commit,
# cut a branch and push, so nothing short of the whole thing is a reset.
reset_repo() {
  rm -rf "$TARGET" "$ORIGIN"
  cp -R "$PRISTINE/target" "$TARGET"
  cp -R "$PRISTINE/origin.git" "$ORIGIN"
  git_q -C "$TARGET" remote set-url origin "$ORIGIN"
  : > "$GH_LOG"
  : > "$GH_BODY"
}

# derived — the Derived-work table out of the body `author` composed,
# in one line. It is what the command read out of the queue files the
# change adds, and nothing else in a cell shows it: the run writes
# nothing, and the exit code is the same table or no table at all.
derived() {
  local out
  out=$(grep '^| ' "$GH_BODY" 2>/dev/null | grep -v '^|---' \
    | sed -e 's/^| //' -e 's/ |$//' -e 's/ | /,/g' | paste -sd';' -)
  printf '%s' "${out:--}"
}

# manifest — every file under work/, and its bytes, in one sorted list.
manifest() {
  (
    cd "$TARGET" || exit 1
    find work -type f | LC_ALL=C sort | while IFS= read -r f; do
      printf '%s %s\n' "$(cksum < "$f" | awk '{print $1 "-" $2}')" "$f"
    done
  )
}

# changed <before> <after> — the paths whose bytes differ, or `-`.
changed() {
  local out
  out=$(diff <(printf '%s\n' "$1") <(printf '%s\n' "$2") \
    | sed -n 's/^[<>] [0-9]*-[0-9]* //p' | LC_ALL=C sort -u | paste -sd, -)
  printf '%s' "${out:--}"
}

# say <file> — the opening of what a run said, with everything that is
# the run's own place or moment taken out of it. Three lines because
# `status` answers in labelled ones and the task is the second.
# The cap is bytes, under LC_ALL=C, because `cut -c` is not one thing:
# BSD counts bytes and GNU counts characters in a UTF-8 locale, so the
# same sentence — these carry em dashes and typographic quotes — is cut
# at two different places on the two platforms and the golden holds only
# one of them.
say() {
  LC_ALL=C sed -e "s|$WORK|<work>|g" -e 's/[0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}T[0-9:]*Z/<stamp>/g' "$1" \
    | LC_ALL=C grep -v '^[[:space:]]*$' | head -n3 | paste -sd/ - | LC_ALL=C tr -s ' ' \
    | LC_ALL=C cut -c1-300
}

# reached_forge — whether the cell talked to the stubbed forge at all.
#
# **Whether, not how many times.** How chatty a check is belongs to the
# kit's scripts, not to this binary: WritRun v0.0.04 raised the author
# cell's calls from 23 to 41 without changing one exit code, one path or
# one sentence in this whole matrix. A count would have made every such
# tag a golden to re-argue, and would have held this suite to a number
# that is not even stable across contexts — the nested run inside the
# release e2e reads a different one for the same behaviour.
reached_forge() {
  if [ -s "$GH_LOG" ]; then printf 'yes'; else printf 'no'; fi
}

# run_cell <state> <id> <command> — one cell of the matrix, answered in
# four ways on one line.
run_cell() {
  local state="$1" id="$2" cmd="$3" code before after refs out="$WORK/cell.out"
  reset_repo
  case "$cmd" in
    status) git_q -C "$TARGET" checkout -q -B "$id" main 2>/dev/null || true ;;
    finish) git_q -C "$TARGET" checkout -q task/0012-queue-reader ;;
    author) git_q -C "$TARGET" checkout -q docs/the-rule ;;
  esac
  before=$(manifest)
  (
    cd "$TARGET" || exit 1
    case "$cmd" in
      amend)  "$WRITRUN" amend "$id" --title "The spec is returned to draft" --slug the-amendment --yes ;;
      finish) "$WRITRUN" finish "$id" --range main...HEAD --yes ;;
      # author takes the whole title, where amend takes only the summary
      # and composes the bracket itself. The observance check the door
      # runs judges what is passed, so this one is written in the style
      # .writrun/settings.json declares.
      author) "$WRITRUN" author --range main...HEAD --title "[Docs][Product] The rule is written" --yes ;;
      status) "$WRITRUN" status ;;
      screen) "$WRITRUN" ;;
    esac
  ) > "$out" 2>&1
  code=$?
  after=$(manifest)
  refs=$(git_q -C "$ORIGIN" for-each-ref --format='%(refname)' | wc -l | tr -d ' ')
  printf '%s|%s|%s|exit=%s|work=%s|gh=%s|refs=%s|table=%s|%s\n' \
    "$state" "$id" "$cmd" "$code" "$(changed "$before" "$after")" \
    "$(reached_forge)" "$refs" "$(derived)" "$(say "$out")"
}

# matrix <file> — every cell, in the order the three lists fix. The id
# axis is inert for `author` and for the screen, which take none: they
# are run once per state, and what varies for `author` is the state of
# the queue files the change adds.
matrix() {
  local state id cmd
  : > "$1"
  for state in $STATES; do
    build_state "$state"
    for cmd in $COMMANDS; do
      case "$cmd" in
        author|screen) run_cell "$state" "-" "$cmd" >> "$1" ;;
        *) for id in $IDS; do run_cell "$state" "$id" "$cmd" >> "$1"; done ;;
      esac
    done
  done
}

make_gh
