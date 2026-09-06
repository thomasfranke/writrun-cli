#!/usr/bin/env bash
# intake_report.sh — a maintainer's label turns an issue into a report.
#
# Usage: intake_report.sh <owner/repo> <issue-number>
#   Run from the repository root, on a checkout of the authority branch
#   with full history. The issue's fields travel through env, never as
#   arguments and never interpolated into anything that executes:
#
#     ISSUE_TITLE       the issue's title, the report's title-to-be
#     ISSUE_BODY        the issue's text — a stranger's, and pure data
#     ISSUE_AUTHOR      the login that opened the issue
#     ISSUE_CREATED_AT  when it was opened (RFC 3339 UTC)
#     LABEL_NAME        the label this event applied
#     BASE_REF          the authority branch (default: main)
#
# Anyone can open an issue, so an issue's arrival writes nothing — an
# intake that minted files on arrival would hand the queue's front door
# to whoever finds the repository. The gate is the label: someone with
# triage rights applying `writrun:report` is the judgement that the
# observation deserves a file, and nothing more — the route stays
# triage's, made after the file exists
# (docs/product/stage-3-github-issues/intake.md).
#
# On that label this script mints the next report id — over the same
# three views the generator reads: the directory, the git history, and
# every open pull request, through the minting stack queue_lib.sh shares
# with the generator — writes `work/reports/report-NNNN-<slug>.md` with
# `status: open`, the issue's title as its title and its text as its
# body, commits it to the authority branch with the same rebase-not-force
# pattern every queue recording uses, then retitles the issue
# `[REPORT-NNNN] <title>`, labels it `status:open`, and comments the
# file's path. From that moment the issue is the report's mirror,
# exactly as if the file had come first.
#
# Two arrivals it declines by design: a label that is not
# `writrun:report` (this event is not the gate), and a title already
# carrying a `[REPORT-` or `[TASK-` tag — an existing mirror is another
# workflow's.
#
# **The file, not the title, is what makes a re-delivered label safe.**
# The retitle is the last write, so a run that died between the push and
# the retitle left the report on the authority branch and the issue
# untagged — and a title is a stranger's to edit besides. Before minting
# anything this script looks for a report already naming this issue (the
# `Issue #N` line it writes into every report it mints); finding one, it
# re-dresses the mirror for that file instead of minting a second id for
# the same observation.
#
# **Two intakes can still race on one id** — the concurrency group in
# the workflow is per issue, so two different issues labelled together
# run together — and so can an intake against a report/ branch merging.
# The rebase alone cannot see that race: a report landing under another
# filename replays cleanly. So after every rebase the tree is re-read,
# and an id another file now claims is dropped and minted again, bounded
# at three attempts.
#
# **The body is data.** It reaches the report through `printf '%s'` of
# an environment variable — no eval, no interpolation into shell or
# YAML — so `$(...)`, backticks and a front-matter block of its own all
# arrive verbatim, as evidence claimed by the reporter and nothing else.
#
# Exit codes: 0 recorded, or nothing to do and it says why; 1 the forge
# refused a write after the file landed; 3 usage error or git failed.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions. See the
# standing rule in docs/technical/decisions/.

set -euo pipefail

REPO="${1:?usage: intake_report.sh <owner/repo> <issue-number>}"
ISSUE="${2:?usage: intake_report.sh <owner/repo> <issue-number>}"
BASE_REF="${BASE_REF:-main}"

TITLE="${ISSUE_TITLE:-}"
BODY="${ISSUE_BODY:-}"
AUTHOR="${ISSUE_AUTHOR:-}"
CREATED="${ISSUE_CREATED_AT:-}"
LABEL="${LABEL_NAME:-}"

. "$(dirname "$0")/../stage-2-pull-requests/queue_lib.sh"

if [ "$LABEL" != "writrun:report" ]; then
  echo "label '${LABEL}' is not the gate — only writrun:report mints a report."
  exit 0
fi

if [ -z "$TITLE" ]; then
  echo "the issue carries no title, and a report is named by one — nothing recorded." >&2
  exit 3
fi

case "$TITLE" in
  "[REPORT-"*|"[TASK-"*)
    echo "the title already carries a mirror tag — that issue is some file's"
    echo "mirror, and its writer is another workflow. Nothing to do."
    exit 0 ;;
esac

# --- the durable no-op guard: the queue itself --------------------------
#
# Every report this script mints opens its body with `Issue #N`, and
# that line — on the authority branch, written by this script and
# nobody else — is the key a re-delivered or retried label event is
# judged against. A hit means the recording already happened; only the
# mirror dressing below may still be owed, so that is all a second run
# does.
report_of_issue() {
  local f
  for f in work/reports/report-*.md work/reports/REPORT-*.md; do
    [ -f "$f" ] || continue
    if grep -q "^Issue #${ISSUE}[.,]" "$f" 2>/dev/null; then
      printf '%s' "$f"
      return 0
    fi
  done
  return 0
}

FILE=$(report_of_issue)
if [ -n "$FILE" ]; then
  RID=$(ql_fm_field id "$FILE")
  if [ -z "$RID" ]; then
    echo "${FILE} names issue #${ISSUE} but carries no id — fix the file by hand." >&2
    exit 3
  fi
  echo "issue #${ISSUE} is already recorded as ${FILE} — re-dressing its mirror only."
