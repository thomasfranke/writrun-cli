#!/usr/bin/env bash
. "$(dirname "$0")/../../harness.sh"
. "$(dirname "$0")/../../cli_lib.sh"

# The screens, driven through a real terminal.
#
# Every other case about the screens drives the model — keys in, actions
# out — and a model answers correctly while the terminal underneath it
# does not. Three defects reached a person that way: the screens stacked
# instead of replacing, their colours were a canvas's rather than the
# reader's, and a session that reopened the screen left a second program
# holding the keyboard, so a confirmation answered itself. None of them
# was visible to a case that never opened a pty.
#
# This tier opens one. It reads only — it moves, it crosses between the
# two screens, and it leaves — because a case that took a task would be
# working the queue it is testing against.
#
# It catches the first and the third: the alternate screen is asked for,
# and `esc` returns rather than leaving. The colours it does not check,
# and nothing else does either — a case that pinned them would be
# asserting escape codes, which says a terminal library works and not
# that a reader can read the screen. That one is still held by looking.
#
# expect drives the pty. CI installs it; a session without it gets a
# named skip rather than a green that proved nothing.
if ! command -v expect >/dev/null 2>&1; then
  echo "ok    the screen answers keys (skipped: expect is not installed)"
  finish
fi

cd "$REPO_ROOT" || exit 1

drive() {
  expect -f - "$WRITRUN" <<'EXPECT'
set timeout 15
set binary [lindex $argv 0]
# A terminal with a size, because a screen is what fills one.
set stty_init "rows 40 columns 100"
spawn -noecho $binary

# The alternate screen, first: a screen replaces what is on the terminal
# and gives it back on the way out. Without this the screens print one
# under another and the reader scrolls through the last one to see this
# one — which is how that defect reached a person.
#
# `-ex` because the sequence is matched as itself: expect globs by
# default, and `[?` there opens a character set instead of naming two
# characters, so a glob would fail on a screen that did ask.
expect {
  -ex "\033\[?1049h" {}
  timeout            { puts "\nFAIL: the screen never asked for the alternate screen"; exit 9 }
  eof                { puts "\nFAIL: the screen never asked for the alternate screen (the screen left first)"; exit 9 }
}

# The entry screen: the groups the drawing gives, and its own footer.
expect {
  "TASKS" {}
  timeout { puts "\nFAIL: the entry screen never named its first group"; exit 10 }
  eof     { puts "\nFAIL: the entry screen never named its first group (the screen left first)"; exit 10 }
}
expect {
  "enter run" {}
  timeout { puts "\nFAIL: the entry screen never drew its footer"; exit 11 }
  eof     { puts "\nFAIL: the entry screen never drew its footer (the screen left first)"; exit 11 }
}

# The cursor moves, and the detail line follows it.
send -- "\033\[B"
expect {
  "take — begin a task" {}
  timeout { puts "\nFAIL: down did not move to take"; exit 12 }
  eof     { puts "\nFAIL: down did not move to take (the screen left first)"; exit 12 }
}
send -- "\033\[A"
expect {
  "list — the queue" {}
  timeout { puts "\nFAIL: up did not move back to list"; exit 13 }
  eof     { puts "\nFAIL: up did not move back to list (the screen left first)"; exit 13 }
}

# enter on `list` opens the queue, which has its own footer and a way back.
send -- "\r"
expect {
  "esc back" {}
  timeout { puts "\nFAIL: enter on list did not open the queue"; exit 14 }
  eof     { puts "\nFAIL: enter on list did not open the queue (the screen left first)"; exit 14 }
}

# esc returns to the entry screen rather than leaving or dispatching.
send -- "\033"
expect {
  "enter run" {}
  timeout { puts "\nFAIL: esc did not return to the entry screen"; exit 15 }
  eof     { puts "\nFAIL: esc did not return to the entry screen (the screen left first)"; exit 15 }
}

# The session: a command runs and the screen comes back.
#
# This is the whole of what a session is, and the only tier that can
# show it. `tea.Exec` releases the terminal so the command can read it,
# and releasing means cancelling the read the program already has in
# flight — which is possible on a terminal and not on an `io.Reader`, so
# a case with a fake reader would race itself rather than test this.
#
# `status` because it only reads. It is also the command that was
# reported as locking the CLI, which is the same defect from the other
# side: it ran, printed, and left nothing to come back to.
send -- "\033\[B"
send -- "\033\[B"
send -- "\033\[B"
expect {
  "status — " {}
  timeout     { puts "\nFAIL: down three times did not reach status"; exit 17 }
  eof         { puts "\nFAIL: down three times did not reach status (the screen left first)"; exit 17 }
}
send -- "\r"
# A command is a screen like the others: it takes the whole terminal
# rather than printing into the scrollback under the one it came from.
# `tea.Exec` gives the terminal back on the normal buffer, so this is
# asked for by hand and has to be asked for again here.
expect {
  -ex "\033\[?1049h" {}
  timeout            { puts "\nFAIL: the command printed inline instead of taking the screen"; exit 20 }
  eof                { puts "\nFAIL: the command closed the CLI instead of taking the screen"; exit 20 }
}
# What is running, before it runs. Expected here, ahead of the command's
# own words, because that ordering is the whole point: a command that
# reaches the forge takes seconds, and an empty terminal in the meantime
# is indistinguishable from a hung one.
expect {
  "running status" {}
  timeout          { puts "\nFAIL: the screen never said what it was running"; exit 21 }
  eof              { puts "\nFAIL: the screen never said what it was running (it left)"; exit 21 }
}
# status reads the queue and asks the forge, so it is given room.
set timeout 90
expect {
  "enter to return" {}
  timeout           { puts "\nFAIL: the command never offered the way back"; exit 18 }
  eof               { puts "\nFAIL: the command closed the CLI instead of returning"; exit 18 }
}
set timeout 15
send -- "\r"
expect {
  "enter run" {}
  timeout     { puts "\nFAIL: the screen did not come back after the command"; exit 19 }
  eof         { puts "\nFAIL: the screen did not come back after the command (it left)"; exit 19 }
}

# q leaves, and leaving is not a failure.
send -- "q"
expect {
  eof     {}
  timeout { puts "\nFAIL: q did not leave the screen"; exit 16 }
}
catch wait result
exit [lindex $result 3]
EXPECT
}

check "the screen answers keys, and the two screens replace one another" 0 "" \
  -- drive

finish
