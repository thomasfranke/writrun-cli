#!/usr/bin/env bash
# take_task.sh — takes a task in one act: the eligibility re-checked, the
# branch cut from a fresh origin/main, given its first commit, pushed,
# and the draft pull request opened.
#
#   bash .writrun/scripts/stage-2-pull-requests/take_task.sh <task-id> \
#     --title "<summary>" [--slug words] [--coauthor "Name <address>"] \
#     [--resume] [--confirm]
#
# **The push and the opening are one act** (conventions/prs.md): the
# branch reaching the forge and the draft opening is the moment work
# becomes public, and it is what an adopter's `auto_push`/`auto_pr` gate.
# A branch on the forge with no pull request is the hiding place the
# taking flow exists to close, so this script never leaves one behind on
# purpose — and names the one it left behind when the forge failed.
#
# **The first commit is empty, and that is the record.** A branch
# identical to origin/main has no commits between the two, and the forge
# refuses a pull request over nothing — so the push has to carry
# something. The take produced no content, and a commit with no diff is
# the honest account of that; the squash-merge discards it, so nothing of
# it reaches main.
#
# What it does not decide: which task to work (list_tasks.sh's), and what
# the title's summary should say (the agent's, validated here). It writes
# no queue file — the status line has one writer, and it is the
# machinery answering the draft this opens. That holds for the first
# commit too: stamping the task file here would make the take a second
# writer on a line that has one.
#
# Exit codes:
#   0  the branch is pushed and the draft pull request is open.
#   1  a refusal — an ineligible task, a title the declared style refuses,
#      a branch that already exists, or an unusable argument. Nothing was
#      created.
#   2  composed and waiting: `auto_commit`, `auto_push` or `auto_pr` is
#      false, so the branch, the first commit's message, the title and
#      the body are printed and nothing was done. Rerun with --confirm
#      to perform exactly the printed act.
#   3  the forge or git failed. Before the branch was cut the repository
#      is untouched; after it, the branch is named wherever it got to,
#      and --resume finishes the act.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/queue_lib.sh"
READ_SETTING="$HERE/read_setting.sh"
OBSERVANCE="$HERE/check_observance.sh"

TASK_ARG=""
TITLE=""
SLUG=""
COAUTHOR=""
RESUME=""
CONFIRM=""

usage() {
  cat >&2 <<'USAGE'
usage: take_task.sh <task-id> --title "<summary>" [--slug words] [--coauthor "Name <address>"] [--resume] [--confirm]
USAGE
}

refuse() { echo "REFUSED: $*" >&2; exit 1; }

# The count is checked before the shift, not after it: `shift 2` with one
# word left shifts *nothing* and merely reports it, so a loop that shrugs
# the report off reads the same word forever. A missing value has to be a
# refusal — a take that hangs is worse than one that fails.
while [ "$#" -gt 0 ]; do
  case "$1" in
    --title)   [ "$#" -ge 2 ] || refuse "--title needs a summary after it"
               TITLE="$2"; shift 2 ;;
    --slug)    [ "$#" -ge 2 ] || refuse "--slug needs words after it"
               SLUG="$2"; shift 2 ;;
    --coauthor) [ "$#" -ge 2 ] || refuse "--coauthor needs a name and address after it"
               COAUTHOR="$2"; shift 2 ;;
    --resume)  RESUME=yes; shift ;;
    --confirm) CONFIRM=yes; shift ;;
    -h|--help) usage; exit 1 ;;
    -*)        echo "REFUSED: unknown option '$1'" >&2; usage; exit 1 ;;
    *)         [ -n "$TASK_ARG" ] && { echo "REFUSED: two task ids given ('$TASK_ARG' and '$1')" >&2; exit 1; }
               TASK_ARG="$1"; shift ;;
  esac
done

[ -n "$TASK_ARG" ] || { usage; exit 1; }

# Everything below reads and writes repository paths, so a run from a
# subdirectory re-roots first rather than reporting a queue it cannot see.
TOP=$(git rev-parse --show-toplevel 2>/dev/null) || {
  echo "REFUSED: not a git repository" >&2; exit 1; }
cd "$TOP" || exit 1

STYLE=$(bash "$READ_SETTING" stage_2.pr_title_style)
AUTO_COMMIT=$(bash "$READ_SETTING" stage_2.auto_commit)
AUTO_PUSH=$(bash "$READ_SETTING" stage_2.auto_push)
AUTO_PR=$(bash "$READ_SETTING" stage_2.auto_pr)
CREDIT=$(bash "$READ_SETTING" stage_2.agent_coauthor)

