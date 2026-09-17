package doctorcmd

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/chapter"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/pointer"
	"github.com/thomasfranke/writrun-cli/internal/queue"
	"github.com/thomasfranke/writrun-cli/internal/requirements"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// stage0 is the environment: one requirement per program the wrapped
// scripts need, named whether it is there or not. The list is
// internal/requirements — `init` probes the same one.
func stage0(d Deps) []requirement {
	missing := map[string]bool{}
	for _, bin := range requirements.Missing(d.LookPath) {
		missing[bin] = true
	}
	var found []requirement
	for _, bin := range requirements.All() {
		r := requirement{stage: 0, name: bin}
		if missing[bin] {
			r.mark = breaks
			r.note = "not on the PATH, and the wrapped scripts require it"
		}
		found = append(found, r)
	}
	return found
}

// splitName is the one requirement the docs/ and work/ split makes,
// named for the four folders it is about. The queue's folders are
// internal/queue's to name (coupling.md, rule 1).
var splitFolders = []string{"docs", queue.TasksDir, queue.SpecsDir, queue.ReportsDir}

// stage1 is the files: the three documents the methodology requires of
// the adopter, the docs/ and work/ split, the gates answered in
// `writrun/gates.md`, the kit's tag recorded, and the two checks whose
// verdict is the repository's own. Nine requirements, each a row
// whichever way it answers. `init` asks its own stage-1 questions in its
// own words; why those are not shared — and which mechanics underneath
// them are — is written at initcmd.checkFiles (task-0019).
func stage1(root string, d Deps) []requirement {
	return []requirement{
		about(d.Files, root),
		theChapter(d.Files, root, "product"),
		theChapter(d.Files, root, "technical"),
		split(d.Files, root),
		agents(d.Files, root),
		gates(d, root),
		kitVersion(d.Files, root),
		script(root, d, frontMatterScript,
			"the queue's front matter is not canonical; every fault it named is below"),
		script(root, d, settingsScript,
			kit.Settings+" does not hold the shape the line-based readers can see"),
	}
}

// about is the About file the methodology requires of every project.
func about(disk vfs.FS, root string) requirement {
	r := requirement{stage: 1, name: "docs/about.md"}
	if !exists(disk, filepath.Join(root, "docs", "about.md")) {
		r.mark = breaks
		r.note = "an About file is required of the project, and none was found"
	}
	return r
}

// theChapter is one of the two documentation halves, which has to hold a
// real doc and not only its README.
func theChapter(disk vfs.FS, root, folder string) requirement {
	r := requirement{stage: 1, name: "docs/" + folder + "/", note: "a chapter beyond the README"}
	if !chapter.In(disk, filepath.Join(root, "docs", folder)) {
		r.mark = breaks
		r.note = fmt.Sprintf("at least one real %s doc is required beyond the README", folder)
	}
	return r
}

// split is the docs/ and work/ folders the methodology keeps its two
// halves in. One requirement, because they are one split: a report
// naming three of the four would read as three separate faults.
func split(disk vfs.FS, root string) requirement {
	r := requirement{stage: 1, name: strings.Join(splitFolders, "/, ") + "/"}
	var gone []string
	for _, rel := range splitFolders {
		if !exists(disk, filepath.Join(root, filepath.FromSlash(rel))) {
			gone = append(gone, rel+"/")
		}
	}
	if len(gone) > 0 {
		r.mark = breaks
		r.note = strings.Join(gone, ", ") + " missing, and the docs/ and work/ split requires every one of them"
	}
	return r
}

// agents reads AGENTS.md — the entry point present, and the stale
// fenced section a kit before v0.0.04 grafted. From v0.0.04 the file is
// the project's whole, so a leftover section is a duplicate of what
// `.writrun/AGENTS.md` now says rather than a broken refresh: it
// advises, it does not break.
func agents(disk vfs.FS, root string) requirement {
	r := requirement{stage: 1, name: "AGENTS.md"}
	raw, err := disk.ReadFile(filepath.Join(root, "AGENTS.md"))
	switch {
	case err != nil:
		r.mark = breaks
		r.note = "the agents' entry point is missing"
	case pointer.Legacy(raw):
		r.mark = advises
		r.note = "a writrun:begin/writrun:end section is stale"
	}
	return r
}

// gatesFile is where a project states who operates each gate. It is the
// kit's own declaration of the question, so the answers are read from
// its rows rather than from a list of gates held here
// (docs/technical/engineering/coupling.md, rule 2).
const gatesFile = kit.Gates

