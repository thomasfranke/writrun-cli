#!/usr/bin/env bash
# check_promise_paths.sh — a promise whose path cannot resolve under
# `docs/`, or which resolves onto a chapter that is not a rule yet, is
# refused where the spec enters.
#
# Usage: check_promise_paths.sh <diff-range>
#
# A path in either Proposed-changes section is read relative to `docs/`
# — the schema says so, and `check_deltas.sh` prefixes every bullet with
# it before comparing. A spec that writes a repository-root path instead
# promises something no diff can ever touch: `tests/harness.sh` is read
# as `docs/tests/harness.sh`, a file that will never appear in any range,
# so the promise is unkeepable rather than merely unkept
# (docs/product/concepts/spec.md#the-doc-delta-contract).
#
# **Resolution is not the whole question.** Two of the three conditions
# below are about it; the third is about the chapter a resolving path
# lands on, and the two must not be run together in a reader's head or in
# a refusal's wording.
#
# On resolution the rule is **shape, never existence**. A spec
# legitimately promises a doc its own change will create, so "is the file
# there" is not askable — at spec entry the promised doc is precisely
# what has not been written yet. What is askable is whether the path is a
# documentation path at all, and two conditions answer that without a
# table of known folders:
#
#   - **Root-relative.** The first segment names an entry at the
#     repository root for which `docs/` holds no counterpart. `tests/`,
#     `kit/`, `.github/`, `.writrun/` and a leading `docs/` are all
#     caught this way; `product/` and `technical/` are not root entries,
#     so a promise into them passes whether or not the file exists yet.
#   - **A folder promise missing its slash.** The path names a directory
#     that is there, written without the trailing slash `check_deltas.sh`
#     reads as covering everything beneath — so the comparison would be
#     against file paths and would never match.
#
# **Neither of them is the file's extension**, and one of them used to
# be. A document is not a file format: a drawing stating what a screen
# renders, a payload shape, a fixture a rule is written against are each
# read by a person, state what the implementation must satisfy, and are
# touched by the diff that changes them. `check_deltas.sh` already reads
# `docs/` whole and always did, so the extension test here left such a
# file UNDECLARED at the completion gate and unpromisable at this one —
# neither touchable nor declarable
# (docs/technical/decisions/pull-requests/0075-a-document-is-not-a-file-format.md).
#
# Both are read off the repository rather than a list, so an adopter
# whose docs tree is shaped differently is judged by their own tree
# (docs/technical/decisions/pull-requests/0065-a-promise-is-judged-by-shape.md).
#
# The third condition is a different question with a different answer:
#
#   - **Not a rule yet.** The path resolves, and the chapter it names
#     declares itself a draft. Nothing derives from a chapter that is not
#     a rule, so the promise is about a chapter no task was allowed to be
#     born from (authoring.md#a-chapter-that-is-not-a-rule-yet).
#
# The fourth is about the bullet rather than the path:
#
#   - **A bullet the gates cannot read.** A top-level bullet under either
#     heading that neither opens with a backticked path nor says `none`.
#     Both gates read a promise off the backtick that follows the dash,
#     so a bullet written any other way — a link around the path, bold
#     before it, prose — used to be no promise at all: this check found
#     nothing to judge and passed, and the completion gate later judged
#     every touched document undeclared against an empty set
#     (report-0044, decision 0078). It is refused here, by name, where
#     the fix is one edit.
#
# Which condition faulted decides what the refusal's closing advice says.
# A path already written exactly as the schema reads it must never be
# answered with "write it as the schema reads it" — that sends the author
# hunting a typo that is not there, the failure this check exists to
# spare them.
#
# **Where this fires is the whole point**, and it is the reason this is
# not left to `writrun-check-spec-deltas`. That gate catches the same
# promise, but against a finished branch, where the fix is an amendment
# under an open pull request and a suspended task — which is what
# `spec-0041` cost on #83. This one reads the specs the range adds or
# modifies, so the fix is one edit and nothing has been assented to yet.
# A spec that reached the base branch already carrying an offending path
# is out of reach here, the same boundary the companions check draws.
#
# Exit codes: 0 every promise read resolves and names a rule, or there
# was no promise to read; 1 a promise cannot resolve, or resolves onto a
# chapter that is not a rule yet, with each named; 3 usage error, a range
# that could not be read, or a changed path this check cannot parse.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
. "$HERE/queue_lib.sh"