# The first commit is a commit in the pull request's range, so it owes
# whatever the flag obliges — and the flag is read in both directions,
# exactly as check_observance.sh reads it. Who is running this is the one
# thing the script cannot know: an agent commits under the same name and
# address as the person who ran it, so the name is given rather than
# guessed at, and a take with no --coauthor is a take by a person.
if [ -n "$COAUTHOR" ] && [ "$CREDIT" = false ]; then
  refuse "--coauthor was given while stage_2.agent_coauthor is false — this project's commits carry no credit trailer"
fi

# The three vocabularies come from check_observance.sh's own assignment
# lines — the machine half of conventions/commits.md, and a single
# source. A title this script accepted and the door then refused would
# be worse than no validation at all, and the same holds for the name
# this script writes into a trailer.
TYPES=$(sed -n 's/^TYPES="\(.*\)"$/\1/p' "$OBSERVANCE" | head -n1)
SCOPES=$(sed -n 's/^SCOPES="\(.*\)"$/\1/p' "$OBSERVANCE" | head -n1)
# CATEGORIES is written over two lines with a continuation, so it is read
# as the span from its assignment to the line that closes the quote.
CATEGORIES=$(awk '/^CATEGORIES="/ { c = 1 }
                  c { print }
                  c && /"[[:space:]]*$/ { exit }' "$OBSERVANCE" \
  | sed 's/^CATEGORIES="//; s/"[[:space:]]*$//; s/\\$//' \
  | tr '\n' ' ' | tr -s ' ')
if [ -z "$TYPES" ] || [ -z "$SCOPES" ] || [ -z "$CATEGORIES" ]; then
  echo "Could not read TYPES/SCOPES/CATEGORIES from ${OBSERVANCE} — a title" >&2
  echo "or a trailer checked against a vocabulary this script had to guess" >&2
  echo "at is not checked." >&2
  exit 3
fi

# The trailer is written onto the commit verbatim, so its shape is
# judged here rather than left to the gate that reads it hours later.
# One line, a name and an address — a value carrying a newline would
# write arbitrary lines into the commit body — and a name the door
# refuses is refused at the door that offers the flag.
if [ -n "$COAUTHOR" ]; then
  if [ "$(printf '%s' "$COAUTHOR" | wc -l | tr -d '[:space:]')" != 0 ] \
     || ! printf '%s' "$COAUTHOR" | grep -q '^[^<>]\{1,\}<[^<>]\{1,\}>$'; then
    refuse "--coauthor takes one line of the form \"Name <address>\" — for example --coauthor \"Claude Opus 5 <noreply@anthropic.com>\""
  fi
  CA_NAME=$(printf '%s' "$COAUTHOR" \
    | sed -e 's/[[:space:]]*<.*$//' -e 's/[[:space:]]*$//' \
    | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
  case " $CATEGORIES " in
    *" $CA_NAME "*)
      refuse "--coauthor names '${CA_NAME}', a category rather than a model — the record has to survive the next model's arrival, so name it: --coauthor \"Claude Opus 5 <...>\"" ;;
  esac
fi

in_list() {
  local n
  n=$(printf '%s' "$1" | tr 'A-Z' 'a-z')
  case " $2 " in *" $n "*) return 0 ;; esac
  return 1
}

# example_title — one valid title in the declared style, printed with
# every title refusal so the fix needs no second document.
example_title() {
  case "$STYLE" in
    bracketed) printf '[Feat][Ci] Debounce the mirror updates' ;;
    *)         printf 'feat(ci): debounce the mirror updates' ;;
  esac
}

# valid_summary <summary> — the same grammar check_observance.sh applies
# to what follows the task tags. Refusing here costs a rerun; refusing at
# the door costs a pull request already open under a title the project
# said it would not have.
valid_summary() {
  local s="$1" t sc
  case "$STYLE" in
    conventional)
      printf '%s' "$s" | grep -qE '^[a-z]+(\([a-z-]+\))?: .+$' || return 1
      t=$(printf '%s' "$s" | sed -E 's/^([a-z]+).*/\1/')
      sc=$(printf '%s' "$s" | sed -nE 's/^[a-z]+\(([a-z-]+)\):.*/\1/p') ;;
    bracketed)
      printf '%s' "$s" | grep -qE '^\[[A-Za-z]+\](\[[A-Za-z-]+\])? .+$' || return 1
      t=$(printf '%s' "$s" | sed -E 's/^\[([A-Za-z]+)\].*/\1/')
      sc=$(printf '%s' "$s" | sed -nE 's/^\[[A-Za-z]+\]\[([A-Za-z-]+)\].*/\1/p') ;;
    *)
      # An unreadable style is check_settings.sh's to name; judging a
      # title against a vocabulary nobody declared would refuse every
      # honest one for a fault in another file.
      return 0 ;;
  esac
  in_list "$t" "$TYPES" || { echo "  the type '${t}' is outside the vocabulary (${TYPES})" >&2; return 1; }
  if [ -n "$sc" ]; then
    in_list "$sc" "$SCOPES" || { echo "  the scope '${sc}' is outside the vocabulary (${SCOPES})" >&2; return 1; }
  fi
  return 0
}

