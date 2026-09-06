package doctorcmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/chapter"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/pointer"
	"github.com/thomasfranke/writrun-cli/internal/requirements"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// stage0 is the environment: every requirement named where it is
// missing, one finding each, so a reader installs all of them in one
// pass rather than one per run. The list is internal/requirements —
// `init` probes the same one.
func stage0(d Deps) []finding {
	var found []finding
	for _, bin := range requirements.Missing(d.LookPath) {
		found = append(found, finding{stage: 0, level: breaks,
			text: bin + " is not on the PATH — the wrapped scripts require it"})
	}
	return found
}

// stage1 is the files: the three documents the methodology requires of
// the adopter, the docs/ and work/ split, the gates answered in
// `.writrun/gates.md`, the kit's tag recorded, and the two checks whose
// verdict is the repository's own. `init` asks its own
// stage-1 questions in its own words; why those are not shared — and
// which mechanics underneath them are — is written at
// initcmd.checkFiles (task-0019).
func stage1(root string, d Deps) []finding {
	var found []finding

	if !exists(d.Files, filepath.Join(root, "docs", "about.md")) {
		found = append(found, finding{stage: 1, level: breaks,
			text: "docs/about.md — an About file is required of the project, and none was found"})
	}
	for _, folder := range []string{"product", "technical"} {
		if !chapter.In(d.Files, filepath.Join(root, "docs", folder)) {
			found = append(found, finding{stage: 1, level: breaks,
				text: fmt.Sprintf("docs/%s/ — at least one real %s doc is required beyond the README", folder, folder)})
		}
	}
	for _, rel := range []string{"docs", "work/tasks", "work/specs", "work/reports"} {
		if !exists(d.Files, filepath.Join(root, filepath.FromSlash(rel))) {
			found = append(found, finding{stage: 1, level: breaks,
				text: rel + "/ — the docs/ and work/ split requires it, and it is missing"})
		}
	}

	found = append(found, agents(d.Files, root)...)
	found = append(found, gates(d.Files, root)...)
	found = append(found, kitVersion(d.Files, root)...)
	found = append(found, script(root, d, frontMatterScript,
		"the queue's front matter is not canonical; every fault it named is below")...)
	found = append(found, script(root, d, settingsScript,
		".writrun/settings.json does not hold the shape the line-based readers can see")...)
	return found
}

// agents reads AGENTS.md — the entry point present, and the stale
// fenced section a kit before v0.0.04 grafted. From v0.0.04 the file is
// the project's whole, so a leftover section is a duplicate of what
// `.writrun/AGENTS.md` now says rather than a broken refresh: it
// advises, it does not break.
func agents(disk vfs.FS, root string) []finding {
	raw, err := disk.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		return []finding{{stage: 1, level: breaks,
			text: "AGENTS.md — the agents' entry point is missing"}}
	}
	if pointer.Legacy(raw) {
		return []finding{{stage: 1, level: advises,
			text: "AGENTS.md — a writrun:begin/writrun:end section is still there; the flow now lives in " + pointer.Target + " and this copy of it is stale"}}
	}
	return nil
}

// gatesFile is where a project states who operates each gate. It is the
// kit's own declaration of the question, so the answers are read from
// its rows rather than from a list of gates held here
// (docs/technical/engineering/coupling.md, rule 2).
const gatesFile = ".writrun/gates.md"

// gates reports every row of that file the project has not answered.
// The transition each row names is the finding's own words, so a gate
// this binary has never seen is judged by the same rule and named by
// the file that states it.
func gates(disk vfs.FS, root string) []finding {
	raw, err := disk.ReadFile(filepath.Join(root, filepath.FromSlash(gatesFile)))
	if err != nil {
		return []finding{{stage: 1, level: breaks,
			text: gatesFile + " — the project's gate answers are missing; every gate the kit states is unanswered"}}
	}
	rows := tableRows(string(raw))
	if len(rows) == 0 {
		return []finding{{stage: 1, level: breaks,
			text: gatesFile + " — no table of gates is readable in it"}}
	}
	var found []finding
	for _, row := range rows {
		if isDivider(row[0]) || isHeader(row[0]) {
			continue
		}
		if unanswered(row[1]) {
			found = append(found, finding{stage: 1, level: breaks,
				text: gatesFile + " — the gate for " + strip(row[0]) + " is unanswered"})
		}
	}
	return found
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

// strip is a transition cell as a finding should read it: without the
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
func kitVersion(disk vfs.FS, root string) []finding {
	tag, err := kittag.Read(disk, root)
	if err != nil {
		return []finding{{stage: 1, level: breaks,
			text: ".writrun/VERSION — the kit's tag is not recorded, so no refresh can tell what is installed"}}
	}
	if !kittag.Readable(tag) {
		return []finding{{stage: 1, level: breaks,
			text: fmt.Sprintf(".writrun/VERSION — %q is not a readable tag; vMAJOR.MINOR.PATCH is expected", tag)}}
	}
	return nil
}

// script runs one of the repository's own checks and turns its verdict
// into a finding. The exit code is the whole answer — this reads it and
// never re-decides it — and what the script said is carried under the
// finding so the reader gets the faults in the script's own words
// (product/rules.md).
func script(root string, d Deps, name, expectation string) []finding {
	var said bytes.Buffer
	if err := d.Scripts(root, &said, &said, nil, name); err != nil {
		return []finding{{stage: 1, level: breaks,
			text: name + " — it refuses: " + expectation, detail: said.String()}}
	}
	return nil
}

// exists reports whether a path is there at all — the question every
// required file and folder asks.
func exists(disk vfs.FS, path string) bool {
	_, err := disk.Stat(path)
	return err == nil
}
