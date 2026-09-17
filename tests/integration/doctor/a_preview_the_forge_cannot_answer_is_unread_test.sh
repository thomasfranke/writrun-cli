#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# A read the preview needs and cannot make is `unread`, never unmet: a
# forge that does not answer has told doctor nothing about the
# repository, and an adopter with no network still exits 0 (spec-0036,
# edge cases).
make_repo 1
cd "$TARGET" || exit 1

# A PATH with no gh on it at all — neither the stub nor whatever the
# machine running the suite has installed. No forge read can be
# attempted, so every previewed requirement is unknown rather than
# unsatisfied.
BARE_PATH=/usr/bin:/bin
nogh() { env PATH="$BARE_PATH" "$WRITRUN" "$@"; }

if command -v gh >/dev/null 2>&1 && env PATH="$BARE_PATH" command -v gh >/dev/null 2>&1; then
  echo "ok    a preview the forge cannot answer is unread (skipped: gh is on $BARE_PATH)"
  finish
fi

nogh doctor > "$WORK/unread.out" 2>&1

check "a preview nobody could make exits 0" 0 "" -- nogh doctor
check "the preview still counts six rows" 0 \
  "Stage 2 — the forge, previewed: 0 of 6 met." -- cat "$WORK/unread.out"
check "the requirement that could not be read says why" 0 \
  "?  gh on the PATH — not on the PATH" -- cat "$WORK/unread.out"
check "every requirement under it is unread" 0 \
  "?  no rule over main refuses the recording push — no forge check was made" \
  -- cat "$WORK/unread.out"
check "the reach line counts them unread, not unmet" 0 \
  "Stage 2 is not within reach: 6 unread." -- cat "$WORK/unread.out"
check "the declared stage still holds" 0 "Every assumption up to stage 1 holds." \
  -- cat "$WORK/unread.out"

finish
