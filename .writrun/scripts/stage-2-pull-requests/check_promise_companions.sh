#!/usr/bin/env bash
# check_promise_companions.sh — a promise includes its mandatory
# companions, refused where the spec enters.
#
# Usage: check_promise_companions.sh <diff-range>
#
# Some documents never change alone: a rule elsewhere makes touching one
# imply touching another, so a promise naming the first without the
# second is not a smaller promise — it is a wrong one
# (docs/product/concepts/spec.md#the-doc-delta-contract).
#
# **Where this fires is the whole point.** The completion gate,
# `writrun-check-spec-deltas`, already catches an incomplete promise —
# but it catches it against a finished branch, where the fix is an
# amendment under an open pull request and a suspended task. That is the
# case this check exists to stop happening again: it reads the specs the
# range adds or modifies, which is the pull request that *creates or
# amends* a spec, where the fix is one edit and nothing has been assented
# to yet. It is not a second completion gate and never judges the diff's
# doc changes — only the promise's own internal completeness.
#
# **Only the pair this range wrote.** A spec that reached the base branch
# already promising an entry and not its index is out of reach here — the
# check reads the range, and history is not re-judged, or a completion
# pull request that merely flips `approved` to `implemented` would be
# refused with a message insisting the cheap fix is still available when
# it is not. So the promise is read at the merge base too, and a pair
# whose halves both predate the range is left to the completion gate that
# already holds it.
#
# **The pair table is a table so a second pair is a row.** One line per
# pair, `<entry-glob> <companion>`, both repository-root relative — the
# shape promises are normalised to below. The refusal prints them back
# the other way, relative to `docs/`, because that is the shape a spec
# writes and the author's fix is to paste the named path into a section.
# The first pair is the one the authoring case named; the next one is a
# line, not a rewrite.
#
# Exit codes: 0 every promise read carries its companions, or there was
# no promise to read; 1 a promise is incomplete, with each named; 3 usage
# error, or a range that could not be read.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

# Spelled out rather than `${1:?…}`, which exits 1 — the code this check
# uses for "a promise is incomplete". A caller cannot be left reading a
# mis-wired invocation as a rule violation.
if [ "$#" -lt 1 ] || [ -z "${1:-}" ]; then
  echo "usage: check_promise_companions.sh <diff-range>" >&2
  exit 3
fi
RANGE="$1"

. "$(dirname "$0")/queue_lib.sh"

# --- the pair table -------------------------------------------------------
#
# The dated decisions log and its chronology index. The index is the only
# part of that folder that is rewritten, and it is rewritten by appending
# a row whenever an entry is added — so an entry promised alone is always
# a promise short of the truth.
#
# The leading `*` in the entry glob is what covers both `decisions_style`
# variants at once: `per-subsystem` puts the entry in a folder of the
# adoption level it concerns, `chronological` puts it in the log's root,
# and a `case` pattern's `*` spans the separator either way. A project
# that spells its layout a third way changes this row; it does not read
# the setting, because the promise is a path and the path is what is
# being judged.
PAIRS="
docs/technical/decisions/*[0-9][0-9][0-9][0-9]-*.md docs/technical/decisions/README.md
"

# --- the range's base -----------------------------------------------------
ql_range_ends "$RANGE"
BASE="$QL_BASE"

# promised_paths — every path both Proposed-changes sections of the spec
# on stdin name, anchor stripped, normalised to repository-root the way
# the schema says to read them: a spec writes `technical/…`, relative to
# `docs/`.
#
# Reads stdin rather than a file so the same parser serves the working
# tree and `git show BASE:<spec>` — one promise reader, two revisions.
#
# A deliberate second reader of the same two sections, not a call into
# the completion gate: that script is a Stage 1 skill and must keep
# running with nothing but `work/` and git, while this one is workflow
# machinery. The scope of what they share is four lines of awk; the cost
# of coupling them is a skill that stops working where it is promised to.
#
# The empty line is dropped *before* the prefix, or a `none —` bullet's
# nothing would arrive as the path `docs/`.
promised_paths() {
  awk '
    /^## Proposed (product|technical) changes/ { inp = 1; next }
    /^## / && inp { inp = 0 }
    inp && /^- `/ { print }
  ' \
    | sed -n 's/^- `\([^`]*\)`.*/\1/p' | sed 's/#.*//' \
    | sed '/^$/d' | sed 's|^|docs/|' | sort -u
}

# promise_covers <path> <promise> — the promise names this exact path, or
# a folder above it. A trailing `/` is what `check_deltas.sh` reads as
# covering everything beneath, so the two readers must agree on it or
# this gate would refuse a promise the completion gate accepts.
promise_covers() {
  local want="$1" list="$2" p
  while IFS= read -r p; do
    [ -n "$p" ] || continue
    case "$p" in
      */) case "$want" in "$p"*) return 0 ;; esac ;;
      *)  if [ "$p" = "$want" ]; then return 0; fi ;;
    esac
  done <<COVER
