#!/usr/bin/env bash
# queue_lib.sh — the helpers both halves of the transition machine share:
# front-matter reads and writes, the resting derivation, the task-file
# resolver, the carried-ids parser, and the tab-delimited row reader.
# Sourced, never executed; the sourcing script owes `set -euo pipefail`
# itself.
#
# One copy on purpose: flip_task_status.sh and record_task_status.sh
# each carried private clones of these until a review caught the clones
# drifting — and caught the resolver pipeline dying under pipefail when
# the last find candidate failed the id filter.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions. See the
# standing rule in docs/technical/decisions/.

# ql_fm_field_in <field> — the field's value from the front-matter block
# alone, read from stdin; a body line spelling `status:` at column 0
# never counts. One awk body, two entry points: this is the primitive,
# and `ql_fm_field` below is its file form. The stdin door exists
# because `git show REV:path |` is the shape a caller cannot avoid when
# the blob may legitimately be absent — a file new in the range has
# nothing at the base, and the pipeline's empty output is the answer,
# never an error to handle.
ql_fm_field_in() {
  awk -v f="$1" '
    NR == 1 { if ($0 != "---") exit; next }
    /^---$/ { exit }
    sub("^" f ": *", "") { sub(/[[:space:]]*$/, ""); print; exit }
  '
}

# ql_fm_field <field> <file> — the same read from a named file.
ql_fm_field() {
  ql_fm_field_in "$1" < "$2"
}

# ql_set_field <file> <field> <value> — front matter only.
ql_set_field() {
  awk -v field="$2" -v value="$3" '
    NR == 1 && $0 == "---" { infm = 1; print; next }
    infm && /^---$/        { infm = 0; print; next }
    infm && index($0, field ":") == 1 { print field ": " value; next }
    { print }
  ' "$1" > "$1.tmp" && mv "$1.tmp" "$1"
}

# ql_task_num <anything> — the task number, zero-padding stripped;
# empty when the input names none.
#
# The zeros go in a step of their own, after the prefix rather than with
# it: `0034` is how every queue file and every [TASK-NNNN] tag spells the
# id, so it is what a person retypes — and stripping the padding only
# when a `task-` prefix carried it made that spelling resolve to nothing.
ql_task_num() {
  printf '%s' "$1" | sed -E 's/^task-//; s/^task\///; s/^0+//; s/[^0-9].*$//'
}

# ql_task_file <task-id-or-number> — the work/tasks file whose id is
# that number, whatever width it was written at; empty when none. The
# filter is an `if` on purpose: a trailing failed `[ … ] &&` would end
# the while loop non-zero and, under pipefail, kill the caller with no
# output at all.
ql_task_file() {
  local num
  num=$(ql_task_num "$1")
  [ -n "$num" ] || return 0
  find work/tasks \( -iname "task-*${num}.md" -o -iname "task-*${num}-*.md" \) 2>/dev/null \
    | while IFS= read -r c; do
        if [ "$(ql_task_num "$(basename "$c" .md)")" = "$num" ]; then
          printf '%s\n' "$c"
        fi
      done | head -n1
  return 0
}

# ql_spec_file <spec-id> — same resolution for a spec.
ql_spec_file() {
  find work/specs \( -iname "$1.md" -o -iname "$1-*.md" \) 2>/dev/null | head -n1
  return 0
}

# ql_resting <task-file> — where a task out of flight belongs: ready, or
# backlog if any spec it references is draft. An empty spec_ref is ready
# by construction — no approval event exists for it, and backlog must
# not be a trap.
ql_resting() {
  local refs ref spec st
  refs=$(ql_fm_field spec_ref "$1" | tr -d '[]' | tr ',' ' ')
  for ref in $refs; do
    [ -n "$ref" ] || continue
    spec=$(ql_spec_file "$ref")
    [ -n "$spec" ] || continue
    st=$(ql_fm_field status "$spec")
    if [ "$st" = "draft" ]; then printf 'backlog'; return 0; fi
  done
  printf 'ready'
}

# QL_CARRIED_MAX — the most distinct tasks one pull request may claim.
# It bounds the carried set below, counted after dedup, because the set
# is what becomes status writes: both routes into it are the author's to
# type, and without a ceiling one title moves the queue by being long.
# A constant, not a setting: the schema requires a key's documented
# default to be the behaviour from before the key existed, and here that
# behaviour is unbounded — the defect itself. Eight sits above five, the
# largest related batch one merge here ever produced, and far below the
# queue (docs/technical/decisions/pull-requests/0068-what-a-pull-request-claims-is-bounded.md).
QL_CARRIED_MAX=8

