#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# The title is the human's free text, and until now nothing judged it
# before the push: the branch reached the forge and the door refused it
# there. The same script now runs over the composition, where a refusal
# costs nothing (spec-0023).
make_repo
authoring_change
cd "$TARGET" || exit 1

check "a title in neither style is refused" 1 "does not read as the declared" \
  -- author --title "just a sentence about the rule" --yes
check "nothing was pushed" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/derived-work
check "the forge was never reached" 1 "" \
  -- grep -q . "$GH_LOG"

# The same call judges the type against the kit's own vocabulary.
check "a type outside the vocabulary is refused" 1 "outside the vocabulary" \
  -- author --title "[Wibble] The declaration is the section" --yes
check "still nothing on the forge" 1 "" \
  -- grep -q . "$GH_LOG"

# And the title the declared style accepts goes through the same door,
# so the two refusals are the vocabulary's doing and not the check's
# silence.
check "the declared style is opened" 0 "pull request open and ready" \
  -- author --title "$TITLE" --yes

finish