else
  # --- the id, minted over the same three views the generator reads -----
  #
  # Pinned to $REPO so the pull request numbers and their file lists are
  # one repository's answer; a scan that cannot answer completely leaves
  # the view local and says so.
  ql_forge_scan "$REPO"

  # The filename's subject: a short kebab echo of the title, at most
  # three words — readability, never identity, exactly as the generator
  # derives it when nobody chose one.
  SLUG=$(ql_slugify "$TITLE")

  # The issue's own timestamp when it is canonical; the moment of minting
  # when it is not — the front-matter check accepts exactly one spelling,
  # and a report refused at the door records nothing.
  if ! printf '%s' "$CREATED" \
    | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$'; then
    CREATED=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
  fi

  # id_claimed_elsewhere — after a rebase, does any other file claim the
  # id this run minted? The front-matter `id:` is the identity; the
  # filename answers for a file whose front matter cannot be read.
  id_claimed_elsewhere() {
    local other base
    for other in work/reports/report-*.md work/reports/REPORT-*.md; do
      [ -f "$other" ] || continue
      [ "$other" = "$FILE" ] && continue
      if [ "$(ql_fm_field id "$other")" = "$RID" ]; then return 0; fi
      base=$(basename "$other" .md | tr '[:upper:]' '[:lower:]')
      case "$base" in "$RID"|"$RID"-*) return 0 ;; esac
    done
    return 1
  }

  git config user.name  "github-actions[bot]"
  git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

  # The recording, with the same identity and the same rebase-not-force
  # pattern as every other commit the machinery makes: an addition to
  # the branch's history, never a replacement of it. Minted, committed,
  # rebased, re-checked — a rebase that brought in a sibling claiming
  # this id drops the commit and mints again over what landed.
  attempt=0
  while :; do
    attempt=$((attempt + 1))
    RID="report-$(ql_next_id work/reports report)"
    FILE="work/reports/${RID}${SLUG:+-$SLUG}.md"
    if [ -e "$FILE" ]; then
      echo "${FILE} already exists — the id scan failed to clear it." >&2
      exit 3
    fi

    mkdir -p work/reports
    {
      printf -- '---\n'
      printf 'id: %s\n' "$RID"
      printf 'status: open\n'
      printf 'task_ref: []\n'
      printf 'doc_ref: null\n'
      printf 'created: %s\n' "$CREATED"
      printf 'triaged: null\n'
      printf -- '---\n'
      printf '\n'
      printf '# %s\n' "$TITLE"
      printf '\n'
      printf 'Issue #%s%s.\n' "$ISSUE" "${AUTHOR:+, opened by @$AUTHOR}"
      if [ -n "$BODY" ]; then
        printf '\n'
        printf '%s\n' "$BODY"
      fi
    } > "$FILE"

    git add work/reports
    git commit -q \
      -m "$(bash "$(dirname "$0")/../stage-2-pull-requests/commit_subject.sh" intake)" \
      -m "Issue #${ISSUE}, labelled writrun:report."

    # A conflicting rebase stops half-applied and leaves conflict
    # markers in the very queue files this script just wrote. Abort back
    # to the recording commit before failing, so what is lost is the
    # push and never the tree's legibility.
    if ! git pull --rebase origin "$BASE_REF"; then
      git rebase --abort || true
      echo "rebase onto ${BASE_REF} failed — the recording was not pushed." >&2
      exit 3
    fi

    if id_claimed_elsewhere; then
      git reset --hard -q HEAD~1
      if [ "$attempt" -ge 3 ]; then
        echo "the id was claimed again on every attempt — giving up after ${attempt}." >&2
        exit 3
      fi
      continue
    fi

    if git push origin "HEAD:${BASE_REF}"; then break; fi
    # The push lost a race the rebase predates. Drop the commit and go
    # around: the next pull sees what beat it.
    git reset --hard -q HEAD~1
    if [ "$attempt" -ge 3 ]; then
      echo "the push was refused on every attempt — giving up after ${attempt}." >&2
      exit 3
    fi
  done

  echo "recorded ${FILE} from issue #${ISSUE}"
  ql_mint_note
fi

# The issue becomes the mirror: the tag in the title marks it as one —
# the label is the live state, and the comment names the file that is
# the authority from here. Values reach the forge as data through -f.
# The durable guard above is the file, so a run that dies here leaves a
# retry with only this dressing to redo.
NUM=$(printf '%s' "$RID" | sed -E 's/^report-0*//')
TAG=$(printf '[REPORT-%04d]' "$NUM")
gh api -X PATCH "repos/${REPO}/issues/${ISSUE}" \
  -f "title=${TAG} ${TITLE}" >/dev/null
gh api -X POST "repos/${REPO}/labels" \
  -f name="status:open" -f color="0e8a16" \
  -f description="Recorded and awaiting triage" >/dev/null 2>&1 || true
gh api -X POST "repos/${REPO}/issues/${ISSUE}/labels" \
  -f "labels[]=status:open" >/dev/null
gh api -X POST "repos/${REPO}/issues/${ISSUE}/comments" \
  -f "body=Recorded as \`${FILE}\` — the file is the authority from here; triage closes this issue." >/dev/null
echo "issue #${ISSUE} is now ${RID}'s mirror (${TAG}, status:open)"