# ql_carried_of <head-branch> <title> — the task ids whose work a pull
# request carries: the head branch's own (task/NNNN-*) plus every
# [TASK-NNNN] tag leading the title, deduplicated. Both arguments are a
# fork's to write, so only digits survive.
#
# Above QL_CARRIED_MAX distinct tasks, the whole set is refused: the
# sentinel `over-ceiling:<count>` is printed alone on stdout, and the
# exit stays 0. A caller tests for it with one `case` before touching
# the ids; ql_carried_from_env passes it through untouched. Exit 0 on
# purpose — every call site assigns this output bare under
# `set -euo pipefail`, and a non-zero substitution would kill such a
# caller with no message at all. The sentinel is a token no task id can
# be, in the stream every caller already reads, so a forgetful caller
# meets a non-id that matches no task file, never a vanished run.
#
# Taking the pair as arguments is what lets a caller ask the question of
# *another* pull request — the amendment check has to, to name the one it
# suspends, and apply_pr_event.sh's survivor query has to, so a close
# finds a survivor by every route it carries a task — while the
# env-reading form below stays the shape CI uses.
ql_carried_of() {
  local carried="" num rest tg
  case "${1:-}" in
    task/[0-9]*)
      num=$(ql_task_num "$1")
      [ -n "$num" ] && carried="task-$num"
      ;;
  esac
  rest="${2:-}"
  while :; do
    rest=$(printf '%s' "$rest" | sed 's/^[[:space:]]*//')
    tg=$(printf '%s' "$rest" | sed -n 's/^\[[Tt][Aa][Ss][Kk]-0*\([0-9][0-9]*\)\].*/\1/p')
    [ -n "$tg" ] || break
    case " $carried " in
      *" task-$tg "*) ;;
      *) carried="${carried:+$carried }task-$tg" ;;
    esac
    rest=$(printf '%s' "$rest" | sed 's/^\[[Tt][Aa][Ss][Kk]-[0-9][0-9]*\]//')
  done
  # The count is of the deduplicated set, never of the tags: the set is
  # what becomes writes, so a title repeating one tag fifty times still
  # claims one task. Word splitting is the count — the ids carry no
  # spaces and no globs.
  # shellcheck disable=SC2086
  set -- $carried
  if [ "$#" -gt "$QL_CARRIED_MAX" ]; then
    printf 'over-ceiling:%s' "$#"
    return 0
  fi
  printf '%s' "$carried"
}

# ql_carried_from_env — the same question about the pull request CI is
# running on, read from env as data (PR_HEAD_REF, PR_TITLE).
ql_carried_from_env() {
  ql_carried_of "${PR_HEAD_REF:-}" "${PR_TITLE:-}"
}

# --- the tab-delimited row ------------------------------------------------
#
# QL_TAB — the delimiter, computed once. `printf '\t'` in every caller was
# four spellings of one character.
QL_TAB=$(printf '\t')

# ql_row_fields <count> <row> — splits one tab-delimited row into exactly
# <count> fields, left to right, and leaves them in QL_F1 … QL_F<count>.
# The last field takes the remainder, tabs and all. Returns 1 when the
# row holds fewer separators than that, so a caller skips a row it cannot
# answer instead of reading it short.
#
# **`read` cannot do this, and that is why the helper exists.** A tab is
# an IFS *whitespace* character, so `IFS="$TAB" read -r a b c` folds a run
# of tabs into one separator: an empty field vanishes and every field
# after it shifts left by one. The empty field is not hypothetical — `gh`
# emits `author.login` as the empty string for a pull request whose
# author deleted their account, and `@tsv` writes that as two adjacent
# tabs. The shift is silent, and what it produced was a listing row about
# a pull request still working a task being read as a row about something
# else.
#
# One reader for three callers, for the reason the helpers above are one
# copy: three correct copies of one parse agree until one of them is
# edited. Named for the parse rather than for the forge, because one of
# the three rows never came from a forge — check_unique_ids.sh assembles
# its own — and a helper named for pull requests would read wrong there.
#
# **A newline does not survive, and cannot.** It is the *row* separator
# every caller splits on before this is reached, so a field carrying one
# has already ended its row. That is why every projection puts the
# free-text field last: a title is the only field a person writes, and
# last is where a broken row costs a title rather than a number.
ql_row_fields() {
  local want="$1" row="$2" i=1 f
  while [ "$i" -lt "$want" ]; do
    case "$row" in
      *"$QL_TAB"*) ;;
      *) return 1 ;;
    esac
    f=${row%%"$QL_TAB"*}
    row=${row#*"$QL_TAB"}
    # eval, because bash 3.2 has no name references and no associative
    # arrays. The value reaches the assignment through $f, never through
    # the evaluated string, so a field holding a quote or a `$` is data.
    eval "QL_F${i}=\$f"
    i=$((i + 1))
  done
  eval "QL_F${want}=\$row"
  return 0
}