# ---------------------------------------------------------- the task

TASK_FILE=$(ql_task_file "$TASK_ARG")
[ -n "$TASK_FILE" ] || refuse "no task in work/tasks/ resolves '${TASK_ARG}'"

ID=$(ql_fm_field id "$TASK_FILE")
NUM=$(ql_task_num "$ID")
STATUS=$(ql_fm_field status "$TASK_FILE")
TAKEN=$(ql_fm_field taken_by "$TASK_FILE")

# A dirty tree would ride into the branch this cuts, and the first commit
# of an implementation is not the place to discover it.
DIRTY=$(git status --porcelain 2>/dev/null)
[ -z "$DIRTY" ] || refuse "the working tree is dirty — commit or stash first:
$(printf '%s\n' "$DIRTY" | head -n 5)"

# The eligibility is re-read against a fresh authority branch: a stale
# origin/main is how a checkout reports a task nobody may take as free.
if ! FETCH_ERR=$(git fetch origin main 2>&1); then
  echo "git fetch origin main failed:" >&2
  printf '%s\n' "$FETCH_ERR" | head -n 3 >&2
  echo "The eligibility below would be read against a stale base, so nothing was done." >&2
  exit 3
fi

# --- selection steps 2–4, re-checked --------------------------------
#
# The lister already applied these, and it is re-applied here because the
# two runs are separated by however long the session took to decide.

[ "$STATUS" = ready ] || refuse "${ID} is '${STATUS}', and only a 'ready' task may be taken"

for d in $(ql_fm_field depends_on "$TASK_FILE" | tr -d '[]' | tr ',' ' '); do
  [ -n "$d" ] || continue
  df=$(ql_task_file "$d")
  ds=missing
  [ -n "$df" ] && ds=$(ql_fm_field status "$df")
  [ "$ds" = done ] || refuse "${ID} waits on ${d}, which is '${ds}'"
done

SPECS=$(ql_fm_field spec_ref "$TASK_FILE" | tr -d '[]' | tr ',' ' ')
for s in $SPECS; do
  [ -n "$s" ] || continue
  sf=$(ql_spec_file "$s")
  ss=missing
  [ -n "$sf" ] && ss=$(ql_fm_field status "$sf")
  case "$ss" in
    approved|implemented) ;;
    *) refuse "${ID}'s ${s} is '${ss}' — a task whose spec is not approved is not authorized work" ;;
  esac
done

[ -z "$TAKEN" ] || [ "$TAKEN" = null ] \
  || refuse "${ID} is already taken by ${TAKEN}"

# --- compose, touching nothing ---------------------------------------

if [ -z "$TITLE" ]; then
  echo "REFUSED: --title is required — the summary after the task tag is yours to write." >&2
  echo "  e.g. --title \"$(example_title)\"" >&2
  exit 1
fi

# The summary is judged before the tag is prepended, so a title that
# already carries one is refused here — and the example printed below
# carries no tag, which reads as "your tag is wrong" unless the refusal
# says whose the tag is.
if ! valid_summary "$TITLE"; then
  echo "REFUSED: '${TITLE}' does not read as the declared '${STYLE}' style." >&2
  echo "  --title takes the summary alone; the leading [TASK-NNNN] tag is this script's to prepend." >&2
  echo "  e.g. --title \"$(example_title)\"" >&2
  exit 1
fi

# The slug a human chose at creation is the filename's subject; deriving
# one from the title is the fallback, exactly as it is in new.sh.
if [ -z "$SLUG" ]; then
  SLUG=$(basename "$TASK_FILE" .md | sed -n 's/^task-[0-9][0-9]*-//p')
fi
if [ -z "$SLUG" ]; then
  SLUG=$(sed -n 's/^# *//p' "$TASK_FILE" | head -n1 \
    | tr 'A-Z' 'a-z' | sed -e 's/[^a-z0-9]\{1,\}/-/g' -e 's/^-//' -e 's/-$//' \
    | cut -d- -f1-3)
fi
SLUG=$(printf '%s' "$SLUG" | tr ' ' '-')

BRANCH=$(printf 'task/%04d-%s' "$NUM" "$SLUG")

