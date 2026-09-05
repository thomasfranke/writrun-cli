package reportcmd

import (
	"fmt"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// The observation reaches the file after the generator, not through
// it. `new.sh report` takes a title, a slug and a doc-ref and writes a
// body that is a `TODO:` placeholder — it has no flag for the
// paragraph, and `.writrun/` is the methodology's copy, refreshed whole
// by `writrun update`, so teaching it one here would be a fork of the
// script this command exists to wrap. What is left is the smallest
// edit that keeps the generator the author of everything schema-shaped:
// the file is minted first, and the reporter's paragraph then stands
// where the placeholder stood.
//
// The substitution is by paragraph, not by the placeholder's wording:
// an adopter's own `.writrun/templates/report.md` writes its own words,
// and only "the body starts at TODO" is common to all of them.

// fill puts the reporter's paragraph where the generator left its
// placeholder. A blank observation leaves the file exactly as the
// generator wrote it — the placeholder is then the truth, and the
// reporter finishes the file by hand.
func fill(files vfs.FS, path, named, observation string) error {
	if strings.TrimSpace(observation) == "" {
		return nil
	}
	raw, err := files.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w — it is recorded; write the observation into it by hand", named, err)
	}
	if err := files.WriteFile(path, []byte(substitute(string(raw), observation)), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w — it is recorded with its placeholder; write the observation into it by hand", named, err)
	}
	return nil
}

// substitute replaces the placeholder paragraph with the observation.
// A file with no placeholder keeps everything it has and takes the
// observation as its last paragraph — a template that dropped the TODO
// is still a report, and the reporter's words are never the part that
// goes missing.
func substitute(file, observation string) string {
	body := strings.TrimRight(observation, " \t\n")
	lines := strings.Split(file, "\n")
	start, end, found := placeholder(lines)
	if !found {
		return strings.TrimRight(file, "\n") + "\n\n" + body + "\n"
	}
	out := append([]string{}, lines[:start]...)
	out = append(out, strings.Split(body, "\n")...)
	out = append(out, lines[end:]...)
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

// placeholder locates the generator's TODO paragraph: the first line
// below the front matter that opens with TODO, through to the blank
// line that ends its paragraph. Front matter is skipped rather than
// searched — a field is not a body, whatever its value reads like.
func placeholder(lines []string) (start, end int, found bool) {
	i := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i = 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				i++
				break
			}
		}
	}
	for ; i < len(lines); i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "TODO") {
			continue
		}
		start = i
		for end = i; end < len(lines); end++ {
			if strings.TrimSpace(lines[end]) == "" {
				break
			}
		}
		return start, end, true
	}
	return 0, 0, false
}
