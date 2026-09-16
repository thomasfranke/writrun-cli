#!/usr/bin/env bash
# check_body_answers.sh — a pull request that is no longer a draft has
# answered every section its body carries.
#
# Usage: check_body_answers.sh <owner/repo> <pr-number>
#   `gh` must be on PATH and authenticated (GH_TOKEN in CI; a stub in the
#   test suite).
#
# The body is the one declaration a reviewer reads before the diff, and
# `take_task.sh` seeds its headings at the taking — before the work
# exists, so before any of them could be answered. Nothing asked whether
# they ever were: #261 and #262 reached `ready` here with `## What`,
# `## Why` and `## How to test` holding nothing but the comments the
# template put there, and every completion gate ran green over them
# (docs/product/stage-2-pull-requests/body.md#a-section-the-body-carries-is-a-section-it-answered).
#
# **What the body carries is read off the body.** Not a list of required
# headings: the template ships three kind-specific sections and tells a
# change to keep the one that applies, and a project may shape the rest
# to suit its reviewers. A fixed list here would be a second copy of the
# template — wrong the first time an adopter edits theirs — and it would
# judge every project by this repository's habits. The same reading
# check_promise_paths.sh takes of a promise: shape, off the thing itself
# (decisions/pull-requests/0065-a-promise-is-judged-by-shape.md).
#
# **Unanswered has one definition, and it covers both shapes.** What
# sits under a heading, with HTML comments and whitespace removed, is
# empty. The seeded instruction is an HTML comment, so a section nobody
# touched is empty by it; a heading standing over nothing is empty
# already. An answer written *inside* a comment is refused, and
# correctly — a comment renders as nothing, so it was never read by the
# reviewer it was owed to.
#
# **Never the prose.** One word passes. A gate that weighed an answer
# would be performing review, which is the line this methodology draws
# everywhere else; what is checkable without understanding a word is
# whether anything is there at all.
#
# **Every unanswered section at once.** A gate that names one of four
# costs four round trips, and the fourth arrives after the reviewer has
# already been asked to look three times.
#
# **A draft is not judged, and says so.** Marking a pull request ready is
# the claim that a reviewer can now read it; before that the body is a
# body written before its own change. So the draft arm exits before the
# body is even fetched — the one read it needs is `isDraft`.
#
# Exit codes: 0 nothing owed, or every section answered; 1 a section the
# body carries is unanswered; 3 the forge did not answer. A call with the
# arguments missing dies on the `${…:?}` below, with bash's own 1 — the
# shape every gate here uses.
#
# Never best-effort: a check that reported "every section answered"
# without having read the body would be asserting the one thing it
# failed to look at. The lie is the verdict, not the exit code — the
# reasoning check_queue_impact.sh states for its own advisory.
#
# Portable bash 3.2, POSIX awk/sed — no gawk extensions, no associative
# arrays. See the standing rule in docs/technical/decisions/.

set -euo pipefail

REPO="${1:?usage: check_body_answers.sh <owner/repo> <pr-number>}"
PR="${2:?usage: check_body_answers.sh <owner/repo> <pr-number>}"

DOC="docs/product/stage-2-pull-requests/body.md"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh is not on PATH, so #${PR}'s body could not be read." >&2
  echo "This check has nothing else to read: a pass here would be a" >&2
  echo "claim about a body nobody fetched." >&2
  exit 3
fi

# Each read lands in a variable before anything looks at it — never
# `gh … | grep -q` or `| head -1`. The reader closes the pipe, gh dies on
# SIGPIPE and `pipefail` turns a correct read into a failure; this suite
# has been bitten by that shape twice (queue_lib.sh's ql_git_read header
# carries the whole hazard).
if ! draft=$(gh pr view "$PR" --repo "$REPO" --json isDraft --jq '.isDraft' 2>/dev/null); then
  echo "The forge did not answer for ${REPO}#${PR}." >&2
  echo "Its body could not be read, so nothing about the sections it" >&2
  echo "carries is known." >&2
  exit 3
fi

case "$draft" in
  true)
    echo "#${PR} is a draft — its body is not judged."
    echo "A draft answers nothing and owes nothing; the claim that a reviewer"
    echo "can read this body is made by marking it ready (${DOC})."
    exit 0
    ;;
  false) ;;
  *)
    echo "The forge answered '${draft}' for #${PR}'s draft state, which is" >&2
    echo "neither true nor false. Refusing rather than guessing at it." >&2
    exit 3
    ;;
esac

