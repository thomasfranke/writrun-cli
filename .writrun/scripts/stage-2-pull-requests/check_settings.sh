#!/usr/bin/env bash
# check_settings.sh — .writrun/settings.json holds the shape a line-based
# reader can see, and only the choices Adoption leaves open.
#
# Usage: check_settings.sh
#   Run from the repository root; the path is relative to it.
#
# JSON permits arbitrary nesting, arrays and free-form whitespace.
# read_setting.sh sees none of that and would misread it in silence, so
# the file is restricted to what such a reader can see — a two-level
# object and nothing deeper. At the top level: scalar pairs and stage
# sections, each opened by a two-space-indented `"stage_N": {` line of its
# own and closed by a two-space `}` line of its own. Inside a section:
# scalar pairs at four spaces. Values are `true`, `false`, an unquoted
# integer, or a double-quoted string. This is what enforces all of it
# (docs/technical/README.md#the-shape-is-a-checked-contract), including
# that every documented key sits in its documented home.
#
# **Strictness is scoped to where the risk is.** `stage` is parsed by the
# workflows, so its shape and its value are both checked. The keys only an
# agent reads are checked for value alone, since an agent reads JSON the
# way it reads prose.
#
# **An absent file passes.** The reader documents its defaults and keeps
# working without one; a check that failed here would make the file
# mandatory, which the schema does not.
#
# **A file left at the old address is named, not read.** The reader
# honours it, flat, under the contract frozen at the move (decision 0053)
# — and this check is the only thing that will tell the adopter to move
# it.
#
# Exit codes: 0 the file is honest; 1 it is not, with every fault named;
# 3 usage error.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

SETTINGS=".writrun/settings.json"
LEGACY=".writrun/conventions/settings.json"

faults=0
fault() { echo "REJECTED: $*" >&2; faults=$((faults + 1)); }

close() {
  echo "" >&2
  echo "The shape is a checked contract because read_setting.sh reads it" >&2
  echo "line by line and would misread anything else in silence" >&2
  echo "(docs/technical/README.md#the-shape-is-a-checked-contract)." >&2
  exit 1
}

if [ ! -f "$SETTINGS" ]; then
  if [ -f "$LEGACY" ]; then
    fault "the settings file is still at ${LEGACY} — it moved to ${SETTINGS}, WritRun's root, and its keys are now sectioned by stage; the reader honours the old file flat meanwhile, but only this check will tell you"
    close
  fi
  echo "No ${SETTINGS} — the documented defaults apply."
  exit 0
fi

if [ -f "$LEGACY" ]; then
  fault "${LEGACY} is left over — ${SETTINGS} is the one address, and it wins; delete the old file rather than leaving two that are free to disagree"
fi

# The vocabularies, as the schema spells them.
STAGES="1 2 3"
TITLE_STYLES="conventional bracketed"
BOOLEANS="true false"
SPEC_REQUIRED="always when-warranted"
DECISIONS_STYLES="per-subsystem chronological"
PRODUCT_LAYOUTS="by-concept by-feature"

# Every documented key and the section it lives in — "" is the top level.
# This is both the present-always list and the homes the check enforces:
# a documented key found anywhere else is homeless, not merely misplaced,
# because the address is its identity.
#
# Two kinds live here. `stage_1` holds the four declarations — the
# variants Adoption orders declared, which gate nothing mechanical and
# are read by agents alone. `stage_2` holds the conduct flags and the
# title style, because that is where the actions they govern begin: git
# starts at Stage 2 (product/adoption.md#three-stages), so below it there
# is neither a commit nor a pull request for a conduct flag to gate.
HOMES=":stage \
stage_1:decisions_style stage_1:product_layout stage_1:provenance_ledger \
stage_1:spec_required \
stage_2:agent_coauthor stage_2:auto_commit stage_2:auto_pr \
stage_2:auto_push stage_2:pr_title_style"

home_of() {   # home_of <name> — the section the schema gives that key
  for h in $HOMES; do
    case "$h" in *:"$1") printf '%s' "${h%%:*}"; return 0 ;; esac
  done
  return 1
}

# A key that would switch off something Adoption lists as core is refused,
# not discouraged (product/adoption.md#mandatory-core-vs-documented-variant).
# Matched as substrings of the key name, because the shapes such a key
# could take are not enumerable.
#
# **This is a tripwire, not a proof.** A key named to evade it evades it;
# what this catches is the honest attempt — someone reaching for a switch
# the methodology does not offer, told where the rule is instead of
# discovering at review that their file was ignored.
CORE_STEMS="audience permanent ephemeral technical_detail proposed_changes identity human_gate gates"

# The comma discipline, per container: an entry carries a comma when
# another follows it in the same container and none when it is the last.
# Both halves are tracked as the container is walked, because anything
# else is invalid JSON and a reader that shrugged at it would be reading
# something no other tool agrees is a settings file.
t_prev_line=0; t_prev_comma=0
s_prev_line=0; s_prev_comma=0

