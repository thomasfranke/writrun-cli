package updatecmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/gitx"

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

// makeSource builds a local WritRun repository with two tags: oldTag,
// then newTag with a reworded skill, an added template and a reworded
// workflow — one of each verb the plan reports.
func makeSource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	gitT(t, src, "init", "-q")
	write(t, src, "template/AGENTS.md", agentsAt("The flow's text."))
	write(t, src, "template/.writrun/VERSION", oldTag+"\n")
	write(t, src, "template/.writrun/settings.json", "{\n  \"stage\": 1\n}\n")
	write(t, src, "template/.writrun/conventions/commits.md", "# Commits\n")
	write(t, src, "template/.writrun/skills/select/SKILL.md", "# Select\n")
	write(t, src, "template/.writrun/scripts/take.sh", "echo take\n")
	write(t, src, "template/.writrun/templates/task.md", "# Task\n")
	for _, wf := range []string{"approve", "check", "issues", "progress"} {
		write(t, src, "template/.github/workflows/writrun-"+wf+".yml", "name: writrun "+wf+"\n")
	}
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "the kit, one release back")
	gitT(t, src, "tag", oldTag)

	write(t, src, "template/AGENTS.md", agentsAt("The flow's text, reworded."))
	write(t, src, "template/.writrun/VERSION", newTag+"\n")
	write(t, src, "template/.writrun/skills/select/SKILL.md", "# Select, reworded\n")
	write(t, src, "template/.writrun/templates/spec.md", "# Spec\n")
	write(t, src, "template/.github/workflows/writrun-check.yml", "name: writrun check\n# reworded\n")
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

// runUpdate drives the command the way the frame does, with every
// question already answered.
func runUpdate(t *testing.T, root, source string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	ctx := &command.Ctx{
		Stdout:   &out,
		Stderr:   &out,
		Terminal: &command.FakeTerminal{},
		Root:     root,
		Adopted:  true,
		Yes:      true,
	}
	err := run(ctx, Deps{Tag: newTag, Source: source, Git: gitx.Run, Files: vfs.OS{}}, args)
	return out.String(), err
}
