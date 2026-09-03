#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# The cut ends at the forge, which runs after the push — an unusable gh
# must abort before anything is mutated, not fail after the tag is
# already public.
release_setup
printf '#!/usr/bin/env bash\nexit 1\n' > "$WORK/stub-bin/gh"
chmod +x "$WORK/stub-bin/gh"

out=$(bash "$RELEASE_SH" 2>&1); code=$?
if [ "$code" -ne 0 ] &&
   printf '%s' "$out" | grep -q 'gh must be installed and authenticated' &&
   [ "$(cat VERSION)" = "v0.0.00" ] &&
   [ -z "$(git tag --list)" ] &&
   [ ! -f "$WORK/calls.log" ]; then
  echo "ok    an unusable gh aborts with nothing mutated"; pass=$((pass + 1))
else
  echo "FAIL  an unusable gh aborts with nothing mutated"
  printf '%s\n' "$out" | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