// gates counts the rows of that file and how many the project answered.
// The transition each row names is the detail's own words, so a gate
// this binary has never seen is judged by the same rule and named by the
// file that states it.
func gates(d Deps, root string) requirement {
	r := requirement{stage: 1, name: gatesFile}
	// The address is the project's; which file answers it is the kit's
	// resolver's to say. A project that never wrote its own gates defers
	// to the kit's default, which answers every gate the cautious way —
	// so reading the address directly would report a stub as a file with
	// no table in it (coupling.md, rule 3).
	inForce, err := kit.Resolve(d.Scripts, root, gatesFile)
	if err != nil {
		// A check that could not be made is not a failed check: the state
		// is the one the forge's unreachable reads already use, so a
		// resolver this kit does not ship never fails a run.
		r.mark, r.note, r.detail = unread, "which file answers it could not be read", err.Error()
		return r
	}
	raw, err := d.Files.ReadFile(filepath.Join(root, filepath.FromSlash(inForce)))
	if err != nil {
		r.mark = breaks
		r.note = "the project's gate answers are missing; every gate the kit states is unanswered"
		return r
	}
	rows := tableRows(string(raw))
	var gates, open []string
	for _, row := range rows {
		if isDivider(row[0]) || isHeader(row[0]) {
			continue
		}
		gates = append(gates, strip(row[0]))
		if unanswered(row[1]) {
			open = append(open, strip(row[0]))
		}
	}
	if len(gates) == 0 {
		r.mark = breaks
		r.note = "no table of gates is readable in it"
		return r
	}
	r.note = fmt.Sprintf("%d gates, %d answered", len(gates), len(gates)-len(open))
	if len(open) > 0 {
		r.mark = breaks
		r.detail = "unanswered: " + strings.Join(open, "; ")
	}
	return r
}

// isDivider reports a markdown table's alignment row, which is the
// table's shape and nobody's gate.
func isDivider(cell string) bool {
	trimmed := strings.Trim(cell, " -:")
	return trimmed == ""
}

// isHeader reports the column titles, which the kit words the same in
// every table it ships.
func isHeader(cell string) bool {
	return strings.EqualFold(strings.TrimSpace(cell), "Transition")
}

// strip is a transition cell as a row should read it: without the
// backticks the file uses for its own emphasis.
func strip(cell string) string {
	return strings.TrimSpace(strings.ReplaceAll(cell, "`", ""))
}

// tableRows reads the two-column rows of every markdown table in a
// document: the transition and who answers it, trimmed.
func tableRows(doc string) [][2]string {
	var rows [][2]string
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 2 {
			continue
		}
		rows = append(rows, [2]string{strings.TrimSpace(cells[0]), strings.TrimSpace(cells[1])})
	}
	return rows
}

// unanswered reports whether a Who cell is still the kit's placeholder
// rather than the project's answer.
func unanswered(who string) bool {
	return strings.TrimSpace(who) == "" || strings.Contains(who, "TODO")
}

// kitVersion reads `.writrun/VERSION` through kittag, which owns that
// file's path and its parsing, and grades what it finds here: a tag no
// refresh could act on breaks a flow, and saying so is doctor's alone.
func kitVersion(disk vfs.FS, root string) requirement {
	r := requirement{stage: 1, name: kittag.Rel}
	tag, err := kittag.Read(disk, root)
	switch {
	case err != nil:
		r.mark = breaks
		r.note = "the kit's tag is not recorded, so no refresh can tell what is installed"
	case !kittag.Readable(tag):
		r.mark = breaks
		r.note = fmt.Sprintf("%q is not a readable tag; vMAJOR.MINOR.PATCH is expected", tag)
	default:
		r.note = tag
	}
	return r
}

// script runs one of the repository's own checks and turns its verdict
// into a requirement. The exit code is the whole answer — this reads it
// and never re-decides it — and what the script said is carried under
// the row so the reader gets the faults in the script's own words
// (product/rules.md).
func script(root string, d Deps, name, expectation string) requirement {
	r := requirement{stage: 1, name: path.Base(name)}
	var said bytes.Buffer
	if err := d.Scripts(root, &said, &said, nil, name); err != nil {
		r.mark = breaks
		r.note = "it refuses: " + expectation
		r.detail = said.String()
	}
	return r
}

// exists reports whether a path is there at all — the question every
// required file and folder asks.
func exists(disk vfs.FS, path string) bool {
	_, err := disk.Stat(path)
	return err == nil
}
