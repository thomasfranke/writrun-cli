#!/usr/bin/env bash
# new.sh — scaffolds a schema-correct task, spec or report file.
#
# Usage:
#   new.sh task "<title>" --origin rule|report \
#                          [--slug words-here] \
#                          [--priority high|medium|low] \
#                          [--depends-on task-nnn[,task-mmm...]] \
#                          [--doc-ref path/to/doc.md#anchor] \
#                          [--milestone name] \
#                          [--from-report report-nnn]
#   new.sh spec task-nnn "<title>" [--slug words-here]
#   new.sh report "<title>" [--slug words-here] [--doc-ref path/to/doc.md#anchor]
#
# Run from the repository root — paths are relative to work/tasks,
# work/specs and work/reports there.
#
# **A report takes neither --origin nor --priority, and both are refused
# by name rather than as unknown flags.** `origin` is a fact about how a
# *task* came to exist and a report is one of its two answers — a report
# has no origin of its own. `priority` orders work, and a report commits
# to none: it says something was seen, and triage decides whether
# anything follows (docs/product/concepts/report.md).
#
# Exists because the schema in docs/technical/schemas.md is precise about
# details easy to get wrong from memory: spec_ref/depends_on are always
# lists even with one element, doc_ref is a full path+anchor resolved
# relative to docs/, and every field is present even when null. A
# hand-written task drifted on exactly these while this methodology's own
# docs/product/ chapters were being drafted (see
# docs/product/concepts/task.md#example) — this script fills them
# mechanically instead of relying on an agent recalling the rules
# correctly every time.
#
# `new.sh spec` also appends the new spec's id to its task's spec_ref list
# — appends, never overwrites existing entries. `new.sh task
# --from-report` is the same edit one kind up: the new task's id joins
# that report's `task_ref`, its `triaged` date is stamped, and a report
# still `open` becomes `tracked` — which is the route the append just
# took. That third write is not optional bookkeeping: `triaged` is null
# while a report is `open`, so stamping the date without moving the
# status would write a file the front-matter check refuses.
#
# The next id is minted above the queue, the history, *and* every open pull
# request — an id is unique across all three
# (docs/technical/README.md#task-schema). Consulting the forge is
# best-effort by design: no `gh`, no network, or no auth mints from this
# checkout alone, exactly as before, and says so. A narrower view is not
# wrong, it is narrower — and a silently narrower scan is how two branches
# cut from the same main both claimed 0009 here.
#
# Body templates resolve in three layers — the project's shape wins:
#   1. .writrun/conventions/templates/{task,spec,report}.md  — the project customized
#   2. .writrun/templates/{task,spec,report}.md   — WritRun's shipped default
#   3. the skeleton built into this script         — safety: works anywhere
# The template's {{id}}, {{title}}, {{task_ref}} and {{references}} are
# substituted.
#
# **References are navigable, not just resolvable**
# (docs/technical/README.md#task-schema). The front matter keeps its
# references as plain strings — it is the machine contract, and every
# reader here is line-based — while the generated *body* carries the same
# references as relative markdown links: a task's body links its
# `doc_ref` and each spec in `spec_ref`, a spec's body links its
# `task_ref`, a report's body links its `doc_ref`, and both appends below
# — a spec onto its task, a task onto its report — add the body link in
# the same edit. A template that drops {{references}} simply gets no links —
# taste, not contract.
#
# The *contract* front matter is always generated here, never templated —
# the machinery reads those fields, and a template that reshaped them
# would blind it silently. A project template may, however, open with a
# front-matter block of its own: those are **extension fields** (owner,
# estimate, whatever the project tracks), appended to the generated
# contract block. A template that redefines a contract field, or writes
# an extension the canonical form would reject, is refused — the same
# pattern as a spec template missing the contract headings (the two
# Proposed-changes sections and Outcome), which generation also refuses:
# a shape that would blind or fail a check is stopped where it is born.
#
# Exit codes: 0 success; 3 usage error or an invalid project template.

set -euo pipefail

# --- what open pull requests already claim --------------------------------
#
# FORGE_VIEW is `forge` once open pull requests have answered, `local`
# otherwise; FORGE_PATHS holds every path they touch. Never called from
# inside a command substitution — the subshell would swallow both.
FORGE_VIEW=local
FORGE_PATHS=""

