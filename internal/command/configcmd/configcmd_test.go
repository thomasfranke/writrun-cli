package configcmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// settings is the shape the kit's checker holds the file to: a
// two-level object, scalars only.
const settings = `{
  "stage": 3,
  "stage_1": {
    "spec_required": "when-warranted"
  },
  "stage_2": {
    "auto_push": true,
    "pr_title_style": "bracketed"
  }
}
`

// scripts stands in for the kit: it answers a key the way the reader
// does, and accepts or refuses the way the checker does. No case tells
// this command what a key or a value is — that is the point.
type scripts struct {
	values map[string]string
	refuse string // the checker's own words, where it refuses
}

func (s *scripts) run(_ string, stdout, stderr io.Writer, _ []string, name string, args ...string) error {
	switch name {
	case kit.ReadSetting:
		fmt.Fprintln(stdout, s.values[args[0]])
		return nil
	case kit.CheckSettings:
		if s.refuse != "" {
			fmt.Fprintln(stderr, s.refuse)
			return exitErr(1)
		}
		return nil
	}
	return fmt.Errorf("unexpected script %s", name)
}

type exitErr int

func (e exitErr) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e exitErr) ExitCode() int { return int(e) }

func repo(t *testing.T) (root string, sc *scripts) {
	t.Helper()
	root = t.TempDir()
	path := filepath.Join(root, filepath.FromSlash(kit.Settings))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(settings), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, &scripts{values: map[string]string{
		"stage":                  "3",
		"stage_1.spec_required":  "when-warranted",
		"stage_2.auto_push":      "true",
		"stage_2.pr_title_style": "bracketed",
	}}
}

func exec(t *testing.T, root string, sc *scripts, args ...string) (string, string, error) {
	t.Helper()
	var out, errb bytes.Buffer
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &errb,
		Terminal: &command.FakeTerminal{}, Root: root, Adopted: true, Yes: true,
	}
	err := run(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}}, args)
	return out.String(), errb.String(), err
}

