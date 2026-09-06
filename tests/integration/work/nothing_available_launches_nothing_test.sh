#!/usr/bin/env bash
. "$(dirname "$0")/../../work_lib.sh"

# With nothing available the command says so in the lister's own words
# and launches nothing (spec-0007, acceptance criteria).
make_repo
configure_agent
cd "$TARGET" || exit 1

# Every task held back: the two ready ones lose their approval.
sed -i.bak 's/^status: approved$/status: draft/' work/specs/spec-0001-a-thing.md
sed -i.bak 's/^status: approved$/status: draft/' work/specs/spec-0002-another-thing.md
rm -f work/specs/*.bak

check "the lister's own answer is shown" 1 "Nothing is available." -- work
check "nothing was launched" 0 "" -- test ! -s "$AGENT_LOG"

# A repository the methodology was never installed in has no queue to
# select from at all.
mkdir -p "$WORK/plain"
cd "$WORK/plain" || exit 1
git_q init -q
check "work runs only where .writrun/ is" 1 "not an adopted repository" -- "$WRITRUN" work

finish
