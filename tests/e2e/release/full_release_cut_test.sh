#!/usr/bin/env bash
. "$(dirname "$0")/../../harness.sh"

# The whole cut, for real: this repository's working tree copied whole,
# `make release patch` run with the real make — so the real suite
# executes inside — and the commit, the tag, and the push land on a
# local bare origin. The forge is the one fake: `gh` is a PATH stub.
# The env guard below is what stops the nested suite from cutting a
# release of its own.
if [ -n "${WHITRUN_CLI_E2E_RUNNING:-}" ]; then
  echo "ok    full release cut (skipped: nested inside another e2e run)"
  exit 0
fi

WORK=$(mktemp -d)
mkdir -p "$WORK/repo"
(cd "$REPO_ROOT" && tar -cf - --exclude .git --exclude .claude .) | tar -xf - -C "$WORK/repo"
cd "$WORK/repo" || exit 1
git init -q .
git symbolic-ref HEAD refs/heads/main
git config user.email t@example.com
git config user.name Test
git add -A >/dev/null
git commit -qm baseline
git tag -a v0.1.0 -m v0.1.0
git init -q --bare "$WORK/origin.git"
git remote add origin "$WORK/origin.git"
git push -q origin main
mkdir -p "$WORK/stub-bin"
# The fake forge accepts only what this e2e is about — `gh release …`
# and the auth guard — and fails everything else.
printf '#!/usr/bin/env bash\ncase "${1:-}" in release|auth) ;; *) exit 1 ;; esac\necho "gh $*" >> "%s/gh.log"\n' "$WORK" > "$WORK/stub-bin/gh"
chmod +x "$WORK/stub-bin/gh"
export PATH="$WORK/stub-bin:$PATH"
export WHITRUN_CLI_E2E_RUNNING=1
unset MAKE

out=$(make release patch 2>&1); code=$?
if [ "$code" -eq 0 ] &&
   git log -1 --format=%s | grep -q 'chore(release): v0.1.1' &&
   git tag --list | grep -qx 'v0.1.1' &&
   grep -q '^## v0.1.1' CHANGELOG.md &&
   git show --stat --format= HEAD | grep -q 'CHANGELOG.md' &&
   git ls-remote --tags origin | grep -q 'refs/tags/v0.1.1' &&
   grep -q 'gh release create v0.1.1' "$WORK/gh.log"; then
  echo "ok    a real make release cuts, tests, commits, tags, pushes, publishes"; pass=$((pass + 1))
else
  echo "FAIL  a real make release cuts, tests, commits, tags, pushes, publishes"
  printf '%s\n' "$out" | grep -A4 '^FAIL\|^release:' | sed 's/^/      | /'
  printf '%s\n' "$out" | tail -4 | sed 's/^/      | /'
  fail=$((fail + 1))
fi

finish
