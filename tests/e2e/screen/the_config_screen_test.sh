#!/usr/bin/env bash
. "$(dirname "$0")/../../harness.sh"
. "$(dirname "$0")/../../cli_lib.sh"

# `writrun config` as a screen, driven through a real terminal.
#
# The command listed its keys and returned; the drawing
# (docs/product/screens/config.excalidraw) shows a screen with a cursor
# and `enter change`. The maintainer found the difference by trying to
# navigate. Only a terminal can be asked whether a screen is there.
#
# It changes nothing. `enter` opens the question and `esc` cancels it,
# so the case reads the settings of the repository it runs in and
# leaves them exactly as they were — a case that wrote one would be
# editing its own repository.
if ! command -v expect >/dev/null 2>&1; then
  echo "ok    the config screen answers keys (skipped: expect is not installed)"
  finish
fi

cd "$REPO_ROOT" || exit 1

before=$(cat writrun/settings.json)

drive() {
  expect -f - "$WRITRUN" <<'EXPECT'
set timeout 20
set binary [lindex $argv 0]
set stty_init "rows 30 columns 100"
spawn -noecho $binary config

# A screen, not a listing: the alternate buffer and the drawn footer.
#
# `-ex` because the sequence is matched as itself: expect globs by
# default, and `[?` there opens a character set instead of naming two
# characters.
expect {
  -ex "\033\[?1049h" {}
  timeout            { puts "\nFAIL: config printed inline instead of opening a screen"; exit 30 }
  eof                { puts "\nFAIL: config closed instead of opening a screen"; exit 30 }
}
expect {
  "enter change" {}
  timeout        { puts "\nFAIL: the screen never drew the footer the drawing gives"; exit 31 }
  eof            { puts "\nFAIL: the screen left before drawing its footer"; exit 31 }
}

# The cursor moves, and the detail line follows it.
send -- "\033\[B"
expect {
  " — " {}
  timeout { puts "\nFAIL: down did not move, or the detail line did not follow"; exit 32 }
  eof     { puts "\nFAIL: down left the screen"; exit 32 }
}

# `enter` asks for a value, and the question names its own way out.
send -- "\r"
expect {
  "The value for" {}
  timeout         { puts "\nFAIL: enter did not ask for a value"; exit 33 }
  eof             { puts "\nFAIL: enter closed the screen"; exit 33 }
}
expect {
  "esc cancels" {}
  timeout       { puts "\nFAIL: the question never named the way out of it"; exit 34 }
  eof           { puts "\nFAIL: the question left before naming a way out"; exit 34 }
}

# Cancelled, so nothing is written, and the screen comes back.
send -- "\033"
expect {
  "enter to return" {}
  timeout           { puts "\nFAIL: a cancelled change offered no way back"; exit 35 }
  eof               { puts "\nFAIL: a cancelled change closed the CLI"; exit 35 }
}
send -- "\r"
expect {
  "enter change" {}
  timeout        { puts "\nFAIL: the screen did not come back after the question"; exit 36 }
  eof            { puts "\nFAIL: the screen did not come back after the question (it left)"; exit 36 }
}

# q leaves, and leaving is not a failure.
send -- "q"
expect {
  eof     {}
  timeout { puts "\nFAIL: q did not leave the screen"; exit 37 }
}
catch wait result
exit [lindex $result 3]
EXPECT
}

check "the config screen answers keys, and a cancelled change writes nothing" 0 "" \
  -- drive

# The settings this ran against are byte-for-byte what they were.
if [ "$before" != "$(cat writrun/settings.json)" ]; then
  echo "FAIL  the case changed the settings of the repository it ran in"
  fail=$((fail + 1))
else
  echo "ok    the settings are exactly as they were"
  pass=$((pass + 1))
fi

finish