${list}
COVER
  return 1
}

# engages <promised-path> <entry-glob> — this promised path is an entry
# the pair speaks about.
#
# A concrete path engages the glob directly. A folder promise engages it
# when a dated entry could live inside the folder, which the probe name
# asks in the only vocabulary a `case` pattern has: promising the folder
# is a legitimate way to promise the entry beneath it, and a pair that
# read concrete paths alone would stay silent on exactly the omission it
# exists to catch — leaving it to surface at the completion gate as an
# undeclared index.
engages() {
  local p="$1" glob="$2" probe
  case "$p" in
    */) probe="${p}0000-probe.md" ;;
    *)  probe="$p" ;;
  esac
  # Unquoted on purpose: the table's field is a pattern, not a path.
  case "$probe" in
    $glob) return 0 ;;
  esac
  return 1
}

# --- the specs this change enters -----------------------------------------

ql_git_read "git diff --name-only ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only "$RANGE" -- 'work/specs/*.md'
touched="$QL_GIT_OUT"

read_specs=0
faults=0
fault() { echo "REJECTED: $*" >&2; faults=$((faults + 1)); }

# Read line by line and never with `for s in $touched`: word splitting
# turns one path containing a space into two paths that exist nowhere,
# each skipped by the `-f` test below — a promise dropped in silence,
# which for a gate is the same failure as reading nothing at all.
#
# Here-documents throughout rather than pipes: a pipeline's subshell
# cannot raise the counter it is counting into.
while IFS= read -r spec; do
  [ -n "$spec" ] || continue

  # git quotes a path holding control characters or non-ASCII bytes. A
  # spec path never needs it, so a quoted one is a path this check cannot
  # parse, and an unparseable input is refused rather than skipped.
  case "$spec" in
    '"'*)
      echo "cannot read the changed path ${spec} — refusing rather than skipping it" >&2
      exit 3
      ;;
  esac

  [ -f "$spec" ] || continue        # deleted on the branch: promises nothing

  promised=$(promised_paths < "$spec")
  [ -n "$promised" ] || continue    # a promise of "none" names no path
  read_specs=$((read_specs + 1))

  # The same promise as the range found it. A spec absent from the base
  # is new, and everything it promises is this range's to answer for —
  # asked with `cat-file -e` so that "not there" stays distinguishable
  # from "git could not say", which ql_git_read still refuses.
  base_promised=""
  if git cat-file -e "${BASE}:${spec}" 2>/dev/null; then
    ql_git_read "git show ${BASE}:${spec}" show "${BASE}:${spec}"
    base_promised=$(printf '%s\n' "$QL_GIT_OUT" | promised_paths)
  fi

  id=$(ql_fm_field id "$spec")
  [ -n "$id" ] || id="$spec"

  while read -r glob companion; do
    [ -n "$glob" ] || continue

    # Asked once per pair rather than once per entry: a promise naming
    # three entries and the index is complete, and asking per entry would
    # only be the same yes three times.
    if promise_covers "$companion" "$promised"; then
      continue
    fi

    while IFS= read -r p; do
      [ -n "$p" ] || continue
      engages "$p" "$glob" || continue

      # A pair whose halves both predate the range is not this gate's to
      # judge — the spec arrived incomplete and the completion gate still
      # holds it. Anything else is this range's doing: an entry it added,
      # or a companion it dropped.
      if promise_covers "$p" "$base_promised" \
         && ! promise_covers "$companion" "$base_promised"; then
        continue
      fi

      # Named back relative to `docs/`, the shape the spec writes: the
      # fix is to paste this path into a Proposed changes section, and a
      # repository-root path pasted there normalises to `docs/docs/…` and
      # is refused again by the same message.
      fault "${id} promises ${p#docs/} and not ${companion#docs/}, which adding an entry implies."
    done <<PROMISED
${promised}
PROMISED
  done <<PAIRTABLE
${PAIRS}
PAIRTABLE
done <<TOUCHED
${touched}
TOUCHED

if [ "$faults" -ne 0 ]; then
  echo "" >&2
  echo "Some documents never change alone, and a promise naming the first" >&2
  echo "without the second is wrong rather than smaller. Add the companion" >&2
  echo "to the same Proposed changes section, spelled as it is named above" >&2
  echo "— relative to docs/ — here, where it is one edit, rather than at" >&2
  echo "the completion gate, where it is an amendment under a finished" >&2
  echo "branch (docs/product/concepts/spec.md#the-doc-delta-contract)." >&2
  exit 1
fi

if [ "$read_specs" -eq 0 ]; then
  echo "No spec this change adds or modifies promises a path — nothing to judge."
  exit 0
fi

echo "OK — ${read_specs} promise(s) read; every mandatory companion is present."
