#!/usr/bin/env bash
# mirror_issues.sh — reconciles the GitHub Issues mirror with a pull
# request's task and report files.
#
# Usage: mirror_issues.sh <owner/repo> <pr-number>
#   The pull request's fields arrive via the environment — never inline
#   interpolation, the same rule check_derived_work.sh follows for the PR
#   body, since a fork controls some of them:
#     PR_STATE               open | closed
#     PR_DRAFT               true | false
#     PR_MERGED              true | false
#     PR_AUTHOR_ASSOCIATION  OWNER | MEMBER | COLLABORATOR | ...
#     PR_HTML_URL            the PR's URL, linked from every mirror body
#
# **Runs in the base branch's checkout, and reads it.** The diff is the
# forge's to tell; `work/reports/` on disk is the branch the pull request
# targets, and it answers the two questions the diff cannot: where a
# modified report's front-matter block ends — a hunk that starts partway
# down the file cannot show it — and, once a pull request closes without
# merging, which of the reports it touched were already there. The pull
# request's own files are never read from disk; they are not checked out
# (writrun-issues.yml runs on pull_request_target and takes the base).
#
# **Two kinds, side by side and never folded into one another.** A task's
# mirror carries `writrun:task`, a report's carries `writrun:report`, and
# the two lists are fetched and reconciled separately — nothing ever
# converts an Issue from one kind into the other
# (docs/product/stage-3-github-issues/labels.md#the-report-mirror).
#
# The file under work/tasks/ or work/reports/ is the authority; the Issue
# is a projection of it, one direction only. This is a reconciliation, not a handler per
# event type: the desired mirror set is a function of the PR's state and
# the task files its diff adds, and every trigger re-syncs to it. That is
# what keeps the late cases honest — a task added in a later push gains
# its mirror on that push, a task a rederivation dropped loses its
# mirror, a reopened PR gets its mirrors back, and a merge creates any
# mirror still missing rather than assuming the open event already did.
#
# **Which mirrors exist is this script's question; what they are labelled
# is not** — past the open event, where `status:proposed` is the one
# state no file can hold. From the merge on, the queue holds the task and
# the label is projected from it by rederive_labels.sh, sequentially in
# the same workflow that writes the queue. Both answers used to be
# derived here, from the pull request's own patch, and the second one was
# wrong every time a merge approved the specs it carried
# (docs/technical/decisions/github-issues/, 0048 and 0060).
#
# The PR's files are read out of the API's own patch text and parsed as
# data — the PR's code is never checked out and never executed. `gh` must
# be on PATH and authenticated (GH_TOKEN in CI; a stub in the test suite).
#
# Output contract: prose on stdout, and — when $GITHUB_OUTPUT is set —
# `tasks=<id ...>` and `reports=<id ...>` appended there, naming every
# task and report whose mirror this pass minted or found to be its own.
# The projection that labels them next reads both, so the set it labels
# is the set that was really minted rather than one re-derived from a
# commit range (see below).
#
# Exit codes: 0 reconciled (including nothing to do); 3 usage error. An
# unexpected forge failure aborts non-zero via set -e.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

REPO="${1:?usage: mirror_issues.sh <owner/repo> <pr-number>}"
PR="${2:?usage: mirror_issues.sh <owner/repo> <pr-number>}"
: "${PR_STATE:?PR_STATE (open|closed) must arrive via the environment}"
: "${PR_DRAFT:?PR_DRAFT (true|false) must arrive via the environment}"
: "${PR_MERGED:?PR_MERGED (true|false) must arrive via the environment}"
: "${PR_AUTHOR_ASSOCIATION:?PR_AUTHOR_ASSOCIATION must arrive via the environment}"
: "${PR_HTML_URL:?PR_HTML_URL must arrive via the environment}"

TAB=$(printf '\t')

# macOS's stock base64 spells decode -D on older releases; feature-detect
# once rather than assume either spelling.
if printf '' | base64 -d >/dev/null 2>&1; then B64_FLAG="-d"; else B64_FLAG="-D"; fi
b64_decode() { base64 "$B64_FLAG"; }

# A draft is not in the queue yet. Only while open, though — a PR that had
# real mirrors and was later drafted-then-closed still needs its cleanup
# pass below.
if [ "$PR_STATE" = "open" ] && [ "$PR_DRAFT" = "true" ]; then
  echo "Draft PR — the mirror waits for ready_for_review."
  exit 0
fi

# The diff, as the API tells it: one row per file — status, path, the
# path it had before, patch. The jq only reshapes; every filter lives
# below, where the tests run. The patch travels base64-encoded so an
# attacker-controlled patch line can never masquerade as a row of this
# stream, and it stays last because it is the widest field.
#
# `previous_filename` is the forge's word for a rename, and a rename is a
# claim on the id it lands on — the third thing a queue file can do to an
# id, after adding and modifying. It travels as "-" where the forge sends
# nothing, never as the empty string: a tab is IFS whitespace, so an empty
# middle field would collapse on read and put the patch where the path
# goes.
FILES=$(gh api "repos/${REPO}/pulls/${PR}/files" --paginate \
  --jq '.[] | [.status, .filename, (.previous_filename // "-"), ((.patch // "") | @base64)] | @tsv')

# The commit this pull request stands at, and the branch its files land
# on. One call, two fields, and **no sixth `PR_*` name** — a name the
# caller never sets or the callee never reads is the miswiring hazard
# `technical/distribution/checks.md` exists to name, and neither half of
# it is loud. The base ref is read rather than assumed: `main` is this
# repository's answer and not every adopter's.
#
# A call that fails leaves both empty, and the link falls back a step at
# a time rather than failing the mirror.
PR_META=$(gh api "repos/${REPO}/pulls/${PR}" --jq '[.head.sha, .base.ref] | @tsv' 2>/dev/null || true)
HEAD_SHA=$(printf '%s' "$PR_META" | cut -f1)
BASE_REF=$(printf '%s' "$PR_META" | cut -f2)

# Front-matter and title, read out of the patch. Every line of an added
# file's patch is a '+' line, so stripping that column reconstructs the
# file without fetching across repositories — which a fork PR would
# otherwise require.
#
# **The front matter is read from the patch's line numbers, never by
# matching a field name wherever it appears.** A report's body quotes
# front matter — that is what evidence looks like, and this repository's
# own reports do it — so a body line at column 0 is indistinguishable
# from a real field to a reader that only greps. `status: declined` in a
# fenced block would have closed the mirror of a report still open.
#
# fm_end_of <file> — the line the front-matter block closes on, in a file
# the base-branch checkout holds. Nothing when the file has no block,
# which is also the answer for a file that is not there.
fm_end_of() {
  awk 'NR == 1 && $0 != "---" { exit } NR > 1 && /^---$/ { print NR; exit }' "$1"
}

