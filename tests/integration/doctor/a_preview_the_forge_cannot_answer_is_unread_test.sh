#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# A read the preview needs and cannot make is `unread`, never unmet: a
# forge that does not answer has told doctor nothing about the
# repository, and an adopter with no network still exits 0 (spec-0036,
# edge cases).
make_repo 1
cd "$TARGET" || exit 1

# A PATH with no gh on it at all, built rather than hoped for. Naming
# two directories and trusting them to hold no gh is a guess about the
# machine — and a wrong one on a GitHub runner, where gh is installed at
# /usr/bin and the preview then reads one requirement it was meant to
# find unreadable. So the directory is made here, holding a link to
# every tool the examined stages need and to nothing else.
#
# The guard that used to stand here could not work: `env PATH=... command
# -v gh` asks env to execute `command`, which is a shell builtin and not
# a program, so the test never skipped and never noticed.
NOGH="$WORK/nogh-bin"
mkdir -p "$NOGH"
for t in git bash sh awk sed grep cat cut tr sort uniq head tail wc ls \
         mkdir rm mv cp find date env printf test dirname basename; do
  p=$(command -v "$t" 2>/dev/null) || continue
  ln -sf "$p" "$NOGH/$t"
done
nogh() { env -i PATH="$NOGH" HOME="$HOME" "$WRITRUN" "$@"; }

if env PATH="$NOGH" sh -c 'command -v gh' >/dev/null 2>&1; then
  echo "FAIL  the no-gh PATH has a gh on it — this case cannot ask its question"
  fail=$((fail + 1))
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
