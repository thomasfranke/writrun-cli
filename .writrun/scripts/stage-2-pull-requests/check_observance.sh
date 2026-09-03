#!/usr/bin/env bash
# check_observance.sh — the settings an agent is told to obey, checked
# where disobedience leaves a trace
# (docs/technical/README.md#observance-is-checked-where-it-leaves-a-trace).
#
# Usage: check_observance.sh <diff-range>
#   The PR title and body arrive via $PR_TITLE and $PR_BODY — through the
#   environment, never inline interpolation: both are attacker-controlled
#   text on a fork PR. Run from the repository root.
#
# Two verifications, and both exist because the flag they enforce is
# otherwise pure trust:
#
#   1. **The title obeys `stage_2.pr_title_style`.** The style is the
#      adopter's choice and the title is the whole of what that key
#      governs, so a title in the other style is the one place the
#      disobedience shows — the queue the project decided would read one
#      way. What the squash lands on `main` is a separate text and a
#      constant; this check never sees it.
#
#   2. **`stage_2.agent_coauthor` is honoured in both directions.** At
#      `false`, a co-author trailer, a session link or a generated-with
#      line in a commit message or the pull request body is exactly the
#      write the flag forbids. At `true`, the flag obliges an artifact
#      with a fixed shape and a fixed place — a `Co-Authored-By:` trailer
#      naming the model — so its absence is visible too.
#
# **What leaves no trace stays instruction-bound.** The three conduct
# flags — `auto_commit`, `auto_push`, `auto_pr` — govern whether the
# agent *asked* before acting, and no diff can show a question that
# wasn't asked. They are not checked, and this script does not pretend
# otherwise: a check that guessed at them would fail honest agents and
# pass the dishonest one that committed silently.
#
# **The machinery's own recording commits are never judged.** They are
# not an agent's action, so no conduct flag reaches them; they are
# skipped by their committer identity, which the workflows set and
# nothing else in a pull request's range carries.
#
# Exit 0: the range and the title observe what the settings declare.
# Exit 1: they do not, with every offence named.
# Exit 3: usage error, or git could not be read.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail
RANGE="${1:?usage: check_observance.sh <diff-range>}"

HERE="$(cd "$(dirname "$0")" && pwd)"
READ_SETTING="$HERE/read_setting.sh"

faults=0
fault() { echo "REJECTED: $*" >&2; faults=$((faults + 1)); }

# The committer both recording workflows write their commits as. Matching
# on the identity rather than the subject is deliberate: the subject is a
# constant (commit_subject.sh) and is therefore text like any other,
# while what makes these commits exempt is who wrote them — and only the
# identity says that.
BOT_COMMITTER="github-actions[bot]"

# git_read <label> <git-args...> — runs git and leaves its stdout in
# GIT_OUT. On failure it prints what git said and exits 3, because a
# check that could not read its input must never report the empty result
# as a clean one: `$(git … || true)` yields the same empty string whether
# nothing matched or nothing ran, and this is a gate.
#
# **Never call this inside a command substitution.** The `exit` would end
# only the subshell, and the caller would go on reading the empty value
# this exists to prevent.
GIT_OUT=""
git_read() {
  local label="$1" err
  shift
  err=$(mktemp "${TMPDIR:-/tmp}/writrun-git.XXXXXX")
  if ! GIT_OUT=$(git "$@" 2>"$err"); then
    echo "${label} failed:" >&2
    head -n 2 "$err" >&2
    rm -f "$err"
    exit 3
  fi
  rm -f "$err"
}

# ---------------------------------------------------------------- title

STYLE=$(bash "$READ_SETTING" stage_2.pr_title_style)
TITLE="${PR_TITLE:-}"

# The vocabularies, as conventions/commits.md spells them. They are the
# adopter's to edit there; this list is the machine half of the same
# statement, and the two are kept in step by hand — the file is prose an
# agent reads, not a format a script can parse.
TYPES="docs feat fix refactor chore"
SCOPES="about product technical tasks specs skills ci tests agents readme setup queue conventions"

