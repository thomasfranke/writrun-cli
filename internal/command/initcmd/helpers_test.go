package initcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// testTag is the tag every test fixture pins; the tests' Deps name it
// too, so the clone and the record agree.
const testTag = "v9.9.9"

// gitT runs one git invocation in a test repository, with an identity
// so commits work on a bare CI runner, and fails the test on error.
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

func rm(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(root, rel)); err != nil {
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

const templateAgents = `# AGENTS.md — entry point for AI agents

<!-- TODO: one paragraph. -->

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. -->

### Picking work

The flow's text.

<!-- writrun:end -->
`

const templateCommits = "# Commits\n\n" +
	"- **Types**: `docs`, `feat`, `fix`, `refactor`, `chore`.\n" +
	"- **Scopes** (optional — omit when a change genuinely spans the\n" +
	"  repository): `about`, `product`, `technical`.\n" +
	"- Example: `docs(product): add a chapter`.\n"

const templateObservance = `#!/usr/bin/env bash
# check_observance.sh — the door.
TYPES="docs feat fix refactor chore"
SCOPES="about product technical"
exit 0
`

const templateSettings = `{
  "stage": 1,
  "stage_2": {
    "auto_commit": false
  }
}
`

// writeTemplate writes the kit's template tree under dir: the files
// init touches, one copy, whether a clone put them there or not.
func writeTemplate(t *testing.T, dir string) {
	t.Helper()
	write(t, dir, "AGENTS.md", templateAgents)
	write(t, dir, "WRITRUN.md", "# This project uses WritRun\n")
	write(t, dir, ".writrun/settings.json", templateSettings)
	write(t, dir, ".writrun/VERSION", testTag+"\n")
	write(t, dir, ".writrun/conventions/commits.md", templateCommits)
	write(t, dir, ".writrun/scripts/stage-2-pull-requests/check_observance.sh", templateObservance)
	if err := os.Chmod(filepath.Join(dir, ".writrun/scripts/stage-2-pull-requests/check_observance.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "docs/product/README.md", "# Product skeleton\n")
	write(t, dir, "docs/technical/README.md", "# Technical skeleton\n")
	write(t, dir, "work/tasks/README.md", "# Tasks\n")
	write(t, dir, "work/specs/README.md", "# Specs\n")
	write(t, dir, "work/reports/README.md", "# Reports\n")
}

// makeTemplate is the template a fake fetch hands over: the same tree
// makeSource commits, without the repository and without the clone.
func makeTemplate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeTemplate(t, dir)
	return dir
}

// fakeKit is the fake fetch every test drives init through, holding a
// template no clone produced.
func fakeKit(t *testing.T) *kitfetch.Fake {
	t.Helper()
	return kitfetch.NewFake(makeTemplate(t))
}

// realKit is the production fetcher — the one case per command that
// checks the fake against the thing it fakes.
func realKit() kitfetch.Clone {
	return kitfetch.Clone{Files: vfs.OS{}, Git: gitx.Run}
}

// makeSource builds a local WritRun repository: a template/ with the
// files init touches, committed and tagged testTag.
func makeSource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	gitT(t, src, "init", "-q")
	writeTemplate(t, filepath.Join(src, "template"))
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "the kit")
	gitT(t, src, "tag", testTag)
	return src
}

// makeTarget builds a repository to adopt: one commit, a clean tree.
func makeTarget(t *testing.T, subjects ...string) string {
	t.Helper()
	target := t.TempDir()
	gitT(t, target, "init", "-q")
	if len(subjects) == 0 {
		subjects = []string{"initial import"}
	}
	write(t, target, "README.md", "# A project\n")
	gitT(t, target, "add", ".")
	for i, s := range subjects {
		if i == 0 {
			gitT(t, target, "commit", "-q", "-m", s)
			continue
		}
		gitT(t, target, "commit", "-q", "--allow-empty", "-m", s)
	}
	return target
}