# Spelled out rather than `${1:?…}`, which exits 1 — the code this check
# uses for "a promise cannot resolve". A caller cannot be left reading a
# mis-wired invocation as a rule violation.
if [ "$#" -lt 1 ] || [ -z "${1:-}" ]; then
  echo "usage: check_promise_paths.sh <diff-range>" >&2
  exit 3
fi
RANGE="$1"

# --- the range's base ------------------------------------------------------
#
# The same resolution the companions check uses. A merge-base that could
# not be computed is not a base of "nothing", it is an unanswered
# question.
#
# What it is for: a promise that already reached the base is out of
# reach here — the fix there is an amendment under an open pull request,
# not the one edit this refusal offers. Without it a completion run that
# merely flips `approved` to `implemented` is refused with a message
# insisting on a cheap fix that no longer exists.
# The ends come from queue_lib.sh's one parse (spec-0086); this check
# reads only the base — its spec bodies come from the checkout, a
# deliberate choice recorded in decision 0071 — so the bare shape's
# working-tree sentinel in QL_HEADREF is simply never read here.
ql_range_ends "$RANGE"
BASE="$QL_BASE"

# The bullets are read by queue_lib.sh's one reader — `ql_promised_paths`
# for the paths **as the spec wrote them** (the `docs/` prefix its sibling
# adds is deliberately not applied: this check judges the written form
# and prints the prefixed form back, so the author sees the reading that
# made the promise unkeepable), and `ql_unreadable_promises` for the
# bullets no reader can take. One copy, so the completion gate cannot
# read a bullet this gate refused, nor pass one it never saw.
TAB=$(printf '\t')

# --- the specs this change enters -----------------------------------------

ql_git_read "git diff --name-only ${RANGE} -- 'work/specs/*.md'" \
  diff --name-only "$RANGE" -- 'work/specs/*.md'
touched="$QL_GIT_OUT"

read_specs=0
faults=0

# The two kinds are counted apart because the closing advice differs:
# printing the resolution advice on a draft fault would tell an author to
# rewrite a path that is already written correctly, which is the typo
# hunt the draft refusal exists to spare them.
#
# The kind is the first argument and there is no unkinded form, so a
# fault added later cannot reach the exit with neither trailer behind it
# — a refusal that names a path and then offers no way out of it is a
# worse answer than either wording.
resolution_faults=0
draft_faults=0
unreadable_faults=0
fault() {
  local kind="$1"
  shift
  echo "REJECTED: $*" >&2
  faults=$((faults + 1))
  case "$kind" in
    resolution) resolution_faults=$((resolution_faults + 1)) ;;
    draft)      draft_faults=$((draft_faults + 1)) ;;
    unreadable) unreadable_faults=$((unreadable_faults + 1)) ;;
    *) echo "internal: fault kind '${kind}' has no closing advice" >&2; exit 3 ;;
  esac
}

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

  promised=$(ql_promised_paths < "$spec")
  unreadable=$(ql_unreadable_promises < "$spec")
  # A promise of "none" names no path and carries no unreadable bullet:
  # nothing to judge, as before.
  [ -n "$promised" ] || [ -n "$unreadable" ] || continue
  read_specs=$((read_specs + 1))

  id=$(ql_fm_field id "$spec")
  [ -n "$id" ] || id="$spec"

  # Condition four — a bullet the gates cannot read. Judged before the
  # paths, because a spec whose only bullets are unreadable has no path
  # to judge and used to pass on exactly that. A bullet the base already
  # carried is history, as a path already promised there is below.
  base_unreadable=$(git show "${BASE}:${spec}" 2>/dev/null | ql_unreadable_promises || true)
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    if [ -n "$base_unreadable" ] && printf '%s\n' "$base_unreadable" | grep -qxF -- "$line"; then
      continue
    fi
    sec="${line%%"$TAB"*}"
    text="${line#*"$TAB"}"
    fault unreadable "${id}'s Proposed ${sec} changes carries a bullet the gates cannot read — \`${text}\` — so it promises nothing, and a promise that reads as nothing passes here and fails at the completion gate."
  done <<UNREADABLE
