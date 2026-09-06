#!/usr/bin/env bash
# check_front_matter.sh — every queue file's front matter is canonical.
#
# Usage: check_front_matter.sh [task-dir] [spec-dir] [docs-dir] [report-dir]
#   Defaults: work/tasks work/specs docs work/reports. Validates every
#   task-*.md, spec-*.md and report-*.md in the working tree; READMEs are
#   skipped. All four are relative to the working directory, and
#   deliberately the same base: a cwd wrong enough to hide `docs/` has
#   already hidden the queue, so the check finds nothing to complain
#   about rather than complaining about everything.
#
#   The report directory comes last because it arrived last, and the
#   three before it are an argument order other callers already pass.
#
# Every reader in this methodology is line-based on purpose — plain
# bash/awk/sed, no YAML parser, no runtime dependency. YAML, though,
# allows the same meaning in forms those readers cannot see: a block
# list under `spec_ref:` reads as an empty list (a task would look
# ready without its approval gate), a quoted value never matches a path
# comparison, a folded scalar reads as nothing. Silently — which is the
# failure mode this repository treats as worse than a wrong answer.
#
# So the canonical form is a checked contract, not an assumption:
#
#   - front matter opens at line 1 with `---` and closes with `---`
#   - one field per line: `key: value` — no continuations, no comments
#   - values are bare: no quotes, no `>` / `|` block scalars, no
#     trailing whitespace
#   - every schema field present exactly once, even when null
#   - lists are inline — `[]` or `[spec-0001, spec-0002]` — and their
#     items are well-formed ids
#   - `id` agrees with the filename; statuses and priority hold only
#     their documented values; `blocked` and `blocked_reason` come
#     paired, both ways; dates are YYYY-MM-DD
#   - `doc_ref` is null or a path under docs/ written *relative to*
#     docs/ — a `docs/` prefix would double when the machinery resolves
#     it
#   - `origin` is `rule` or `report`, always present on a task — how it
#     came to exist is a fact, and there is no third answer
#   - a report's `status` is the route triage took, not a lifecycle:
#     `open` and the five ends, and `triaged` is set exactly when the
#     status is one of the five. Null while `open` because nothing has
#     been decided; set when something has, because a judgement with no
#     date is a judgement nothing can be ordered against. The two ends
#     this file can follow carry what names their outcome with them:
#     `tracked` a non-empty `task_ref`, `authored` a `doc_ref` — while
#     `routed` sent the work upstream, so a `task_ref` here would claim
#     work this queue does not own
#     (docs/product/concepts/report.md#statuses--the-route-not-a-lifecycle)
#   - `provenance` is the one field allowed to open a block list, and the
#     only shape it may take is a dash-opened line per entry, each entry a
#     YAML flow mapping written whole on that line. `provenance: []` is
#     the other legal spelling. An entry opened as a block mapping — its
#     keys on lines of their own — is refused, because that is precisely
#     the shape the line-based readers cannot see
#     (docs/technical/schemas/task.md#task-schema).
#
# Unknown keys in canonical shape are allowed — an adopter may extend.
# The schemas themselves: docs/technical/schemas/README.md.
#
# Exit codes: 0 every file canonical (or nothing to validate); 1 a file
# is malformed; 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions.

set -euo pipefail

TASK_DIR="${1:-work/tasks}"
SPEC_DIR="${2:-work/specs}"
DOCS_DIR="${3:-docs}"
REPORT_DIR="${4:-work/reports}"

status=0
checked=0

fail() {   # fail <file> <reason>
  echo "MALFORMED: $1: $2" >&2
  status=1
}

# fm_block <file> — prints the lines between the opening and closing
# `---`; fails when either fence is missing.
fm_block() {
  awk '
    NR == 1 { if ($0 != "---") exit 1; next }
    /^---$/ { closed = 1; exit }
    { print }
    END { exit closed ? 0 : 1 }
  ' "$1"
}

get() {   # get <block> <field> — first value, as a line-based reader sees it
  printf '%s\n' "$1" | sed -n "s/^$2: *//p" | head -n1
}

count() {   # count <block> <field>
  printf '%s\n' "$1" | grep -c "^$2:" || true
}

require_once() {   # require_once <file> <block> <field>
  local n
  n=$(count "$2" "$3")
  [ "$n" = "1" ] || fail "$1" "field '$3' must appear exactly once (found $n)"
}