# The tag leads the title in both styles, and how it meets the summary is
# the style's to spell: `bracketed` runs the brackets together, and
# `conventional` keeps one space
# (technical/settings/titles.md#pr_title_style). One format string
# carrying the space always titled every bracketed pull request against
# the card a session is told to run for a value, and an agent reading an
# open pull request to learn the form learned the disagreement.
#
# An unreadable style keeps the space, the same direction valid_summary
# takes: check_settings.sh is where a style nobody declared is named, and
# composing the tighter spelling on a guess would be this script deciding
# it. The door's tolerance for the space is untouched — every title
# already open stays valid.
case "$STYLE" in
  bracketed) TAG_GAP='' ;;
  *)         TAG_GAP=' ' ;;
esac
PR_TITLE=$(printf '[TASK-%04d]%s%s' "$NUM" "$TAG_GAP" "$TITLE")

TEMPLATE=".writrun/templates/pull_request_template.md"

# repo_blob_url — `https://github.com/<owner>/<repo>/blob/main/`, or the
# empty string. A body is a page and not a file in the tree, so its links
# are absolute and on the authority branch
# (product/stage-2-pull-requests/body.md).
#
# The empty string is a real answer, not a failure: this blob path shape
# is GitHub's, and composing one for a remote somewhere else would write
# a dead link that reads as live until it is clicked. An unlinked bullet
# still carries the id and the title; a wrong link carries less than
# nothing. Under-linking is the safe direction.
#
# The remote is read through an overridable name for one reason: the act
# fetches `origin main` long before it composes anything, so a suite
# wanting to present a github.com remote would have to make that fetch
# reach github.com. `url.<base>.insteadOf` cannot help — `git remote
# get-url` applies the rewrite too, so the case under test disappears.
# The override is the suite's seam and nothing else sets it, the same
# shape as WRITRUN_MIRROR_REFRESH_WAIT in rederive_labels.sh.
repo_blob_url() {
  local url owner_repo
  url="${WRITRUN_ORIGIN_URL:-}"
  [ -n "$url" ] || url=$(git remote get-url origin 2>/dev/null) || return 0
  case "$url" in
    git@github.com:*)       owner_repo=${url#git@github.com:} ;;
    ssh://git@github.com/*) owner_repo=${url#ssh://git@github.com/} ;;
    https://github.com/*)   owner_repo=${url#https://github.com/} ;;
    http://github.com/*)    owner_repo=${url#http://github.com/} ;;
    *) return 0 ;;
  esac
  # A remote may carry a trailing slash, a .git suffix, or both.
  owner_repo=${owner_repo%/}
  owner_repo=${owner_repo%.git}
  owner_repo=${owner_repo%/}
  case "$owner_repo" in
    */*/*) return 0 ;;
    */*)   ;;
    *)     return 0 ;;
  esac
  printf 'https://github.com/%s/blob/main/' "$owner_repo"
}

# spec_title <spec-file> — the spec's own first heading, without the
# `spec-NNNN — ` the file repeats there. A heading that is only the id
# yields nothing rather than a title that restates the id beside it.
spec_title() {
  local h
  h=$(sed -n 's/^# *//p' "$1" | head -n1)
  case "$h" in
    spec-[0-9]*" — "*) h=${h#*" — "} ;;
    spec-[0-9]*)       h="" ;;
  esac
  printf '%s' "$h"
}

# spec_bullets — one bullet per spec_ref entry, in that order: the id,
# the spec's title, and a link that opens the file on main.
#
# **Composition never fails a take.** Every part degrades to the one
# above it — no URL leaves id and title, no file leaves the bare id —
# because a body is what the act produces and the act is the branch
# reaching the forge with a draft on it. A take that refused over a
# heading it could not parse would trade the whole act for a bullet.
spec_bullets() {
  local s sf t line out=""
  for s in $SPECS; do
    [ -n "$s" ] || continue
    sf=$(ql_spec_file "$s")
    t=""
    [ -n "$sf" ] && t=$(spec_title "$sf")
    if [ -n "$BLOB" ] && [ -n "$sf" ]; then
      line="- [${s}](${BLOB}${sf})"
    else
      line="- ${s}"
    fi
    [ -n "$t" ] && line="${line} — ${t}"
    out="${out}${line}
"
  done
  printf '%s' "$out"
}

BLOB=$(repo_blob_url)
SPEC_BULLETS=$(spec_bullets)
[ -n "$SPEC_BULLETS" ] || SPEC_BULLETS="No spec — the task body and its doc_ref are the brief."

