package kit

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The vocabulary of commit types and scopes is the adopter's, declared
// in `writrun/settings.json` and read there by the checks. A copy in Go
// would be a second authority over one statement, which is what
// product/rules.md forbids and what amend and author avoid by handing
// the script the composed title instead of judging it
// (spec-0023, task-0024).
//
// It lived in check_observance.sh and in conventions/commits.md until
// WritRun v0.0.07 gave it one home: a kit file a refresh replaces could
// not hold an adopter's answer, and a convention explains a vocabulary
// rather than carrying it (spec-0033).
//
// This case lives beside the runner because the runner is why no copy
// is needed: everything that has to know the list asks the script.
//
// It reads the shipped tree and not the tests. `initcmd`'s fixtures
// quote the list on purpose — `applyVocabulary` writes that very line
// into the copied kit, and a case proving it cannot assert about text
// it may not spell. What the invariant is about is the binary that
// runs.
func TestNoShippedGoFileHoldsTheKitsVocabulary(t *testing.T) {
	root := filepath.Join("..", "..")
	script := filepath.Join(root, filepath.FromSlash(Settings))
	declared, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("reading %s: %v", script, err)
	}

	lists := map[string]string{}
	for _, line := range strings.Split(string(declared), "\n") {
		line = strings.TrimSpace(line)
		for _, key := range []string{"commit_types", "commit_scopes"} {
			prefix := `"` + key + `": "`
			if strings.HasPrefix(line, prefix) {
				lists[key] = strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(line, prefix), `,`), `"`)
			}
		}
	}
	if len(lists) != 2 {
		t.Fatalf("%s declares %d of the two vocabulary lines; the case cannot read what it is about", script, len(lists))
	}

	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".writrun" || d.Name() == "writrun" {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".go" || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for key, list := range lists {
			if strings.Contains(string(content), list) {
				t.Errorf("%s carries the kit's %s list — the vocabulary has one home, and it is %s", p, key, script)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the Go tree: %v", err)
	}
}