# The `[TASK-NNNN]` tag is not the settable part: one bracket per task,
# uppercase, no separator, leading the title. Stripping it leaves the
# summary, which is what the style governs. Authoring and reporting
# titles carry no tag, so for them the summary is the whole title — the
# same grammar, checked tagless.
summary="$TITLE"
while :; do
  case "$summary" in
    \[TASK-[0-9][0-9][0-9][0-9]\]*) summary="${summary#\[TASK-????\]}" ;;
    *) break ;;
  esac
done
# One space is allowed between the last tag and the summary, and only
# because a title with none reads as one word to a human scanning it.
summary="${summary# }"

in_list() {   # in_list <needle> <space-separated haystack> — case-folded
  local n
  n=$(printf '%s' "$1" | tr 'A-Z' 'a-z')
  case " $2 " in *" $n "*) return 0 ;; esac
  return 1
}

if [ -z "$TITLE" ]; then
  echo "No PR title given — the title check needs \$PR_TITLE and has none."
elif [ -z "$summary" ]; then
  fault "the title is nothing but task tags: '${TITLE}' — the style '${STYLE}' asks for a summary after them"
else
  case "$STYLE" in
    conventional)
      # type(scope): subject — the scope optional, omitted when a change
      # genuinely spans the repository.
      if printf '%s' "$summary" \
        | grep -qE '^[a-z]+(\([a-z-]+\))?: .+$'; then
        t=$(printf '%s' "$summary" | sed -E 's/^([a-z]+).*/\1/')
        sc=$(printf '%s' "$summary" | sed -nE 's/^[a-z]+\(([a-z-]+)\):.*/\1/p')
        in_list "$t" "$TYPES" \
          || fault "the title's type '${t}' is outside the vocabulary (${TYPES}): '${TITLE}'"
        if [ -n "$sc" ]; then
          in_list "$sc" "$SCOPES" \
            || fault "the title's scope '${sc}' is outside the vocabulary (${SCOPES}): '${TITLE}'"
        fi
      else
        fault "the title does not read as the declared 'conventional' style — 'type(scope): subject' after any task tags: '${TITLE}'"
      fi ;;
    bracketed)
      # [Type][Scope] Sentence — the scope optional for the same reason.
      # Case inside the brackets is not judged: the convention writes
      # `[Docs][Product]` for an implementing title and `[DOCS]` for an
      # authoring one, and a check that picked one would reject the
      # project's own examples.
      if printf '%s' "$summary" \
        | grep -qE '^\[[A-Za-z]+\](\[[A-Za-z-]+\])? .+$'; then
        t=$(printf '%s' "$summary" | sed -E 's/^\[([A-Za-z]+)\].*/\1/')
        sc=$(printf '%s' "$summary" | sed -nE 's/^\[[A-Za-z]+\]\[([A-Za-z-]+)\].*/\1/p')
        in_list "$t" "$TYPES" \
          || fault "the title's type '${t}' is outside the vocabulary (${TYPES}): '${TITLE}'"
        if [ -n "$sc" ]; then
          in_list "$sc" "$SCOPES" \
            || fault "the title's scope '${sc}' is outside the vocabulary (${SCOPES}): '${TITLE}'"
        fi
      else
        fault "the title does not read as the declared 'bracketed' style — '[Type][Scope] Sentence' after any task tags: '${TITLE}'"
      fi ;;
    *)
      # An unreadable style is check_settings.sh's fault to name. Judging
      # the title against a vocabulary nobody declared would fail every
      # honest title for a fault in another file.
      echo "pr_title_style is '${STYLE}', which this check does not know — check_settings.sh names that; no title judged." ;;
  esac
fi

# --------------------------------------------------------------- credit

CREDIT=$(bash "$READ_SETTING" stage_2.agent_coauthor)

# Trailers and whole lines, never a subject: a title that *mentions* a
# trailer ("remove the Co-Authored-By trailer") is prose about the rule,
# not an instance of it. Anchored to the start of a line for the same
# reason. Both directions read these.
CREDIT_LINES='^[[:space:]]*(Co-[Aa]uthored-[Bb]y:|Claude-Session:|Generated-[Bb]y:|Co-authored-by:)'
CREDIT_PROSE='(Generated with \[?Claude|🤖 Generated with|https://claude\.ai/code/session)'
TRAILER='^[[:space:]]*[Cc]o-[Aa]uthored-[Bb]y:'