if [ -f "$TEMPLATE" ]; then
  # The kit ships one template for all three kinds of pull request and
  # says to keep the section that applies. This is an implementing one,
  # so the authoring and reporting sections go — headings, marker
  # comments and all.
  #
  # The bullets arrive through the environment rather than -v: they are
  # several lines and may carry a backslash, which -v would read as an
  # escape.
  BODY=$(SPEC_BULLETS="$SPEC_BULLETS" awk '
    /^<!--$/ && NR == 1 { skip = 1 }
    skip { if ($0 ~ /-->/) skip = 0; next }
    /^## Derived work$/  { drop = 1 }
    /^## Report$/        { drop = 1 }
    /^## Spec$/          { drop = 0 }
    /^## How to verify$/ { drop = 0 }
    drop { next }
    /^- \[spec-NNNN\]/ { print ENVIRON["SPEC_BULLETS"]; next }
    { print }
  ' "$TEMPLATE")
else
  # The same sections in the same order, so a project whose template is
  # missing is not handed a different contract.
  BODY="## What

## Why

<!-- writrun:begin -->

## Spec

${SPEC_BULLETS}

## How to verify

## How to test

<!-- writrun:end -->

## Notes"
fi

# The first commit's message is composed here, beside the branch, the
# title and the body — `auto_commit: false` gates the action and never
# the work, so what it holds has to be presentable in full before the
# word is given, and the message that gets presented is the one the
# commit is later written with.
FIRST_MSG=$(printf 'chore(tasks): take task-%04d' "$NUM")
[ -n "$COAUTHOR" ] && FIRST_MSG="${FIRST_MSG}

Co-Authored-By: ${COAUTHOR}"

show_composition() {
  printf 'branch: %s\n' "$BRANCH"
  printf 'title:  %s\n' "$PR_TITLE"
  printf 'first commit (empty):\n'
  printf '%s\n' "$FIRST_MSG" | sed 's/^/  | /'
  printf 'body:\n'
  printf '%s\n' "$BODY" | sed 's/^/  | /'
}

# resume_command — the rerun that finishes a half-done act, written in
# one place so the two failure paths cannot print different acts.
#
# It carries every argument that decided the branch and the act, because
# both are load-bearing: a hint without `--slug` names a *different*
# branch than the one just cut, and --resume then refuses a branch that
# "does not exist locally"; a hint without `--confirm`, on a run the
# conduct flags only let through because --confirm was given, walks back
# into the conduct gate and exits 2 having done nothing — leaving the
# pushed branch without its pull request, the one state this act must
# not leave behind.
rerun_command() {
  printf '  bash %s %s --title "%s" --slug %s%s%s\n' \
    "$0" "$TASK_ARG" "$TITLE" "$SLUG" \
    "${COAUTHOR:+ --coauthor \"${COAUTHOR}\"}" "${1:+ $1}"
}
resume_command()  { rerun_command "--resume${CONFIRM:+ --confirm}"; }
confirm_command() { rerun_command "--confirm"; }
# fresh_command — the take again, from nothing. What a refusal prints
# where the recovery is a new take rather than a resume: a refusal that
# names no runnable line leaves the operator with neither path.
fresh_command()   { rerun_command "${CONFIRM:+--confirm}"; }

# --- the branch must not already exist -------------------------------

branch_local=""
git rev-parse --verify --quiet "refs/heads/${BRANCH}" >/dev/null 2>&1 && branch_local=yes
branch_remote=""
git rev-parse --verify --quiet "refs/remotes/origin/${BRANCH}" >/dev/null 2>&1 && branch_remote=yes

if [ -n "$RESUME" ]; then
  # The one carve-out: a local branch no pull request carries is the
  # leftover of an interrupted take, and finishing it is not a second
  # take — wherever the branch got to. The push is idempotent, so how
  # far the interruption let the act get does not change what finishing
  # it costs; what would make the resume wrong is a pull request that
  # already exists, and the forge reads below are what answer that. The
  # remote-tracking ref is deliberately not consulted: it is a cache
  # saying this checkout once pushed, not that the forge holds the
  # branch now.
  [ -n "$branch_local" ] || refuse "--resume was given but ${BRANCH} does not exist locally — a take that left nothing behind is a take, not a resume"
else
  [ -z "$branch_local" ] || refuse "${BRANCH} already exists locally — resuming is not taking; --resume finishes an interrupted take"
  [ -z "$branch_remote" ] || refuse "${BRANCH} already exists on the forge — resuming is not taking"
fi

# --- the conduct flags ------------------------------------------------

if [ "$AUTO_COMMIT" != true ] || [ "$AUTO_PUSH" != true ] || [ "$AUTO_PR" != true ]; then
  if [ -z "$CONFIRM" ]; then
    # The arg list is empty by here — every argument was shifted off —
    # so it is free to hold the flags that held, in declaration order.
    set --
    [ "$AUTO_COMMIT" != true ] && set -- "$@" auto_commit
    [ "$AUTO_PUSH" != true ] && set -- "$@" auto_push
    [ "$AUTO_PR" != true ] && set -- "$@" auto_pr
    case "$#" in
      1) which="$1" ;;
      2) which="$1 and $2" ;;
      *) which="$1, $2 and $3" ;;
    esac
    echo "Composed, and waiting: ${which} is false, so nothing was committed, nothing was pushed and no pull request was opened."
    echo
    show_composition
    echo
    echo "Rerun with --confirm once the word is given, and this exact act is performed:"
    confirm_command
    exit 2
  fi
