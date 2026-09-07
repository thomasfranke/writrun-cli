package initcmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// fakeDeps is the checks' wiring with every external faked: PATH has
// everything, and gh answers from a script of canned replies.
func fakeDeps(gh map[string]string, ghErr map[string]error) Deps {
	return Deps{
		Files:    vfs.OS{},
		LookPath: func(name string) (string, error) { return "/usr/bin/" + name, nil },
		Gh: func(args ...string) (string, error) {
			key := strings.Join(args, " ")
			if err, ok := ghErr[key]; ok {
				return "", err
			}
			return gh[key], nil
		},
	}
}

// adoptedFixture is a target that already went through apply: the kit
// files present, the project's three documents in place.
func adoptedFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, "docs/about.md", "# About\n")
	write(t, root, "docs/product/rules.md", "# Rules\n")
	write(t, root, "docs/technical/architecture.md", "# Architecture\n")
	write(t, root, "work/tasks/README.md", "# Tasks\n")
	write(t, root, "work/specs/README.md", "# Specs\n")
	write(t, root, "work/reports/README.md", "# Reports\n")
	write(t, root, "AGENTS.md", strings.ReplaceAll(templateAgents, "<!-- TODO: one paragraph. -->", "A project."))
	write(t, root, ".writrun/VERSION", testTag+"\n")
	write(t, root, ".writrun/settings.json", `{"stage": 1}`)
	write(t, root, ".writrun/gates.md", strings.ReplaceAll(templateGates,
		"<!-- TODO — default: human reviews -->", "Thomas reviews before merge."))
	return root
}

func TestCheckStagesAllClear(t *testing.T) {
	root := adoptedFixture(t)
	d := fakeDeps(map[string]string{
		"api repos/{owner}/{repo} --jq .allow_squash_merge":                                        "true\n",
		"api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions": "write\n",
		"api repos/{owner}/{repo} --jq .has_issues":                                                "true\n",
	}, nil)
	if gaps := checkStages(root, 3, d); len(gaps) != 0 {
		t.Errorf("gaps = %v, want none", gaps)
	}
}

func TestCheckStagesNamesTheFileGaps(t *testing.T) {
	root := adoptedFixture(t)
	cases := []struct {
		name   string
		breaks func()
		names  string
	}{
		{"a missing About file", func() { rm(t, root, "docs/about.md") }, "docs/about.md"},
		{"a product folder of only READMEs", func() { rm(t, root, "docs/product/rules.md") }, "docs/product/"},
		{"a missing queue folder", func() { rm(t, root, "work/reports") }, "work/reports/"},
		{"an unanswered TODO in AGENTS.md", func() { write(t, root, "AGENTS.md", templateAgents) }, "a TODO remains"},
		{"no section linking the kit's flow", func() { write(t, root, "AGENTS.md", "# AGENTS.md\nnothing of WritRun's\n") }, "no section links"},
		{"gates still the kit's TODOs", func() { write(t, root, ".writrun/gates.md", templateGates) }, "still the kit's TODOs"},
		{"a missing gates file", func() { rm(t, root, ".writrun/gates.md") }, ".writrun/gates.md"},
		{"an unrecorded kit version", func() { write(t, root, ".writrun/VERSION", "\n") }, ".writrun/VERSION"},
		{"settings that do not parse", func() { write(t, root, ".writrun/settings.json", "{") }, "not canonical"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root = adoptedFixture(t)
			tc.breaks()
			gaps := checkStages(root, 1, fakeDeps(nil, nil))
			if len(gaps) != 1 || !strings.Contains(gaps[0].Text, tc.names) {
				t.Errorf("gaps = %v, want one naming %q", gaps, tc.names)
			}
		})
	}
}

func TestCheckStagesNamesTheForgeGaps(t *testing.T) {
	root := adoptedFixture(t)
	t.Run("an unauthenticated gh is the one gap named", func(t *testing.T) {
		d := fakeDeps(nil, map[string]error{"auth status": errors.New("not logged in")})
		gaps := checkStages(root, 2, d)
		if len(gaps) != 1 || !strings.Contains(gaps[0].Text, "not authenticated") {
			t.Errorf("gaps = %v, want only the authentication gap", gaps)
		}
	})
	t.Run("squash off and read-only permissions are both named", func(t *testing.T) {
		d := fakeDeps(map[string]string{
			"api repos/{owner}/{repo} --jq .allow_squash_merge":                                        "false\n",
			"api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions": "read\n",
		}, nil)
		gaps := checkStages(root, 2, d)
		if len(gaps) != 2 {
			t.Fatalf("gaps = %v, want two", gaps)
		}
	})
	t.Run("disabled issues gap at stage 3", func(t *testing.T) {
		d := fakeDeps(map[string]string{
			"api repos/{owner}/{repo} --jq .allow_squash_merge":                                        "true\n",
			"api repos/{owner}/{repo}/actions/permissions/workflow --jq .default_workflow_permissions": "write\n",
			"api repos/{owner}/{repo} --jq .has_issues":                                                "false\n",
		}, nil)
		gaps := checkStages(root, 3, d)
		if len(gaps) != 1 || !strings.Contains(gaps[0].Text, "Issues are disabled") {
			t.Errorf("gaps = %v, want the issues gap", gaps)
		}
	})
}

func TestCheckStagesNamesAMissingBinary(t *testing.T) {
	root := adoptedFixture(t)
	d := fakeDeps(nil, nil)
	d.LookPath = func(name string) (string, error) {
		if name == "awk" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + name, nil
	}
	gaps := checkStages(root, 1, d)
	if len(gaps) != 1 || !strings.Contains(gaps[0].Text, "awk is not on the PATH") {
		t.Errorf("gaps = %v, want the awk gap", gaps)
	}
}

func TestCheckStagesStopsAtTheForgeGapAtStageThree(t *testing.T) {
	root := adoptedFixture(t)
	d := fakeDeps(nil, nil)
	d.LookPath = func(name string) (string, error) {
		if name == "gh" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + name, nil
	}
	// A gh that cannot answer makes the stage-3 read a restatement of
	// the stage-2 gap, not a finding of its own.
	gaps := checkStages(root, 3, d)
	if len(gaps) != 1 || !strings.Contains(gaps[0].Text, "gh is not on the PATH") {
		t.Errorf("gaps = %v, want only the missing-gh gap", gaps)
	}
}
