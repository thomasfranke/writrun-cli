#!/usr/bin/env bash
# commit_subject.sh — prints the subject of a commit the machinery makes.
#
# Usage: commit_subject.sh <merge|forge|intake>
#
#   commit_subject.sh merge    what `writrun approve` records after a merge
#   commit_subject.sh forge    what `writrun progress` records from an event
#   commit_subject.sh intake   what `writrun intake` records from a label
#
# **The subject is a constant, and `pr_title_style` is not consulted.**
# That key governs the pull request title, which is read in a queue by
# the people working it; these commits land on `main`, which is read by
# bisect, by release tooling and by whoever arrives a year later — an
# audience that is the same in every project, and is not served by a
# per-project grammar. So the grammar here is Conventional Commits
# everywhere (docs/technical/decisions/pull-requests/0063-title-and-subject-are-two-texts.md).
#
# **The file still exists because three workflows write these commits.**
# One writer, three callers, no drift: a literal in each workflow would
# be three places to edit and nothing squashes these, so a subject that
# diverged would sit on `main` for good. It is not checked anywhere, and
# could not usefully be: `writrun check` reads pull request titles at the
# door, and these commits pass no door.
#
# The scope is `queue` — these commits record what happened to `work/`,
# and nothing else (.writrun/conventions/commits.md).
#
# Exit codes: 0 always, except 3 for a usage error.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

EVENT="${1:-}"
case "$EVENT" in
  merge)  printf 'chore(queue): record what the merge decided\n' ;;
  forge)  printf 'chore(queue): record what the forge just did\n' ;;
  intake) printf 'chore(queue): record what the label let in\n' ;;
  *) echo "usage: commit_subject.sh <merge|forge|intake>" >&2; exit 3 ;;
esac