fi

# --- the act, one motion ---------------------------------------------
#
# The forge is verified *before* the branch is cut: a repository left
# with a branch nobody can push is the state this ordering avoids.

if ! command -v gh >/dev/null 2>&1; then
  echo "gh is not on PATH — the draft pull request is half of the act, so nothing was done." >&2
  exit 3
fi
if ! GH_ERR=$(gh auth status 2>&1); then
  echo "gh cannot reach the forge:" >&2
  printf '%s\n' "$GH_ERR" | head -n 3 >&2
  echo "This run committed nothing, pushed nothing and opened nothing." >&2
  exit 3
fi

# --- what the forge says --------------------------------------------
#
# The same two reads list_tasks.sh makes, and for the same reason: the
# queue file cannot see a draft opened seconds ago, nor an amendment
# still riding an open pull request. A gate weaker than the lister it
# re-checks would hand back the task the lister held.
#
# They sit **after** the conduct gate on purpose: a run the flags hold
# composes and stops, and asking the forge about work that is not about
# to happen would spend a network call — and leave a trace on someone
# else's server — for an act the adopter has not yet allowed. The word
# is what starts this, and the rerun makes both reads before it acts.

pr_lines=""
pr_source=none
if [ -n "${WRITRUN_PR_LIST:-}" ]; then
  pr_lines="$WRITRUN_PR_LIST"; pr_source=supplied
elif command -v gh >/dev/null 2>&1; then
  if pr_lines=$(gh pr list --state open --limit 200 \
        --json number,headRefName,author,title \
        --jq '.[] | "\(.number)\t\(.headRefName)\t\(.author.login)\t\(.title)"' 2>/dev/null); then
    pr_source=gh
  fi
fi

idless=""
if [ "$pr_source" != none ]; then
  while IFS="$(printf '\t')" read -r num branch author ptitle; do
    [ -n "$branch" ] || continue
    carried=$(ql_carried_of "$branch" "${ptitle:-}")
    case "$carried" in
      over-ceiling:*)
        # Skipped, never emptied: an empty set is what routes a row
        # into the amendment-candidate list below and its files probe,
        # so "carrying nothing" would hand a hostile title a paginated
        # files read and a suspended take. Somebody else's over-long
        # title must not stop this take either — its own check is red.
        echo "pull request #${num} claims ${carried#over-ceiling:} distinct tasks — over the ceiling of ${QL_CARRIED_MAX}; its row is skipped"
        continue
        ;;
    esac
    if [ -z "$carried" ]; then
      idless="${idless}${num}"$'\n'
      continue
    fi
    for c in $carried; do
      if [ "$(ql_task_num "$c")" = "$NUM" ]; then
        refuse "${ID} is already in flight on pull request #${num} (by @${author}) — resuming is not taking"
      fi
    done
  done <<EOF
$pr_lines
EOF
fi

# An open pull request carrying no task id that touches one of this
# task's specs is an amendment: the approval it rides on is in question,
# so the take waits for it rather than implementing through a suspension.
if [ -n "$idless" ] && [ -n "$SPECS" ]; then
  for num in $idless; do
    [ -n "$num" ] || continue
    files=""
    if [ -n "${WRITRUN_PR_FILES:-}" ]; then
      files=$(printf '%s\n' "$WRITRUN_PR_FILES" | awk -F'\t' -v n="$num" '$1 == n { print $2 }')
    elif [ "$pr_source" = gh ]; then
      files=$(gh api "repos/{owner}/{repo}/pulls/${num}/files" --paginate \
        --jq '.[].filename' 2>/dev/null) || continue
    fi
    for s in $SPECS; do
      [ -n "$s" ] || continue
      sn=$(printf '%s' "$s" | sed 's/^spec-0*//')
      printf '%s\n' "$files" \
        | sed -n 's|^work/specs/spec-0*\([0-9][0-9]*\).*|\1|p' \
        | grep -qx "$sn" \
        && refuse "${ID} is suspended: pull request #${num} amends ${s}; the take waits for that pull request"
    done
  done
fi