${unreadable}
UNREADABLE

  # What the same spec already promised at the base. A path in this list
  # is history — refusing it here would offer a one-edit fix for
  # something that is now an amendment under an open pull request. A
  # spec the base does not have promises nothing there, and `git show`
  # failing is exactly that, so its stderr is dropped rather than read
  # as an unanswered question.
  base_promised=$(git show "${BASE}:${spec}" 2>/dev/null | ql_promised_paths || true)

  while IFS= read -r p; do
    [ -n "$p" ] || continue

    if [ -n "$base_promised" ] && printf '%s\n' "$base_promised" | grep -qxF -- "$p"; then
      continue
    fi

    # The reading `check_deltas.sh` will take, and the one the refusal
    # prints: the author's fix is to write a path that reads as a real
    # documentation path, so they are shown the one theirs read as.
    as_read="docs/${p}"

    # A leading slash is refused before anything reads a segment: it is
    # not a root-relative promise but an unreadable one, prefixed into
    # `docs//…`, which no diff can match — and its empty first segment
    # would otherwise slip past condition one untested.
    case "$p" in
      /*)
        fault resolution "${id} promises \`${p}\`, read as ${as_read} — a leading slash makes a path no diff can match."
        continue
        ;;
    esac

    # Condition one — root-relative. The first segment is compared
    # against the repository's own top level, so nothing here is a list
    # to keep in step with the tree. `docs/` holding a counterpart is
    # what makes the documentation reading win where the two collide:
    # a project with both `docs/tests/` and `tests/` promises into its
    # docs, which is the only reading that can ever be kept.
    #
    # Only a path carrying a slash has a first segment to compare. A
    # promise naming a doc at `docs/`'s own top level — `about.md` —
    # has none, and reading its whole text as a segment would compare it
    # against the repository's root *files* and refuse `README.md`,
    # which is a promise the criterion says shall pass.
    case "$p" in
      */*) first="${p%%/*}" ;;
      *)   first="" ;;
    esac
    if [ -n "$first" ] && [ -e "$first" ] && [ ! -e "docs/${first}" ]; then
      fault resolution "${id} promises \`${p}\`, read as ${as_read} — \`${first}\` is a repository-root entry and docs/${first} does not exist, so no diff can ever touch it."
      continue
    fi

    # Condition two — a folder promise missing its slash.
    #
    # **The extension is not the question, and never was.** This case
    # read `*/|*.md` and faulted everything else as "not a document",
    # which made the file format the test for whether something is
    # documentation. It is not: a drawing that states what a screen must
    # render, a JSON file stating the shape of a payload, a fixture a
    # rule is written against — each is read by a person, states what the
    # implementation must satisfy, and is touched by the diff that
    # changes it, which is everything this contract asks of a document
    # except its extension (report-0041, decision 0075).
    #
    # The `.md` test was standing in for "is this a documentation path at
    # all", in a discussion entirely about repository-root paths
    # (0065). Condition one is what answers that, and it is untouched —
    # `tests/harness.sh` is still refused, by the segment that names it.
    # What the extension test cost was the far side: `check_deltas.sh`
    # reads `docs/` whole, so a non-Markdown file under it was
    # UNDECLARED at the completion gate and unpromisable here, and could
    # be neither touched nor declared.
    #
    # **What is left is the refusal the widening makes newly available.**
    # A trailing `/` is the folder promise `check_deltas.sh` reads as
    # covering everything beneath. Without one, a path that names a
    # directory that *is there* is a folder promise missing its slash —
    # `check_deltas.sh` would match it against file paths and never find
    # it. That is knowable here, and saying so costs one edit where the
    # completion gate would report a promise that can never be kept.
    #
    # Existence is asked about the *directory* only, which is why this
    # does not reintroduce the test 0065 ruled out: a spec legitimately
    # promises the file its own change creates, so an absent path stays
    # legal and is judged by shape alone.
    case "$p" in
      */) ;;
      *)
        if [ -d "$as_read" ]; then
          fault resolution "${id} promises \`${p}\`, read as ${as_read} — that is a folder, and a folder promise is written with a trailing slash."
        fi
        ;;
    esac

    # Condition three — not a rule yet. A chapter that declares itself a
    # draft derives nothing, so promising to change one is a promise
    # about a chapter no task was allowed to be born from
    # (docs/product/stage-1-tasks-and-specs/authoring.md#a-chapter-that-is-not-a-rule-yet).
    #
    # Unlike the two above, this one is not about resolution — the path
    # resolves perfectly, and saying otherwise would send the author
    # hunting a typo. The tree is read the way condition one already
    # reads it: what is checked out is the version whose promise is being
    # judged. A folder promise names no chapter and is left alone.
    case "$p" in
      */) ;;
      *.md)
        if ql_doc_is_draft "$as_read"; then
          fault draft "${id} promises \`${p}\`, read as ${as_read} — that chapter declares itself a draft, and nothing derives from a chapter that is not a rule yet."
        fi
        ;;
    esac
  done <<PROMISED