# The ledger's vocabulary. `by` and `login` are the two an entry always
# carries — who did the work and who answers for it; `model` belongs to an
# agent's entry alone, and the four counts are the platform's own numbers,
# kept as counts because a stored currency figure becomes a lie about the
# past the next time a price changes (docs/technical/schemas/task.md#task-schema).
LEDGER_KEYS="by model login input output cache_read cache_write"
LEDGER_COUNTS="input output cache_read cache_write"

# **A category is not a model** — the same refusal check_observance.sh
# makes of the commit trailer, for the same reason: `model: ai` satisfies
# every shape check and answers nothing a quarter later, which is the one
# question the field exists for. A tripwire, not a proof: a name written
# to evade it evades it.
MODEL_CATEGORIES="ai llm agent model assistant bot claude gpt gemini llama opus sonnet haiku fable"

check_entry() {   # check_entry <file> <field> <line> — one ledger entry
  local f="$1" field="$2" line="$3"
  local inner rest pair key val seen="" by="" login="" model="" counts=""
  case "$line" in
    "  - {"*"}")
      inner="${line#  - \{}"
      inner="${inner%\}}"
      ;;
    *)
      fail "$f" "'$field' entry is not a flow mapping: '$line' — an entry is written whole on one line as '  - {key: value, ...}', never opened as a block mapping"
      return 0
      ;;
  esac

  rest="$inner"
  while [ -n "$rest" ]; do
    case "$rest" in
      *", "*) pair="${rest%%, *}"; rest="${rest#*, }" ;;
      *)      pair="$rest"; rest="" ;;
    esac
    case "$pair" in
      *": "*) key="${pair%%: *}"; val="${pair#*: }" ;;
      *) fail "$f" "'$field' entry holds '$pair', which is not 'key: value'"; continue ;;
    esac
    case " $LEDGER_KEYS " in
      *" $key "*) ;;
      *) fail "$f" "'$field' entry holds '$key', which the ledger's vocabulary does not have: ${LEDGER_KEYS}"; continue ;;
    esac
    case " $seen " in
      *" $key "*) fail "$f" "'$field' entry repeats '$key'"; continue ;;
    esac
    seen="$seen $key"
    case "$val" in
      ""|\"*|\'*|*\{*|*\}*) fail "$f" "'$field' entry's '$key' is '$val' — values are written bare"; continue ;;
    esac
    case " $LEDGER_COUNTS " in
      *" $key "*)
        counts="$counts $key"
        printf '%s' "$val" | grep -qE '^[0-9]+$' \
          || fail "$f" "'$field' entry's '$key' is '$val' — a count is a bare non-negative integer, never a converted sum"
        ;;
    esac
    case "$key" in
      by)    by="$val" ;;
      login) login="$val" ;;
      model) model="$val" ;;
    esac
  done

  case "$by" in
    agent|human) ;;
    "") fail "$f" "'$field' entry names no actor — every entry carries 'by: agent' or 'by: human'"; return 0 ;;
    *)  fail "$f" "'$field' entry's 'by' is '$by' — an entry names a person or an agent and nothing else" ;;
  esac

  if [ -z "$login" ]; then
    fail "$f" "'$field' entry carries no login — every entry names who answers for it, which on an agent's entry is the person who ran it"
  else
    printf '%s' "$login" | grep -qE '^[A-Za-z0-9-]+(\[bot\])?$' \
      || fail "$f" "'$field' entry's login '$login' is not a bare forge login"
  fi

  if [ "$by" = "agent" ]; then
    if [ -z "$model" ]; then
      fail "$f" "'$field' agent entry names no model — the record survives the next model's arrival only by naming this one"
    else
      case " $MODEL_CATEGORIES " in
        *" $(printf '%s' "$model" | tr '[:upper:]' '[:lower:]') "*)
          fail "$f" "'$field' entry's model '$model' is a category, not a model id — a category answers nothing a quarter later" ;;
      esac
    fi
  fi

  if [ "$by" = "human" ]; then
    [ -z "$model" ] \
      || fail "$f" "'$field' human entry carries model '$model' — a person's entry names no model"
    [ -z "$counts" ] \
      || fail "$f" "'$field' human entry carries counts (${counts# }) — a person's entry carries none, and no check reads that absence as a gap"
  fi
  return 0
}

