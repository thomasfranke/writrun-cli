package doctorcmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/chapter"
	"github.com/thomasfranke/writrun-cli/internal/fence"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
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
// AGENTS.md, the fence intact, the kit's tag recorded, and the two
// checks whose verdict is the repository's own. `init` asks its own
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
	found = append(found, kitVersion(d.Files, root)...)
	found = append(found, script(root, d, frontMatterScript,
		"the queue's front matter is not canonical; every fault it named is below")...)
	found = append(found, script(root, d, settingsScript,
		".writrun/settings.json does not hold the shape the line-based readers can see")...)
	return found
}

// agents reads AGENTS.md: the entry point present, the fence the kit
// refreshes through intact, and the four gates answered.
func agents(disk vfs.FS, root string) []finding {
	raw, err := disk.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		return []finding{{stage: 1, level: breaks,
			text: "AGENTS.md — the agents' entry point is missing"}}
	}
	doc := string(raw)
	var found []finding
	begin := strings.Index(doc, fence.Begin)
	end := strings.Index(doc, fence.End)
	if begin < 0 || end < begin {
		found = append(found, finding{stage: 1, level: breaks,
			text: "AGENTS.md — the fenced writrun:begin/writrun:end markers are damaged; a refresh rewrites nothing without them"})
	}
	return append(found, gates(doc)...)
}

// gate is one of the four answers the methodology requires of a
// project. The key is the phrase the transition cell carries in every
// wording the kit has shipped, and the words are how the finding names
// the gate that went unanswered.
type gate struct {
	key   string
	names string
}

// theGates are the four, in the order the table states them
// (spec-0004, step 3).
var theGates = []gate{
	{"under `docs/`", "who writes or reviews a change under docs/"},
	{"authored rule", "who declares an authored rule finished"},
	{"approved", "who moves a spec from draft to approved"},
	{"spec_ref", "who acts on a task carrying no spec"},
}

// gates reports the gates the table leaves unanswered. The table is
// read from the Human gates section where AGENTS.md still heads it that
// way, and from the whole document otherwise — a project may retitle
// the section, and a gate answered under another heading is answered.
func gates(doc string) []finding {
	rows := tableRows(section(doc, "Human gates"))
	var found []finding
	for _, g := range theGates {
		who, stated := answer(rows, g.key)
		switch {
		case !stated:
			found = append(found, finding{stage: 1, level: breaks,
				text: "AGENTS.md — the human gates table states no row for " + g.names})
		case unanswered(who):
			found = append(found, finding{stage: 1, level: breaks,
				text: "AGENTS.md — the gate for " + g.names + " is still a placeholder; it must be answered, not left as a TODO"})
		}
	}
	return found
}

// section cuts the markdown heading named by title and everything under
// it, up to the next heading of any level. An absent heading yields the
// whole document, because a project that retitled the section still
// wrote its answers somewhere in this file.
func section(doc, title string) string {
	lines := strings.Split(doc, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "#") && strings.Contains(line, title) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return doc
	}
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "#") {
			return strings.Join(lines[start:i], "\n")
		}
	}
	return strings.Join(lines[start:], "\n")
}

// tableRows reads the two-column rows of every markdown table in a
// stretch of document: the transition and who answers it, trimmed. The
// alignment row is one of them and matches no gate, so it needs no
// special case.
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

// answer finds the row whose transition carries key and returns what it
// says about who. The match is case-insensitive: the wording is the
// project's, and a capital is not a missing gate.
func answer(rows [][2]string, key string) (string, bool) {
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row[0]), strings.ToLower(key)) {
			return row[1], true
		}
	}
	return "", false
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
