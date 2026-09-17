package doctorcmd

import (
	"strings"
	"unicode"
)

// requirementsTable is the table `docs/product/adoption/doctor.md`
// carries under "## The requirements", held here line for line.
//
// # Why it is copied rather than embedded
//
// `//go:embed` reads only files at or under the package's own
// directory, and the document is three directories above it. The
// guarantee the rule asks for is kept by the suite instead: a unit
// test fails unless every line below is in the document, in this
// order — so a sentence edited in one place and not the other is a
// red build and never a screen that disagrees with the document
// (docs/technical/engineering/coupling.md, rule 5).
var requirementsTable = strings.Join([]string{
	"| Requirement | What it is | Why the stage needs it | What clears it |",
	"|---|---|---|---|",
	"| `git` | the version control the wrapped scripts run | every flow reads and writes history through it | installing git and putting it on the `PATH` |",
	"| `bash` | the shell every kit script is written for | the scripts are the execution authority, and they are bash | installing bash and putting it on the `PATH` |",
	"| `awk` | the text processor the kit's readers use | the front-matter and settings readers are awk | installing awk and putting it on the `PATH` |",
	"| `sed` | the stream editor the kit's writers use | the queue's edits are sed | installing sed and putting it on the `PATH` |",
	"| `docs/about.md` | the About file every project owes | a reader arriving at the repository is told what it is before anything else | writing `docs/about.md` |",
	"| `docs/product/` | the product chapter, beyond its README | the product half is where rules a task may derive from live | writing one real product doc beside the README |",
	"| `docs/technical/` | the technical chapter, beyond its README | the technical half is where the machinery is stated | writing one real technical doc beside the README |",
	"| `docs/, work/tasks/, work/specs/, work/reports/` | the docs and work split | documents and queue are two halves, and every flow addresses them by these names | creating the folder that is missing |",
	"| `AGENTS.md` | the agents' entry point at the repository root | an agent reads this file first, and the flow is reached from it | writing `AGENTS.md`, and deleting a `writrun:begin`/`writrun:end` section a kit before v0.0.04 left |",
	"| `writrun/gates.md` | who operates each gate the methodology names | the rows are that file's own, so a gate a newer kit adds is named without this binary knowing it | answering the rows still carrying the kit's placeholder |",
	"| `.writrun/VERSION` | the kit tag this repository has installed | no refresh can tell what is installed without it | running `writrun update`, which records the tag it installs |",
	"| `writrun/settings.json` | the adopter's own answers, the declared stage first | every flow reads the stage and the conduct flags from it | writing the file, or running `writrun config` to correct the value it refuses |",
	"| `check_front_matter.sh` | the kit's own sweep over the queue's front matter | the queue is read by line-based readers that need canonical fields | fixing each fault the script named under the row |",
	"| `check_settings.sh` | the kit's own check on the settings file's shape | the same readers read the settings | fixing each fault the script named under the row |",
	"| `gh on the PATH` | the forge client the flows ask the forge through | from stage 2 every forge read and the pull request go through it | installing the GitHub CLI |",
	"| `gh authenticated` | a forge client with credentials | an unauthenticated client answers no read | running `gh auth login` |",
	"| `squash merging is on` | the repository's merge setting | the methodology lands every pull request as one commit | turning squash merging on in the repository settings |",
	"| `the recording push can write to main` | from stage 2 the workflows record the queue's state by pushing to main | a push with no right to write leaves the queue's state unrecorded | setting the Actions workflow permissions to read-and-write, or raising `contents: write` in that file |",
	"| `main is governed by a ruleset` | branch protection over the branch every flow lands on | the methodology recommends protecting it | adding a ruleset that targets `main` |",
	"| `no rule over main refuses the recording push` | whether any rule over main stops the Actions bot | a rule the bot cannot get past stops the recording push | taking the rule off `main`, or putting the Actions bot on the ruleset's bypass list |",
	"| `Issues are enabled` | somewhere for the upstream mirror to land | from stage 3 the flows open issues on this repository | enabling Issues in the repository settings |",
}, "\n")

// explanation is one row of that table: what the requirement is, why
// the stage makes it, and what clears it where it is unmet.
type explanation struct {
	what   string
	why    string
	clears string
}

// explanations is the table read once, keyed by requirement name.
var explanations = readExplanations(requirementsTable)

// readExplanations parses the document's table. A cell holding only a
// dash says the column has nothing to add for that row.
func readExplanations(table string) map[string]explanation {
	out := map[string]explanation{}
	for _, line := range strings.Split(table, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 4 {
			continue
		}
		name := strings.Trim(strings.TrimSpace(cells[0]), "`")
		if name == "" || name == "Requirement" || strings.Trim(name, "- :") == "" {
			continue
		}
		out[name] = explanation{
			what:   cell(cells[1]),
			why:    cell(cells[2]),
			clears: cell(cells[3]),
		}
	}
	return out
}

// cell is one table cell as a sentence, with the placeholder dash read
// as nothing to say.
func cell(s string) string {
	s = strings.TrimSpace(s)
	if s == "—" || s == "-" {
		return ""
	}
	return s
}

// stateWords name each mark for the footer, which tells a reader what
// this repository answered rather than what the glyph is called.
var stateWords = map[mark]string{met: "Met", breaks: "Not met", advises: "Advised", unread: "Unread"}

// explain is the sentence under the cursor: the document's own words
// about the requirement, then what this run found.
//
// **A requirement the table does not name is explained by the check's
// own sentence and nothing else.** A check a newer kit adds is named
// without this binary knowing it, and never guessed at (spec-0037).
func explain(r requirement) string {
	e, named := explanations[r.name]
	if !named {
		return strings.TrimSpace(sentence(r.note, r.detail))
	}
	parts := []string{e.what}
	if e.why != "" {
		parts = append(parts, e.why)
	}
	state := stateWords[r.mark]
	if found := sentence(r.note, r.detail); found != "" {
		state += ": " + found
	}
	parts = append(parts, state)
	if r.mark != met && e.clears != "" {
		parts = append(parts, "Clear it by "+e.clears)
	}
	return join(parts)
}

// sentence is what this run found about a requirement: the row's own
// note, then whatever a script or the forge said, on one line.
func sentence(note, detail string) string {
	parts := []string{note}
	parts = append(parts, detailLines(detail)...)
	return join(parts)
}

// join runs the parts together as sentences, each closed with a full
// stop it does not already carry. Every sentence after the first opens
// with a capital, so a table cell written in the document's own
// lower-case reads as a sentence where the footer quotes it.
func join(parts []string) string {
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasSuffix(p, ".") {
			p += "."
		}
		if len(out) > 0 {
			p = capitalise(p)
		}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

// capitalise raises a sentence's first letter, and leaves a sentence
// opening with a path or a backtick exactly as it was written.
func capitalise(s string) string {
	r := []rune(s)
	if len(r) == 0 || !unicode.IsLower(r[0]) {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