# --- draft chapters -------------------------------------------------------
#
# QL_DRAFT_MARKER — the line a chapter puts first to say it is **not a
# rule yet**. Nothing derives from a chapter carrying it: no task, no
# spec, no recorded gap
# (docs/product/stage-1-tasks-and-specs/authoring.md#a-chapter-that-is-not-a-rule-yet).
QL_DRAFT_MARKER='/// writrun:draft'

# ql_doc_is_draft <path> [ref] — does that chapter declare itself a
# draft? With a ref, the file as that ref holds it; without one, the file
# in the checkout. A file that is not there is not a draft, which is the
# answer that keeps every caller strict: an absent chapter never earns an
# exemption.
#
# **The first line, and no other.** A chapter that *documents* the marker
# names it in its prose — `authoring.md` does, twice — so a reader that
# searched the whole file would mark the methodology's own docs as
# drafts. This repository has been bitten from the other direction
# already: a report quoting front matter in its body was read as carrying
# that front matter, and a mirror was closed on it
# (tests/…/quoted_front_matter_is_not_a_status_test.sh).
#
# Trailing whitespace is trimmed — a line an editor touched is the same
# declaration — and nothing else is: a marker indented, or on line two,
# or inside a fence, is prose about the marker rather than the marker.
#
# **The blob is captured before it is read.** `git show … | sed -n 1p`
# would be fine, but `| head -1` or an early-exiting awk would not: the
# reader closes the pipe, git dies on SIGPIPE and `pipefail` turns a
# correct read into a failure. This suite has been bitten by that shape
# twice, so the whole blob lands in a variable first.
ql_doc_is_draft() {
  local blob first
  if [ -n "${2:-}" ]; then
    blob=$(git show "${2}:${1}" 2>/dev/null) || return 1
  else
    [ -f "$1" ] || return 1
    blob=$(cat "$1")
  fi
  first=${blob%%$'\n'*}
  first=$(printf '%s' "$first" | sed 's/[[:space:]]*$//')
  [ "$first" = "$QL_DRAFT_MARKER" ]
}

# ql_range_ends <range> — the range's two ends, derived once, into
# QL_BASE and QL_HEADREF. The three shapes `git diff` accepts at the
# stage-2 gates:
#
#   A...B  QL_BASE is the merge base, QL_HEADREF the right end. A
#          merge-base that could not be computed is not a base of
#          "nothing", it is an unanswered question — and the callers
#          are gates, so it exits 3, loudly.
#   A..B   the two ends as written; an omitted end is HEAD.
#   A      QL_BASE is the ref, and the bare shape's diff compares it to
#          the **working tree** — so QL_HEADREF is empty. A caller that
#          probes blobs honours the sentinel by reading the checkout
#          (spec-0084); one that reads commits replaces it:
#          `TIP="${QL_HEADREF:-HEAD}"`.
#
# Lifted from three private copies that had already diverged in more
# than spelling: the bare-ref bug spec-0084 fixed had to be found in
# one copy out of three, and the next range-shape bug should not get
# that chance.
QL_BASE=""
QL_HEADREF=""
ql_range_ends() {
  local left right
  case "$1" in
    *...*)
      left="${1%%...*}"
      right="${1##*...}"
      QL_HEADREF="${right:-HEAD}"
      if ! QL_BASE=$(git merge-base "${left:-HEAD}" "${right:-HEAD}" 2>&1); then
        echo "git merge-base ${left:-HEAD} ${right:-HEAD} failed:" >&2
        printf '%s\n' "$QL_BASE" | head -n 2 >&2
        exit 3
      fi
      ;;
    *..*) QL_BASE="${1%%..*}"; QL_HEADREF="${1##*..}"; QL_HEADREF="${QL_HEADREF:-HEAD}" ;;
    *)    QL_BASE="$1"; QL_HEADREF="" ;;
  esac
}

