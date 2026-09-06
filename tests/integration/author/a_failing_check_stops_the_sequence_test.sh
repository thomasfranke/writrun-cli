#!/usr/bin/env bash
. "$(dirname "$0")/../../author_lib.sh"

# The three checks run in the one order the methodology fixed, and the
# first non-zero verdict is the whole answer: nothing later runs, no
# branch is cut, nothing reaches the forge (spec-0009, step 1 and
# acceptance criteria).
make_repo

# 1 — the front matter. A quoted value is not the canonical form, and
# the check says so before any other reader is asked anything.
on_branch docs/broken-front-matter
rule_written
work_derived
sed -i.bak 's|^doc_ref: product/rules.md$|doc_ref: "product/rules.md"|' \
  "$TARGET/work/tasks/task-0001-derived-work.md"
rm -f "$TARGET/work/tasks/task-0001-derived-work.md.bak"
committed "docs(product): the declaration is the section"
cd "$TARGET" || exit 1

check "the front-matter check refuses in its own words" 1 "MALFORMED" \
  -- author --title "$TITLE" --yes
check "the doc-shapes check never ran" 1 "" \
  -- grep -q "shown shape(s)" "$AUTHOR_OUT"
check "the state check never ran" 1 "" \
  -- grep -q "no forbidden lifecycle transition" "$AUTHOR_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q . "$GH_LOG"

# 2 — the shapes a doc shows. The front matter passed, so this is the
# second verdict and the state check is the one that must not run.
git_q -C "$TARGET" checkout -q main
on_branch docs/broken-shape
rule_written
work_derived
cat >> "$TARGET/docs/product/rules.md" <<'DOC'

```yaml
---
id: task-0002
status: backlog
---
```
DOC
committed "docs(product): the declaration is the section"

check "the doc-shapes check refuses in its own words" 1 "REJECTED" \
  -- author --title "$TITLE" --yes
check "the front-matter check ran first" 0 "" \
  -- grep -q "all canonical" "$AUTHOR_OUT"
check "the state check never ran" 1 "" \
  -- grep -q "no forbidden lifecycle transition" "$AUTHOR_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q . "$GH_LOG"

# 3 — the lifecycle. No legitimate path produces a spec born past both
# gates, so a branch that adds one is refused at the last check.
git_q -C "$TARGET" checkout -q main
on_branch docs/born-implemented
rule_written
task_file 0001 derived-work backlog spec-0001 "Declare the derived work"
spec_file 0001 derived-work task-0001 implemented "The declaration is the section"
committed "docs(product): the declaration is the section"

check "the state check refuses in its own words" 1 "spec-0001" \
  -- author --title "$TITLE" --yes
check "the two checks before it ran" 0 "" \
  -- grep -q "shown shape(s)" "$AUTHOR_OUT"
check "the forge was never reached" 1 "" \
  -- grep -q . "$GH_LOG"
check "no branch reached the origin" 1 "" \
  -- git -C "$ORIGIN" rev-parse --verify --quiet refs/heads/docs/born-implemented

finish
