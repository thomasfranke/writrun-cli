#!/usr/bin/env bash
. "$(dirname "$0")/../../harness.sh"
. "$(dirname "$0")/../../cli_lib.sh"

# The first run, driven through a real terminal.
#
# Where `.writrun/` is absent the binary used to print thirteen commands
# of which twelve refuse to run there. It opens a screen instead: the
# wordmark, what the environment answers, and `init`
# (docs/product/screens/first-run.excalidraw, spec-0038).
#
# Only a terminal can be asked whether a screen was drawn, so this tier
# opens one. It reads only — it looks and it leaves — because a case
# that pressed `enter` would adopt the repository it is testing in.
#
# expect drives the pty. CI installs it; a session without it gets a
# named skip rather than a green that proved nothing.
if ! command -v expect >/dev/null 2>&1; then
  echo "ok    the first run opens where the kit is absent (skipped: expect is not installed)"
  finish
fi

# A git repository with no kit in it — which is what every repository
# looks like the moment before it is adopted.
BARE="$WORK/unadopted"
mkdir -p "$BARE"
git -c init.defaultBranch=main init -q "$BARE"
git -C "$BARE" -c user.name=suite -c user.email=suite@test commit -q --allow-empty -m "first"

drive() {
  expect -f - "$WRITRUN" "$BARE" <<'EXPECT'
set timeout 15
set binary [lindex $argv 0]
set repo [lindex $argv 1]
set stty_init "rows 40 columns 100"
cd $repo
spawn -noecho $binary

expect {
  -ex "\033\[?1049h" {}
  timeout            { puts "\nFAIL: the first run never asked for the alternate screen"; exit 9 }
  eof                { puts "\nFAIL: the binary printed the help instead of opening a screen"; exit 9 }
}

# The wordmark, drawn here and on no other screen: this is the one
# screen a reader reaches without having adopted anything.
expect {
  "W R I T R U N" {}
  timeout { puts "\nFAIL: the first run never introduced the product"; exit 10 }
  eof     { puts "\nFAIL: the first run left before introducing the product"; exit 10 }
}
expect {
  "NO KIT HERE" {}
  timeout { puts "\nFAIL: the context line never said no kit is here"; exit 11 }
  eof     { puts "\nFAIL: the screen left before naming its context"; exit 11 }
}

# The environment, answered with doctor's own marks, and `init` as the
# one adoption row.
expect {
  "ENVIRONMENT" {}
  timeout { puts "\nFAIL: the environment was never answered"; exit 12 }
  eof     { puts "\nFAIL: the screen left before answering the environment"; exit 12 }
}
expect {
  "init " {}
  timeout { puts "\nFAIL: init was never offered"; exit 13 }
  eof     { puts "\nFAIL: the screen left before offering init"; exit 13 }
}

# A screen with nowhere to go back to offers `q quit` alone.
expect {
  "q quit" {}
  timeout { puts "\nFAIL: the first run drew no footer"; exit 14 }
  eof     { puts "\nFAIL: the screen left before drawing its footer"; exit 14 }
}

send -- "q"
expect {
  eof     {}
  timeout { puts "\nFAIL: q did not leave the first run"; exit 15 }
}
catch wait result
exit [lindex $result 3]
EXPECT
}

check "the first run opens where the kit is absent, and q leaves it" 0 "" \
  -- drive

finish