if ! body=$(gh pr view "$PR" --repo "$REPO" --json body --jq '.body' 2>/dev/null); then
  echo "The forge did not answer with #${PR}'s body, so whether its sections" >&2
  echo "are answered is not something this check knows." >&2
  exit 3
fi

# The body reaches awk on stdin, never as a `-v` assignment and never
# interpolated into the program: it is attacker-controlled text on a fork
# pull request, and `-v` would read a backslash in it as an escape.
#
# Two records come back, on stdout, prefixed so a section title can never
# be mistaken for one of them: `#sections <n>` once, and `#empty <name>`
# per section standing over nothing.
#
# **Fences are tracked, because `^## ` inside one is content.** A body
# quoting a template — this repository's own bodies do — would otherwise
# be split at a heading that is being shown rather than used, and the
# check would invent an empty section out of the quotation.
report=$(printf '%s\n' "$body" | awk '
  {
    line = $0
    sub(/\r$/, "", line)     # a CRLF body is the same body

    # The comment stripper, character-run by character-run rather than a
    # per-line regex: an HTML comment spans lines, so the state that says
    # "still inside one" has to survive the record. What survives the
    # strip is `vis` — what a reader would actually see on the page,
    # which is the only text that can be an answer.
    rest = line
    vis = ""
    while (length(rest) > 0) {
      if (incomment) {
        p = index(rest, "-->")
        if (p == 0) { rest = "" }
        else { rest = substr(rest, p + 3); incomment = 0 }
      } else if (infence) {
        # No comment parsing inside a fence: what is fenced is shown, not
        # interpreted, and a fenced `<!--` that never closes would
        # swallow every section after it.
        vis = vis rest
        rest = ""
      } else {
        p = index(rest, "<!--")
        if (p == 0) { vis = vis rest; rest = "" }
        else {
          vis = vis substr(rest, 1, p - 1)
          rest = substr(rest, p + 4)
          incomment = 1
        }
      }
    }

    # A heading, and only outside a fence. `^## ` is the shape, plus the
    # bare `##` a writer leaves behind when the title goes; `###` and
    # deeper are content of the section they sit in, which is what they
    # are on the rendered page too.
    if (!infence && (vis ~ /^## / || vis ~ /^##[ \t]*$/)) {
      name = vis
      sub(/^##[ \t]*/, "", name)
      sub(/[ \t]+$/, "", name)
      nsec++
      title[nsec] = (name == "" ? "(untitled)" : name)
      filled[nsec] = 0
      next
    }

    # Anything visible under a heading is an answer. Length is not the
    # test and quality never was: "nothing runnable ships here" is a
    # complete answer to how to test, and so is one word.
    if (nsec > 0) {
      t = vis
      gsub(/[ \t]/, "", t)
      if (t != "") filled[nsec] = 1
    }

    # The fence toggle goes last, so the line that opens a fence is
    # counted as the content it is.
    f = vis
    sub(/^[ \t]*/, "", f)
    if (f ~ /^```/ || f ~ /^~~~/) infence = !infence
  }
  END {
    print "#sections " nsec + 0
    for (i = 1; i <= nsec; i++) if (!filled[i]) print "#empty " title[i]
  }
')

sections=$(printf '%s\n' "$report" | sed -n 's/^#sections //p')
empty=$(printf '%s\n' "$report" | sed -n 's/^#empty //p')

# **No sections is a complete answer, not a silent one.** Someone
# replaced the template wholesale; nothing is carried, so nothing is
# owed — and the line says which of the two happened, because a check
# that passes without a word is indistinguishable from one that did not
# run.
if [ "${sections:-0}" -eq 0 ]; then
  echo "#${PR}'s body carries no '## ' section — nothing is owed."
  echo "What this reads is the headings a body has, never a list of the"
  echo "headings it could have (${DOC})."
  exit 0
fi

if [ -n "$empty" ]; then
  echo "#${PR} is not a draft, and these sections of its body stand over" >&2
  echo "nothing — a seeded comment, or an empty heading:" >&2
  while IFS= read -r name; do
    [ -n "$name" ] || continue
    echo "  ## ${name}" >&2
  done <<EOF
$empty
EOF
  echo "" >&2
  echo "Marking a pull request ready is the claim that a reviewer can read" >&2
  echo "its body. Answer each section named above, or delete the ones this" >&2
  echo "change has no use for — a body owes nothing for a heading it does" >&2
  echo "not carry. An answer of one line is an answer; an empty section and" >&2
  echo "a forgotten one look identical, which is the whole of it" >&2
  echo "(${DOC})." >&2
  exit 1
fi

echo "Every one of #${PR}'s ${sections} sections is answered."
