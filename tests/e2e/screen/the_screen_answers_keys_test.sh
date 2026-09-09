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

# The session, first way: a command that asks nothing is captured, and
# its answer becomes a screen.
#
# The program never gives up the terminal for it, which is what buys
# both the spinner while it waits and the scrolling afterwards. `status`
# because it only reads — and because it is the command that was
# reported as locking the CLI, which is this same path before it
# existed.
send -- "\033\[B"
send -- "\033\[B"
send -- "\033\[B"
expect {
  "status — " {}
  timeout     { puts "\nFAIL: down three times did not reach status"; exit 17 }
  eof         { puts "\nFAIL: down three times did not reach status (the screen left first)"; exit 17 }
}
send -- "\r"
# The spinner, before the answer. A command that reaches the forge takes
# seconds, and a still screen in the meantime is indistinguishable from
# a hung one.
expect {
  "running status" {}
  timeout          { puts "\nFAIL: the screen never said what it was running"; exit 21 }
  eof              { puts "\nFAIL: the screen never said what it was running (it left)"; exit 21 }
}
# status reads the queue and asks the forge, so it is given room.
set timeout 90
expect {
  "esc back" {}
  timeout    { puts "\nFAIL: the command's answer never became a screen"; exit 18 }
  eof        { puts "\nFAIL: the command closed the CLI instead of answering"; exit 18 }
}
set timeout 15
send -- "\033"
expect {
  "enter run" {}
  timeout     { puts "\nFAIL: the screen did not come back from the answer"; exit 19 }
  eof         { puts "\nFAIL: the screen did not come back from the answer (it left)"; exit 19 }
}

# The session, second way: a command that asks keeps the terminal, and
# is cancelled here rather than answered.
#
# There is no capturing a question: it would wait on a reader who
# cannot see it. So this half runs on the released terminal, takes the
# alternate screen by hand, and hands it back when the reader has read
# it.
#
# It is cancelled, never completed: a case that filed a report would be
# writing into the queue it is testing against.
# Four rows on from `status`: finish, author, amend, report.
send -- "\033\[B"
send -- "\033\[B"
send -- "\033\[B"
send -- "\033\[B"
expect {
  "report — " {}
  timeout     { puts "\nFAIL: down four times did not reach report"; exit 22 }
  eof         { puts "\nFAIL: down four times did not reach report (the screen left)"; exit 22 }
}
send -- "\r"
expect {
  -ex "\033\[?1049h" {}
  timeout            { puts "\nFAIL: the asking command printed inline instead of taking the screen"; exit 20 }
  eof                { puts "\nFAIL: the asking command closed the CLI"; exit 20 }
}
expect {
  "running report" {}
  timeout          { puts "\nFAIL: the screen never said it was running report"; exit 27 }
  eof              { puts "\nFAIL: the screen never said it was running report (it left)"; exit 27 }
}
expect {
  "What did you notice" {}
  timeout               { puts "\nFAIL: report never asked its first question"; exit 23 }
  eof                   { puts "\nFAIL: report closed the CLI instead of asking"; exit 23 }
}
# The question names the way out. huh spells its footer from the field's
# own bindings and leaves the form's `ctrl+c` unnamed, so a reader who
# did not want to answer had no key they could see — which is how this
# command came to be reported as having no way out. Only a terminal can
# be asked whether it was drawn: huh writes nothing to anything else.
expect {
  "esc cancels" {}
  timeout       { puts "\nFAIL: the question never named the way out of it"; exit 26 }
  eof           { puts "\nFAIL: the question left before naming a way out"; exit 26 }
}
send -- "\033"
expect {
  "enter to return" {}
  timeout           { puts "\nFAIL: a cancelled question offered no way back"; exit 24 }
  eof               { puts "\nFAIL: a cancelled question closed the CLI"; exit 24 }
}
send -- "\r"
expect {
  "enter run" {}
  timeout     { puts "\nFAIL: the way back did not hear the return after a question"; exit 25 }
  eof         { puts "\nFAIL: the screen did not come back after a question"; exit 25 }
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