entered() {   # entered <lineno> <has-comma> — an entry in the open container
  if [ "$section" = "" ]; then
    [ "$t_prev_line" -eq 0 ] || [ "$t_prev_comma" -eq 1 ] \
      || fault "line ${t_prev_line} has another pair after it and ends without a comma — invalid JSON"
    t_prev_line=$1; t_prev_comma=$2
  else
    [ "$s_prev_line" -eq 0 ] || [ "$s_prev_comma" -eq 1 ] \
      || fault "line ${s_prev_line} has another pair after it and ends without a comma — invalid JSON"
    s_prev_line=$1; s_prev_comma=$2
  fi
}

lineno=0
section=""
section_line=0
section_pairs=0
addresses=""
found=""
while IFS= read -r line || [ -n "$line" ]; do
  lineno=$((lineno + 1))

  if [ "$lineno" -eq 1 ]; then
    [ "$line" = "{" ] || fault "line 1 is '${line}' — the file opens with a bare '{'"
    continue
  fi

  # A trailing blank line is the file's newline, not a line.
  [ -n "$line" ] || continue

  # This guard comes before the closing brace is recognised, so a second
  # bare '}' is caught as trailing content rather than read as another
  # close. The object ends once; a file that closes twice is invalid JSON,
  # and jq would say so even though the line-based reader shrugs.
  if [ -n "${closed:-}" ]; then
    fault "line ${lineno} follows the closing '}' — the object ends once"
    continue
  fi

  if [ "$line" = "}" ] && [ "$section" = "" ]; then
    [ "$t_prev_line" -eq 0 ] || [ "$t_prev_comma" -eq 0 ] \
      || fault "line ${t_prev_line} is the last pair of the object and ends with a comma — invalid JSON"
    closed=$lineno
    continue
  fi

  if [ "$section" != "" ]; then
    # Inside a section: its closing line, or a scalar pair at four spaces.
    case "$line" in
      "  }" | "  },")
        [ "$s_prev_line" -eq 0 ] || [ "$s_prev_comma" -eq 0 ] \
          || fault "line ${s_prev_line} is the last pair of '${section}' and ends with a comma — invalid JSON"
        [ "$section_pairs" -gt 0 ] \
          || fault "'${section}' opens at line ${section_line} and holds nothing — a section exists only when it holds a documented key, never as an empty placeholder"
        section=""
        case "$line" in *,) entered "$lineno" 1 ;; *) entered "$lineno" 0 ;; esac
        continue ;;
    esac
    if printf '%s' "$line" \
      | grep -qE '^    "[A-Za-z_][A-Za-z0-9_]*": \{[[:space:]]*$'; then
      fault "line ${lineno} opens a third nesting level inside '${section}' — the canon is two levels and nothing deeper: ${line}"
      continue
    fi
    if ! printf '%s' "$line" \
      | grep -qE '^    "[A-Za-z_][A-Za-z0-9_]*": (true|false|[0-9]+|"[^"]*")(,?)$'; then
      fault "line ${lineno} is not one canonical '\"key\": value' pair at four spaces inside '${section}': ${line}"
      continue
    fi
    key=$(printf '%s' "$line" | sed -E 's/^    "([^"]*)".*/\1/')
    val=$(printf '%s' "$line" | sed -E 's/^    "[^"]*": (.*)$/\1/')
    indent="$section"
  else
    # At the top level: a stage section opening, or a scalar pair.
    if printf '%s' "$line" | grep -qE '^  "[A-Za-z_][A-Za-z0-9_]*": \{[[:space:]]*$'; then
      section=$(printf '%s' "$line" | sed -E 's/^  "([^"]*)".*/\1/')
      section_line=$lineno
      section_pairs=0
      s_prev_line=0; s_prev_comma=0
      case "$section" in
        stage_[0-9]|stage_[0-9][0-9]) ;;
        *) fault "'${section}' opens a section at line ${lineno}, and only a '\"stage_N\"' section may — the file's one nesting level is the stage split, not free structure" ;;
      esac
      case " $addresses " in
        *" ${section} "*) fault "'${section}' appears more than once" ;;
      esac
      addresses="$addresses $section"
      continue
    fi
    if ! printf '%s' "$line" \
      | grep -qE '^  "[A-Za-z_][A-Za-z0-9_]*": (true|false|[0-9]+|"[^"]*")(,?)$'; then
      fault "line ${lineno} is not one canonical '\"key\": value' pair: ${line}"
      continue
    fi
    key=$(printf '%s' "$line" | sed -E 's/^  "([^"]*)".*/\1/')
    val=$(printf '%s' "$line" | sed -E 's/^  "[^"]*": (.*)$/\1/')
    indent=""
  fi

  case "$val" in *,) has_comma=1 ;; *) has_comma=0 ;; esac
  [ "$indent" = "" ] || section_pairs=$((section_pairs + 1))
  entered "$lineno" "$has_comma"

  val=${val%,}
  case "$val" in \"*\") val=${val#\"}; val=${val%\"} ;; esac

  if [ "$indent" = "" ]; then address="$key"; else address="${indent}.${key}"; fi
  case " $addresses " in *" $address "*) fault "'${address}' appears more than once" ;; esac
  addresses="$addresses $address"

  for stem in $CORE_STEMS; do
    case "$key" in
      *"$stem"*)
        fault "'${key}' names a rule Adoption lists as core — those are not settable (product/adoption.md#mandatory-core-vs-documented-variant)" ;;
    esac
  done

  # A documented key that is not in its documented home is homeless: the
  # address is its identity, so the reader looking there finds nothing and
  # falls back to the default, silently.
  if want_section=$(home_of "$key"); then
    if [ "$want_section" != "$indent" ]; then
      here=${indent:-the top level}
      there=${want_section:-the top level}
      fault "'${key}' sits at line ${lineno} in ${here} — its home is ${there}, and the address is a key's identity, so a reader will never find it here"
    else
      found="$found $address"
    fi
  fi

  case "$key" in
    level)
      fault "'level' was renamed: declare 'stage' (1|2|3) instead — the reader honours the old value meanwhile, but only this check will tell you" ;;
    credit_ai)
      fault "'credit_ai' was renamed: declare 'agent_coauthor' (true|false) instead — the key names the artifact it obliges, a Co-Authored-By: trailer naming the model, rather than the platform it came from; the reader honours the old value meanwhile, so it carries over unchanged, but only this check will tell you" ;;
    stage)
      case " $STAGES " in
        *" $val "*) ;;
        *) fault "stage '${val}' is outside its vocabulary: ${STAGES}" ;;
      esac ;;
    pr_title_style)
      case " $TITLE_STYLES " in
        *" $val "*) ;;
        *) fault "pr_title_style '${val}' is outside its vocabulary: ${TITLE_STYLES}" ;;
      esac ;;
    agent_coauthor|auto_commit|auto_pr|auto_push|provenance_ledger)
      case " $BOOLEANS " in
        *" $val "*) ;;
        *) fault "${key} '${val}' is outside its vocabulary: ${BOOLEANS}" ;;
      esac ;;
    spec_required)
      case " $SPEC_REQUIRED " in
        *" $val "*) ;;
        *) fault "spec_required '${val}' is outside its vocabulary: ${SPEC_REQUIRED}" ;;
      esac ;;
    decisions_style)
      case " $DECISIONS_STYLES " in
        *" $val "*) ;;
        *) fault "decisions_style '${val}' is outside its vocabulary: ${DECISIONS_STYLES}" ;;
      esac ;;
    product_layout)
      case " $PRODUCT_LAYOUTS " in
        *" $val "*) ;;
        *) fault "product_layout '${val}' is outside its vocabulary: ${PRODUCT_LAYOUTS}" ;;
      esac ;;
  esac