# --- what only a resume asks -----------------------------------------
#
# A resume's one guard is the forge's answer: without it the run cannot
# tell the state it recovers from the state it refuses, and opening a
# second pull request over a branch that has one is the failure this
# whole act exists to avoid. The fresh path is left as it is — its own
# local guards stand without that answer.

if [ -n "$RESUME" ]; then
  if [ "$pr_source" = none ]; then
    echo "The open pull request list went unanswered, and whether one already carries this take is the one question a resume turns on." >&2
    echo "This run pushed nothing and opened nothing. Once the forge answers, finish the act with:" >&2
    resume_command >&2
    exit 3
  fi
  # One question the fresh path never asks: which pull requests ever
  # carried this branch, in any state. It is asked branch-scoped, so it
  # answers where the repo-wide open list above cannot — past its cap of
  # 200 this take's own pull request falls off that list, and these rows
  # still hold it.
  #
  # `--head` matches a head branch's *name*, and forks share that
  # namespace: a fork's pull request on `task/0007-widget` is not this
  # take's flight, so the cross-repository rows are dropped.
  if ! head_lines=$(gh pr list --head "$BRANCH" --state all --limit 200 \
        --json number,state,headRefOid,closedAt,author,isCrossRepository \
        --jq '.[] | "\(.number)\t\(.state)\t\(.headRefOid)\t\(.closedAt)\t\(.author.login)\t\(.isCrossRepository)"' 2>/dev/null); then
    echo "Whether a pull request ever carried ${BRANCH} went unanswered, and it is the one question a resume turns on." >&2
    echo "This run pushed nothing and opened nothing. Once the forge answers, finish the act with:" >&2
    resume_command >&2
    exit 3
  fi
  BRANCH_TIP=$(git rev-parse --verify --quiet "refs/heads/${BRANCH}")
  while IFS="$(printf '\t')" read -r hnum hstate hoid hclosed hauthor hcross; do
    [ -n "$hnum" ] || continue
    [ "$hcross" = true ] && continue
    if [ "$hstate" = OPEN ]; then
      refuse "${ID} is already in flight on pull request #${hnum} (by @${hauthor}) — resuming is not taking"
    fi
    # An ended flight is finished by a fresh take, never resumed — but
    # only this branch's own flight ends this branch. Branch names are
    # deterministic, so a name an ended pull request once used is the
    # name the next take cuts again, and refusing on the name alone
    # would burn it for every take that follows.
    #
    # The commit the pull request ended on is the evidence: a branch
    # carrying it is that flight continued. Where this clone no longer
    # has that commit, the close's timestamp stands in — a tip made
    # after the flight ended was never part of it. With neither in hand
    # the run cannot tell, and refuses.
    if [ -n "$hoid" ] && [ "$hoid" != null ] \
       && git cat-file -e "${hoid}^{commit}" 2>/dev/null; then
      git merge-base --is-ancestor "$hoid" "$BRANCH_TIP" 2>/dev/null || continue
      hwhy="carries that flight's commits"
    elif [ -n "$hclosed" ] && [ "$hclosed" != null ]; then
      [ -z "$(git rev-list -1 --since="$hclosed" "$BRANCH_TIP" 2>/dev/null)" ] || continue
      hwhy="was last committed to before that flight ended"
    else
      hwhy="cannot be told apart from that flight"
    fi
    # The state is named as the forge gave it: a merged flight ended more
    # conclusively than a closed one, and calling it closed would be the
    # same kind of over-claim these sentences exist to stop. So is the
    # reason — three refusals stand on three different pieces of
    # evidence, and each says which one it has.
    #
    # The way out is printed in full, because there is no other. A resume
    # is refused here and a fresh take is refused on the branch that
    # still exists, so a refusal naming only the first of those leaves
    # the operator with nothing to run.
    hended=$(printf '%s' "$hstate" | tr 'A-Z' 'a-z')
    echo "REFUSED: pull request #${hnum} carried ${BRANCH} and is ${hended} — an ended flight is finished by a fresh take, never resumed." >&2
    echo "This checkout's ${BRANCH} ${hwhy}. The fresh take needs the name back:" >&2
    echo "  git switch main && git branch -D ${BRANCH}" >&2
    echo "  git push origin --delete ${BRANCH}   # where the forge still holds it" >&2
    fresh_command >&2
    exit 1
  done <<EOF
$head_lines
EOF
fi


if [ -z "$RESUME" ]; then
  if ! CUT_ERR=$(git switch -c "$BRANCH" origin/main 2>&1); then
    echo "git switch -c ${BRANCH} origin/main failed:" >&2
    printf '%s\n' "$CUT_ERR" | head -n 3 >&2
    exit 3
  fi