func readSettings(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(kit.Settings)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The keys come from the settings file, so a kit that documents one
// more shows one more — with no list in Go to edit.
func TestTheKeysComeFromTheFile(t *testing.T) {
	root, sc := repo(t)
	sc.values["stage_2.commit_types"] = "docs feat"
	path := filepath.Join(root, filepath.FromSlash(kit.Settings))
	if err := os.WriteFile(path, []byte(strings.Replace(settings,
		`    "auto_push": true,`,
		`    "auto_push": true,`+"\n"+`    "commit_types": "docs feat",`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := exec(t, root, sc)
	if err != nil {
		t.Fatalf("config = %v", err)
	}
	if !strings.Contains(out, "commit_types") || !strings.Contains(out, "docs feat") {
		t.Errorf("a key the file declares is missing from the listing:\n%s", out)
	}
}

func TestAnAcceptedChangeIsKept(t *testing.T) {
	root, sc := repo(t)
	if _, _, err := exec(t, root, sc, "stage_2.pr_title_style", "conventional"); err != nil {
		t.Fatalf("config = %v", err)
	}
	if got := readSettings(t, root); !strings.Contains(got, `"pr_title_style": "conventional"`) {
		t.Errorf("the change was not written:\n%s", got)
	}
}

// The adopter's file is never left in a shape its own kit calls
// invalid, and the checker's words are the ones the reader sees.
func TestARefusedChangeRestoresTheFileByteForByte(t *testing.T) {
	root, sc := repo(t)
	sc.refuse = "REJECTED: pr_title_style 'shouty' is outside its vocabulary"
	before := readSettings(t, root)

	_, errOut, err := exec(t, root, sc, "stage_2.pr_title_style", "shouty")
	if err == nil {
		t.Fatal("a refused change reported no error")
	}
	if got := readSettings(t, root); got != before {
		t.Errorf("the file was not restored:\ngot  %q\nwant %q", got, before)
	}
	if !strings.Contains(errOut, "outside its vocabulary") {
		t.Errorf("the checker's own words did not reach the user: %q", errOut)
	}
}

// A value already held is reported and not written: an untouched file
// keeps its mtime, and the checker is not asked about a change nobody
// made.
func TestAValueAlreadyHeldIsNotWritten(t *testing.T) {
	root, sc := repo(t)
	out, _, err := exec(t, root, sc, "stage_2.auto_push", "true")
	if err != nil {
		t.Fatalf("config = %v", err)
	}
	if !strings.Contains(out, "already") {
		t.Errorf("the run did not say the value was already held:\n%s", out)
	}
}

func TestAKeyTheFileDoesNotDeclareIsRefused(t *testing.T) {
	root, sc := repo(t)
	_, _, err := exec(t, root, sc, "stage_2.invented", "x")
	if err == nil || !strings.Contains(err.Error(), "not a key") {
		t.Errorf("err = %v, want a refusal naming the key", err)
	}
}

// Writing the project's home is adoption's act. A command that created
// the file would be answering for a project that never answered.
func TestAnAbsentSettingsFileOffersNoEdit(t *testing.T) {
	root := t.TempDir()
	_, _, err := exec(t, root, &scripts{values: map[string]string{}}, "stage", "2")
	if err == nil || !strings.Contains(err.Error(), "absent") {
		t.Errorf("err = %v, want a refusal naming the absent file", err)
	}
}

// screened runs `writrun config` with a terminal at both ends and the
// keys the reader would press, so the screen opens for real.
func screened(t *testing.T, root string, sc *scripts, typed string, args ...string) (string, error) {
	t.Helper()
	var out, errb bytes.Buffer
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &errb,
		Stdin:    strings.NewReader(typed),
		Terminal: &command.FakeTerminal{In: true, Out: true},
		Root:     root, Adopted: true, Yes: true,
	}
	err := run(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}}, args)
	return out.String(), err
}

// With no argument and a terminal at both ends, the settings open as
// the screen the drawing gives — not the listing.
func TestNoArgumentOpensTheScreenOnATerminal(t *testing.T) {
	root, sc := repo(t)
	out, err := screened(t, root, sc, "q")
	if err != nil {
		t.Fatalf("config = %v", err)
	}
	if !strings.Contains(out, "enter change") {
		t.Errorf("the screen's footer is absent; it printed:\n%s", out)
	}
	if strings.Contains(out, "writrun config <key> <value> changes one") {
		t.Error("the listing's closing line was printed, so this was the listing and not the screen")
	}
}

// Without a terminal it prints, exactly as it did — a script reading
// `writrun config` must keep reading it.
func TestNoArgumentStillPrintsWithoutATerminal(t *testing.T) {
	root, sc := repo(t)
	out, _, err := exec(t, root, sc)
	if err != nil {
		t.Fatalf("config = %v", err)
	}
	if !strings.Contains(out, "writrun config <key> <value> changes one") {
		t.Errorf("the listing is not what a non-terminal got:\n%s", out)
	}
	if strings.Contains(out, "enter change") {
		t.Error("a screen was opened where there is no terminal")
	}
}

// The screen and the command line reach the same write.
//
// `one` is what a key chosen on the screen runs: ask for the value,
// then write and be judged. A second write path would be a second set
// of rules about what happens to the adopter's file.
func TestAKeyChosenOnTheScreenTakesTheSameWrite(t *testing.T) {
	root, sc := repo(t)
	var out, errb bytes.Buffer
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &errb,
		Terminal: &command.FakeTerminal{In: true, Out: true, InputAnswer: "conventional"},
		Root:     root, Adopted: true, Yes: true,
	}
	path := filepath.Join(root, filepath.FromSlash(kit.Settings))
	if err := one(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}}, path, "stage_2.pr_title_style"); err != nil {
		t.Fatalf("one = %v", err)
	}
	if !strings.Contains(readSettings(t, root), "conventional") {
		t.Error("the value the reader typed was not written")
	}
	if !strings.Contains(out.String(), "pr_title_style is conventional") {
		t.Errorf("the write was not reported:\n%s", out.String())
	}
}

// A refusal from the checker reaches the reader through the same path,
// and the file is exactly as it was.
func TestAKeyChosenOnTheScreenCarriesTheCheckersRefusal(t *testing.T) {
	root, sc := repo(t)
	sc.refuse = "pr_title_style 'nonsense' is outside its vocabulary: conventional bracketed"
	before := readSettings(t, root)

	var out, errb bytes.Buffer
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &errb,
		Terminal: &command.FakeTerminal{In: true, Out: true, InputAnswer: "nonsense"},
		Root:     root, Adopted: true, Yes: true,
	}
	path := filepath.Join(root, filepath.FromSlash(kit.Settings))
	err := one(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}}, path, "stage_2.pr_title_style")
	if err == nil {
		t.Fatal("a refused change answered success")
	}
	if !strings.Contains(errb.String(), "outside its vocabulary") {
		t.Errorf("the checker's own words did not reach the reader:\n%s", errb.String())
	}
	if readSettings(t, root) != before {
		t.Error("a refused change left the settings altered")
	}
}

// The stage has no section of its own in the file, and is shown above
// the ones that do.
func TestTheStageIsHeadedWithoutASection(t *testing.T) {
	if got := heading(""); got != "THE STAGE" {
		t.Errorf("heading(\"\") = %q, want the stage's own label", got)
	}
	if got := heading("stage_2"); got != "STAGE 2" {
		t.Errorf("heading(\"stage_2\") = %q", got)
	}
}