check_shape() {   # check_shape <file> <block> [ledger-field]
  # Every line is `key: value`, with exactly one exception: the ledger
  # field opens a block list of one-line entries. The exception is named
  # by the caller — a spec has no ledger, so a block there is a fault as
  # it always was.
  local line key val ledger="${3:-}" in_ledger=0 entries=0
  while IFS= read -r line; do
    [ -n "$line" ] || { fail "$1" "front matter holds an empty line"; continue; }

    if [ -n "$ledger" ] && [ "$line" = "${ledger}:" ]; then
      in_ledger=1; entries=0; continue
    fi
    if [ "$in_ledger" = "1" ]; then
      case "$line" in
        "  - "*) entries=$((entries + 1)); check_entry "$1" "$ledger" "$line"; continue ;;
        [[:space:]]*|-*)
          fail "$1" "'$ledger' holds '$line' — an entry is one line, opened by '  - ' and written whole as a flow mapping"
          continue ;;
        *)
          [ "$entries" -gt 0 ] \
            || fail "$1" "'$ledger' opens a block list with no entries — an empty ledger is written inline as []"
          in_ledger=0 ;;
      esac
    fi
    case "$line" in
      *": "*) key="${line%%: *}"; val="${line#*: }" ;;
      *:)     fail "$1" "field '${line%:}' has no inline value — block forms are outside the contract"; continue ;;
      *)      fail "$1" "not a 'key: value' line: '$line' — continuations and list items are outside the contract"; continue ;;
    esac
    printf '%s' "$key" | grep -qE '^[A-Za-z_][A-Za-z0-9_-]*$' \
      || { fail "$1" "malformed key '$key'"; continue; }
    case "$val" in
      \"*|\'*) fail "$1" "field '$key' is quoted — values are written bare" ;;
      '>'|'|'|'>-'|'|-') fail "$1" "field '$key' uses a block scalar — outside the contract" ;;
    esac
    printf '%s' "$val" | grep -q '[[:space:]]$' \
      && fail "$1" "field '$key' carries trailing whitespace"
  done <<EOF
$2
EOF
  if [ "$in_ledger" = "1" ] && [ "$entries" -eq 0 ]; then
    fail "$1" "'$ledger' opens a block list with no entries — an empty ledger is written inline as []"
  fi
  return 0
}

check_list() {   # check_list <file> <block> <field> <item-prefix>
  local val inner it
  val=$(get "$2" "$3")
  case "$val" in
    \[*\]) ;;
    *) fail "$1" "field '$3' must be an inline list — [] or [${4}-001, ...]"; return 0 ;;
  esac
  inner=$(printf '%s' "$val" | sed 's/^\[//; s/\]$//' | tr ',' ' ')
  for it in $inner; do
    printf '%s' "$it" | grep -qE "^${4}-[0-9]+$" \
      || fail "$1" "field '$3' item '$it' is not a ${4} id"
  done
  return 0
}

check_date() {   # check_date <file> <block> <field> <null-ok>
  # An RFC 3339 UTC timestamp, and `Z` is the only spelling accepted.
  # That is the load-bearing half: every reader here is line-based, and
  # with one spelling a lexicographic sort of these strings is a
  # chronological sort. Allowing `+02:00` alongside `Z` would keep `sort`
  # looking correct while being wrong for exactly the entries that
  # crossed a timezone (docs/technical/decisions/0049-...). A bare date
  # is rejected for the reason that rule exists: it cannot order two
  # entries made the same day, which is most of them in an active queue.
  local val
  val=$(get "$2" "$3")
  [ "$4" = "null-ok" ] && [ "$val" = "null" ] && return 0
  printf '%s' "$val" \
    | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' \
    || fail "$1" "field '$3' is '$val' — expected YYYY-MM-DDTHH:MM:SSZ$([ "$4" = "null-ok" ] && printf ' or null')"
  return 0
}


# doc_declares_draft <file> — does that chapter's **first line** declare
# it a draft? Trailing whitespace trimmed, nothing else: a marker
# indented, on line two, or inside a fence is prose about the marker
# rather than the marker. A chapter that documents it names it in prose,
# so presence anywhere in the file is not the question.
doc_declares_draft() {
  local blob first
  [ -f "$1" ] || return 1
  blob=$(cat "$1")
  first=${blob%%$'\n'*}
  first=$(printf '%s' "$first" | sed 's/[[:space:]]*$//')
  [ "$first" = '/// writrun:draft' ]
}