# forge_scan — asks the forge for the paths open pull requests touch,
# added *or* modified. Coarser than the collision check downstream, and
# deliberately: a generator needs an upper bound, not an accusation. A
# modified queue file's id is already on the branch this checkout reads,
# so folding it in can only agree with what the tree said.
#
# The open numbers, then each pull request's file list — the same
# question check_unique_ids.sh asks, minus the per-file `status` it needs
# and this does not. One call per open pull request instead of one call
# total, because the single `gh pr list --json files` this replaced was
# cheaper and wrong: that field stops at 100 files per pull request and
# says nothing when it does, so a larger diff hides every queue file it
# adds past the cut. spec-0010's Outcome called the coarser question
# "strictly safe"; the cap is the case that argument missed.
#
# All or nothing: a call that fails leaves the view local rather than
# quietly narrow, since a scan that under-reports without saying so is
# the exact failure the uniqueness rule exists to prevent.
forge_scan() {
  command -v gh >/dev/null 2>&1 || return 0
  local numbers paths files n
  # gh defaults to 30 open pull requests, and the id this misses is
  # exactly the one worth seeing.
  numbers=$(gh pr list --state open --limit 200 --json number \
    --jq '.[].number' 2>/dev/null) || return 0
  paths=""
  for n in $numbers; do
    # --paginate is the point: a pull request's own file list is paged
    # too, and the queue file may sit on any page of it.
    files=$(gh api "repos/{owner}/{repo}/pulls/${n}/files" --paginate \
      --jq '.[].filename' 2>/dev/null) || return 0
    paths="${paths}${files}
"
  done
  FORGE_VIEW=forge
  FORGE_PATHS="$paths"
  return 0
}

# mint_report — what the id above was minted against, printed after the
# file so an id claimed elsewhere is never reported as simply "created".
mint_report() {
  if [ "$FORGE_VIEW" = forge ]; then
    echo "Minted above the queue, its history, and every open pull request."
  else
    echo "Minted from this checkout only — no forge answered, so an id an" >&2
    echo "open pull request already claims would not have been seen." >&2
  fi
}

next_id() {
  # next_id <dir> <prefix>  — e.g. next_id work/tasks task
  local dir="$1" prefix="$2" max=0 f n
  bump() {
    # The id is the digits after the prefix; a filename subject slug
    # follows them (task-0004-file-naming) and is not part of identity.
    n=$(basename "$1" .md | tr '[:upper:]' '[:lower:]' \
      | sed -E "s/^${prefix}-0*([0-9]+).*/\1/")
    [[ "$n" =~ ^[0-9]+$ ]] || return 0
    (( n > max )) && max=$n
    return 0
  }

  # Scan case-insensitively so historical uppercase IDs contribute to the
  # next number while newly generated filenames remain lowercase.
  # A directory that is not there is zero files, not a failure: an adopter
  # who has never recorded a report has no work/reports/, and minting the
  # first id is exactly the moment that is still true.
  while IFS= read -r f; do bump "$f"; done \
    < <(find "$dir" -maxdepth 1 -type f -iname "${prefix}-*.md" -print 2>/dev/null)

  # An id is never reused, including after its file was deleted — and a
  # deleted file is invisible to the scan above. Every id this directory
  # ever held is recoverable from the history, so ask it too. Outside a git
  # repository the filesystem is all there is, which is the correct answer
  # there: nothing was ever deleted from a history that doesn't exist.
  if git rev-parse --git-dir >/dev/null 2>&1; then
    while IFS= read -r f; do
      case "$f" in "$dir"/*) bump "$f" ;; esac
    done < <(git log --diff-filter=A --name-only --pretty=format: -- "$dir" 2>/dev/null)
  fi

  # An open pull request holds numbers no branch here can see: it may be
  # a fork's, and even from this repository it reaches this checkout only
  # once fetched. Its paths are the third input, and the one the tree and
  # the history cannot stand in for.
  if [ -n "$FORGE_PATHS" ]; then
    while IFS= read -r f; do
      case "$f" in "$dir"/*) bump "$f" ;; esac
    done <<EOF
$FORGE_PATHS
EOF
  fi

  printf "%04d" $((max + 1))
}

# slugify <title> — the filename's subject, **derived**: an extremely short
# kebab-case echo of the title, at most three words. This is the fallback,
# not the default — whoever creates the file chooses those words with
# --slug, because "which task is this, among these" is a judgement about
# the queue rather than a string operation on the title
# (docs/technical/README.md#task-schema). Identity lives in the id, so a
# derived slug that loses nuance costs nothing; it exists to make a
# directory listing readable. Prints nothing when the title has no usable
# word, and the caller then writes a bare-id filename.
slugify() {
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | sed 's/[^a-z0-9]\{1,\}/-/g; s/^-*//; s/-*$//' \
    | awk -F- '{
        n = 0
        for (i = 1; i <= NF && n < 3; i++) {
          if ($i == "") continue
          s = (n == 0 ? $i : s "-" $i)
          n++
        }
        print s
      }'
}

