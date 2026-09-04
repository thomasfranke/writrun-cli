#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# With --stage given, nothing is asked and that stage is written
# (spec-0002, acceptance criteria). There is no terminal here at all,
# so a question would abort rather than hang — completing is the proof
# none was asked.
make_target "$TARGET"
cd "$TARGET" || exit 1

check "init completes without a terminal under --stage and --yes" 0 "" \
  -- "$WRITRUN" init --stage 2 --yes
check "the settings record the given stage" 0 '"stage": 2' \
  -- cat .writrun/settings.json

finish
