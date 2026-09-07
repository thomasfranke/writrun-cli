#!/usr/bin/env bash
# brief.sh — the whole brief of one task, in one call.
#
#   bash .writrun/skills/writrun-select-next-task/brief.sh <task-id> \
#     [task-dir] [spec-dir] [docs-dir]
#
# Step 7 of the selection algorithm (docs/technical/selection/algorithm.md) says to
# read the task's body, every spec in `spec_ref`, and the section
# `doc_ref` anchors — before writing any code. Done by hand that is four
# to six whole-file reads to get at one section of each. This prints
# exactly that brief, in that order, behind `== <path> ==` dividers.
#
# It is a **reader**: no git, no network, no writes, and no judgement.
# Whether a `doc_ref` section now contradicts its spec is the reader's
# call, and eligibility stays list_tasks.sh's.
#
# Exit codes: 0 the brief is complete; 1 the task id resolves to no file;
# 2 the brief is partial — every part that resolved is printed and the
# ones that did not are named.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -uo pipefail

WANT="${1:-}"
TASK_DIR="${2:-work/tasks}"
SPEC_DIR="${3:-work/specs}"
DOCS_DIR="${4:-docs}"

if [ -z "$WANT" ]; then
  echo "usage: brief.sh <task-id> [task-dir] [spec-dir] [docs-dir]" >&2
  exit 1
fi

field() { sed -n "s/^$2: *//p" "$1" | head -n1; }

# list_field <file> <name> — a [a, b] front-matter list as bare words.
list_field() { field "$1" "$2" | tr -d '[]' | tr ',' ' '; }

# num_file <dir> <prefix> <spelling> — the queue file whose id is that
# number, whatever width either side wrote it at. A person types `34`,
# a task file says `task-0034`, and both name the one file.
num_file() {
  local want f n
  want=$(printf '%s' "$3" | tr '[:upper:]' '[:lower:]' \
    | sed -n "s/^\(${2}-\)\{0,1\}0*\([0-9][0-9]*\).*$/\2/p")
  [ -n "$want" ] || return 0
  for f in "$1"/"$2"-*.md; do
    [ -f "$f" ] || continue
    n=$(basename "$f" .md | tr '[:upper:]' '[:lower:]' \
      | sed -n "s/^$2-0*\([0-9][0-9]*\).*/\1/p")
    [ -n "$n" ] || continue
    if [ "$n" -eq "$want" ] 2>/dev/null; then printf '%s' "$f"; return 0; fi
  done
  return 0
}