# check_slug <slug> — the shape the filename contract allows: lowercase
# alphanumerics and single interior hyphens, no leading or trailing one.
# Refused here, where it is typed, rather than written and discovered by
# check_front_matter.sh at the merge — the same reason a project template
# that would fail that check is refused at generation.
#
# Leading digits followed by a hyphen are refused separately, and the
# message says why: `task-0004-2-of-3.md` reads as id 4 to a human and to
# every prefix resolver in this repository, which take the digits after
# the prefix and stop at the first hyphen. A slug may hold digits — it
# may not open with the one shape that is already the id's.
check_slug() {
  local slug="$1"
  if [ -z "$slug" ]; then
    echo "--slug was given an empty string — omit the flag to derive one from the title" >&2
    exit 3
  fi
  if printf '%s' "$slug" | grep -qE '^[0-9]+-'; then
    echo "--slug '${slug}' opens with digits and a hyphen, which reads as a continuation of the id" >&2
    echo "The id is the digits after the prefix, up to the first hyphen — a slug in that shape is unresolvable." >&2
    exit 3
  fi
  if ! printf '%s' "$slug" | grep -qE '^[a-z0-9]+(-[a-z0-9]+)*$'; then
    echo "--slug '${slug}' is outside the filename contract" >&2
    echo "Lowercase alphanumerics and single interior hyphens, no leading or trailing hyphen." >&2
    exit 3
  fi
}

# queue_file <dir> <prefix> <id> — the file whose id is <id>, whatever
# its filename subject and whatever width its number was written at.
# Prints nothing when no file matches.
queue_file() {
  local dir="$1" prefix="$2" want f n
  want=$(printf '%s' "$3" | tr '[:upper:]' '[:lower:]' \
    | sed -E "s/^${prefix}-0*([0-9]+)$/\1/")
  [[ "$want" =~ ^[0-9]+$ ]] || return 0
  for f in "$dir"/"$prefix"-*.md; do
    [[ -f "$f" ]] || continue
    n=$(basename "$f" .md | tr '[:upper:]' '[:lower:]' \
      | sed -E "s/^${prefix}-0*([0-9]+).*/\1/")
    [[ "$n" =~ ^[0-9]+$ ]] || continue
    if (( n == want )); then printf '%s' "$f"; return 0; fi
  done
  return 0
}