# patch_fm <base-file> — a unified diff on stdin, the front matter of the
# file *after* the patch on stdout. Two ways to know where the block ends,
# and which one applies is decided by whether the file existed before:
#
#   - **Added**: the patch is the whole file, so line 1 of it opens the
#     block and the next `---` closes it. Nothing is read from a file
#     whose first line is not `---`; it has no front matter to read.
#   - **Modified**: the hunks show a window into the file, and the window
#     rarely starts at line 1. The bound comes from the base checkout
#     instead — the block's closing line in the file as the authority
#     branch holds it — and a line counts as front matter when its
#     position in *that* file is inside it. Added lines carry the position
#     they are inserted at, which is the position that decides them.
#
# A patch that shows neither prints nothing, and nothing is the honest
# answer: this diff says where the body went, not where the status is.
# The caller's "says nothing about its status" branch is where that
# belongs.
patch_fm() {
  local bound
  bound=0
  if [ -n "$1" ] && [ -f "$1" ]; then bound=$(fm_end_of "$1"); fi
  awk -v bound="${bound:-0}" '
    /^@@/ {
      # A second hunk in an added file means a gap, and a block still
      # open across it cannot be proven contiguous.
      if (started) exit
      o = $0; sub(/^@@[^-]*-/, "", o); sub(/[ ,].*/, "", o); oln = o + 0
      n = $0; sub(/^@@[^+]*\+/, "", n); sub(/[ ,].*/, "", n); nln = n + 0
      inhunk = 1
      next
    }
    !inhunk { next }
    /^\\/  { next }                       # "\ No newline at end of file"
    {
      ch = substr($0, 1, 1)
      line = ($0 == "" ? "" : substr($0, 2))
      if (ch == "-") { oln++; next }      # not in the file any more
      if (bound > 0) {
        if (oln <= bound && line != "---") print line
      } else if (nln == 1) {
        if (line != "---") exit
        started = 1
      } else if (started) {
        if (line == "---") exit
        print line
      }
      if (ch != "+") oln++                # context advances both files
      nln++
    }
  '
}

fm_first() {   # fm_first <front-matter> <name>
  printf '%s\n' "$1" | sed -n "s/^$2: *//p" | head -n1 | sed 's/[[:space:]]*$//'
}
first_heading() {
  printf '%s\n' "$1" | sed -n 's/^# //p' | head -n1 | sed 's/[[:space:]]*$//'
}

# base_file_of <report-id> — the report as the base-branch checkout holds
# it, which is what this workflow runs in (writrun-issues.yml checks the
# base out). Nothing when the branch does not carry the report: for an
# added file that is the normal answer, and for a modified one it means
# the bound below has to come from the patch alone.
#
# The id is matched on its number, not its text, for the reason every
# other lookup here does: `report-004` and `report-0004` are one report.
base_file_of() {
  local want f n
  want=$(num_of_id "$1")
  [ -n "$want" ] || return 0
  for f in work/reports/report-*.md; do
    [ -f "$f" ] || continue
    n=$(num_of_id "$(basename "$f" | tr '[:upper:]' '[:lower:]' \
      | sed -n 's/^\(report-[0-9][0-9]*\).*/\1/p')")
    [ -n "$n" ] || continue
    if [ "$n" -eq "$want" ] 2>/dev/null; then printf '%s' "$f"; return 0; fi
  done
  return 0
}

# base_status_of <report-file> — the status the authority branch records,
# read from the front-matter block alone for the same reason the patch is.
base_status_of() {
  [ -f "$1" ] || return 0
  awk '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub(/^status: */, "") { sub(/[[:space:]]*$/, ""); print; exit }
  ' "$1"
}

# file_fm <file> / file_heading <file> — the front matter and the first
# heading of a file the base-branch checkout holds, read whole rather
# than out of a patch. They exist for one status: **a rename carries no
# patch.** Renumbering a queue file changes only its path, so the forge
# sends an empty patch, and a reader that keeps the patch-only rule to
# the letter learns nothing about a file it can plainly see.
#
# The rule is not relaxed — it is told what a rename *is*: the same file,
# at a new id. The file is read at the path it had **before** the rename,
# because that is the path the base branch holds and this workflow checks
# out the base and never the pull request's code (writrun-issues.yml).
# Where the rename also carries a patch, the patch's fields are put
# first and win, since fm_first reads the first line that names a field.
#
# The patch-only rule keeps its whole meaning for `modified`, which is
# where it was written for: a status the patch does not carry means
# "this diff says nothing about where the report is".
file_fm() {
  [ -f "$1" ] || return 0
  awk 'NR == 1 { if ($0 != "---") exit; next } /^---$/ { exit } { print }' "$1"
}
file_heading() {
  [ -f "$1" ] || return 0
  awk '/^# / { sub(/^# /, ""); sub(/[[:space:]]*$/, ""); print; exit }' "$1"
}

# renamed_from <status> <previous-filename> — the base-branch path a
# renamed row came from, and nothing for every other row. "-" is the
# absence the tuple carries; a caller reading it as a path would look up
# a file named "-".
renamed_from() {
  [ "$1" = renamed ] || return 0
  [ -n "$2" ] && [ "$2" != "-" ] || return 0
  printf '%s' "$2"
}

# file_url <path> — where a mirror's opening sentence points, which the
# sentence says is the file itself.
#
# **The file is not on the base branch while the pull request is open**,
# which is why that sentence used to link the diff — a nine-file changed
# -files view, making the reader do the lookup the sentence exists to
# save. But the file does exist, on the head commit, from the moment the
# mirror is born. So the open window gets a permalink at that sha: it
# resolves the instant the Issue is created and keeps resolving after the
# branch is deleted. Verified against a real fork pull request — a fork's
# head commit is reachable from the base repository, so `blob/<sha>` on
# the base repository resolves it too.
#
# The merge moves it to the base ref, because a mirror outlives its pull
# request and a reader arriving a year later wants the file, not a
# snapshot of a branch that is gone. A mirror created *at* merge — the
# catch-up path — is born on the base ref and never on a sha.
#
# **This does not reverse
# `decisions/pull-requests/0067-a-body-link-points-at.md`.** That entry
# is about a *pull request body*, composed by `take_task.sh` at take time
# on an empty branch: there is no commit to point at, and no later writer
# to move the link off a revision the next push supersedes. A mirror is
# born from a commit that exists and is rewritten at merge, so both
# objections 0067 names are answered here rather than ignored.
#
# The host comes from `PR_HTML_URL` rather than a literal: the same
# forge that served the pull request serves its blobs, and an adopter on
# an Enterprise host has neither hardcoded.
file_url() {
  local ref=""
  [ "${merged:-false}" = "true" ] || ref="$HEAD_SHA"
  [ -n "$ref" ] || ref="$BASE_REF"
  if [ -n "$ref" ]; then
    printf '%s/blob/%s/%s' "${PR_HTML_URL%/pull/*}" "$ref" "$1"
  else
    # Neither ref could be read. The diff is where the sentence pointed
    # before this, so it is what a run with no answer falls back to: a
    # mirror that points somewhere beats one that fails to be written.
    printf '%s/files' "$PR_HTML_URL"
  fi
}

# mirror_line <path> — the opening sentence, the one place the two
# writers and the merge rewrite have to agree on character for character.
mirror_line() {
  printf 'Mirrors [`%s`](%s), which is the authority.' "$1" "$(file_url "$1")"
}

# A mirror's title names its task, and that is how a mirror is found —
# there is no stored number anywhere. The rule now spells the name as the
# tag a pull request title carries, `[TASK-NNNN] <task title>`
# (docs/product/stage-3-github-issues/README.md), so one search for the tag
# finds the task in the queue, in the PR, and in the mirror at once.
#
# Every lookup below still reads the `task-NNNN — ` prefix that predates
# the rule. A mirror minted before it must be *found*, because a lookup
# that only knows the new shape does not report a miss — it mints a
# second mirror for a task that already has one.

# id_of_title <title> [kind] — the id a mirror's title names, lowercased
# and at whatever width the title spells it; nothing for a title that
# names none, which is every foreign Issue the label filter let through.
# The kind is `task` unless said otherwise, because that is what every
# caller predating the report mirror means.
#
# The title is lowercased before the match rather than matched with a
# character class per letter: the result was already lowercased on the
# way out, so the two are the same answer and one of them is legible.
id_of_title() {
  local kind="${2:-task}"
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -n \
    -e "s/^\[\(${kind}-[0-9][0-9]*\)\].*/\1/p" \
    -e "s/^\(${kind}-[0-9][0-9]*\)[[:space:]].*/\1/p" \
    | head -n1
}

