package amendcmd

import (
	"fmt"
	"strings"
)

// template is the adopter's pull-request body, the one home the
// conventions give it. Absent, the fallback below carries the same
// headings — a project that deleted the template still gets a body the
// checks can read.
const template = ".writrun/templates/pull_request_template.md"

// slugify turns a spec's file name or a phrase into branch words.
func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// specSlug is the subject words of a spec's file name — `spec-0011-`
// stripped off `spec-0011-amend-command.md`. It is the branch's default
// subject: the amendment is about that spec, and the human who named
// the file already chose the words.
func specSlug(rel string) string {
	name := rel
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	name = strings.TrimSuffix(name, ".md")
	name = strings.TrimPrefix(name, "spec-")
	if i := strings.Index(name, "-"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// branchName is the amendment's branch, and it carries no task id on
// purpose: an amendment records that an approval is in question, it
// does not work the task riding on it, and a `task/NNNN-` name would
// make the machinery read it as flight (conventions/branches.md).
func branchName(kind, slug string) string { return kind + "/" + slug }

// title composes the pull request's title in the declared style, with
// no task tags for the same reason the branch carries no id. The scope
// is `specs` because `work/specs/` is the whole of what an amendment
// touches (conventions/commits.md).
func title(style, kind, summary string) string {
	if style == "conventional" {
		return fmt.Sprintf("%s(%s): %s", kind, amendScope, summary)
	}
	// `bracketed` is the declared default here, and an unreadable style
	// is check_settings.sh's to name — composing in the style nobody
	// declared would be this command judging another file's fault.
	return fmt.Sprintf("[%s][%s] %s", capitalise(kind), capitalise(amendScope), summary)
}

// amendScope is the one scope an amendment can honestly carry.
const amendScope = "specs"

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// subject is the commit's, and it is Conventional Commits whatever the
// title style says: the title is read in the queue of open pull
// requests, the subject is read on `main` for good
// (conventions/commits.md).
func subject(kind, specID string) string {
	return fmt.Sprintf("%s(%s): return %s to draft", kind, amendScope, specID)
}

// suspension is one task this amendment holds up, and the pull request
// it rides. Number 0 says the forge did not name one.
type suspension struct {
	task   string
	number int
}

// line is the sentence check_amendment_reference.sh accepts, spelled
// the way that script itself spells it when it asks for one. A pull
// request the forge could not number is named by its task instead: the
// reference is still composed from the queue, and the body says what a
// person must finish (spec-0011, acceptance criteria).
func (s suspension) line() string {
	if s.number > 0 {
		return fmt.Sprintf("Suspends #%d — %s waits on this amendment.", s.number, s.task)
	}
	return fmt.Sprintf("Suspends the open pull request working %s — the forge did not answer, "+
		"so its number must be written in by hand.", s.task)
}

// statement is what the pull request says it does, under `## Spec`. It
// never says the spec is re-approved: that is the merge's, and the
// merge is the maintainer's (product/rules.md).
func statement(specID string, susp []suspension) string {
	lines := []string{
		fmt.Sprintf("Returns %s to `draft`. Re-approval is the merge's, and the maintainer's.", specID),
	}
	if len(susp) == 0 {
		lines = append(lines, "",
			"No task referencing it is in flight, so this suspends nothing.")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "")
	for _, s := range susp {
		lines = append(lines, s.line())
	}
	return strings.Join(lines, "\n")
}

// verification is what a reviewer re-reads by hand.
const verification = "`check_amendment_reference.sh` reads this body for the pull request the\n" +
	"amendment suspends; `check_front_matter.sh` reads the spec's own front matter."

// fallbackBody is the body of a repository whose template is gone. It
// carries the same headings, `## Derived work` excepted: an amendment
// derives nothing, and an empty declaration and a forgotten one look
// the same.
const fallbackBody = `## What

## Why

<!-- writrun:begin -->

## Spec

Implements spec-NNNN.

## How to verify

<!-- writrun:end -->

## Notes`

// body composes the pull request's body from the adopter's template:
// the guidance comments dropped, the authoring half dropped, the
// `Implements spec-NNNN.` line replaced by what this change actually
// does, and the three sections a reviewer reads filled.
func body(tmpl, specID, why string, susp []suspension) string {
	src := tmpl
	if strings.TrimSpace(src) == "" {
		src = fallbackBody
	}
	lines := strip(strings.Split(src, "\n"))
	lines = replaceImplements(lines, statement(specID, susp))
	lines = fill(lines, "## What", fmt.Sprintf("Returns %s to `draft` so its approval can be reconsidered.", specID))
	lines = fill(lines, "## Why", why)
	lines = fill(lines, "## How to verify", verification)
	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
}

// strip drops the template's guidance comments and its authoring half.
// The two `writrun:` markers survive: they are the fence the kit reads,
// not guidance for whoever fills the form.
func strip(lines []string) []string {
	var out []string
	inComment, inDerived := false, false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch {
		case inComment:
			if strings.Contains(l, "-->") {
				inComment = false
			}
			continue
		case strings.HasPrefix(trimmed, "<!-- writrun:"):
			// The fence, kept whole.
		case strings.HasPrefix(trimmed, "<!--"):
			if !strings.Contains(l, "-->") {
				inComment = true
			}
			continue
		}
		if trimmed == "## Derived work" {
			inDerived = true
			continue
		}
		if inDerived {
			if strings.HasPrefix(trimmed, "## ") {
				inDerived = false
			} else {
				continue
			}
		}
		out = append(out, l)
	}
	return squeeze(out)
}

// replaceImplements swaps the template's placeholder sentence for the
// amendment's own. A template that carries no such line takes the
// statement under `## Spec` instead, and one with neither keeps
// everything it has and takes the statement at the end — the reference
// the gate reads is never the part that goes missing.
func replaceImplements(lines []string, text string) []string {
	for i, l := range lines {
		if strings.TrimSpace(l) == "Implements spec-NNNN." {
			out := append([]string{}, lines[:i]...)
			out = append(out, strings.Split(text, "\n")...)
			return append(out, lines[i+1:]...)
		}
	}
	if filled := fill(lines, "## Spec", text); !sameLines(filled, lines) {
		return filled
	}
	return append(append([]string{}, lines...), "", text)
}

// fill puts text under a heading that has nothing under it. A section
// somebody already wrote is left exactly as they wrote it.
func fill(lines []string, heading, text string) []string {
	for i, l := range lines {
		if strings.TrimSpace(l) != heading {
			continue
		}
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		// A heading's section ends at the next heading or at either
		// fence marker; anything else under it is content somebody
		// wrote, and it is left as they wrote it.
		if j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "## ") &&
			!strings.HasPrefix(strings.TrimSpace(lines[j]), "<!-- writrun:") {
			return lines
		}
		out := append([]string{}, lines[:i+1]...)
		out = append(out, "")
		out = append(out, strings.Split(text, "\n")...)
		out = append(out, lines[i+1:]...)
		return squeeze(out)
	}
	return lines
}

// squeeze collapses runs of blank lines to one — the template's
// dropped comments leave gaps, and a body full of them reads as
// unfinished.
func squeeze(lines []string) []string {
	var out []string
	blank := false
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			if blank {
				continue
			}
			blank = true
		} else {
			blank = false
		}
		out = append(out, l)
	}
	for len(out) > 0 && strings.TrimSpace(out[0]) == "" {
		out = out[1:]
	}
	return out
}

func sameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
