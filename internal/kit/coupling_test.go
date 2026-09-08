package kit_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// adopterNames are the pieces of an adopter-owned path. Each is a whole
// address or the segment a filepath.Join would carry, because both are
// how one gets written — and the segment form is the one that escaped
// task-0027's consolidation: a grep for the joined path found nothing,
// four packages kept their own copy, and the addresses those copies
// named went stale the day WritRun moved them.
var adopterNames = []string{
	"settings.json",
	"gates.md",
	"writrun/conventions",
}

// TestAdopterPathsAreNamedInThisPackageAlone guards the rule that makes
// a kit release cheap: a path the adopter owns is declared in
// internal/kit, and every other package references that name
// (docs/technical/engineering/coupling.md, rule 3).
//
// Comments are not read — only string literals — so prose may still
// name a file to explain it. What may not exist is a second address the
// compiler will happily keep after a tag moves the first.
func TestAdopterPathsAreNamedInThisPackageAlone(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			switch e.Name() {
			case ".git", "dist", "worktrees", "kit":
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case !strings.HasSuffix(path, ".go"):
			return nil
		case strings.HasSuffix(path, "_test.go"):
			// A fixture writes these addresses on purpose: it is
			// building the repository the code then reads.
			return nil
		case filepath.Dir(path) == filepath.Join(root, "internal", "kit"):
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			for _, name := range adopterNames {
				if strings.Contains(lit.Value, name) {
					rel, _ := filepath.Rel(root, path)
					t.Errorf("%s:%d names %q in a string literal — declare it in internal/kit and reference the name, so the next tag that moves it is one edit",
						filepath.ToSlash(rel), fset.Position(lit.Pos()).Line, name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking the source: %v", err)
	}
}