# num_of_id <id> — the number in `task-0006` or `report-0003`, leading
# zeros dropped. Comparisons are on the number, never the text: the id is
# the number, not how many zeroes precede it, so a mirror titled at one
# width is still found by an id spelled at another.
#
# The kind is not carried here because it never has to be: each list is
# filtered by its own label before any number is compared, so a task's 3
# and a report's 3 are never in the same list to be confused.
num_of_id() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -n \
    -e 's/^task-0*\([0-9][0-9]*\)$/\1/p' \
    -e 's/^report-0*\([0-9][0-9]*\)$/\1/p'
}

# tag_of_id <task-id> — the title's prefix: the id uppercased in brackets,
# character for character the tag its pull request title carries. The id's
# own width is kept rather than padded to four — `task-004` is that task's
# id, and a mirror is a projection of the file, never a correction of it.
tag_of_id() {
  printf '[%s]' "$(printf '%s' "$1" | tr '[:lower:]' '[:upper:]')"
}

# **This script derives no `status:` label past the open event.** It once
# read "ready" out of the spec statuses in this pull request's own patch,
# where the merge has not yet flipped them — right for what a merge
# carried, wrong for what it caused, and it overwrote the correct write
# seconds after the approve workflow made it. The merged close now has
# one owner, and the label is the queue's to project
# (rederive_labels.sh; docs/product/stage-3-github-issues/labels.md).
# What is left here is the half only the diff can answer: which tasks the
# pull request puts in the queue, and therefore which mirrors must exist.
#
# The task files the diff adds, one tab-separated record each: id,
# filename, priority, milestone, origin, title.
TASKS=""
while IFS="$TAB" read -r fstatus fname fprev fpatch; do
  case "$fstatus" in added|renamed) ;; *) continue ;; esac
  printf '%s' "$fname" | tr '[:upper:]' '[:lower:]' \
    | grep -qE '^work/tasks/task-[0-9]+(-[a-z0-9-]+)?\.md$' || continue
  patch=$(printf '%s' "$fpatch" | b64_decode)
  prev=$(renamed_from "$fstatus" "$fprev")
  # **A task file arrives added, or renamed onto its id.** An added one
  # brings its whole file as the patch, so the block opens on line 1 of
  # it and there is no base file to bound with. A renamed one brings no
  # patch at all — it is the same file at a new id — so it is read from
  # the base checkout at the path it left, with whatever the patch does
  # carry put first and winning.
  tfm=$(printf '%s\n' "$patch" | patch_fm "$prev")
  body=$(printf '%s\n' "$patch" | sed -n 's/^+//p')
  if [ -n "$prev" ]; then
    tfm=$(printf '%s\n%s\n' "$tfm" "$(file_fm "$prev")")
  fi
  # The id of a renamed task is the one its **filename** lands on. Front
  # matter that travelled unchanged still names the id the file left, and
  # the id a change claims is the one in the path — which is what the
  # uniqueness check reads and what the mirror is found by.
  if [ -n "$prev" ]; then
    tid=$(printf '%s' "$fname" | tr '[:upper:]' '[:lower:]' \
      | sed -n 's|^work/tasks/\(task-[0-9][0-9]*\).*|\1|p')
  else
    tid=$(fm_first "$tfm" id)
  fi
  ttitle=$(first_heading "$body")
  [ -n "$ttitle" ] || ttitle=$(file_heading "$prev")
  if [ -z "$tid" ] || [ -z "$ttitle" ]; then
    echo "WARNING: Could not parse ${fname}; skipping."
    continue
  fi
  # Tab is IFS whitespace, so an empty middle field would collapse on
  # read — a missing `origin` travels as "-" instead. The schema requires
  # that field and the front-matter check refuses a task without it; this
  # runs after the merge, though, and a repository that does not gate on
  # the check can land one anyway. Read it as absent rather than trip
  # over it.
  torigin=$(fm_first "$tfm" origin)
  [ -n "$torigin" ] || torigin="-"
  TASKS="${TASKS}${tid}${TAB}${fname}${TAB}$(fm_first "$tfm" priority)${TAB}$(fm_first "$tfm" milestone)${TAB}${torigin}${TAB}${ttitle}"$'\n'
done <<EOF
$FILES
EOF

# The reports the diff touches, one tab-separated record each: id,
# filename, status, title.
#
# **Modified counts here, and only here.** A task's mirror is minted from
# the files a pull request *adds*, because a task already on the
# authority branch has a mirror the machinery keeps in step from forge
# events. A report has no such events — its status is a judgement, so a
# human or an agent writes it, at every stage — which means a triage that
# lands as an edit to a file already on `main` is visible in the diff and
# nowhere else.
#
# The id comes from the **filename**, not the front matter, for the same
# reason: an edit that changes only the status line carries no `id:` in
# its patch, and the row's path is the one field always there.
#
# The status is read from the patch's front-matter block and from nowhere
# else (patch_fm). A status the patch does not carry reads as empty, and
# an empty status is "this diff says nothing about where the report is" —
# left alone rather than guessed at.
REPORTS=""
while IFS="$TAB" read -r fstatus fname fprev fpatch; do
  case "$fstatus" in added|modified|renamed) ;; *) continue ;; esac
  lname=$(printf '%s' "$fname" | tr '[:upper:]' '[:lower:]')
  printf '%s' "$lname" \
    | grep -qE '^work/reports/report-[0-9]+(-[a-z0-9-]+)?\.md$' || continue
  rid=$(printf '%s' "$lname" \
    | sed -n 's|^work/reports/\(report-[0-9][0-9]*\).*|\1|p')
  [ -n "$rid" ] || continue
  rpatch=$(printf '%s' "$fpatch" | b64_decode)
  rprev=$(renamed_from "$fstatus" "$fprev")
  # A renamed report is bounded by the file it left, not by a file at the
  # id it lands on — the base branch has never heard of that id — and
  # what the patch does not carry is read from that file whole.
  if [ -n "$rprev" ]; then
    rfm=$(printf '%s\n%s\n' "$(printf '%s\n' "$rpatch" | patch_fm "$rprev")" \
      "$(file_fm "$rprev")")
  else
    rfm=$(printf '%s\n' "$rpatch" | patch_fm "$(base_file_of "$rid")")
  fi
  rbody=$(printf '%s\n' "$rpatch" | sed -n 's/^+//p')
  # Tab is IFS whitespace, so an empty middle field collapses on read and
  # every field after it shifts one left — the title landing in the status
  # the loop below branches on. The task loop above carries "-" for the
  # same reason, and a report's status is empty far more often than a
  # task's origin: every body-only edit produces one.
  rstatus=$(fm_first "$rfm" status)
  [ -n "$rstatus" ] || rstatus="-"
  rtitle=$(first_heading "$rbody")
  [ -n "$rtitle" ] || rtitle=$(file_heading "$rprev")
  REPORTS="${REPORTS}${rid}${TAB}${fname}${TAB}${rstatus}${TAB}${rtitle}"$'\n'
done <<EOF
$FILES
EOF

# One list, fetched once, two lookups on it. Identity — the id prefix in
# the title, never a stored number — decides whether a mirror exists at
# all. Ownership — the "Introduced by" line this script writes into every
# body — decides whether this PR may reopen or retire it, and the line is
# only worth as much as the pull request it names: a mirror another *open*
# pull request owns is named in the log and never touched, while one whose
# owner is gone is adopted (docs/product/stage-3-github-issues/README.md —
# the file is the authority and the mirror is a projection of it, so a
# task with a file and no reachable mirror is the one state this
# reconciliation may not leave behind).
ISSUES=$(gh api "repos/${REPO}/issues?labels=writrun:task&state=all&per_page=100" \
  --paginate \
  --jq '.[] | [.number, .state, ((.labels // []) | map(.name) | join(",")), (.title | @base64), ((.body // "") | @base64)] | @tsv')