# render_template <file> <id> <title> <task_ref> <references> — prints
# the template with {{id}}, {{title}}, {{task_ref}} and {{references}}
# substituted. Pure parameter expansion: titles may carry any character
# sed would trip on.
#
# An empty <references> takes the placeholder's whole line with it,
# blank separator included: a task with no `doc_ref` and no spec has
# nothing to link, and an empty References heading is noise.
render_template() {
  local content
  content=$(<"$1")
  content=${content//'{{id}}'/$2}
  content=${content//'{{title}}'/$3}
  content=${content//'{{task_ref}}'/$4}
  if [[ -n "${5:-}" ]]; then
    content=${content//'{{references}}'/$5}
  else
    content=${content//'{{references}}'$'\n\n'/}
    content=${content//'{{references}}'/}
  fi
  printf '%s\n' "$content"
}

# refs_line <link>... — the body's References line, or nothing when there
# is nothing to link. One line, so the `spec_ref` append has one place to
# add to.
refs_line() {
  local out="" l
  for l in "$@"; do
    [ -n "$l" ] || continue
    if [ -z "$out" ]; then out="$l"; else out="$out · $l"; fi
  done
  [ -n "$out" ] || return 0
  printf '**References:** %s' "$out"
}

# body_template_for <task|spec> — the project's template wins, then the
# shipped default; empty output means fall back to the built-in skeleton.
body_template_for() {
  if [[ -f ".writrun/conventions/templates/$1.md" ]]; then
    printf '.writrun/conventions/templates/%s.md' "$1"
  elif [[ -f ".writrun/templates/$1.md" ]]; then
    printf '.writrun/templates/%s.md' "$1"
  fi
}

# --- editing a file that already exists -------------------------------
#
# Three edits, shared by every subcommand that makes them: a spec appends
# its id to its task, a task appends its id to the report it was triaged
# from, and that triage moves the report's status and date. They were
# written twice, once per subcommand, and the two copies are the reason
# this section exists — the anchoring below had to be added to both, and
# a fix that lands in one of two identical awk programs is a fix that
# lands in neither.
#
# **Every one of them is bounded to the front-matter block.** A body that
# quotes front matter is what evidence looks like — this repository's own
# reports paste `status: open` into a fenced block — and an unanchored
# `/^status: /` rewrites it, silently falsifying the evidence while every
# check still passes, because the file is still canonical. The block is
# the only place these fields mean anything.
#
# Portable single-arg sub() only: no gawk-only 3-arg match (see
# writrun-check-spec-deltas/check_deltas.sh's own history with this).

# append_list_field <file> <field> <id> — add an id to an inline list,
# keeping what is there. Append, never overwrite: `spec_ref` and
# `task_ref` are lists because a task can need several specs and triage
# can split one finding into several tasks, and the second run must find
# the first one's id still in place.
append_list_field() {
  local file field newid
  file="$1"; field="$2"; newid="$3"
  awk -v field="$field" -v newid="$newid" '
    NR == 1 { infm = ($0 == "---"); print; next }
    infm && /^---$/ { infm = 0; print; next }
    infm && index($0, field ": [") == 1 {
      line = $0
      sub("^" field ": \\[", "", line)
      sub(/\]$/, "", line)
      if (line == "") { print field ": [" newid "]" }
      else { print field ": [" line ", " newid "]" }
      next
    }
    { print }
  ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

# set_fm_field <file> <field> <value> — replace a field's value.
set_fm_field() {
  local file field value
  file="$1"; field="$2"; value="$3"
  awk -v field="$field" -v value="$value" '
    NR == 1 { infm = ($0 == "---"); print; next }
    infm && /^---$/ { infm = 0; print; next }
    infm && index($0, field ": ") == 1 { print field ": " value; next }
    { print }
  ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

# append_body_link <file> <kind> <link> — the matching body link, in the
# same run as the front-matter append: the two must never disagree about
# which ids a file carries. An existing References line gains the link; a
# file that had nothing to link gets the line — under the body's first
# heading, whatever its level, or straight after the front matter when
# the body opens without one. That last fallback belongs to the
# invariant: an insert that finds no anchor must not quietly not happen,
# and a body shaped unlike the ones we imagined is still a body the front
# matter has to agree with.
#
# Unless the project asked for no links at all. A <kind> template without
# {{references}} opted its bodies out — taste, not contract, as the
# header says — and an opt-out that lasted until the file's first
# reference would be no opt-out. The front matter carries the id either
# way: that half is the machine contract, and never the template's to
# shape.
append_body_link() {
  local file kind link tpl
  file="$1"; kind="$2"; link="$3"
  tpl=$(body_template_for "$kind")
  if [[ -n "$tpl" ]] && ! grep -q '{{references}}' "$tpl"; then return 0; fi
  if grep -q '^\*\*References:\*\* ' "$file"; then
    awk -v link="$link" '
      !done && /^\*\*References:\*\* / { print $0 " · " link; done = 1; next }
      { print }
    ' "$file" > "${file}.tmp"
  elif grep -qE '^#+ ' "$file"; then
    awk -v link="$link" '
      !done && /^#+ / { print; print ""; print "**References:** " link; done = 1; next }
      { print }
    ' "$file" > "${file}.tmp"
  else
    awk -v link="$link" '
      fm < 2 && /^---$/ {
        fm++; print
        if (fm == 2) { print ""; print "**References:** " link }
        next
      }
      { print }
    ' "$file" > "${file}.tmp"
  fi
  mv "${file}.tmp" "$file"
}

# The contract fields — the script's to write, never a template's.
TASK_CONTRACT="id status blocked_reason taken_by spec_ref doc_ref origin priority depends_on milestone created completed provenance"
SPEC_CONTRACT="id task_ref status created"
REPORT_CONTRACT="id status task_ref doc_ref created triaged"

# template_extensions — stdin: a rendered template. Prints the lines of
# its leading front-matter block (the extension fields); nothing when the
# template has no block.
template_extensions() {
  awk 'NR == 1 { if ($0 != "---") exit; next } /^---$/ { exit } { print }'
}

# template_body — stdin: a rendered template. Prints everything after the
# front-matter block (minus one separating blank line), or the whole
# input when there is no block.
template_body() {
  awk '
    NR == 1 && $0 == "---" { infm = 1; next }
    infm && /^---$/        { infm = 0; sep = 1; next }
    infm                   { next }
    sep && $0 == ""        { sep = 0; next }
    { sep = 0; print }
  '
}

# validate_extensions <template-file> <ext-lines> <contract-field-list> —
# refuses a template whose front-matter block redefines a contract field
# or writes a line the canonical form (check_front_matter.sh) would
# reject at the merge. Refused here, where the shape is born, so the
# generated file never carries the failure forward.
validate_extensions() {
  local tpl="$1" lines="$2" contract="$3" line key val seen=""
  [[ -n "$lines" ]] || return 0
  while IFS= read -r line; do
    if [[ -z "$line" ]]; then
      echo "${tpl}: front-matter extensions must not hold empty lines" >&2; exit 3
    fi
    case "$line" in
      *": "*) key="${line%%: *}"; val="${line#*: }" ;;
      *) echo "${tpl}: extension line is not 'key: value': '$line'" >&2; exit 3 ;;
    esac
    if ! printf '%s' "$key" | grep -qE '^[A-Za-z_][A-Za-z0-9_-]*$'; then
      echo "${tpl}: malformed extension key '$key'" >&2; exit 3
    fi
    case " $contract " in *" $key "*)
      echo "${tpl}: template redefines the contract field '$key' — contract front matter is generated, never templated" >&2
      exit 3 ;;
    esac
    case " $seen " in *" $key "*)
      echo "${tpl}: extension field '$key' appears twice" >&2; exit 3 ;;
    esac
    seen="$seen $key"
    case "$val" in
      \"*|\'*)
        echo "${tpl}: extension '$key' is quoted — values are written bare" >&2; exit 3 ;;
      '>'|'|'|'>-'|'|-')
        echo "${tpl}: extension '$key' uses a block scalar — outside the canonical form" >&2; exit 3 ;;
    esac
    if printf '%s' "$val" | grep -q '[[:space:]]$'; then
      echo "${tpl}: extension '$key' carries trailing whitespace" >&2; exit 3
    fi
  done <<EOF
$lines
EOF
  return 0
}

usage() {
  echo "Usage: new.sh task \"<title>\" --origin rule|report [--slug words-here] [--priority high|medium|low] [--depends-on task-nnn,...] [--doc-ref path#anchor] [--milestone name] [--from-report report-nnn]" >&2
  echo "       new.sh spec task-nnn \"<title>\" [--slug words-here]" >&2
  echo "       new.sh report \"<title>\" [--slug words-here] [--doc-ref path#anchor]" >&2
  exit 3
}

cmd="${1:-}"
[[ -z "$cmd" ]] && usage

case "$cmd" in
  task)
    shift
    title="${1:-}"
    [[ -z "$title" ]] && usage
    shift
    priority=medium
    depends_on=""
    doc_ref=null
    origin=""
    origin_given=""
    milestone=null
    slug_given=""
    from_report=""
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --slug) slug_given="${2-}"; slug_chosen=1; shift 2 ;;
        --priority) priority="$2"; shift 2 ;;
        --depends-on) depends_on="$2"; shift 2 ;;
        --doc-ref) doc_ref="$2"; shift 2 ;;
        --origin) origin="${2-}"; origin_given=1; shift 2 ;;
        --milestone) milestone="$2"; shift 2 ;;
        --from-report) from_report="${2-}"; shift 2 ;;
        *) echo "Unknown flag: $1" >&2; exit 3 ;;
      esac
    done

    case "$priority" in
      high|medium|low) ;;
      *) echo "Invalid --priority '$priority' — expected high, medium, or low" >&2; exit 3 ;;
    esac

    # **`--from-report` states the origin rather than defaulting it.** The
    # flag names the report this task was born from, which is the whole
    # content of `origin: report` — so the fact is typed, once, and the
    # two cannot disagree. `--origin rule` alongside it is refused for the
    # same reason: a task cannot be derived from a rule and born from a
    # report, and picking a winner silently is the wrong fact nobody typed.
    if [[ -n "$from_report" ]]; then
      if [[ -n "$origin_given" && "$origin" != "report" ]]; then
        echo "--from-report names the report this task was born from, which is 'origin: report' — --origin '$origin' contradicts it" >&2
        exit 3
      fi
      origin=report
    fi

    # No default, deliberately. `origin` is a fact about how the task came
    # to exist, written once and never rewritten — and a default would
    # record one of the two silently, for whichever kind of change happened
    # not to say. A wrong fact nobody typed is the failure the field exists
    # to prevent, so an unstated origin refuses (docs/technical/README.md#task-schema).
    case "$origin" in
      rule|report) ;;
      "") echo "--origin is required — 'rule' for a task derived from an authored rule, 'report' for one born from a report (or --from-report, which states it)" >&2; exit 3 ;;
      *) echo "Invalid --origin '$origin' — expected rule or report" >&2; exit 3 ;;
    esac

    # **The tracked route travels on a change of its own, and the cheapest
    # place to say so is before anything is written.** A task with
    # `origin: report` is what that route puts in the queue, and what
    # enters the queue passes a gate: the reporting change's own pull
    # request, whose squash-merge is the assent that the finding deserves
    # the work (docs/product/concepts/report.md). `check_state.sh` refuses
    # such a task on any other branch — so a run that minted it here would
    # flip the report to `tracked`, stamp `triaged`, write the task and
    # hand back a change CI cannot pass, whose undo is by hand across
    # three files. The generator that drives the route is where the branch
    # is still free to change.
    #
    # The stage condition is the rule's own: below Stage 2 there is no
    # pull request to be the vehicle, so the route runs wherever the
    # project works. A name that cannot be read (detached HEAD, no
    # `HEAD_REF`) is not judged — CI's copy of the rule sees the branch
    # this one could not.
    if [[ "$origin" = "report" ]]; then
      tracked_stage=$(bash "$(cd "$(dirname "$0")" && pwd)/../../scripts/stage-2-pull-requests/read_setting.sh" stage 2>/dev/null || printf '3')
      case "$tracked_stage" in 1|2|3) ;; *) tracked_stage=3 ;; esac
      tracked_branch="${HEAD_REF:-}"
      if [[ -z "$tracked_branch" ]]; then
        tracked_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
        if [[ "$tracked_branch" = "HEAD" ]]; then tracked_branch=""; fi
      fi
      if [[ "$tracked_stage" -ge 2 && -n "$tracked_branch" && "$tracked_branch" != report/* ]]; then
        echo "'${tracked_branch}' is not a report/ branch — a task with 'origin: report' is the tracked route, and that route never rides" >&2
        echo "Deriving a task from a report is a reporting change of its own: its pull request presents the report, the task and the spec together, and the maintainer's squash-merge is the assent that the finding deserves the work (docs/product/concepts/report.md#recording-rides-any-change--routing-to-the-queue-does-not)." >&2
        echo "Open that change and run this again — 'git switch -c report/<short-name>'. Recording a report still rides anything; so do the fixed and declined ends." >&2
        exit 3
      fi
    fi

    # Validate before the forge is consulted: a refusal must cost nothing
    # and touch nothing, and an id minted for a file never written is an
    # id the next run would mint again anyway.
    if [[ -n "${slug_chosen:-}" ]]; then check_slug "$slug_given"; fi

    # The report is resolved here, before anything is written, for the same
    # reason: a task created against a report that does not resolve leaves
    # the link half made, and nothing downstream would ever notice.
    report_file=""
    report_status=""
    if [[ -n "$from_report" ]]; then
      report_file=$(queue_file work/reports report "$from_report")
      [[ -n "$report_file" && -f "$report_file" ]] \
        || { echo "No such report: ${from_report} — --from-report names a report that is already recorded" >&2; exit 3; }
      report_status=$(sed -n 's/^status: *//p' "$report_file" | head -n1 | sed 's/[[:space:]]*$//')
      case "$report_status" in
        open|tracked) ;;
        authored|fixed|declined)
          echo "${report_file} is '${report_status}' — a terminal report is never re-routed" >&2
          echo "Triage ended it, and a recurrence is a new report: ids are never reused and a second sighting keeps its own date (docs/product/concepts/report.md)." >&2
          exit 3 ;;
        *)
          echo "${report_file} carries status '${report_status}', which is not a report status" >&2
          exit 3 ;;
      esac
    fi

    forge_scan
    id="task-$(next_id work/tasks task)"
    if [[ -n "${slug_chosen:-}" ]]; then slug="$slug_given"; else slug=$(slugify "$title"); fi
    file="work/tasks/${id}${slug:+-$slug}.md"
    [[ -e "$file" ]] && { echo "$file already exists" >&2; exit 3; }

    if [[ -n "$depends_on" ]]; then
      depends_list="[$(echo "$depends_on" | sed 's/,/, /g')]"
    else
      depends_list="[]"
    fi

    # A task's body links its doc_ref, resolved from work/tasks/. The
    # front-matter value stays exactly as typed — relative to docs/ —
    # because that is what the machinery reads; the link is the reader's
    # half of the same fact. `spec_ref` is empty at creation, so the
    # spec links arrive through the append below, never here.
    doc_link=""
    if [[ "$doc_ref" != "null" ]]; then
      doc_link="[${doc_ref}](../../docs/${doc_ref})"
    fi
    references=$(refs_line "$doc_link")

    # Resolve and validate the template before writing anything — a
    # refusal must leave no half-written file behind.
    tpl=$(body_template_for task)
    tpl_ext=""
    tpl_body=""
    if [[ -n "$tpl" ]]; then
      rendered=$(render_template "$tpl" "$id" "$title" "" "$references")
      tpl_ext=$(printf '%s\n' "$rendered" | template_extensions)
      tpl_body=$(printf '%s\n' "$rendered" | template_body)
      validate_extensions "$tpl" "$tpl_ext" "$TASK_CONTRACT"
    fi

    {
      cat <<EOF
---
id: ${id}
status: backlog
blocked_reason: null
taken_by: null
spec_ref: []
doc_ref: ${doc_ref}
origin: ${origin}
priority: ${priority}
depends_on: ${depends_list}
milestone: ${milestone}
created: $(date -u +%Y-%m-%dT%H:%M:%SZ)
queued: null
completed: null
merged: null
provenance: []
EOF
      if [[ -n "$tpl_ext" ]]; then printf '%s\n' "$tpl_ext"; fi
      printf '%s\n\n' "---"
      if [[ -n "$tpl" ]]; then
        printf '%s\n' "$tpl_body"
      else
        printf '# %s\n\n' "$title"
        [[ -n "$references" ]] && printf '%s\n\n' "$references"
        cat <<EOF
TODO: what to do, and why it matters. No technical detail — that belongs
in the spec.
EOF
      fi
    } > "$file"

    # --- the mechanical half of triage ------------------------------------
    #
    # The link runs one way: the report names the tasks triage produced,
    # and the task schema is untouched
    # (docs/technical/decisions/tasks-and-specs/0064-a-report-is-an-artefact.md).
    # So this edit belongs here rather than in an agent's memory — the
    # generator is the only place that knows both ids at once.
    if [[ -n "$report_file" ]]; then
      append_list_field "$report_file" task_ref "$id"

      # `open` becomes `tracked` and the date is stamped, in the same run.
      # A report already `tracked` keeps both: the date belongs to the
      # first task triage produced, and a second one does not re-date the
      # judgement. `triaged` is null exactly while a report is `open`, so
      # these two move together or the file stops being canonical.
      if [[ "$report_status" = "open" ]]; then
        set_fm_field "$report_file" status tracked
        set_fm_field "$report_file" triaged "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
      fi

      append_body_link "$report_file" report "[${id}](../tasks/$(basename "$file"))"

      if [[ "$report_status" = "open" ]]; then
        echo "Created ${file} (${id}), appended to ${report_file}'s task_ref — it is now tracked"
      else
        echo "Created ${file} (${id}), appended to ${report_file}'s task_ref"
      fi
      mint_report
      exit 0
    fi

    echo "Created ${file} (${id})"
    mint_report
    ;;

  spec)
    shift
    task_id="${1:-}"
    title="${2:-}"
    [[ -z "$task_id" || -z "$title" ]] && usage
    shift 2
    slug_given=""
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --slug) slug_given="${2-}"; slug_chosen=1; shift 2 ;;
        *) echo "Unknown flag: $1" >&2; exit 3 ;;
      esac
    done
    if [[ -n "${slug_chosen:-}" ]]; then check_slug "$slug_given"; fi

    task_file=$(queue_file work/tasks task "$task_id")
    [[ -n "$task_file" && -f "$task_file" ]] \
      || { echo "No such task: ${task_id} — a spec is never created before its task" >&2; exit 3; }
    # The task's file is the authority on how its id is written — the
    # argument only had to identify it, and `task-1` or a historical
    # `task-001` must still record the reference the queue actually holds.
    task_id=$(sed -n 's/^id: *//p' "$task_file" | head -n1 | sed 's/[[:space:]]*$//')

    forge_scan
    id="spec-$(next_id work/specs spec)"
    if [[ -n "${slug_chosen:-}" ]]; then slug="$slug_given"; else slug=$(slugify "$title"); fi
    file="work/specs/${id}${slug:+-$slug}.md"
    [[ -e "$file" ]] && { echo "$file already exists" >&2; exit 3; }

    # A spec's body links back to its task, resolved from work/specs/.
    # The link targets the filename, which never changes — identity is
    # never order, and a retitled task is never renamed — so no link
    # here ever needs maintaining.
    references=$(refs_line "[${task_id}](../tasks/$(basename "$task_file"))")

    # Resolve and validate the template before writing anything — a
    # refusal must leave no half-written file behind.
    tpl=$(body_template_for spec)
    tpl_ext=""
    tpl_body=""
    if [[ -n "$tpl" ]]; then
      # A template may reshape everything except the contract: the delta
      # check greps these headings, and the completion flow fills Outcome
      # — a template without them blinds both silently.
      for h in "## Proposed product changes" "## Proposed technical changes" "## Outcome"; do
        if ! grep -qxF "$h" "$tpl"; then
          echo "${tpl} is missing the contract heading: ${h}" >&2
          echo "Add it back — the machinery reads these headings (see the public contract)." >&2
          exit 3
        fi
      done
      rendered=$(render_template "$tpl" "$id" "$title" "$task_id" "$references")
      tpl_ext=$(printf '%s\n' "$rendered" | template_extensions)
      tpl_body=$(printf '%s\n' "$rendered" | template_body)
      validate_extensions "$tpl" "$tpl_ext" "$SPEC_CONTRACT"
    fi

    {
      cat <<EOF
---
id: ${id}
task_ref: ${task_id}
status: draft
created: $(date -u +%Y-%m-%dT%H:%M:%SZ)
EOF
      if [[ -n "$tpl_ext" ]]; then printf '%s\n' "$tpl_ext"; fi
      printf '%s\n\n' "---"
    } > "$file"
    if [[ -n "$tpl" ]]; then
      printf '%s\n' "$tpl_body" >> "$file"
      skip_default_body=1
    else
      skip_default_body=0
    fi
    if [[ "$skip_default_body" -eq 0 ]]; then
      printf '# %s — %s\n\n' "$id" "$title" >> "$file"
      [[ -n "$references" ]] && printf '%s\n\n' "$references" >> "$file"
    fi
    [[ "$skip_default_body" -eq 1 ]] || cat >> "$file" <<EOF