# ql_git_read <label> <git-args...> — runs git and leaves its stdout in
# QL_GIT_OUT. On failure it prints what git said and exits 3, because a
# check that could not read its input must never report the empty
# result as a clean one: `$(git … || true)` yields exactly the same
# empty string whether nothing matched or nothing ran, and the callers
# are gates (spec-0013).
#
# **The label must spell the command actually run, flags included.** A
# diagnostic naming a different command sends whoever reproduces the
# failure to a different answer.
#
# **Never call this inside a command substitution.** The `exit` would
# end only the subshell, and the caller would go on reading the empty
# value this exists to prevent — the very shape of the bug being
# removed.
QL_GIT_OUT=""
ql_git_read() {
  local label="$1" err
  shift
  err=$(mktemp "${TMPDIR:-/tmp}/writrun-git.XXXXXX")
  if ! QL_GIT_OUT=$(git "$@" 2>"$err"); then
    echo "${label} failed:" >&2
    head -n 2 "$err" >&2
    rm -f "$err"
    exit 3
  fi
  rm -f "$err"
}

# --- minting ------------------------------------------------------------
#
# The id and the filename subject, shared by the two writers that mint
# queue ids: the generator (skills/writrun-create-task-and-spec/new.sh)
# and the intake (scripts/stage-3-github-issues/intake_report.sh). One
# copy for the same reason the helpers above are one copy: the intake
# opened with a private clone of this stack, and by its first review the
# clone had already dropped the outside-a-repository guard and the
# generator's collision check.
#
# QL_FORGE_VIEW names how much of the forge answered: `forge` when both
# halves did, `open-pull-requests` when the file lists answered and the
# mirrors did not, `local` when neither did. QL_FORGE_PATHS holds every
# path those pull requests touch, QL_FORGE_MIRROR_IDS every id a mirror
# carries. Never call ql_forge_scan from inside a command substitution —
# the subshell would swallow all three.
QL_FORGE_VIEW=local
QL_FORGE_PATHS=""
QL_FORGE_MIRROR_IDS=""

# ql_forge_scan [owner/repo] — asks the forge for the paths open pull
# requests touch. With an argument, every call is pinned to that
# repository; without one, gh answers for the checkout's own remote —
# the two channels must agree, or the scan reads one repository's pull
# request numbers and another's file lists.
#
# One call per open pull request instead of one call total, because the
# single `gh pr list --json files` this replaced was cheaper and wrong:
# that field stops at 100 files per pull request and says nothing when
# it does, so a larger diff hides every queue file it adds past the cut.
#
# All or nothing: a call that fails — or a listing that filled its own
# limit, which is a listing that may have been cut — leaves the view
# local rather than quietly narrow, since a scan that under-reports
# without saying so is the exact failure the uniqueness rule exists to
# prevent.
#
# **The mirrors are the second half, and the only record of an id that
# outlives the branch that spent it.** A queue file minted on a branch
# and dropped before that branch merged is in no tree, no history and no
# open pull request — but its Issue is on the forge, titled with the id,
# and a number that has been published is spent whatever became of the
# file (decisions/tasks-and-specs/0070). So both labels are listed, in
# every state, and the ids their titles spell are carried out.
#
# All or nothing holds per half, because the halves answer different
# questions. The file lists failing leaves the view local, as it always
# did; the mirror listing failing narrows it to `open-pull-requests`
# rather than discarding an answer the run already had. A listing that
# succeeds and returns nothing is a complete answer, not a failure — an
# adopter below Stage 3 has no mirrors, and the forge answers a label it
# does not know with an empty list rather than a refusal.
ql_forge_scan() {
  QL_FORGE_VIEW=local
  QL_FORGE_PATHS=""
  QL_FORGE_MIRROR_IDS=""
  command -v gh >/dev/null 2>&1 || return 0
  local repo="${1:-}" numbers paths files n count
  if [ -n "$repo" ]; then
    numbers=$(gh pr list -R "$repo" --state open --limit 200 --json number \
      --jq '.[].number' 2>/dev/null) || return 0
  else
    # gh defaults to 30 open pull requests, and the id this misses is
    # exactly the one worth seeing.
    numbers=$(gh pr list --state open --limit 200 --json number \
      --jq '.[].number' 2>/dev/null) || return 0
  fi
  count=$(printf '%s\n' "$numbers" | sed '/^$/d' | wc -l | tr -d ' ')
  [ "$count" -ge 200 ] && return 0
  paths=""
  for n in $numbers; do
    # --paginate is the point: a pull request's own file list is paged
    # too, and the queue file may sit on any page of it.
    if [ -n "$repo" ]; then
      files=$(gh api "repos/${repo}/pulls/${n}/files" --paginate \
        --jq '.[].filename' 2>/dev/null) || return 0
    else
      files=$(gh api "repos/{owner}/{repo}/pulls/${n}/files" --paginate \
        --jq '.[].filename' 2>/dev/null) || return 0
    fi
    paths="${paths}${files}
"
  done
  QL_FORGE_PATHS="$paths"
  QL_FORGE_VIEW=open-pull-requests

  # Both labels, whatever kind is being minted, and both fetched even for
  # a spec mint that can use neither. ql_next_id runs inside a command
  # substitution, so it cannot make a forge call of its own without
  # risking a stray line landing in the id — every call in this stack
  # belongs to this one pre-pass. A kind argument here would save one
  # listing and add a second thing every caller must get right silently,
  # where a wrong value produces exactly the reuse this input closes.
  #
  # --paginate for the reason the file list above gives in its own
  # words: the listing is ordered by issue number and the id lives in
  # the title, so the highest id is not the newest Issue, and a first
  # page read alone would be a guess.
  local kind titles mirrors=""
  for kind in task report; do
    if [ -n "$repo" ]; then
      titles=$(gh api "repos/${repo}/issues?labels=writrun:${kind}&state=all&per_page=100" \
        --paginate --jq '.[].title' 2>/dev/null) || return 0
    else
      titles=$(gh api "repos/{owner}/{repo}/issues?labels=writrun:${kind}&state=all&per_page=100" \
        --paginate --jq '.[].title' 2>/dev/null) || return 0
    fi
    mirrors="${mirrors}$(ql_mirror_ids "$kind" "$titles")
"
  done
  QL_FORGE_MIRROR_IDS="$mirrors"
  QL_FORGE_VIEW=forge
  return 0
}