# The report mirrors, fetched the same way and kept apart. Unconditional,
# not lazy: the orphan sweep at the bottom needs the list even when this
# pull request's diff carries no report at all — a report added in an
# earlier push and removed in this one is exactly the mirror nobody else
# would ever close.
REPORT_ISSUES=$(gh api "repos/${REPO}/issues?labels=writrun:report&state=all&per_page=100" \
  --paginate \
  --jq '.[] | [.number, .state, ((.labels // []) | map(.name) | join(",")), (.title | @base64), ((.body // "") | @base64)] | @tsv')

# --- duplicates -----------------------------------------------------------
#
# Two runs one second apart minted two mirrors for one record — #234 and
# #235, report-0038 — and nothing downstream retired the loser: every
# lookup here stops at its first match, and the list arrives newest
# first, so the younger duplicate answered every later pass while the
# elder stood open. The workflow's concurrency group closes that window
# for one pull request's own events (decision 0073); this pass is the
# half a group cannot give — duplicates minted across pull requests, or
# standing from before the group existed, are met here, where the
# reconciler already holds the whole list.
#
# The oldest open mirror survives, because it is the one references had
# the longest to accumulate; each younger one is closed naming it.
# Closed rows are left as they lie — a closed duplicate is history, not
# a lie in the tracker — and the list is filtered afterwards so every
# lookup below reads the healed forge, not the race's leftovers.

# dup_pairs <kind> — "dup-number<TAB>survivor-number" per open duplicate
# in the list on stdin, oldest number surviving. Pure derivation: the
# writes are the caller's, so a run with no duplicates derives nothing
# and writes nothing.
dup_pairs() {
  local kind="$1" num state labels tb bb t idn
  while IFS="$TAB" read -r num state labels tb bb; do
    [ -n "$num" ] || continue
    [ "$state" = "open" ] || continue
    t=$(printf '%s' "$tb" | b64_decode)
    idn=$(num_of_id "$(id_of_title "$t" "$kind")")
    [ -n "$idn" ] || continue
    printf '%s\t%s\n' "$idn" "$num"
  done | sort -n -k1,1 -k2,2 | awk -F'\t' '
    $1 == prev { print $2 "\t" surv; next }
    { prev = $1; surv = $2 }'
}

# drop_rows <dup-numbers> — the list on stdin without those issue rows,
# so the retired duplicates stop answering lookups this same run.
drop_rows() {
  awk -F'\t' -v dups="$1" '
    BEGIN { n = split(dups, a, " "); for (i = 1; i <= n; i++) d[a[i]] = 1 }
    !($1 in d)'
}

# retire_dups <kind> <pairs> — the writes: a comment naming the
# survivor, then the close. Not planned, because a duplicate never was.
retire_dups() {
  local kind="$1" pairs="$2" dup surv note
  while IFS="$TAB" read -r dup surv; do
    [ -n "$dup" ] || continue
    note="Duplicate mirror of one record — #${surv} is the mirror. Retired by the reconciliation (report-0038)."
    [ -n "${GITHUB_RUN_ID:-}" ] && note="${note} Run ${GITHUB_RUN_ID}."
    gh api -X POST "repos/${REPO}/issues/${dup}/comments" -f body="$note" >/dev/null
    gh api -X PATCH "repos/${REPO}/issues/${dup}" \
      -f state=closed -f state_reason=not_planned >/dev/null
    echo "#${dup} retired as a duplicate ${kind} mirror — #${surv} survives"
  done <<EOF
$pairs
EOF
}

DUP_TASKS=$(printf '%s\n' "$ISSUES" | dup_pairs task)
if [ -n "$DUP_TASKS" ]; then
  retire_dups task "$DUP_TASKS"
  ISSUES=$(printf '%s\n' "$ISSUES" | drop_rows "$(printf '%s\n' "$DUP_TASKS" | cut -f1 | tr '\n' ' ')")
fi
DUP_REPORTS=$(printf '%s\n' "$REPORT_ISSUES" | dup_pairs report)
if [ -n "$DUP_REPORTS" ]; then
  retire_dups report "$DUP_REPORTS"
  REPORT_ISSUES=$(printf '%s\n' "$REPORT_ISSUES" | drop_rows "$(printf '%s\n' "$DUP_REPORTS" | cut -f1 | tr '\n' ' ')")
fi

issue_row_of() {   # issue_row_of <task-id> — "number<TAB>state<TAB>labels<TAB>body-b64"
  local num state labels tb bb t tn want
  want=$(num_of_id "$1")
  [ -n "$want" ] || return 0
  while IFS="$TAB" read -r num state labels tb bb; do
    [ -n "$num" ] || continue
    t=$(printf '%s' "$tb" | b64_decode)
    tn=$(num_of_id "$(id_of_title "$t")")
    [ -n "$tn" ] || continue
    if [ "$tn" -eq "$want" ] 2>/dev/null; then
      # Labels travel with the row because one of them must survive every
      # rewrite below: `origin:` is a fact about the task's birth, so a
      # relabelling pass re-states it rather than dropping it
      # (docs/product/stage-3-github-issues/labels.md).
      printf '%s\t%s\t%s\t%s\n' "$num" "$state" "$labels" "$bb"; return 0
    fi
  done <<EOF
$ISSUES
EOF
  return 0
}

# The title travels out as a fifth field, base64-encoded like the body
# beside it. It is already fetched, matched on and dropped here; carrying
# it costs no forge call, and it is the one thing that can tell a report
# coming back from an *id* coming back (spec-0080). Appended, never
# inserted: the caller's four reads are unchanged.
report_row_of() {   # report_row_of <report-id> — same row shape, report list
  local num state labels tb bb t tn want
  want=$(num_of_id "$1")
  [ -n "$want" ] || return 0
  while IFS="$TAB" read -r num state labels tb bb; do
    [ -n "$num" ] || continue
    t=$(printf '%s' "$tb" | b64_decode)
    tn=$(num_of_id "$(id_of_title "$t" report)")
    [ -n "$tn" ] || continue
    if [ "$tn" -eq "$want" ] 2>/dev/null; then
      printf '%s\t%s\t%s\t%s\t%s\n' "$num" "$state" "$labels" "$bb" "$tb"
      return 0
    fi
  done <<EOF
$REPORT_ISSUES
EOF
  return 0
}

OWN_LINE="| Introduced by | #${PR} |"
# `|` is literal in a basic regex, so the line matches as written.
OWN_RE='^| Introduced by | #[0-9][0-9]* |'
is_mine() {   # is_mine <body-b64>
  printf '%s' "$1" | b64_decode | grep -qF "$OWN_LINE"
}

# owner_of <body-b64> — the pull request number the mirror's ownership
# line names. Nothing when the body carries no such line, which is the
# same answer as "nobody": a line this script did not write is a line
# nobody is working behind.
owner_of() {
  printf '%s' "$1" | b64_decode \
    | sed -n 's/^| Introduced by | #\([0-9][0-9]*\) |.*/\1/p' | head -n1
}

# pr_is_open <number> — does the forge still call that pull request open?
# **This is the whole ownership question.** A mirror belongs to the pull
# request that introduced it only while that pull request is live; once it
# is closed or merged, nobody is working behind the line it left, and a
# refusal to touch the mirror only means the task never gets one. A number
# the forge does not know answers the same way, for the same reason.
pr_is_open() {
  local st
  st=$(gh api "repos/${REPO}/pulls/${1}" --jq '.state' 2>/dev/null) || return 1
  [ "$st" = "open" ]
}