# **A category is not a model.** `Co-Authored-By: AI` satisfies any
# trailer regex and answers nothing a quarter later, which is the whole
# reason the trailer is worth reading. This is a tripwire and not a
# proof — a name written to evade it evades it, exactly as the core-rule
# stems in check_settings.sh do — and what it catches is the honest
# attempt reaching for a category because nobody said not to.
CATEGORIES="ai an-ai the-ai agent an-agent the-agent bot a-bot the-bot \
assistant an-assistant the-assistant llm an-llm the-llm model a-model \
the-model claude gpt copilot"

# The commits in the range, minus the machinery's own. The committer
# name is read rather than the message, so the skip is by identity and
# no subject text can imitate it. Both directions walk this list.
#
# `A...B` reaches this script written for `git diff`, where it means
# "B since the merge base" — every sibling check is handed the same
# string. `git log` reads the same three dots as the *symmetric
# difference*, so it would also hand back the commits the base gained
# since the branch point: work that landed in another pull request,
# judged as if this one had written it. The two ends are resolved here
# and one side is asked for, the same derivation check_derived_work.sh
# makes when it needs a base.
  case "$RANGE" in
    *...*)
      left="${RANGE%%...*}"
      right="${RANGE##*...}"
      # A merge-base that could not be computed is not a base of
      # "nothing", it is an unanswered question — and this is a gate.
      if ! BASE=$(git merge-base "${left:-HEAD}" "${right:-HEAD}" 2>&1); then
        echo "git merge-base ${left:-HEAD} ${right:-HEAD} failed:" >&2
        printf '%s\n' "$BASE" | head -n 2 >&2
        exit 3
      fi
      TIP="${right:-HEAD}"
      ;;
    *..*)
      BASE="${RANGE%%..*}"
      TIP="${RANGE##*..}"
      TIP="${TIP:-HEAD}"
      ;;
    *)
      BASE="$RANGE"
      TIP="HEAD"
      ;;
  esac

  git_read "git log --format ${BASE}..${TIP}" \
    log --format='%h%x09%cn%x09%p' "${BASE}..${TIP}"
  shas=$(printf '%s\n' "$GIT_OUT" \
    | grep -vF "	${BOT_COMMITTER}	" \
    | cut -f1 || true)

  # **A merge commit is nobody's writing.** Its parents field carries two
  # ids, and its message is composed by git — the work it joins already
  # carried whatever credit it owed, in the commits that did the writing.
  # The forge builds one of these for every pull request it tests, so a
  # direction that judged them would fault every branch that ever caught
  # up with its base. Same reason as the recording commit's exemption:
  # the flag reaches what an agent *wrote*.
  #
  # Space-separated, never newline-separated: the membership test below is
  # a `case` glob over `" $merges "`, and a newline between two ids is not
  # the space the pattern looks for. With one merge in the range the two
  # forms are indistinguishable, which is exactly how a broken version of
  # this passed locally and faulted in CI, where the forge's own synthetic
  # merge makes a second one.
  merges=$(printf '%s\n' "$GIT_OUT" \
    | awk -F'	' 'NF >= 3 && $3 ~ / / { print $1 }' | tr '\n' ' ' || true)

body="${PR_BODY:-}"

case "$CREDIT" in
false)
  for sha in $shas; do
    [ -n "$sha" ] || continue
    git_read "git log -1 --format=%B ${sha}" log -1 --format='%B' "$sha"
    hit=$(printf '%s\n' "$GIT_OUT" \
      | grep -E "$CREDIT_LINES|$CREDIT_PROSE" | head -n 1 || true)
    [ -z "$hit" ] || fault "commit ${sha} carries platform credit while agent_coauthor is false: ${hit}"
  done

  if [ -n "$body" ]; then
    hit=$(printf '%s\n' "$body" \
      | grep -E "$CREDIT_LINES|$CREDIT_PROSE" | head -n 1 || true)
    [ -z "$hit" ] || fault "the pull request body carries platform credit while agent_coauthor is false: ${hit}"
  fi
  ;;

