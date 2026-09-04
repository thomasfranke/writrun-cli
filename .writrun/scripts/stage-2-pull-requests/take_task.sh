#!/usr/bin/env bash
# take_task.sh — takes a task in one act: the eligibility re-checked, the
# branch cut from a fresh origin/main, pushed, and the draft pull request
# opened.
#
#   bash .writrun/scripts/stage-2-pull-requests/take_task.sh <task-id> \
#     --title "<summary>" [--slug words] [--resume] [--confirm]
#
# **The push and the opening are one act** (conventions/prs.md): the
# branch reaching the forge and the draft opening is the moment work
# becomes public, and it is what an adopter's `auto_push`/`auto_pr` gate.
# A branch on the forge with no pull request is the hiding place the
# taking flow exists to close, so this script never leaves one behind on
# purpose — and names the one it left behind when the forge failed.
#
# What it does not decide: which task to work (list_tasks.sh's), and what
# the title's summary should say (the agent's, validated here). It writes
# no queue file — the status line has one writer, and it is the
# machinery answering the draft this opens.
#
# Exit codes:
#   0  the branch is pushed and the draft pull request is open.
#   1  a refusal — an ineligible task, a title the declared style refuses,
#      a branch that already exists, or an unusable argument. Nothing was
#      created.
#   2  composed and waiting: `auto_push` or `auto_pr` is false, so the
#      branch, the title and the body are printed and nothing was done.
#      Rerun with --confirm to perform exactly the printed act.
#   3  the forge or git failed. Before the branch was cut the repository
#      is untouched; after it, the branch is named and --resume finishes
#      the act.
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
RESUME=""
CONFIRM=""

usage() {
  cat >&2 <<'USAGE'
usage: take_task.sh <task-id> --title "<summary>" [--slug words] [--resume] [--confirm]
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
AUTO_PUSH=$(bash "$READ_SETTING" stage_2.auto_push)
AUTO_PR=$(bash "$READ_SETTING" stage_2.auto_pr)

# The two vocabularies come from check_observance.sh's own assignment
# lines — the machine half of conventions/commits.md, and a single
# source. A title this script accepted and the door then refused would
# be worse than no validation at all.
TYPES=$(sed -n 's/^TYPES="\(.*\)"$/\1/p' "$OBSERVANCE" | head -n1)
SCOPES=$(sed -n 's/^SCOPES="\(.*\)"$/\1/p' "$OBSERVANCE" | head -n1)
if [ -z "$TYPES" ] || [ -z "$SCOPES" ]; then
  echo "Could not read TYPES/SCOPES from ${OBSERVANCE} — a title checked" >&2
  echo "against a vocabulary this script had to guess at is not checked." >&2
  exit 3
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

if ! valid_summary "$TITLE"; then
  echo "REFUSED: '${TITLE}' does not read as the declared '${STYLE}' style." >&2
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
PR_TITLE=$(printf '[TASK-%04d] %s' "$NUM" "$TITLE")

TEMPLATE=".writrun/templates/pull_request_template.md"
implements="Implements $(printf '%s' "$SPECS" | tr ' ' '\n' | sed '/^$/d' | paste -sd, - | sed 's/,/, /g')."
[ -n "$SPECS" ] || implements="No spec — the task body and its doc_ref are the brief."

if [ -f "$TEMPLATE" ]; then
  BODY=$(awk -v impl="$implements" '
    # The kit ships one template for both kinds of pull request and says
    # to keep the half that applies. This is an implementing one, so the
    # authoring half goes — heading, marker comment and all.
    /^<!--$/ && NR == 1 { skip = 1 }
    skip { if ($0 ~ /-->/) skip = 0; next }
    /^## Derived work$/ { drop = 1 }
    /^## Spec$/ { drop = 0 }
    drop { next }
    /^Implements spec-NNNN\.$/ { print impl; next }
    { print }
  ' "$TEMPLATE")
else
  BODY="## What

## Why

<!-- writrun:begin -->

## Spec

${implements}

## How to verify

<!-- writrun:end -->

## Notes"
fi

show_composition() {
  printf 'branch: %s\n' "$BRANCH"
  printf 'title:  %s\n' "$PR_TITLE"
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
resume_command() {
  printf '  bash %s %s --title "%s" --slug %s --resume%s\n' \
    "$0" "$TASK_ARG" "$TITLE" "$SLUG" "${CONFIRM:+ --confirm}"
}

# --- the branch must not already exist -------------------------------

branch_local=""
git rev-parse --verify --quiet "refs/heads/${BRANCH}" >/dev/null 2>&1 && branch_local=yes
branch_remote=""
git rev-parse --verify --quiet "refs/remotes/origin/${BRANCH}" >/dev/null 2>&1 && branch_remote=yes

if [ -n "$RESUME" ]; then
  # The one carve-out: a local branch with no upstream and no open pull
  # request is the leftover of an interrupted take, and finishing it is
  # not a second take. Everything else stays a refusal.
  [ -n "$branch_local" ] || refuse "--resume was given but ${BRANCH} does not exist locally — a take that left nothing behind is a take, not a resume"
  [ -z "$branch_remote" ] || refuse "${BRANCH} is already on the forge — what --resume finishes is a branch that never reached it"
  # A pull request for the task means the act completed and something
  # else is going on; --resume is for the half-finished one only. The
  # forge read below is what answers that, on the acting path.
else
  [ -z "$branch_local" ] || refuse "${BRANCH} already exists locally — resuming is not taking; --resume finishes an interrupted take"
  [ -z "$branch_remote" ] || refuse "${BRANCH} already exists on the forge — resuming is not taking"
fi

# --- the conduct flags ------------------------------------------------

if [ "$AUTO_PUSH" != true ] || [ "$AUTO_PR" != true ]; then
  if [ -z "$CONFIRM" ]; then
    which=""
    [ "$AUTO_PUSH" != true ] && which="auto_push"
    [ "$AUTO_PR" != true ] && which="${which:+${which} and }auto_pr"
    echo "Composed, and waiting: ${which} is false, so nothing was pushed and no pull request was opened."
    echo
    show_composition
    echo
    echo "Rerun with --confirm once the word is given, and this exact act is performed:"
    echo "  bash $0 ${TASK_ARG} --title \"${TITLE}\"${SLUG:+ --slug ${SLUG}} --confirm"
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
  echo "Nothing was done; the repository is untouched." >&2
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

if ! PUSH_ERR=$(git push -u origin "$BRANCH" 2>&1); then
  echo "git push failed:" >&2
  printf '%s\n' "$PUSH_ERR" | head -n 3 >&2
  echo "${BRANCH} is kept local. Finish the act with:" >&2
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