# adopt_mirror <issue> <body-b64> — rewrite the ownership line to this
# pull request. Only the line: the body is the mirror's, and adopting is
# taking responsibility for it, not rewriting what it says.
adopt_mirror() {
  local body
  body=$(printf '%s' "$2" | b64_decode)
  if printf '%s\n' "$body" | grep -q "$OWN_RE"; then
    body=$(printf '%s\n' "$body" | sed "s/^| Introduced by | #[0-9][0-9]* |.*/${OWN_LINE}/")
  else
    # A mirror with no line yet is intake-born: its body is the
    # reporter's own prose, which the intake never rewrites. The line
    # goes on as a paragraph of its own — glued to their last line it
    # renders as a broken table row inside their text.
    body="${body}"$'\n\n'"${OWN_LINE}"
  fi
  gh api -X PATCH "repos/${REPO}/issues/${1}" -f "body=${body}" >/dev/null
  # Left where the relink below can read it: two writers on one body in
  # one pass, and the second must not PATCH the first one away.
  ADOPTED_BODY="$body"
}

ADOPTED_BODY=""

# relink_mirror <issue> <body> <path> — at merge, move the opening
# sentence's link off the head commit and onto the base ref, where the
# file now lives.
#
# **The first line only, and only when it is the sentence this script
# writes.** A body somebody edited by hand is left exactly as it is: the
# rewrite is here to keep a link true, not to reclaim a maintainer's
# text. A line already saying what it should is not written either — the
# forge would record an edit that changed nothing.
relink_mirror() {
  local body first want
  body="$2"
  first=${body%%$'\n'*}
  # The shape this script writes, and nothing else: an opening `Mirrors [`
  # and a closing `), which is the authority.` with a link between them.
  # The path is not matched — a rename moved it, and the whole point of
  # the rewrite is to land on wherever the file now is.
  case "$first" in
    'Mirrors ['*'), which is the authority.') ;;
    *) return 0 ;;
  esac
  want=$(mirror_line "$3")
  [ "$first" = "$want" ] && return 0
  body=$(printf '%s\n' "$body" | sed "1s|.*|${want}|")
  gh api -X PATCH "repos/${REPO}/issues/${1}" -f "body=${body}" >/dev/null
  return 0
}

# clear_status <issue> <labels-csv> — a retired mirror keeps every label
# except its place in the pipeline, for the same reason a completed one
# does (docs/product/stage-3-github-issues/labels.md): the close and its
# reason are the terminal state, and a `status:` label left on top of them
# contradicts it.
clear_status() {
  local kept l args
  kept=$(printf '%s\n' "$2" | tr ',' '\n' | grep -v '^status:' | sed '/^$/d' || true)
  args=()
  while IFS= read -r l; do
    [ -n "$l" ] || continue
    args+=(-f "labels[]=$l")
  done <<EOF
$kept
EOF
  if [ "${#args[@]}" -eq 0 ]; then
    gh api -X DELETE "repos/${REPO}/issues/${1}/labels" >/dev/null
    return 0
  fi
  gh api -X PUT "repos/${REPO}/issues/${1}/labels" "${args[@]}" >/dev/null
}

# put_report_labels <issue> <labels-csv> <status-label> — the mirror's
# place in the pipeline replaced, and everything else it wears kept.
#
# **Kept is the whole point.** A set-replacing PUT is the only call the
# forge offers that can remove a label, and one written from the kind and
# the new status alone deletes whatever a reviewer put there — on every
# push of every open pull request, silently. clear_status keeps them on
# the way out; a mirror still open has no weaker claim to them.
put_report_labels() {
  local kept l args
  kept=$(printf '%s\n' "$2" | tr ',' '\n' \
    | grep -v '^status:' | grep -v '^writrun:report$' | sed '/^$/d' || true)
  args=(-f "labels[]=writrun:report" -f "labels[]=$3")
  while IFS= read -r l; do
    [ -n "$l" ] || continue
    args+=(-f "labels[]=$l")
  done <<EOF
$kept
EOF
  gh api -X PUT "repos/${REPO}/issues/${1}/labels" "${args[@]}" >/dev/null
}

ensure_label() {   # ensure_label <name> <color> <description>
  local out
  if ! out=$(gh api -X POST "repos/${REPO}/labels" \
      -f "name=$1" -f "color=$2" -f "description=$3" 2>&1); then
    # 422 = already exists; anything else is a real failure.
    printf '%s\n' "$out" | grep -q "HTTP 422" \
      || { printf '%s\n' "$out" >&2; exit 1; }
  fi
}

# The origin label — `origin:rule` or `origin:report`, projecting the
# task's stored `origin`. Unlike `status:` it is never changed and never
# removed: origin is a fact about how the task came to exist, so it stays
# on the mirror through every state, closed included
# (docs/product/stage-3-github-issues/labels.md).
#
# origin_label <file-origin> <labels-csv> — the label to carry. A label
# the mirror already wears wins over the file, because the field is
# written once and a disagreement means this diff is the stale one; a
# task that arrives without the field — one the front-matter check should
# have refused — leaves the label empty rather than guessing, and the next
# recording commit adds it from the queue.
origin_label() {
  local worn
  worn=$(printf '%s\n' "$2" | tr ',' '\n' | grep '^origin:' | head -n1 || true)
  if [ -n "$worn" ]; then printf '%s' "$worn"; return 0; fi
  case "$1" in
    rule)   printf 'origin:rule' ;;
    report) printf 'origin:report' ;;
  esac
}

# ensure_origin_label <label> — created on first use like every other,
# in the vocabulary an Issues reader already knows: GitHub's stock bug
# red for a report, its documentation blue for a rule.
ensure_origin_label() {
  case "$1" in
    origin:rule)   ensure_label "origin:rule" "0075ca" "Derived from an authored rule" ;;
    origin:report) ensure_label "origin:report" "d73a4a" "Born from a report of work an existing rule authorizes" ;;
  esac
}

# put_status_labels <issue-number> <status-label> <origin-label> — the
# mirror's whole label set, rewritten. Every rewrite replaces the set, so
# the origin label is re-stated in each of them: leaving it out would be
# a removal, and this one is never removed.
#
# Only the open path calls it. `status:proposed` is the one label the
# queue cannot project, because the file is not on the authority branch
# yet — every other one is written from the queue, after the merge.
#
# Creating the label in the repository happens here, at the write, and
# not where the label is computed. The paths that decide *not* to touch a
# mirror — somebody else's and still open, or a pull request closed
# without merging — would otherwise leave a label behind in a repository
# where nothing wears it.
put_status_labels() {
  local args=()
  if [ -n "$3" ]; then
    ensure_origin_label "$3"
    args=(-f "labels[]=$3")
  fi
  gh api -X PUT "repos/${REPO}/issues/${1}/labels" \
    -f "labels[]=writrun:task" -f "labels[]=${2}" \
    ${args[@]+"${args[@]}"} >/dev/null
}

open=false;   [ "$PR_STATE" = "open" ]  && open=true
merged=false; [ "$PR_MERGED" = "true" ] && merged=true

# Mirrors are created at open only for an author the forge recognizes —
# the same trio every other authority check in this repository uses.
# Anyone else's tasks get their mirror at merge, when the queue really
# gains them: deferred, never denied — and a drive-by PR cannot spray
# Issues.
authorized=false
case "$PR_AUTHOR_ASSOCIATION" in
  OWNER|MEMBER|COLLABORATOR) authorized=true ;;
esac