check_doc_ref() {   # check_doc_ref <file> <block>
  # Relative to docs/ — a docs/ prefix would double when the machinery
  # prefixes it back (queue impact, the delta check). One rule, one copy:
  # a task and a report carry the same field under the same contract —
  # "the doc this is answered by" — so a second implementation of it
  # would be a second thing to keep true.
  local ref target
  ref=$(get "$2" doc_ref)
  [ "$ref" != "null" ] || return 0
  case "$ref" in
    docs/*) fail "$1" "doc_ref starts with docs/ — paths are written relative to docs/" ;;
    *.md|*.md#*)
      # **The path, never the anchor.** Reverse traceability is only a
      # grep if the file is really there, and a doc_ref pointing at
      # nothing passed every check until now. The anchor is left
      # unverified on purpose: a heading can be renamed without moving
      # the file, and matching one means parsing markdown, which every
      # reader in this methodology refuses to do. So this proves the
      # file, not the section — a limit worth stating rather than a gap
      # to close later.
      target="${DOCS_DIR}/${ref%%#*}"
      [ -f "$target" ] \
        || fail "$1" "doc_ref '$ref' names no file — ${target} does not exist"
      # **Resolving is no longer the whole question.** A chapter that
      # declares itself a draft is not a rule, and nothing derives from
      # one — so a doc_ref into it is derivation from a rule the project
      # has not made
      # (product/stage-1-tasks-and-specs/authoring.md#a-chapter-that-is-not-a-rule-yet).
      # The refusal says that rather than "names no file": the file is
      # right there, and a resolution message would send the reader
      # looking for a typo that is not there.
      #
      # Read here rather than borrowed: this check is a skill, standalone
      # so it runs at every adoption stage, and the copy in
      # scripts/stage-2-pull-requests/queue_lib.sh is on the far side of
      # that boundary. Two copies of four lines, and the boundary is the
      # reason — the two stage-2 readers share one.
      # **The refusal binds what still derives.** "Nothing derives from
      # a draft chapter" is present tense: a done or dropped task, or a
      # report triage already routed, derives nothing — its doc_ref
      # records what it derived from, as the chapter stood then. Refusing
      # those would make the sanctioned demotion (rule to draft, with a
      # declaration — check_derived_work.sh's fourth row) poison the
      # queue retroactively: every finished record into the chapter
      # failing every later sweep, repairable only by editing history,
      # the one repair this methodology forbids. The line is the one
      # conflicts.md draws — an edit under docs/ answers to the
      # *non-completed* tasks pointing into it. Task and report statuses
      # are disjoint vocabularies, so one case serves both callers; a
      # status this case does not know is judged live — strict by
      # default, and it already failed the status check on its own.
      case "$(get "$2" status)" in
        done|dropped|tracked|authored|fixed|declined|routed) ;;
        *)
          if [ -f "$target" ] && doc_declares_draft "$target"; then
            fail "$1" "doc_ref '$ref' names a draft chapter — nothing derives from a chapter that is not a rule yet"
          fi
          ;;
      esac
      ;;
    *) fail "$1" "doc_ref '$ref' is not null or a .md path (optionally with #anchor)" ;;
  esac
  return 0
}

check_id() {   # check_id <file> <block> — id agrees with the filename's id
  # A queue file is named <id> or <id>-<subject>: identity is the id, the
  # subject slug only makes a directory listing readable. Both shapes are
  # canonical, and anything else disagrees with its own name.
  local want got
  want=$(basename "$1" .md | tr '[:upper:]' '[:lower:]')
  got=$(get "$2" "id" | tr '[:upper:]' '[:lower:]')
  case "$want" in
    "$got"|"$got"-*) return 0 ;;
  esac
  fail "$1" "id '$(get "$2" id)' is not the filename's id — a queue file is named <id>.md or <id>-<subject>.md"
  return 0
}

check_task() {   # check_task <file>
  local block f st reason pr org ref prov
  f="$1"
  if ! block=$(fm_block "$f"); then
    fail "$f" "front matter must open at line 1 with --- and close with ---"
    return 0
  fi
  check_shape "$f" "$block" provenance
  for field in id status blocked_reason taken_by spec_ref doc_ref origin priority depends_on milestone created queued completed merged provenance; do
    require_once "$f" "$block" "$field"
  done
  check_id "$f" "$block"

  st=$(get "$block" status)
  case "$st" in
    backlog|ready|in-progress|in-review|done|blocked|dropped) ;;
    *) fail "$f" "status '$st' is not a task status (backlog|ready|in-progress|in-review|done|blocked|dropped)" ;;
  esac

  # blocked and blocked_reason come paired, both ways: a blocked task
  # states its own unblock condition, and a reason on an unblocked task
  # is a status disagreeing with itself.
  reason=$(get "$block" blocked_reason)
  if [ "$st" = "blocked" ] && [ "$reason" = "null" ]; then
    fail "$f" "status is blocked but blocked_reason is null — a blocked task names what unblocks it"
  fi
  if [ "$st" != "blocked" ] && [ -n "$reason" ] && [ "$reason" != "null" ]; then
    fail "$f" "blocked_reason is set but status is '$st' — null unless blocked"
  fi

  # taken_by is the machinery's record of who has the task: a forge
  # login while a pull request works it, kept on done as who completed
  # it, null everywhere else — a login on a task nobody has is a claim
  # the forge never made.
  taken=$(get "$block" taken_by)
  if [ "$taken" != "null" ]; then
    printf '%s' "$taken" | grep -qE '^[A-Za-z0-9-]+(\[bot\])?$' \
      || fail "$f" "taken_by '$taken' is not a bare forge login or null"
    case "$st" in
      in-progress|in-review|done) ;;
      *) fail "$f" "taken_by is set but status is '$st' — a login only while a PR works the task, or on done" ;;
    esac
  fi

  pr=$(get "$block" priority)
  case "$pr" in
    high|medium|low) ;;
    *) fail "$f" "priority '$pr' is not high, medium, or low" ;;
  esac

  # How the task came to exist, and there are only two answers: derived
  # from an authored rule, or born from a report of work an existing rule
  # already authorizes. Written once at creation and never rewritten, so
  # the only thing left to hold is that it is there and says one of the
  # two (docs/technical/schemas/task.md#task-schema).
  org=$(get "$block" origin)
  case "$org" in
    rule|report) ;;
    *) fail "$f" "origin '$org' is not rule or report" ;;
  esac

  check_list "$f" "$block" spec_ref spec
  check_list "$f" "$block" depends_on task

  check_doc_ref "$f" "$block"

  # Four dates, one shape. `created` is the only one a task always has;
  # the machinery's two and `completed` are null until the event each
  # records happens (docs/product/pipeline.md#flows-and-statuses).
  check_date "$f" "$block" created strict
  check_date "$f" "$block" queued null-ok
  check_date "$f" "$block" completed null-ok
  check_date "$f" "$block" merged null-ok

  # The ledger inline is only ever the empty list: entries are written one
  # to the line, and check_shape has already read each of them. A project
  # that declares no ledger carries `[]` here forever, which is a complete
  # statement and not a gap (product/concepts/provenance.md).
  prov=$(get "$block" provenance)
  if [ -n "$prov" ] && [ "$prov" != "[]" ]; then
    fail "$f" "provenance '$prov' — an inline ledger is only ever [], and entries are written one to the line beneath 'provenance:'"
  fi
  return 0
}

check_spec() {   # check_spec <file>
  local block f st ref
  f="$1"
  if ! block=$(fm_block "$f"); then
    fail "$f" "front matter must open at line 1 with --- and close with ---"
    return 0
  fi
  check_shape "$f" "$block"
  for field in id task_ref status created; do
    require_once "$f" "$block" "$field"
  done
  check_id "$f" "$block"

  st=$(get "$block" status)
  case "$st" in
    draft|approved|implemented) ;;
    *) fail "$f" "status '$st' is not a spec status (draft|approved|implemented)" ;;
  esac

  ref=$(get "$block" task_ref)
  printf '%s' "$ref" | grep -qE '^task-[0-9]+$' \
    || fail "$f" "task_ref '$ref' is not a task id — a spec belongs to exactly one task"

  check_date "$f" "$block" created strict
  return 0
}

check_report() {   # check_report <file>
  local block f st tri
  f="$1"
  if ! block=$(fm_block "$f"); then
    fail "$f" "front matter must open at line 1 with --- and close with ---"
    return 0
  fi
  # No ledger: a report records an observation, and nothing was spent
  # producing it. A block list here is the fault it always was.
  check_shape "$f" "$block"
  for field in id status task_ref doc_ref created triaged; do
    require_once "$f" "$block" "$field"
  done
  check_id "$f" "$block"

  # The route triage took, not a lifecycle: one non-terminal value and
  # the five ends. There is no `resolved` — whether the underlying work
  # is done is the task's status, one hop away through task_ref, and a
  # second copy of that fact would need a second writer to stay true
  # (docs/product/concepts/report.md).
  st=$(get "$block" status)
  case "$st" in
    open|tracked|authored|fixed|declined|routed) ;;
    *) fail "$f" "status '$st' is not a report status (open|tracked|authored|fixed|declined|routed)" ;;
  esac

  # A list even with one element, because triage can split one finding
  # into several tasks — and it is the only link between the two kinds.
  check_list "$f" "$block" task_ref task

  check_doc_ref "$f" "$block"

  # `triaged` and a terminal status are the same fact written twice, so
  # they are paired both ways — the shape blocked/blocked_reason has, for
  # the same reason. A date on an `open` report says a judgement was made
  # that the status denies; a terminal one with no date is a judgement
  # nothing can be ordered against, and ordering these is what a
  # line-based reader does with a timestamp.
  check_date "$f" "$block" created strict
  check_date "$f" "$block" triaged null-ok
  tri=$(get "$block" triaged)
  if [ "$st" = "open" ] && [ "$tri" != "null" ]; then
    fail "$f" "status is open but triaged is '$tri' — a report carries the date only once triage has ended it"
  fi
  case "$st" in
    tracked|authored|fixed|declined|routed)
      [ "$tri" != "null" ] \
        || fail "$f" "status is '$st' but triaged is null — the date is when triage decided, and every end has one"
      ;;
  esac

  # An end and the field that names its outcome are one judgement written
  # twice, so they are paired the way `triaged` is. The table in
  # concepts/report.md is the contract: `tracked` is named by `task_ref`,
  # `authored` by the `doc_ref` the rule was written into. The other
  # ends name their outcome where this checker cannot follow — `fixed` in
  # the git history, `declined` and `routed` in the body (the reason, the
  # upstream issue) — so they are not paired here, and that asymmetry is
  # the concept's, not an omission. `routed` is bounded from the other
  # side instead: it sent the work to the repository that owns the
  # defect, so a task named here would claim work this queue never
  # gained.
  #
  # Unpaired, `status: tracked` with `task_ref: []` passed every gate: a
  # report permanently claiming work that nothing carries, and a mirror
  # closed `completed` on the strength of it.
  case "$st" in
    tracked)
      [ "$(get "$block" task_ref)" != "[]" ] \
        || fail "$f" "status is tracked but task_ref is empty — tracked means a task now carries the work, and task_ref is what names it" ;;
    authored)
      [ "$(get "$block" doc_ref)" != "null" ] \
        || fail "$f" "status is authored but doc_ref is null — authored means a rule was written, and doc_ref is what names it" ;;
    routed)
      [ "$(get "$block" task_ref)" = "[]" ] \
        || fail "$f" "status is routed but task_ref names a task — routed sent the work upstream, and the body names the issue it became" ;;
  esac
  return 0
}

for f in "$TASK_DIR"/*.md; do
  [ -f "$f" ] || continue
  case "$(basename "$f" | tr '[:upper:]' '[:lower:]')" in
    readme.md) continue ;;
    task-*.md) ;;
    *) continue ;;
  esac
  check_task "$f"
  checked=$((checked + 1))
done

for f in "$SPEC_DIR"/*.md; do
  [ -f "$f" ] || continue
  case "$(basename "$f" | tr '[:upper:]' '[:lower:]')" in
    readme.md) continue ;;
    spec-*.md) ;;
    *) continue ;;
  esac
  check_spec "$f"
  checked=$((checked + 1))
done

# A directory that is not there is zero reports, never an error: an
# adopter who has never recorded one has no work/reports/, and that is a
# complete state rather than a broken checkout (spec-0043).
for f in "$REPORT_DIR"/*.md; do
  [ -f "$f" ] || continue
  case "$(basename "$f" | tr '[:upper:]' '[:lower:]')" in
    readme.md) continue ;;
    report-*.md) ;;
    *) continue ;;
  esac
  check_report "$f"
  checked=$((checked + 1))
done

if [ "$status" -eq 0 ]; then
  echo "OK — ${checked} queue file(s), all canonical."
fi
exit "$status"
