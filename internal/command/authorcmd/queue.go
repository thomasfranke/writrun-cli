package authorcmd

import "strings"

// The queue's front matter is read here the way every reader in this
// methodology reads it — line-based, one `key: value` per line, the
// block above the body and nothing below it (check_front_matter.sh
// makes that form a checked contract, so a parser would be a second
// authority on a shape already guaranteed).

// frontMatter returns the lines of the leading `---` block. A file that
// does not open with one has no front matter at all, and every reader
// here says so rather than guessing where it ends.
func frontMatter(content []byte) ([]string, bool) {
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return lines[1:i], true
		}
	}
	return nil, false
}

// field reads one front-matter field. A body line spelling `status:` at
// column 0 is prose, so only the block above counts.
func field(content []byte, name string) string {
	lines, ok := frontMatter(content)
	if !ok {
		return ""
	}
	for _, l := range lines {
		if rest, found := strings.CutPrefix(l, name+":"); found {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// list reads an inline list field — `[]`, or `[spec-0009, spec-0010]`.
// `null` and the empty brackets are both a complete answer of nothing.
func list(content []byte, name string) []string {
	raw := strings.TrimSpace(field(content, name))
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	var ids []string
	for _, part := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(part); id != "" && id != "null" {
			ids = append(ids, id)
		}
	}
	return ids
}

// heading is the file's `# ` title — what the table's third column
// says a derived task or spec is for. A spec's title carries its id and
// an em dash before the sentence; the sentence is the part the column
// wants, and the id is already in the column beside it.
func heading(content []byte) string {
	for _, l := range strings.Split(string(content), "\n") {
		rest, found := strings.CutPrefix(l, "# ")
		if !found {
			continue
		}
		title := strings.TrimSpace(rest)
		if _, after, split := strings.Cut(title, " — "); split {
			title = strings.TrimSpace(after)
		}
		return title
	}
	return ""
}