# Only what this script still writes. `status:backlog` and `status:ready`
# are the projection's, created where they are written — a label declared
# here and never worn is one more thing to keep in sync for nothing.
if [ "$open" = "true" ]; then
  ensure_label "writrun:task" "1d76db" "Mirrors a work/tasks/ entry"
  ensure_label "status:proposed" "ededed" "A pull request proposes this task; it is not in the queue yet"
fi

# Every task the diff adds gets a mirror in the right state.
LIVE_NUMS=""

# And what this pass answered, as ids. Which mirrors exist is derived
# here from the pull request's files; which get labelled is derived by
# the caller from a commit range, and the two part on a rebase merge —
# `merge_commit_sha` is only the last rebased commit, so a task file
# added in an earlier one is minted and falls outside the range. Minted
# and never labelled is the one outcome no later event corrects, so the
# pass that minted reports what it minted
# (docs/technical/decisions/github-issues/, 0060).
#
# A mirror this pass refused to touch — another open pull request's — is
# left out: refusing it and then labelling it anyway would be the same
# defect at one remove.
MIRRORED=""
note_mirrored() {
  case " $MIRRORED " in *" $1 "*) ;; *) MIRRORED="${MIRRORED:+$MIRRORED }$1" ;; esac
}
while IFS="$TAB" read -r tid fname priority milestone torigin ttitle; do
  [ -n "$tid" ] || continue
  [ "$torigin" = "-" ] && torigin=""
  LIVE_NUMS="${LIVE_NUMS} $(num_of_id "$tid")"

  # Asked before the lookup, so it covers every write this loop makes and
  # not only the creating one — a mirror that already exists is adopted
  # and relabelled further down, and an unrecognized author may do neither
  # (the report loop below carries the same gate, and the same reason).
  if [ "$open" = "true" ] && [ "$authorized" != "true" ]; then
    echo "${tid}: author lacks authority — mirror deferred to merge."
    continue
  fi

  row=$(issue_row_of "$tid")

  if [ -z "$row" ]; then
    # Closed unmerged: the queue never gained this task, so no mirror is
    # owed. Open or merged: create it — on merge this is the catch-up for
    # a task whose earlier events were missed, and it is born ready.
    if [ "$open" != "true" ] && [ "$merged" != "true" ]; then continue; fi
    # An open pull request only *proposes* the task — it may still close
    # unmerged, and the mirror retires with it, so the queue does not
    # hold it yet, and no reader but this one can say so. A merged one
    # puts the task in the queue, and from there the label is derived
    # from the file: minted bare here, labelled by the projection that
    # runs next (docs/product/stage-3-github-issues/labels.md).
    status_args=()
    if [ "$open" = "true" ]; then
      status_args=(-f "labels[]=status:proposed")
    fi
    body=$(printf '%s\n' \
      "$(mirror_line "$fname")" \
      "Edits made here are **not** written back to the file." \
      "" \
      "| | |" \
      "|---|---|" \
      "| Priority | \`${priority}\` |" \
      "| Milestone | \`${milestone}\` |" \
      "${OWN_LINE}" \
      "" \
      "Becomes ready for development when #${PR} merges and every" \
      "spec in its \`spec_ref\` is \`approved\`.")
    olbl=$(origin_label "$torigin" "")
    olabel_args=()
    if [ -n "$olbl" ]; then
      ensure_origin_label "$olbl"
      olabel_args=(-f "labels[]=${olbl}")
    fi
    gh api -X POST "repos/${REPO}/issues" \
      -f "title=$(tag_of_id "$tid") ${ttitle}" \
      -f "labels[]=writrun:task" \
      ${status_args[@]+"${status_args[@]}"} \
      ${olabel_args[@]+"${olabel_args[@]}"} \
      -f "body=${body}" >/dev/null
    if [ "$merged" = "true" ]; then
      echo "Created issue for ${tid} — its label is the projection's"
    else
      echo "Created issue for ${tid}"
    fi
    note_mirrored "$tid"
    continue
  fi

  num=$(printf '%s' "$row" | cut -f1)
  istate=$(printf '%s' "$row" | cut -f2)
  ilabels=$(printf '%s' "$row" | cut -f3)
  ibody=$(printf '%s' "$row" | cut -f4)

  # The origin label the open path's rewrite carries, worked out here
  # where the mirror's worn labels are in hand. Nothing is created in the
  # repository yet: that is put_status_labels' job, at the write. The
  # merged path writes no labels at all, so a mirror missing its
  # `origin:` gains it from the projection instead.
  olbl=$(origin_label "$torigin" "$ilabels")

  # Three answers to "whose mirror is this", not two. Mine: proceed.
  # Somebody's, and that somebody is still open: refuse, exactly as
  # before — two live pull requests must never fight over one mirror.
  # Nobody's — the introducing pull request closed, merged, or never
  # existed: adopt it, because refusing leaves the task with no mirror at
  # all and nothing ever creates one.
  adopted=false
  ADOPTED_BODY=""
  if ! is_mine "$ibody"; then
    owner=$(owner_of "$ibody")
    if [ -n "$owner" ] && pr_is_open "$owner"; then
      echo "WARNING: ${tid} is mirrored by #${owner}, which is still open — not touching it."
      continue
    fi
    adopt_mirror "$num" "$ibody"
    adopted=true
    if [ -n "$owner" ]; then
      echo "${tid}: adopted stale mirror #${num} — #${owner} is no longer open."
    else
      echo "${tid}: adopted unowned mirror #${num} — no pull request introduced it."
    fi
  fi

  # The merge is when the file stops being a proposal and becomes what
  # the base branch holds, so it is when the mirror's link has to move
  # there. A mirror this pass just adopted is relinked from the body the
  # adoption wrote, never from the one it replaced.
  if [ "$merged" = "true" ]; then
    relink_mirror "$num" \
      "${ADOPTED_BODY:-$(printf '%s' "$ibody" | b64_decode)}" "$fname"
  fi

  if [ "$open" = "true" ]; then
    # A reopened PR finds its mirrors closed as orphans; they are not
    # orphans any more. An adopted mirror gets the same treatment
    # whatever state it was in: its labels were the old owner's, and
    # this pass is what re-derives them.
    if [ "$istate" = "closed" ] || [ "$adopted" = "true" ]; then
      # Reopened means open, and open means proposed — the task is back
      # to being offered, not back in the queue.
      gh api -X PATCH "repos/${REPO}/issues/${num}" -f state=open >/dev/null
      put_status_labels "$num" "status:proposed" "$olbl"
      [ "$adopted" = "true" ] || echo "${tid} reopened with #${PR}"
    else
      echo "${tid} already mirrored; nothing to do."
    fi
    note_mirrored "$tid"
    continue
  fi

  if [ "$merged" = "true" ]; then
    # A mirror adopted while closed is reopened here — the projection
    # writes a label next, and a closed issue wearing one says where the
    # task is twice and disagrees with itself.
    if [ "$adopted" = "true" ] && [ "$istate" = "closed" ]; then
      gh api -X PATCH "repos/${REPO}/issues/${num}" -f state=open >/dev/null
    fi
    # And no label. The merge is the moment the file became the truth,
    # so the label is the queue's to project — this pass has answered
    # the only question that was its own, which is whether the mirror
    # exists at all.
    echo "${tid} is in the queue; its label is the projection's"
    note_mirrored "$tid"
  fi
done <<EOF
$TASKS
EOF

# ------------------------------------------------------------ the reports
#
# The same reconciliation, one kind over, and deliberately its own loop:
# a report carries no `origin:` label, its two live labels are not a
# task's, and the five ways triage ends it collapse into two closes that
# no task status maps onto. Folding the two would have meant a branch per
# difference inside one body that agreed with neither.

