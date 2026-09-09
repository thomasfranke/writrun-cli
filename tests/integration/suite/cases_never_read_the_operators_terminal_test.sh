#!/usr/bin/env bash
. "$(dirname "$0")/../../harness.sh"

# The suite reads from /dev/null, never from whoever ran it.
#
# A case that asks a question stands a terminal up itself, through
# WRITRUN_TTY_IN. A case checking what happens *without* one is
# asserting about the suite's own stdin — and inherited from a person's
# terminal that case asks its question for real and waits.
#
# It waited for half an hour inside `make release`, where the suite runs
# nested and its output is captured, so the question was asked somewhere
# nobody could see it. The release could not end and could not say why.
#
# This drives the runner rather than reading it: a grep for `/dev/null`
# would pass on a line that had been commented out.

# expect stands the terminal up. Where it is absent the case says so by
# name rather than failing: a runner without it is not a broken suite,
# and `release readiness` is such a runner
# (tests/e2e/screen/ skips the same way).
if ! command -v expect >/dev/null 2>&1; then
  echo "ok    a case gets /dev/null (skipped: expect is not installed)"
  finish
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/tests/probe/keyboard"

# A case that hangs if it is given a terminal, and passes if it is not.
cat > "$work/tests/probe/keyboard/reads_stdin_test.sh" <<'CASE'
#!/usr/bin/env bash
if [ -t 0 ]; then
  # A real suite would ask its question here and never return.
  echo "FAIL  the case was handed a terminal"
  exit 1
fi
echo "ok    the case reads no terminal"
exit 0
CASE
chmod +x "$work/tests/probe/keyboard/reads_stdin_test.sh"

cp tests/run.sh tests/harness.sh "$work/tests/"

# Run the runner with a real terminal on its own stdin, which is what a
# person at a keyboard gives it.
out=$(cd "$work" && expect -c '
  set timeout 30
  spawn -noecho bash tests/run.sh
  expect {
    "case files passed" {}
    timeout { puts "TIMEOUT" }
    eof {}
  }
  expect eof' 2>&1)

if printf '%s\n' "$out" | grep -q 'the case reads no terminal'; then
  echo "ok    a case gets /dev/null even when the suite was run from a terminal"
  pass=$((pass + 1))
else
  echo "FAIL  a case gets /dev/null even when the suite was run from a terminal"
  printf '%s\n' "$out" | sed 's/^/      | /' | head -8
  fail=$((fail + 1))
fi

finish