- **Goal:** TODO

## Scope

TODO

## Steps

1. TODO

## Acceptance criteria (EARS)

- TODO

## Edge cases

- TODO

## Tests required

TODO

## Definition of Done

- [ ] TODO

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- none — no machinery change

## Outcome

_(fill after execution)_
EOF

    append_list_field "$task_file" spec_ref "$id"
    append_body_link "$task_file" task "[${id}](../specs/$(basename "$file"))"

    echo "Created ${file} (${id}), appended to ${task_file}'s spec_ref"
    mint_report
    ;;

  report)
    shift
    title="${1:-}"
    [[ -z "$title" ]] && usage
    shift
    doc_ref=null
    slug_given=""
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --slug) slug_given="${2-}"; slug_chosen=1; shift 2 ;;
        --doc-ref) doc_ref="$2"; shift 2 ;;
        # Named refusals, not "unknown flag". Both of these are real flags
        # one kind over, so a person reaching for them has the wrong model
        # of what a report is — and the message is the cheapest place to
        # correct it (docs/product/concepts/report.md).
        --origin)
          echo "--origin is a task's — a report has none: it is what 'origin: report' refers to" >&2
          exit 3 ;;
        --priority)
          echo "--priority is a task's — a report commits to no work, so nothing prioritises it" >&2
          echo "Whether anything follows is triage's answer, and it is written as the report's status." >&2
          exit 3 ;;
        *) echo "Unknown flag: $1" >&2; exit 3 ;;
      esac
    done

    if [[ -n "${slug_chosen:-}" ]]; then check_slug "$slug_given"; fi

    forge_scan
    id="report-$(next_id work/reports report)"
    if [[ -n "${slug_chosen:-}" ]]; then slug="$slug_given"; else slug=$(slugify "$title"); fi
    file="work/reports/${id}${slug:+-$slug}.md"
    [[ -e "$file" ]] && { echo "$file already exists" >&2; exit 3; }

    # A report's body links its doc_ref, resolved from work/reports/ — the
    # same two hops up a task's takes, because the two directories sit
    # side by side. The front-matter value stays exactly as typed, relative
    # to docs/, because that is what the machinery reads. `task_ref` is
    # empty at creation: the links arrive through --from-report, at triage.
    doc_link=""
    if [[ "$doc_ref" != "null" ]]; then
      doc_link="[${doc_ref}](../../docs/${doc_ref})"
    fi
    references=$(refs_line "$doc_link")

    tpl=$(body_template_for report)
    tpl_ext=""
    tpl_body=""
    if [[ -n "$tpl" ]]; then
      rendered=$(render_template "$tpl" "$id" "$title" "" "$references")
      tpl_ext=$(printf '%s\n' "$rendered" | template_extensions)
      tpl_body=$(printf '%s\n' "$rendered" | template_body)
      validate_extensions "$tpl" "$tpl_ext" "$REPORT_CONTRACT"
    fi

    mkdir -p work/reports
    {
      cat <<EOF
---
id: ${id}
status: open
task_ref: []
doc_ref: ${doc_ref}
created: $(date -u +%Y-%m-%dT%H:%M:%SZ)
triaged: null
EOF
      if [[ -n "$tpl_ext" ]]; then printf '%s\n' "$tpl_ext"; fi
      printf '%s\n\n' "---"
      if [[ -n "$tpl" ]]; then
        printf '%s\n' "$tpl_body"
      else
        printf '# %s\n\n' "$title"
        [[ -n "$references" ]] && printf '%s\n\n' "$references"
        cat <<EOF
TODO: what was observed, and the evidence at hand. What should be done
about it is triage's output, never this file's content.
EOF
      fi
    } > "$file"
    echo "Created ${file} (${id})"
    mint_report
    ;;

  *)
    usage
    ;;
esac