# close_reason_of <report-status> — the reason triage's end implies, or
# nothing while the report is still open. Four ends were acted on and
# one was not, which is the whole distinction the close carries; the
# file says which of the four, and a `route:` label saying it again
# would be one more thing to keep true
# (docs/product/stage-3-github-issues/labels.md#the-report-mirror).
close_reason_of() {
  case "$1" in
    tracked|authored|fixed|routed) printf 'completed' ;;
    declined)                      printf 'not_planned' ;;
  esac
}

# ensure_report_status_label <label> — the label's colour and description
# live here, once, rather than in a conditional at each of the three
# places that create it. The two answers are the two states a live report
# mirror can be in, and nothing else may be passed.
ensure_report_status_label() {
  case "$1" in
    status:proposed)
      ensure_label "status:proposed" "ededed" \
        "A pull request proposes this report; it is not on the authority branch yet" ;;
    status:open)
      ensure_label "status:open" "0e8a16" "Recorded and awaiting triage" ;;
  esac
}

LIVE_REPORT_NUMS=""
MIRRORED_REPORTS=""
note_report() {
  case " $MIRRORED_REPORTS " in
    *" $1 "*) ;;
    *) MIRRORED_REPORTS="${MIRRORED_REPORTS:+$MIRRORED_REPORTS }$1" ;;
  esac
}

if [ -n "$REPORTS" ] && [ "$open" = "true" ]; then
  ensure_label "writrun:report" "5319e7" "Mirrors a work/reports/ entry"
fi

