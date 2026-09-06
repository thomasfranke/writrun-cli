package kit

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The vocabulary of commit types and scopes is the kit's, declared in
// check_observance.sh and in conventions/commits.md. A copy in Go would
// be a third authority over one statement, which is what
// product/rules.md forbids and what amend and author avoid by handing
// the script the composed title instead of judging it
// (spec-0023, task-0024).
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
	script := filepath.Join(root, ".writrun", "scripts", "stage-2-pull-requests", "check_observance.sh")
	declared, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("reading %s: %v", script, err)
	}

	lists := map[string]string{}
	for _, line := range strings.Split(string(declared), "\n") {
		for _, key := range []string{"TYPES", "SCOPES"} {
			if strings.HasPrefix(line, key+`="`) && strings.HasSuffix(line, `"`) {
				lists[key] = strings.TrimSuffix(strings.TrimPrefix(line, key+`="`), `"`)
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
			if d.Name() == ".git" || d.Name() == ".writrun" {
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
