#!/usr/bin/env bash
. "$(dirname "$0")/../../init_lib.sh"

# Adopt at one tag, update to the next: the kit-owned paths move, and
# what the project owns is byte-identical afterwards (spec-0003,
# acceptance criteria).
make_target "$TARGET"
cd "$TARGET" || exit 1
"$WRITRUN" init --stage 1 --yes > /dev/null 2>&1 || { echo "FAIL  the fixture could not adopt"; exit 1; }

# The project answers its gates and inverts the deriving default —
# these are the lines the refresh must carry across.
sed -i.bak 's#| Writing docs | <!-- TODO — default: human reviews --> |#| Writing docs | Thomas reviews before merge. |#' AGENTS.md
sed -i.bak 's#Present the derived tasks in the session before opening the PR.#Open the derived pull request directly.#' AGENTS.md
rm -f AGENTS.md.bak

# The project's own record, which the refresh may not reach.
printf 'id: task-0001\n' > work/tasks/task-0001-a-task.md
printf '# Our own chapter\n' > docs/product/a-chapter.md
printf '# Our commits\n' > .writrun/conventions/commits.md

age_kit "$TARGET"
git_q add -A && git_q commit -q -m "chore: the kit, one release back"

for f in work/tasks/task-0001-a-task.md docs/product/a-chapter.md \
         .writrun/conventions/commits.md .writrun/settings.json; do
  cp "$f" "$WORK/$(basename "$f").before"
done

check "update reports the move and refreshes" 0 "$OLD_TAG → $TAG" \
  -- "$WRITRUN" update --yes

check "the tag is recorded" 0 "$TAG" -- cat .writrun/VERSION
check "the refreshed skill landed" 0 "reworded" -- cat .writrun/skills/writrun-select-next-task/SKILL.md
check "a file the new tag adds is written" 0 "Spec template" -- cat .writrun/templates/spec.md
check "a reworded workflow is rewritten" 0 "reworded" -- cat .github/workflows/writrun-check.yml

for f in work/tasks/task-0001-a-task.md docs/product/a-chapter.md \
         .writrun/conventions/commits.md .writrun/settings.json; do
  check "untouched: $f" 0 "" -- cmp "$WORK/$(basename "$f").before" "$f"
done

check "the project's gates answer survived" 0 "Thomas reviews before merge" -- cat AGENTS.md
check "the project's deriving default survived" 0 "Open the derived pull request directly" -- cat AGENTS.md
check "the kit's empty gates row did not come back" 1 "" \
  -- grep -q "TODO — default: human reviews" AGENTS.md

finish