else
  if ! CUT_ERR=$(git switch "$BRANCH" 2>&1); then
    echo "git switch ${BRANCH} failed:" >&2
    printf '%s\n' "$CUT_ERR" | head -n 3 >&2
    exit 3
  fi
fi

# --- the first commit -------------------------------------------------
#
# Guarded on the range rather than on --resume, because what makes a
# second commit wrong is that one is already there: an interrupted take
# that got as far as committing is finished, not given a second marker,
# and one interrupted before it is completed here.
# A range git could not answer is not a range of nothing. Defaulting an
# unreadable count to zero would put the failure on the branch that
# *commits*, which is the second marker this guard exists to prevent —
# so it fails closed, the same posture the TYPES/SCOPES read takes.
if ! AHEAD=$(git rev-list --count origin/main..HEAD 2>&1); then
  echo "git rev-list --count origin/main..HEAD failed:" >&2
  printf '%s\n' "$AHEAD" | head -n 3 >&2
  echo "Whether ${BRANCH} already carries its first commit is the one thing" >&2
  echo "this step must not guess at. ${BRANCH} is kept local. Finish the act with:" >&2
  resume_command >&2
  exit 3
fi
if [ "$AHEAD" = 0 ]; then
  if [ "$CREDIT" = true ] && [ -z "$COAUTHOR" ]; then
    echo "No --coauthor given, so the first commit carries no trailer."
    echo "An agent taking this owes one: --coauthor \"Model Name <address>\"."
  fi
  if ! COMMIT_ERR=$(git commit --allow-empty -q -m "$FIRST_MSG" 2>&1); then
    echo "git commit --allow-empty failed:" >&2
    printf '%s\n' "$COMMIT_ERR" | head -n 3 >&2
    echo "${BRANCH} is kept local and carries nothing. Finish the act with:" >&2
    resume_command >&2
    exit 3
  fi
fi

# LC_ALL=C because the arms below read git's own words. Under any other
# locale git translates them, every pattern misses, and a failure whose
# meaning is known falls to the arm that establishes least.
if ! PUSH_ERR=$(LC_ALL=C git push -u origin "$BRANCH" 2>&1); then
  echo "git push failed:" >&2
  printf '%s\n' "$PUSH_ERR" | head -n 3 >&2
  # The sentence claims only what the failure proves. A non-fast-forward
  # is the forge answering over a branch it already holds; a ref the
  # forge received and declined — branch protection, a pre-receive hook
  # — says nothing about whether it holds the branch at all; a remote
  # that never answered moved nothing, so the branch stays where it was;
  # and anything else — a connection that dropped mid-push — proves only
  # that this push did not complete.
  case "$PUSH_ERR" in
    *'[remote rejected]'*)
      echo "The forge received this push and declined the ref, so ${BRANCH} did not move there. Settle what declined it — a protection rule, or a hook — then finish with:" >&2 ;;
    *non-fast-forward*|*'[rejected]'*)
      echo "The forge holds ${BRANCH} and refused this push over it — the divergence is real, and no force push finishes this act. Reconcile the two, then finish with:" >&2 ;;
    *'Could not read from remote repository'*|*'does not appear to be a git repository'* \
    |*'Could not resolve host'*|*'Could not resolve proxy'* \
    |*'Failed to connect to'*|*'Connection refused'*|*'Connection timed out'* \
    |*'Network is unreachable'*|*'No route to host'*|*'Operation timed out'*)
      echo "The remote never answered, so this push moved nothing and ${BRANCH} is kept local. Finish the act with:" >&2 ;;
    *)
      echo "The push did not complete, and where ${BRANCH} stands on the forge is not established. Finish the act with:" >&2 ;;
  esac
  resume_command >&2
  exit 3
fi

BODY_FILE=$(mktemp "${TMPDIR:-/tmp}/writrun-pr.XXXXXX")
printf '%s\n' "$BODY" > "$BODY_FILE"
if ! PR_ERR=$(gh pr create --draft --base main --head "$BRANCH" \
      --title "$PR_TITLE" --body-file "$BODY_FILE" 2>&1); then
  rm -f "$BODY_FILE"
  echo "gh pr create failed:" >&2
  printf '%s\n' "$PR_ERR" | head -n 3 >&2
  echo "${BRANCH} is pushed but has no pull request, which is the one state this act must not leave behind." >&2
  echo "Finish it with:" >&2
  resume_command >&2
  exit 3
fi
rm -f "$BODY_FILE"

printf '%s\n' "$PR_ERR"
echo "Took ${ID}: ${BRANCH} pushed, draft pull request open."
echo "The machinery writes 'in-progress' and taken_by from that draft — never a branch."
exit 0
