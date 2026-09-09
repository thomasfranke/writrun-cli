package configcmd

import (
	"fmt"
	"regexp"
	"strings"
)

// key is one setting: the section that owns it, and its name. The
// dotted form is what the kit's reader takes.
type key struct {
	section string // "" for a top-level key, "stage_1", "stage_2"
	name    string
}

func (k key) dotted() string {
	if k.section == "" {
		return k.name
	}
	return k.section + "." + k.name
}

// keysOf reads the file's own keys, in the order it writes them.
//
// # Why the file is the list
//
// The kit's checker requires every documented key present, always, in
// its documented home — so a canonical settings file *is* the key list,
// and a tag that adds a key adds it here the moment the checker starts
// demanding it. A list in Go would be a second answer about which keys
// exist, and the stale one (docs/technical/engineering/coupling.md).
//
// # Why the shape may be read
//
// The file is restricted to what a line-based reader can see, and
// `check_settings.sh` exists to enforce exactly that — the shape is a
// checked contract, so reading it line by line is reading what the kit
// guarantees rather than guessing at JSON.
func keysOf(content string) []key {
	var keys []key
	section := ""
	for _, line := range strings.Split(content, "\n") {
		if m := sectionRE.FindStringSubmatch(line); m != nil {
			section = m[1]
			continue
		}
		if strings.TrimSpace(line) == "}" && section != "" {
			section = ""
			continue
		}
		if m := scalarRE.FindStringSubmatch(line); m != nil {
			keys = append(keys, key{section: section, name: m[1]})
		}
	}
	return keys
}

var (
	sectionRE = regexp.MustCompile(`^\s*"([a-z_0-9]+)":\s*\{`)
	scalarRE  = regexp.MustCompile(`^\s*"([a-z_0-9]+)":\s*(?:"[^"]*"|true|false|-?\d+)`)
)

func named(keys []key, dotted string) bool {
	_, ok := find(keys, dotted)
	return ok
}

func find(keys []key, dotted string) (key, bool) {
	for _, k := range keys {
		if k.dotted() == dotted {
			return k, true
		}
	}
	return key{}, false
}

// rewrite replaces one key's value where the file states it, leaving
// every other byte alone. A miss is an error rather than a silent
// no-op: the run would otherwise report a value the file does not hold.
func rewrite(content string, k key, value string) (string, error) {
	quoted := !numericOrBool(value)
	section := ""
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if m := sectionRE.FindStringSubmatch(line); m != nil {
			section = m[1]
			continue
		}
		if strings.TrimSpace(line) == "}" && section != "" {
			section = ""
			continue
		}
		m := scalarRE.FindStringSubmatch(line)
		if m == nil || section != k.section || m[1] != k.name {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		tail := ""
		if strings.HasSuffix(strings.TrimRight(line, " \t"), ",") {
			tail = ","
		}
		written := value
		if quoted {
			written = `"` + value + `"`
		}
		lines[i] = indent + `"` + k.name + `": ` + written + tail
		return strings.Join(lines, "\n"), nil
	}
	return "", fmt.Errorf("no %q line in the settings to write into", k.dotted())
}

// numericOrBool says a value is written bare. The kit's shape allows
// `true`, `false`, an unquoted integer, or a quoted string; which one a
// key takes is the checker's to judge, and a value that lands in the
// wrong form is refused there rather than guessed at here.
func numericOrBool(v string) bool {
	if v == "true" || v == "false" {
		return true
	}
	return regexp.MustCompile(`^-?\d+$`).MatchString(v)
}
