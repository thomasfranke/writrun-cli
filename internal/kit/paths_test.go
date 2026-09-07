package kit_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// writeFile is os.WriteFile with the mode this suite writes with.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

// declaring are the packages that may hold a literal path into the
// adopted repository — one each for the act it owns
// (docs/technical/engineering/coupling.md, rule 1):
//
//	internal/kit       the scripts it runs and the kit files it reads
//	internal/queue     the queue's folders
//	internal/kittag    the recorded tag
//	internal/pointer   the one file whose shape the binary knows
//	internal/kitpaths  the adopter's own paths, which nothing calls
//
// Anywhere else, a path is referenced. The spec named the first four;
// the fifth is the inventory, whose whole subject is naming paths the
// binary never calls — `.writrun/conventions`, `docs`, `work`.
var declaring = map[string]bool{
	"internal/kit":      true,
	"internal/queue":    true,
	"internal/kittag":   true,
	"internal/pointer":  true,
	"internal/kitpaths": true,
}

// prefixes are what makes a literal a path into the adopted repository.
var prefixes = []string{".writrun/", "work/"}

// isPath separates a declaration from a sentence that happens to open
// with one. `".writrun/VERSION"` is a path the code acts on;
// `".writrun/VERSION records no tag"` is a command's own words about
// it, and rule 1 governs what the binary calls, not what it prints.
// Whitespace is the whole test, because a path has none.
func isPath(lit string) bool {
	for _, prefix := range prefixes {
		if !strings.HasPrefix(lit, prefix) {
			continue
		}
		return !strings.ContainsAny(lit, " \t\n")
	}
	return false
}

// TestNoPackageOutsideTheDeclaringFiveNamesAPath is what keeps the
// paths collapsed. Ten of them were declared in two to five packages
// each, under as many as four names for one file: a rename that
// updated three of the four copies compiled, and the fourth command
// failed at run time against a repository that was correct
// (task-0027).
//
// Test files are not walked. A fixture naming `work/tasks/task-0011-…`
// names a document, not a folder, and a suite that could not write one
// down would be worse than the duplication this ends.
func TestNoPackageOutsideTheDeclaringFiveNamesAPath(t *testing.T) {
	root := repoRoot(t)
	var offences []string

	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			pkg := filepath.ToSlash(filepath.Dir(rel))
			if declaring[pkg] {
				return nil
			}
			for _, lit := range literals(t, path) {
				if isPath(lit) {
					offences = append(offences, filepath.ToSlash(rel)+": "+strconv.Quote(lit))
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", dir, err)
		}
	}

	if len(offences) > 0 {
		t.Errorf("%d path(s) declared outside the five packages that may hold one:\n  %s\n\n"+
			"Reference the name instead — internal/kit for a script or a kit file, "+
			"internal/queue for a queue folder, internal/kittag for the tag "+
			"(docs/technical/engineering/coupling.md, rule 1).",
			len(offences), strings.Join(offences, "\n  "))
	}
}

// TestTheCheckSeesAPlantedPath proves the walk above can fail: a check
// that cannot is a green that proved nothing.
func TestTheCheckSeesAPlantedPath(t *testing.T) {
	dir := t.TempDir()
	planted := filepath.Join(dir, "planted.go")
	src := "package planted\n\nconst lister = \".writrun/skills/writrun-select-next-task/list_tasks.sh\"\n"
	if err := writeFile(planted, src); err != nil {
		t.Fatal(err)
	}
	found := literals(t, planted)
	var saw bool
	for _, lit := range found {
		if isPath(lit) {
			saw = true
		}
	}
	if !saw {
		t.Errorf("the reader found no path in a file that declares one: %v", found)
	}
}

// TestASentenceIsNotADeclaration holds the line the check draws: a
// message naming a file is the command's own words, and moving it into
// a constant would put a sentence in the inventory.
func TestASentenceIsNotADeclaration(t *testing.T) {
	if isPath(".writrun/VERSION records no tag") {
		t.Error("a sentence opening with a path is read as a declaration")
	}
	if !isPath(".writrun/VERSION") {
		t.Error("a bare path is not read as a declaration")
	}
	if !isPath("work/tasks") {
		t.Error("a queue folder is not read as a declaration")
	}
	if isPath("docs/product/rules.md") {
		t.Error("a path outside the kit and the queue is read as one")
	}
}

// literals is every string literal in one Go file, unquoted.
func literals(t *testing.T, path string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	var out []string
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		value, unquoteErr := strconv.Unquote(lit.Value)
		if unquoteErr == nil {
			out = append(out, value)
		}
		return true
	})
	return out
}

// repoRoot is the module's own directory, two levels above this
// package — the tests run from the package directory, and the walk is
// over the whole tree.
func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving the repository root: %v", err)
	}
	return abs
}