done < "$SETTINGS"

[ "$section" = "" ] || fault "'${section}' opens at line ${section_line} and never closes — a section ends on a two-space '}' line of its own"
[ -n "${closed:-}" ] || fault "the object never closes — the file ends without a bare '}'"

# Every documented key is present, always, in its documented home — the
# same reason front matter carries null fields rather than omitting them:
# a reader sees the whole configuration without knowing the defaults.
for h in $HOMES; do
  want_section="${h%%:*}"; want_key="${h#*:}"
  if [ "$want_section" = "" ]; then want="$want_key"; else want="${want_section}.${want_key}"; fi
  case " $found " in
    *" $want "*) ;;
    *) fault "'${want}' is missing — every documented key is present, always, in its documented home" ;;
  esac
done

[ "$faults" -eq 0 ] || close

R=".writrun/scripts/stage-2-pull-requests/read_setting.sh"
echo "OK — ${SETTINGS} is canonical:"
echo "  stage=$(bash "$R" stage)"
echo "  stage_1.decisions_style=$(bash "$R" stage_1.decisions_style) stage_1.product_layout=$(bash "$R" stage_1.product_layout)"
echo "  stage_1.provenance_ledger=$(bash "$R" stage_1.provenance_ledger) stage_1.spec_required=$(bash "$R" stage_1.spec_required)"
echo "  stage_2.agent_coauthor=$(bash "$R" stage_2.agent_coauthor) stage_2.auto_commit=$(bash "$R" stage_2.auto_commit)"
echo "  stage_2.auto_pr=$(bash "$R" stage_2.auto_pr) stage_2.auto_push=$(bash "$R" stage_2.auto_push) stage_2.pr_title_style=$(bash "$R" stage_2.pr_title_style)"