${promised}
PROMISED
done <<TOUCHED
${touched}
TOUCHED

if [ "$faults" -ne 0 ]; then
  if [ "$resolution_faults" -ne 0 ]; then
    echo "" >&2
    echo "A Proposed-changes path is read relative to docs/, so a path" >&2
    echo "written from the repository root promises a file no diff can" >&2
    echo "reach. Write it as the schema reads it — \`technical/…\`," >&2
    echo "\`product/…\` — here, where it is one edit, rather than at the" >&2
    echo "completion gate, where it is an amendment under a finished" >&2
    echo "branch (docs/product/concepts/spec.md#the-doc-delta-contract)." >&2
  fi
  if [ "$unreadable_faults" -ne 0 ]; then
    echo "" >&2
    echo "A Proposed-changes bullet is read off the backtick that follows" >&2
    echo "the dash — \`- \\\`product/…\\\` — note\`. Written any other way — a" >&2
    echo "link around the path, bold before it, prose — it is no promise to" >&2
    echo "either gate: this one passes over it and the completion gate then" >&2
    echo "judges every touched document undeclared against nothing. Open the" >&2
    echo "bullet with the path in backticks" >&2
    echo "(docs/product/concepts/spec.md#the-doc-delta-contract)." >&2
  fi
  if [ "$draft_faults" -ne 0 ]; then
    echo "" >&2
    echo "A path above resolves exactly as written — the chapter is there" >&2
    echo "and the path is right. What is missing is the rule: that chapter" >&2
    echo "still declares itself a draft, and nothing derives from one. The" >&2
    echo "fix is to the chapter or to the promise, never to the spelling —" >&2
    echo "remove the marker in the authoring change that makes it a rule," >&2
    echo "or promise a chapter the project has already committed to" >&2
    echo "(docs/product/stage-1-tasks-and-specs/authoring.md#a-chapter-that-is-not-a-rule-yet)." >&2
  fi
  exit 1
fi

if [ "$read_specs" -eq 0 ]; then
  echo "No spec this change adds or modifies promises a path — nothing to judge."
  exit 0
fi

echo "OK — ${read_specs} spec(s) read; every path resolves under docs/, onto a chapter that is a rule."
