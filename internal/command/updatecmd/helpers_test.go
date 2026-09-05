package updatecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

const (
	// oldTag is what an adopted kit records; newTag is what the binary
	// pins and therefore refreshes to.
	oldTag = "v9.9.8"
	newTag = "v9.9.9"
)

func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-c", "user.name=suite", "-c", "user.email=suite@test", "-c", "commit.gpgsign=false"}, args...)
	out, err := gitx.Run(dir, full...)
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(content)
}

// agentsAt is the kit's fenced section as a tag ships it: two `yours`
// markers, one governing the block after it and one the block before.
func agentsAt(flow string) string {
	return `# AGENTS.md — entry point for AI agents

Prose the project wrote.

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

` + flow + `

### Human gates

<!-- yours: this table is the project's own answers; it survives updates. -->

| Transition | Who |
|---|---|
| Writing docs | <!-- TODO — default: human reviews --> |

### Deriving work

Present the derived tasks in the session before opening the PR.
<!-- yours: keep, invert, or drop this default — it is the project's. -->

<!-- writrun:end -->
`
}

// writeTemplateOld writes the template tree oldTag ships, under dir.
func writeTemplateOld(t *testing.T, dir string) {
	t.Helper()
	write(t, dir, "AGENTS.md", agentsAt("The flow's text."))
	write(t, dir, ".writrun/VERSION", oldTag+"\n")
	write(t, dir, ".writrun/settings.json", "{\n  \"stage\": 1\n}\n")
	write(t, dir, ".writrun/conventions/commits.md", "# Commits\n")
	write(t, dir, ".writrun/skills/select/SKILL.md", "# Select\n")
	write(t, dir, ".writrun/scripts/take.sh", "echo take\n")
	write(t, dir, ".writrun/templates/task.md", "# Task\n")
	for _, wf := range []string{"approve", "check", "issues", "progress"} {
		write(t, dir, ".github/workflows/writrun-"+wf+".yml", "name: writrun "+wf+"\n")
	}
}

// writeTemplateNew writes the template tree newTag ships: oldTag's,
// with a reworded skill, an added template and a reworded workflow —
// one of each verb the plan reports.
func writeTemplateNew(t *testing.T, dir string) {
	t.Helper()
	writeTemplateOld(t, dir)
	write(t, dir, "AGENTS.md", agentsAt("The flow's text, reworded."))
	write(t, dir, ".writrun/VERSION", newTag+"\n")
	write(t, dir, ".writrun/skills/select/SKILL.md", "# Select, reworded\n")
	write(t, dir, ".writrun/templates/spec.md", "# Spec\n")
	write(t, dir, ".github/workflows/writrun-check.yml", "name: writrun check\n# reworded\n")
}

// makeTemplate is the template newTag ships, as a fake fetch hands it
// over: the same tree makeSource commits, without the repository and
// without the clone.
func makeTemplate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeTemplateNew(t, dir)
	return dir
}

// makeSource builds a local WritRun repository with two tags: oldTag,
// then newTag.
func makeSource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	gitT(t, src, "init", "-q")
	writeTemplateOld(t, filepath.Join(src, "template"))
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "the kit, one release back")
	gitT(t, src, "tag", oldTag)

	writeTemplateNew(t, filepath.Join(src, "template"))
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "the kit")
	gitT(t, src, "tag", newTag)
	return src
}

// makeAdopted builds a repository carrying the kit as oldTag shipped
// it, plus files of the project's own that a refresh may not reach.
func makeAdopted(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitT(t, root, "init", "-q")
	write(t, root, "AGENTS.md", agentsAt("The flow's text."))
	write(t, root, ".writrun/VERSION", oldTag+"\n")
	write(t, root, ".writrun/settings.json", "{\n  \"stage\": 1\n}\n")
	write(t, root, ".writrun/conventions/commits.md", "# Our commits\n")
	write(t, root, ".writrun/skills/select/SKILL.md", "# Select\n")
	write(t, root, ".writrun/scripts/take.sh", "echo take\n")
	write(t, root, ".writrun/templates/task.md", "# Task\n")
	for _, wf := range []string{"approve", "check", "issues", "progress"} {
		write(t, root, ".github/workflows/writrun-"+wf+".yml", "name: writrun "+wf+"\n")
	}
	write(t, root, "docs/product/a-chapter.md", "# Our own chapter\n")
	write(t, root, "work/tasks/task-0001-a-task.md", "id: task-0001\n")
	gitT(t, root, "add", ".")
	gitT(t, root, "commit", "-q", "-m", "adopt")
	return root
}

// fakeKit is the fake fetch every test drives update through, holding
// the template newTag ships and no clone produced.
func fakeKit(t *testing.T) *kitfetch.Fake {
	t.Helper()
	return kitfetch.NewFake(makeTemplate(t))
}

// realKit is the production fetcher — the one case per command that
// checks the fake against the thing it fakes.
func realKit() kitfetch.Clone {
	return kitfetch.Clone{Files: vfs.OS{}, Git: gitx.Run}
}

// runUpdate drives the command the way the frame does, with every
// question already answered. The fetch is faked unless the case asked
// for the real one: driving update end to end is not a reason to clone
// (spec-0016).
func runUpdate(t *testing.T, root string, d Deps, args ...string) (string, error) {
	t.Helper()
	if d.Tag == "" {
		d.Tag = newTag
	}
	if d.Source == "" {
		d.Source = sourceDefault
	}
	if d.Git == nil {
		d.Git = gitx.Run
	}
	if d.Files == nil {
		d.Files = vfs.OS{}
	}
	if d.Kit == nil {
		d.Kit = fakeKit(t)
	}
	var out bytes.Buffer
	ctx := &command.Ctx{
		Stdout:   &out,
		Stderr:   &out,
		Terminal: &command.FakeTerminal{},
		Root:     root,
		Adopted:  true,
		Yes:      true,
	}
	err := run(ctx, d, args)
	return out.String(), err
}