# ql_mirror_ids <kind> <titles> — the ids a mirror listing's titles name,
# one `task-0011` or `report-0003` per line. The grammar is the one
# mirror_issues.sh's id_of_title reads, because it is the one that script
# writes; it lives twice because that script shares no library with this
# one, and a title is the whole contract between them.
#
# A title with no tag contributes nothing and is not an error: an
# Issue a maintainer labelled by hand, waiting for the intake to retitle
# it, names no id yet.
ql_mirror_ids() {
  printf '%s\n' "$2" | tr '[:upper:]' '[:lower:]' | sed -n \
    -e "s/^\[\($1-[0-9][0-9]*\)\].*/\1/p" \
    -e "s/^\($1-[0-9][0-9]*\)[[:space:]].*/\1/p"
}

# ql_mint_note — what the id was minted against, printed after the file
# so an id claimed elsewhere is never reported as simply "created". Three
# shapes for the three views, because a run that had part of the forge
# should claim that part and no more.
ql_mint_note() {
  case "$QL_FORGE_VIEW" in
    forge)
      echo "Minted above the queue, its history, every open pull request, and every mirror."
      ;;
    open-pull-requests)
      echo "Minted above the queue, its history, and every open pull request —" >&2
      echo "the mirrors went unanswered, so an id only a closed mirror still holds" >&2
      echo "would not have been seen." >&2
      ;;
    *)
      echo "Minted from this checkout and its history only — no forge answered," >&2
      echo "so an id an open pull request already claims would not have been seen." >&2
      ;;
  esac
}

# ql_next_id <dir> <prefix> — e.g. ql_next_id work/tasks task
ql_next_id() {
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
  if [ -n "$QL_FORGE_PATHS" ]; then
    while IFS= read -r f; do
      case "$f" in "$dir"/*) bump "$f" ;; esac
    done <<EOT
$QL_FORGE_PATHS
EOT
  fi

  # And a mirror holds a number all three of those can lose: a queue file
  # minted on a branch and dropped before it merged left the tree, the
  # history and the pull request list, and left its Issue standing. The
  # fourth input is that Issue's title, already reduced to an id.
  #
  # Guarded by the prefix, so a task mint reads no report's number — and
  # a spec mint reads nothing at all, because only tasks and reports are
  # mirrored. These are ids and not paths, deliberately: QL_FORGE_PATHS
  # is filtered by the caller's own directory, and a synthesized
  # `work/tasks/…` would be wrong for any adopter whose queue lives
  # elsewhere.
  if [ -n "$QL_FORGE_MIRROR_IDS" ]; then
    while IFS= read -r f; do
      case "$f" in "$prefix"-*) bump "$f" ;; esac
    done <<EOT
$QL_FORGE_MIRROR_IDS
EOT
  fi

  printf "%04d" $((max + 1))
}

# ql_slugify <title> — the filename's subject, **derived**: an extremely
# short kebab-case echo of the title, at most three words. Identity
# lives in the id, so a derived slug that loses nuance costs nothing; it
# exists to make a directory listing readable. Prints nothing when the
# title has no usable word, and the caller then writes a bare-id
# filename.
ql_slugify() {
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
