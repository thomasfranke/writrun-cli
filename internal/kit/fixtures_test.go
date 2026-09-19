package kit_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// stagedTrees are the two kit directories whose files source one
// another. A script in either reaches sideways — `read_setting.sh` and
// `check_settings.sh` read `queue_lib.sh`, `list_tasks.sh` reads it too
// — so a fixture that copies one of them alone stages a script that
// cannot run.
//
// `.writrun/VERSION` and a kit's `AGENTS.md` are not here on purpose.
// They are leaves: nothing sources them, they have no directory of
// peers to be copied with, and naming one costs nothing a release can
// collect on.
var stagedTrees = []string{
	".writrun/scripts/",
	".writrun/skills/",
}

// assignment is a shell variable taking a value, which is how both
// fixtures that ever held this defect wrote the path: the `cp` line
// names `$LISTER`, and the address is three lines above it. A check
// reading only the `cp` line's own text would have seen nothing on the
// exact file it exists to catch.
var assignment = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

// TestFixturesDiscoverTheKit guards the rule that keeps a kit release
// from landing in the suite: a fixture stages the kit by copying its
// trees, and never by naming the files inside them
// (docs/technical/testing/suites.md).
//
// It reads `cp` and nothing else. Naming a kit script in order to run
// it is what the cases do — `a_stage_raise_is_previewed_test.sh` names
// `$SETTINGS_CHECK` to invoke it — and that is a use, not a staging.
// Writing a file is not staging either: `init_lib.sh` composes a stub
// `read_setting.sh` with a heredoc, building a fake kit for the
// adoption cases rather than copying this repository's.
//
// The cost of not having this test is on the record. WritRun v0.0.09
// made `read_setting.sh` source `queue_lib.sh`; `doctor_lib.sh` had
// named four kit files and staged none of their neighbours, and
// thirty-nine integration cases and the `release` e2e case failed on
// the bump — the one moment nobody is reading fixtures. report-0047 and
// task-0038 are the second time the same shape was found by hand.
func TestFixturesDiscoverTheKit(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "..", "tests", "*_lib.sh"))
	if err != nil {
		t.Fatalf("listing the fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixture was read — tests/*_lib.sh matched nothing, so this test proves nothing")
	}

	for _, path := range fixtures {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		name := filepath.ToSlash(filepath.Join("tests", filepath.Base(path)))

		lines := strings.Split(string(body), "\n")
		staged := stagedVars(lines)
		for i, line := range lines {
			for _, cmd := range commands(line) {
				if tree, ok := copiesOneKitFile(cmd, staged); ok {
					t.Errorf("%s:%d stages a single file under %s — copy the tree instead, so a script the next tag teaches to source a neighbour still runs:\n\t%s",
						name, i+1, tree, strings.TrimSpace(line))
				}
			}
		}
	}
}

// stagedVars are the file's own variables holding a path inside one of
// the staged trees, so a `cp` through one is read as the address it is.
func stagedVars(lines []string) map[string]string {
	vars := map[string]string{}
	for _, line := range lines {
		m := assignment.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if tree, ok := namesTree(m[2]); ok {
			vars[m[1]] = tree
		}
	}
	return vars
}

// commands cuts a line into the commands it runs, so a `cp` behind a
// `&&` is read like one that opens the line.
func commands(line string) []string {
	if i := strings.Index(line, "#"); i >= 0 {
		line = line[:i]
	}
	return strings.FieldsFunc(line, func(r rune) bool {
		return r == ';' || r == '|' || r == '&'
	})
}

// copiesOneKitFile reports whether a command copies a single file out
// of a staged tree, and which tree it came from. A recursive copy is
// the shape this test asks for, so it is the one thing that passes.
func copiesOneKitFile(cmd string, staged map[string]string) (string, bool) {
	fields := strings.Fields(cmd)
	if len(fields) == 0 || fields[0] != "cp" {
		return "", false
	}
	for _, arg := range fields[1:] {
		if strings.HasPrefix(arg, "-") {
			if strings.ContainsAny(arg[1:], "Rr") {
				return "", false
			}
			continue
		}
		if tree, ok := namesTree(arg); ok {
			return tree, true
		}
		for v, tree := range staged {
			if strings.Contains(arg, "$"+v) || strings.Contains(arg, "${"+v+"}") {
				return tree, true
			}
		}
	}
	return "", false
}

// namesTree reports whether a word reaches inside a staged tree — the
// tree's own directory does not, which is what `cp -R` is given.
func namesTree(word string) (string, bool) {
	for _, tree := range stagedTrees {
		if i := strings.Index(word, tree); i >= 0 && len(strings.TrimRight(word[i+len(tree):], `"'`)) > 0 {
			return tree, true
		}
	}
	return "", false
}