true)
  # **The unit is the pull request, because nothing here can be a
  # commit.** The flag obliges an artifact per commit, but deciding whose
  # commit it is needs a signal that does not exist: an agent commits
  # under whoever ran it, same name and same email as any other work of
  # theirs, and this script is handed a title, a body and a range.
  #
  # So the declaration is read where one exists. At `true` the flag
  # obliges a credit line in the body, and that line is the pull request
  # saying an agent worked it. When it is there, every commit that is not
  # the machinery's owes the trailer; when nothing declares agent work
  # anywhere, there is nothing to judge and this says so rather than
  # inventing a verdict.
  #
  # What this catches is partial compliance — three commits trailered of
  # five, or the body line written and none of the trailers. What it
  # cannot catch is an agent that credits itself nowhere, which is the
  # blind spot absence always leaves.
  declared=""
  if [ -n "$body" ]; then
    declared=$(printf '%s\n' "$body" \
      | grep -E "$CREDIT_LINES|$CREDIT_PROSE" | head -n 1 || true)
  fi

  # The trailer half walks the authored commits alone.
  authored=""
  for sha in $shas; do
    [ -n "$sha" ] || continue
    case " $merges " in *" $sha "*) continue ;; esac
    authored="$authored $sha"
  done

  if [ -z "$declared" ]; then
    echo "agent_coauthor is true and nothing in the pull request body declares agent work — no commit judged."
  else
    for sha in $authored; do
      [ -n "$sha" ] || continue
      git_read "git log -1 --format=%B ${sha}" log -1 --format='%B' "$sha"
      trailers=$(printf '%s\n' "$GIT_OUT" | grep -E "$TRAILER" || true)
      if [ -z "$trailers" ]; then
        fault "commit ${sha} carries no Co-Authored-By: trailer while agent_coauthor is true and this pull request declares agent work"
        continue
      fi
      # **Every trailer is read, never only the first.** A commit that
      # credits a person and a model carries two of these, in whichever
      # order the writer typed them, and a tripwire that stopped at line
      # one would pass `Co-Authored-By: AI` for having a human above it
      # while rejecting the same commit for having the human below —
      # a verdict decided by line order, which is not a rule anyone
      # could obey. The obligation is that *a* trailer names a model, so
      # every trailer is asked, and each category found is named.
      #
      # **What a non-category name buys is the tripwire, not a proof.**
      # Nothing here can tell a model's name from a person's, so a commit
      # trailered only `Co-Authored-By: Jane Doe` passes — the same blind
      # spot absence leaves, recorded in spec-0035 rather than papered
      # over with a list of known model names that would fault the next
      # model to ship.
      while IFS= read -r trailer; do
        [ -n "$trailer" ] || continue
        # The name is what sits between the colon and the address. Folded
        # to lowercase with spaces as hyphens, so "an AI" and "An  Ai" are
        # one token against the vocabulary.
        name=$(printf '%s\n' "$trailer" \
          | sed -E 's/^[[:space:]]*[Cc]o-[Aa]uthored-[Bb]y:[[:space:]]*//; s/[[:space:]]*<.*$//; s/[[:space:]]+$//' \
          | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
        case " $CATEGORIES " in
          *" $name "*)
            fault "commit ${sha}'s trailer names '${name}', a category rather than a model — the record has to survive the next model's arrival, so name it: Co-Authored-By: Claude Opus 5 <...>" ;;
        esac
      done <<TRAILERS
${trailers}
TRAILERS
    done
  fi
  ;;

*)
  # The same posture the title check takes for a style it does not know:
  # check_settings.sh is where a value outside the vocabulary is named,
  # and judging credit against a flag nobody declared would fault honest
  # commits for a fault in another file.
  echo "agent_coauthor is '${CREDIT}', which this check does not know — check_settings.sh names that; no credit judged."
  ;;
esac

# ---------------------------------------------------------------- close

if [ "$faults" -ne 0 ]; then
  echo "" >&2
  echo "These are settings the project declared and an agent was told to" >&2
  echo "obey. What leaves a trace is checked rather than trusted" >&2
  echo "(docs/technical/README.md#observance-is-checked-where-it-leaves-a-trace)." >&2
  exit 1
fi

echo "OK — the title observes '${STYLE}', and agent_coauthor is honoured."