while IFS="$TAB" read -r rid fname rstatus rtitle; do
  [ -n "$rid" ] || continue
  LIVE_REPORT_NUMS="${LIVE_REPORT_NUMS} $(num_of_id "$rid")"

  # What the diff says about where this report is. An edit that touched
  # neither the status line nor the file's creation says nothing, and
  # nothing is what this pass then does — the mirror already reflects
  # whatever the last pass that could read it wrote.
  case "$rstatus" in
    open|tracked|authored|fixed|declined|routed) ;;
    *)
      echo "${rid}: this diff says nothing about its status — leaving its mirror alone."
      continue ;;
  esac
  want_close=$(close_reason_of "$rstatus")
  # `status:proposed` while a pull request only offers the report;
  # `status:open` once it is on the authority branch and really waiting
  # for someone to triage it. The second is the state the mirror exists
  # for — a report nobody is prompted to read is a report that rots.
  want_label=""
  if [ -z "$want_close" ]; then
    if [ "$open" = "true" ]; then want_label="status:proposed"; else want_label="status:open"; fi
  fi

  # **Authority is asked before the mirror is looked up, not after.** The
  # gate used to sit inside the create path, where it reads as "a drive-by
  # pull request cannot spray Issues" — but the write it was guarding is
  # not the only one this loop makes. A report already on the authority
  # branch has a mirror, so an unauthorized patch claiming `status:
  # declined` skipped the create path entirely and went on to adopt that
  # mirror and close it: the project's own report, retired by someone the
  # forge does not recognize. Deferred to merge, for every write, is what
  # the rule always meant.
  if [ "$open" = "true" ] && [ "$authorized" != "true" ]; then
    echo "${rid}: author lacks authority — mirror deferred to merge."
    continue
  fi

  row=$(report_row_of "$rid")

  if [ -z "$row" ]; then
    if [ "$open" != "true" ] && [ "$merged" != "true" ]; then continue; fi
    if [ -z "$rtitle" ]; then
      echo "WARNING: ${fname} carries no title in this diff; skipping."
      continue
    fi
    [ "$open" = "true" ] || ensure_label "writrun:report" "5319e7" "Mirrors a work/reports/ entry"
    body=$(printf '%s\n' \
      "$(mirror_line "$fname")" \
      "Edits made here are **not** written back to the file." \
      "" \
      "${OWN_LINE}" \
      "" \
      "A report records what was observed. It is never worked — it is" \
      "triaged, and triage closes this Issue.")
    label_args=()
    if [ -n "$want_label" ]; then
      ensure_report_status_label "$want_label"
      label_args=(-f "labels[]=${want_label}")
    fi
    inum=$(gh api -X POST "repos/${REPO}/issues" \
      -f "title=$(tag_of_id "$rid") ${rtitle}" \
      -f "labels[]=writrun:report" \
      ${label_args[@]+"${label_args[@]}"} \
      -f "body=${body}" --jq '.number')
    if [ -n "$want_close" ]; then
      # **Born closed.** Recording rides any change, so a report may
      # arrive already triaged — the ordinary case, not an edge. It is
      # created carrying no `status:` label and closed in the same pass,
      # so it never stands as an open item asking somebody to read it.
      # Two calls because the forge has no third: creating an Issue takes
      # no state.
      gh api -X PATCH "repos/${REPO}/issues/${inum}" \
        -f state=closed -f "state_reason=${want_close}" >/dev/null
      echo "Created issue for ${rid}, closed ${want_close} — it arrived triaged"
    else
      echo "Created issue for ${rid} (${want_label})"
    fi
    note_report "$rid"
    continue
  fi

  num=$(printf '%s' "$row" | cut -f1)
  istate=$(printf '%s' "$row" | cut -f2)
  ilabels=$(printf '%s' "$row" | cut -f3)
  ibody=$(printf '%s' "$row" | cut -f4)
  ititle=$(printf '%s' "$row" | cut -f5 | b64_decode)

  # Whose mirror this is — the same three answers a task's gets, and for
  # the same reason: two live pull requests must never fight over one
  # mirror, and refusing an unowned one leaves a report with none.
  #
  # The state the mirror was adopted *in* travels with the message. A
  # closed mirror adopted is half the evidence that an id has come back,
  # and the line that omitted it said nothing a reader could act on.
  adopted=""
  ADOPTED_BODY=""
  if ! is_mine "$ibody"; then
    owner=$(owner_of "$ibody")
    if [ -n "$owner" ] && pr_is_open "$owner"; then
      echo "WARNING: ${rid} is mirrored by #${owner}, which is still open — not touching it."
      continue
    fi
    adopt_mirror "$num" "$ibody"
    adopted=yes
    if [ -n "$owner" ]; then
      echo "${rid}: adopted stale mirror #${num}, ${istate} — #${owner} is no longer open."
    else
      echo "${rid}: adopted unowned mirror #${num}, ${istate} — no pull request introduced it."
    fi
  fi

  # The same move the task loop makes, and for the same reason: at merge
  # the report is on the base branch, and that is where its mirror points.
  if [ "$merged" = "true" ]; then
    relink_mirror "$num" \
      "${ADOPTED_BODY:-$(printf '%s' "$ibody" | b64_decode)}" "$fname"
  fi

  # Closed without a merge: what this pull request's diff said about the
  # report died with it, so nothing here is written from the diff. The
  # reconciliation at the bottom answers these mirrors instead, from the
  # base branch — which is the only reader left that can tell a report
  # that never landed from one that was already there.
  if [ "$open" != "true" ] && [ "$merged" != "true" ]; then continue; fi

  if [ -n "$want_close" ]; then
    if [ "$istate" = "open" ]; then
      clear_status "$num" "$ilabels"
      gh api -X PATCH "repos/${REPO}/issues/${num}" \
        -f state=closed -f "state_reason=${want_close}" >/dev/null
      echo "${rid} triaged — mirror #${num} closed as ${want_close}"
    else
      echo "${rid} is triaged and its mirror is already closed; nothing to do."
    fi
    note_report "$rid"
    continue
  fi

  # Still open. **This is the path the "triaged while still proposed"
  # case runs the other way through**, and the reason this loop updates
  # an existing mirror rather than only creating missing ones: a report
  # recorded `open` in one commit and triaged in a later one of the same
  # pull request has a mirror already labelled, and no other reader can
  # reach it — project_pr_tasks.sh learns its ids from the head branch
  # name and the title's [TASK-NNNN] tags, and a report has neither by
  # design.
  # A mirror already open and already wearing the label this pass would
  # write is a mirror this pass has nothing to say about — and saying it
  # anyway costs a label create and a set-replacing PUT on every push of
  # every open pull request that carries a report. The task loop above
  # reaches the same conclusion the same way.
  if [ "$istate" = "open" ]; then
    case ",${ilabels}," in
      *",${want_label},"*)
        echo "${rid} already mirrored ${want_label}; nothing to do."
        note_report "$rid"
        continue ;;
    esac
  fi

  ensure_report_status_label "$want_label"
  if [ "$istate" = "closed" ]; then
    # **A reopen is ordinary; a reopen onto a different title is an id
    # coming back.** Two reopens are legitimate — a pull request that
    # closed unmerged and was reopened, and a report moved to a terminal
    # status and back inside one pull request's life — and on both the
    # mirror is a projection of the same file, so its title still names
    # the same finding. When the *id* has come back the title describes
    # something else, and that mismatch is exactly what makes the Issue
    # in front of a maintainer wrong.
    #
    # So the title is the discriminator, and it is also the thing worth
    # printing. Only a mirror this pass adopted is compared: one this
    # pull request already owns is one it created, and its title moving
    # is the pull request editing its own report. A diff carrying no
    # title compares nothing — an absent title is not a differing one.
    #
    # The pass names it and projects it anyway. This script is the
    # best-effort write half (decisions/pull-requests/0010), and a
    # collision named and projected beats one neither named nor
    # projected — a report with no mirror at all is the one state this
    # reconciliation may not leave behind.
    if [ -n "$adopted" ] && [ -n "$rtitle" ]; then
      mtitle=$(printf '%s' "$ititle" | sed 's/^\[[Rr][Ee][Pp][Oo][Rr][Tt]-[0-9][0-9]*\] *//')
      mtitle=$(printf '%s' "$mtitle" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
      dtitle=$(printf '%s' "$rtitle" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
      if [ "$mtitle" != "$dtitle" ]; then
        echo "WARNING: ${rid} reopens closed mirror #${num}, which names a different finding."
        if [ -n "$owner" ]; then
          echo "  #${num} was introduced by #${owner} and titled: ${mtitle}"
        else
          echo "  #${num} was introduced by no pull request and titled: ${mtitle}"
        fi
        echo "  this diff carries: ${dtitle}"
        echo "  An id is never reused — if these are two findings, one of them holds the wrong number."
      fi
    fi
    gh api -X PATCH "repos/${REPO}/issues/${num}" -f state=open >/dev/null
  fi
  put_report_labels "$num" "$ilabels" "$want_label"
  echo "${rid} → ${want_label}"
  note_report "$rid"
done <<EOF
$REPORTS
EOF

# Orphans: a mirror this PR introduced whose task is no longer in the
# diff (a rederivation dropped the file), and every mirror of a PR closed
# without merging.
closed_unmerged=false
[ "$open" != "true" ] && [ "$merged" != "true" ] && closed_unmerged=true

while IFS="$TAB" read -r num istate labels tb bb; do
  [ -n "$num" ] || continue
  [ "$istate" = "open" ] || continue
  is_mine "$bb" || continue
  oid=$(id_of_title "$(printf '%s' "$tb" | b64_decode)")
  [ -n "$oid" ] || continue
  if [ "$closed_unmerged" != "true" ]; then
    case " $LIVE_NUMS " in *" $(num_of_id "$oid") "*) continue ;; esac
  fi
  clear_status "$num" "$labels"
  gh api -X PATCH "repos/${REPO}/issues/${num}" \
    -f state=closed -f state_reason=not_planned >/dev/null
  if [ "$closed_unmerged" = "true" ]; then
    echo "${oid} closed — #${PR} was not merged"
  else
    echo "${oid} closed — its task left the diff"
  fi
done <<EOF
$ISSUES
EOF

# The report mirrors this pull request owns and the diff no longer
# proposes, plus every one of them when the pull request closed unmerged.
#
# **The two halves ask different questions, and only the first is a
# sweep.** While the pull request is live, a mirror it owns that has left
# the diff is an orphan and retires; a mirror closed by triage is not an
# orphan, it is finished, so `istate = open` filters those out.
#
# When the pull request closed unmerged there is no diff left to be in.
# Everything it wrote about a report was provisional — a mirror closed
# because *this branch* triaged the report, a mirror adopted from an
# on-branch report, a mirror minted for a report that never landed — and
# the only authority still standing is the base branch this workflow is
# checked out at. So each owned mirror is reconciled against the file
# there, closed ones included: filtering on `istate = open` was what left
# a report living on `main` with a mirror closed by a branch nobody
# merged, and no later event ever named that report again.
while IFS="$TAB" read -r num istate labels tb bb; do
  [ -n "$num" ] || continue
  is_mine "$bb" || continue
  oid=$(id_of_title "$(printf '%s' "$tb" | b64_decode)" report)
  [ -n "$oid" ] || continue

  if [ "$closed_unmerged" != "true" ]; then
    [ "$istate" = "open" ] || continue
    case " $LIVE_REPORT_NUMS " in *" $(num_of_id "$oid") "*) continue ;; esac
    clear_status "$num" "$labels"
    gh api -X PATCH "repos/${REPO}/issues/${num}" \
      -f state=closed -f state_reason=not_planned >/dev/null
    echo "${oid} closed — its report left the diff"
    continue
  fi

  bstatus=""
  bfile=$(base_file_of "$oid")
  [ -n "$bfile" ] && bstatus=$(base_status_of "$bfile")

  if [ -z "$bstatus" ]; then
    # The branch does not carry the report, so nothing is owed a mirror.
    # Written even to a mirror already closed: a report born triaged in a
    # pull request nobody merged left one closed `completed`, which says
    # the queue acted on something the queue never received.
    clear_status "$num" "$labels"
    gh api -X PATCH "repos/${REPO}/issues/${num}" \
      -f state=closed -f state_reason=not_planned >/dev/null
    echo "${oid} closed — #${PR} was not merged"
    continue
  fi

  wreason=$(close_reason_of "$bstatus")
  if [ -n "$wreason" ]; then
    # Triaged on the branch, whatever this pull request thought.
    if [ "$istate" = "open" ]; then
      clear_status "$num" "$labels"
      gh api -X PATCH "repos/${REPO}/issues/${num}" \
        -f state=closed -f "state_reason=${wreason}" >/dev/null
      echo "${oid} closed ${wreason} — the branch has it triaged"
    fi
    continue
  fi

  # Still recorded and still untriaged on the branch: the mirror is owed
  # to it, open and asking to be read, however this pull request left it.
  if [ "$istate" = "open" ]; then
    case ",${labels}," in *",status:open,"*) continue ;; esac
  fi
  ensure_report_status_label "status:open"
  if [ "$istate" = "closed" ]; then
    gh api -X PATCH "repos/${REPO}/issues/${num}" -f state=open >/dev/null
  fi
  put_report_labels "$num" "$labels" "status:open"
  echo "${oid} → status:open — #${PR} was not merged and the branch still holds it"
done <<EOF
$REPORT_ISSUES
EOF

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  printf 'tasks=%s\n' "$MIRRORED" >> "$GITHUB_OUTPUT"
  printf 'reports=%s\n' "$MIRRORED_REPORTS" >> "$GITHUB_OUTPUT"
fi

exit 0
