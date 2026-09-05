package workcmd

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

// availableHeader opens the only group the selection offers. Steps 2–6
// of the algorithm decided what is in it and in what order, and they
// ran in the lister — this command takes the first line, it does not
// re-rank them (docs/technical/selection.md).
const availableHeader = "Available"

// headers are the sections the lister prints, matched on their opening
// phrase because the rest of each heading is the skill's wording and
// may be reworded without this binary losing the section.
var headers = []string{
	"In progress",
	availableHeader,
	"In flight",
	"Held back",
	"Open reports",
}

// taskID is a task named the ways a person spells one: `task-0007`,
// `TASK-7`, `0007`, `7`. The zero padding is how the queue writes an
// id and no part of the number.
var taskID = regexp.MustCompile(`^(?i:task-)?0*([0-9]+)$`)

// entry is one line of a section: the id it opens with, that id's
// number, and the line as the lister wrote it — with any continuation
// under it, so a pause travels with the task it qualifies.
type entry struct {
	id   string
	num  int
	line string
}

// section is one of the lister's sections, in its own words.
type section struct {
	header  string
	entries []entry
}

// listing is the lister's answer, cut into sections. Nothing is
// interpreted here beyond which line sits under which heading: the
// reasons are the lister's, and a refusal quotes them rather than
// restating them (spec-0007, steps).
type listing []section

// parse cuts the lister's output into its sections. An indented line
// is a row of the section above it; an unindented line that is not a
// heading — "Order is a suggestion…", the closing note — ends the
// section it follows and belongs to no group.
func parse(out string) listing {
	var l listing
	inside := false
	for _, line := range strings.Split(out, "\n") {
		indented := strings.HasPrefix(line, "  ")
		if !indented {
			if isHeader(line) {
				l = append(l, section{header: strings.TrimSpace(line)})
				inside = true
			} else if strings.TrimSpace(line) != "" {
				inside = false
			}
			continue
		}
		if !inside {
			continue
		}
		s := &l[len(l)-1]
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		id := strings.Fields(text)[0]
		if num, ok := idNumber(id); ok {
			s.entries = append(s.entries, entry{id: id, num: num, line: line})
			continue
		}
		// A continuation — the `paused — …` line the lister prints
		// under an in-flight task. It qualifies the entry above it and
		// is kept with it.
		if len(s.entries) > 0 {
			last := &s.entries[len(s.entries)-1]
			last.line += "\n" + line
		}
	}
	return l
}

// isHeader reports whether a line opens one of the lister's sections.
func isHeader(line string) bool {
	for _, h := range headers {
		if strings.HasPrefix(line, h) {
			return true
		}
	}
	return false
}

// idNumber reads the number a task id carries; ok is false for
// anything that does not name a task.
func idNumber(s string) (int, bool) {
	m := taskID.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return n, true
}

// available is the ids of the Available section, in the lister's order.
func (l listing) available() []string {
	var ids []string
	for _, s := range l {
		if !strings.HasPrefix(s.header, availableHeader) {
			continue
		}
		for _, e := range s.entries {
			ids = append(ids, e.id)
		}
	}
	return ids
}

// locate finds the task in whatever section holds it, and returns that
// section's own heading with it.
func (l listing) locate(num int) (string, entry, bool) {
	for _, s := range l {
		for _, e := range s.entries {
			if e.num == num {
				return s.header, e, true
			}
		}
	}
	return "", entry{}, false
}

// selectTask is step 2: the first available task, or the named one
// verified against the same answer. The lister is the authority on
// both — this command runs it once and reads its sections.
func selectTask(ctx *command.Ctx, d Deps, want string) (string, error) {
	var out bytes.Buffer
	err := d.Scripts(ctx.Root, &out, ctx.Stderr, listScript)
	// Exit 1 is the lister's "nothing is available", which is an answer;
	// anything else is a lister that could not answer at all, and its
	// output is shown rather than summarised.
	if err != nil && exitCode(err) != 1 {
		fmt.Fprint(ctx.Stdout, out.String())
		return "", fmt.Errorf("running %s: %w", listScript, err)
	}
	l := parse(out.String())

	if want == "" {
		ids := l.available()
		if len(ids) == 0 {
			fmt.Fprint(ctx.Stdout, out.String())
			return "", errors.New("nothing is available to work on — no agent was launched")
		}
		return ids[0], nil
	}

	num, ok := idNumber(want)
	if !ok {
		return "", fmt.Errorf("%q does not name a task", want)
	}
	held, found, ok := l.locate(num)
	switch {
	case ok && strings.HasPrefix(held, availableHeader):
		return found.id, nil
	case ok:
		// The reason is the lister's line under the lister's heading,
		// quoted whole. A rephrasing here would be a second opinion on
		// eligibility, and eligibility has one authority (spec-0007).
		return "", fmt.Errorf("%s is not available — %s\n%s", found.id, held, found.line)
	}
	fmt.Fprint(ctx.Stdout, out.String())
	return "", fmt.Errorf("task-%04d is in no section of the lister's answer, so it is not available to work on", num)
}