# section <file> <anchor> — the file's section under the heading whose
# GitHub slug is <anchor>, up to the next heading of the same or higher
# level. Empty output means the anchor names no heading there.
#
# The slug rule is GitHub's own — lowercase, spaces to hyphens, backticks
# dropped, punctuation stripped *except* hyphens and underscores,
# duplicate heading text taking -1/-2 in document order — because that is
# what every `doc_ref` in a queue already targets. A strip-everything
# rule would resolve neither `#pr_title_style` nor `#blocked-vs-depends_on`.
section() {
  awk -v want="$2" '
    function slug(t,   s) {
      s = tolower(t)
      gsub(/`/, "", s)
      gsub(/ /, "-", s)
      gsub(/[^a-z0-9_-]/, "", s)
      return s
    }
    # A fenced block is content, not structure. A shell comment or a
    # schema example at column 0 reads exactly like a heading, and a scan
    # that believes one ends the section early — and silently, because
    # what did print still looks like a whole answer. GitHub gives those
    # lines no anchor either, so skipping them is also what keeps `seen[]`
    # in step with the -1/-2 suffixes a real duplicate heading takes. The
    # closing marker has to match the opening one, or a ~~~ inside a ```
    # block would reopen what it never opened.
    /^[ \t]*(```|~~~)/ {
      mark = $0
      sub(/^[ \t]*/, "", mark)
      mark = substr(mark, 1, 3)
      if (!fence) { fence = 1; opener = mark }
      else if (mark == opener) fence = 0
      if (printing) print
      next
    }
    fence { if (printing) print; next }
    /^#+[ \t]/ {
      line = $0
      sub(/^#+[ \t]+/, "", line)
      sub(/[ \t]+$/, "", line)
      hashes = $0
      sub(/[^#].*$/, "", hashes)
      level = length(hashes)
      s = slug(line)
      n = seen[s]++
      if (n > 0) s = s "-" n
      if (printing && level <= start) printing = 0
      if (!printing && s == want) { printing = 1; start = level }
    }
    printing { print }
  ' "$1"
}

# stub_target <file> <anchor> — where a router stub sends a reader, as
# `path#anchor` relative to the stub file. A stub is a heading whose
# whole body is one link line (README.md's shape after the technical
# reference became a router), and following it once is what keeps a
# brief from being complete-looking while holding nothing but a link.
stub_target() {
  section "$1" "$2" | awk '
    NR == 1 { next }                       # the heading itself
    /^[ \t]*$/ { next }                    # blank lines
    { body++; if (body == 1) line = $0 }
    END {
      if (body != 1) exit 1
      if (match(line, /\]\([^)]+\)/) == 0) exit 1
      t = substr(line, RSTART + 2, RLENGTH - 3)
      if (t ~ /^https?:/) exit 1
      print t
    }
  '
}

divider() { printf '\n== %s ==\n\n' "$1"; }

TASK_FILE=$(num_file "$TASK_DIR" task "$WANT")
if [ -z "$TASK_FILE" ]; then
  echo "No task in ${TASK_DIR}/ resolves '${WANT}' — looked for ${TASK_DIR}/task-<nnnn>*.md" >&2
  exit 1
fi

partial=0
missing=""

ID=$(field "$TASK_FILE" id)
STATUS=$(field "$TASK_FILE" status)
PRIORITY=$(field "$TASK_FILE" priority)
DOC_REF=$(field "$TASK_FILE" doc_ref)

# The header is one line: what the task is, and where every spec on it
# stands — the same cross-check step 4 makes, shown rather than fetched.
header="${ID}  ${STATUS}  ${PRIORITY}  specs:"
specs=""
dupes=""
for s in $(list_field "$TASK_FILE" spec_ref); do
  [ -n "$s" ] || continue
  case " $specs " in
    *" $s "*) dupes="${dupes}${s} "; continue ;;
  esac
  specs="${specs}${s} "
  f=$(num_file "$SPEC_DIR" spec "$s")
  if [ -n "$f" ]; then
    header="${header} ${s} $(field "$f" status),"
  else
    header="${header} ${s} MISSING,"
  fi
done
[ -n "$specs" ] || header="${header} none,"
printf '%s\n' "${header%,}"
[ -n "$dupes" ] && printf 'note: %sis listed twice in spec_ref; printed once.\n' "$dupes"

divider "$TASK_FILE"
cat "$TASK_FILE"

if [ -z "$specs" ]; then
  divider "spec_ref: none — the task body and doc_ref are the whole brief"
else
  for s in $specs; do
    f=$(num_file "$SPEC_DIR" spec "$s")
    if [ -z "$f" ]; then
      divider "${s}: no file in ${SPEC_DIR}/"
      missing="${missing}${s} "
      partial=1
      continue
    fi
    divider "$f"
    cat "$f"
  done
fi

if [ -z "$DOC_REF" ] || [ "$DOC_REF" = null ]; then
  divider "doc_ref: none — this task points at no permanent doc"
else
  path="${DOC_REF%%#*}"
  anchor=""
  case "$DOC_REF" in *#*) anchor="${DOC_REF#*#}" ;; esac
  # A `doc_ref` is written relative to `docs/` — check_front_matter.sh
  # refuses one carrying the prefix — so the file read is docs/<path>,
  # never <path> from the repository root.
  file="${DOCS_DIR}/${path}"
  if [ ! -f "$file" ]; then
    divider "${DOC_REF}: no file at ${file}"
    missing="${missing}${DOC_REF} "
    partial=1
  elif [ -z "$anchor" ]; then
    divider "$file"
    cat "$file"
  else
    body=$(section "$file" "$anchor")
    if [ -z "$body" ]; then
      divider "${file}#${anchor}: no heading with that anchor"
      missing="${missing}${DOC_REF} "
      partial=1
    else
      target=$(stub_target "$file" "$anchor")
      if [ -n "$target" ]; then
        # A router stub: one hop, and the divider shows both ends so the
        # reader knows the section came from the chapter and not the stub.
        tpath="${target%%#*}"
        tanchor=""
        case "$target" in *#*) tanchor="${target#*#}" ;; esac
        tfile="$(dirname "$file")/${tpath}"
        tbody=""
        if [ -f "$tfile" ]; then
          if [ -n "$tanchor" ]; then tbody=$(section "$tfile" "$tanchor"); else tbody=$(cat "$tfile"); fi
        fi
        if [ -n "$tbody" ]; then
          divider "${file}#${anchor} -> ${tfile}${tanchor:+#$tanchor}"
          printf '%s\n' "$tbody"
        else
          divider "${file}#${anchor} -> ${tfile}${tanchor:+#$tanchor}: the stub's link resolves to nothing"
          printf '%s\n' "$body"
          missing="${missing}${target} "
          partial=1
        fi
      else
        divider "${file}#${anchor}"
        printf '%s\n' "$body"
      fi
    fi
  fi
fi

if [ "$partial" -eq 1 ]; then
  printf '\nIncomplete brief — unresolved: %s\n' "${missing% }" >&2
  exit 2
fi
exit 0
