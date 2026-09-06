#!/usr/bin/env bash
. "$(dirname "$0")/../../work_lib.sh"

# writrun never guesses which agent is installed: with the key unset the
# command aborts showing the exact line that sets it, and nothing is
# launched (spec-0007, acceptance criteria).
make_repo
cd "$TARGET" || exit 1

check "the abort names the git config line" 1 "git config writrun.agent" \
  -- work

check "nothing was launched" 0 "" -- test ! -s "$AGENT_LOG"

# A key set to whitespace is no answer either.
git_q config writrun.agent "   "
check "an empty value is no agent at all" 1 "git config writrun.agent" -- work
check "still nothing was launched" 0 "" -- test ! -s "$AGENT_LOG"

finish
