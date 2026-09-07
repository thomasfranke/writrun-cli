package kitpaths

import "testing"

func TestTheAdoptersOwnPathsAreUntouched(t *testing.T) {
	for _, rel := range []string{
		".writrun/settings.json",
		".writrun/gates.md",
		".writrun/conventions/commits.md",
		"AGENTS.md",
		"CLAUDE.md",
		"docs/product/rules.md",
		"work/tasks/task-0001-a-thing.md",
	} {
		if !Untouched(rel) {
			t.Errorf("Untouched(%q) = false — a refresh would overwrite the project's own file", rel)
		}
	}
}

func TestTheKitsOwnFilesAreTouchedWhereverTheyLive(t *testing.T) {
	for _, rel := range []string{
		".writrun/AGENTS.md",
		".writrun/README.md",
		".writrun/VERSION",
		".writrun/skills/writrun-check-task-state/check_state.sh",
		".github/workflows/writrun-intake.yml",
		".github/ISSUE_TEMPLATE/writrun-report.yml",
		"WRITRUN.md",
		// The one file the kit owns under an untouchable path.
		"docs/writrun-instructions.md",
	} {
		if Untouched(rel) {
			t.Errorf("Untouched(%q) = true — a refresh would leave the kit's own file at the tag that installed it", rel)
		}
	}
}

func TestASeededFileIsAlsoUntouchable(t *testing.T) {
	// Seeding writes it once; untouchable keeps every later refresh off
	// it. A seed that were not untouchable would be overwritten on the
	// next run, which is the answer this pair exists to avoid.
	for _, rel := range Seeded {
		if !Seeds(rel) {
			t.Errorf("Seeds(%q) = false", rel)
		}
		if !Untouched(rel) {
			t.Errorf("Untouched(%q) = false — a later refresh would overwrite the seed", rel)
		}
	}
}

func TestTheSettingsAreNotSeeded(t *testing.T) {
	// Its shipped default declares `stage: 1`, which is an answer a
	// refresh may not give on the project's behalf.
	if Seeds(".writrun/settings.json") {
		t.Error("settings.json is seeded — a refresh would declare a stage the project did not choose")
	}
}

func TestOnlyTheKitsOwnFilesAreRemovable(t *testing.T) {
	removable := []string{
		".writrun/skills/writrun-select-next-task/list_tasks.sh",
		".writrun/AGENTS.md",
		".github/workflows/writrun-progress.yml",
		".github/ISSUE_TEMPLATE/writrun-report.yml",
	}
	for _, rel := range removable {
		if !Removable(rel) {
			t.Errorf("Removable(%q) = false — a file the tag dropped would linger", rel)
		}
	}
	for _, rel := range []string{
		".github/workflows/tests.yml",
		".github/workflows/release.yml",
		".writrun/settings.json",
		".writrun/gates.md",
		".writrun/conventions/prs.md",
		"docs/about.md",
		"work/tasks/task-0001-a-thing.md",
	} {
		if Removable(rel) {
			t.Errorf("Removable(%q) = true — a refresh would delete what is not the kit's", rel)
		}
	}
}

func TestTheNamespaceIsReadOneFolderDeep(t *testing.T) {
	if !Namespaced(".github/workflows/writrun-check.yml") {
		t.Error("a workflow the kit names is not recognised")
	}
	if Namespaced(".github/workflows/nested/writrun-check.yml") {
		t.Error("a nested file is recognised — the namespace is one folder deep")
	}
	if Namespaced(".github/workflows/release.yml") {
		t.Error("a workflow the project wrote is recognised as the kit's")
	}
}

func TestTheKeepSetIsNotRemoved(t *testing.T) {
	remove := map[string]bool{}
	for _, rel := range RemoveFiles() {
		remove[rel] = true
	}
	for _, keep := range Keep {
		if remove[keep] {
			t.Errorf("%s is in the removal set although it is the project's", keep)
		}
		for _, dir := range RemoveDirs {
			if dir == keep {
				t.Errorf("%s is removed whole although it is the project's", keep)
			}
		}
	}
	// The one file under docs/ the kit does own.
	if !remove["docs/writrun-instructions.md"] {
		t.Error("the kit's own instructions doc is not removed")
	}
}
