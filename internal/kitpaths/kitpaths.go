// Package kitpaths is the inventory of what `writrun init` installs,
// written down once. update refreshes a part of it, uninstall removes
// all of it, and neither may reach what the project owns — three
// answers that have to agree, so they read from one list.
package kitpaths

// Workflows are the four workflow files the kit installs. They live
// outside `.writrun/` because the forge only runs them from here.
var Workflows = []string{
	".github/workflows/writrun-approve.yml",
	".github/workflows/writrun-check.yml",
	".github/workflows/writrun-issues.yml",
	".github/workflows/writrun-progress.yml",
}

// RefreshDirs are the kit-owned directories `writrun update` replaces
// whole: everything in them is the kit's, so a hand edit inside one is
// overwritten by design — shown in the diff first (spec-0003).
var RefreshDirs = []string{
	".writrun/skills",
	".writrun/scripts",
	".writrun/templates",
}

// RefreshFiles are the single files a refresh rewrites. `AGENTS.md` is
// not among them: only its fenced section is refreshed, and that is
// the fence package's work, never a whole-file copy.
func RefreshFiles() []string {
	return append([]string{".writrun/VERSION"}, Workflows...)
}

// Untouchable is what a refresh never writes, whatever the new tag
// carries: the adopter's conventions and settings, the project's docs,
// and the queue (spec-0003, scope).
var Untouchable = []string{
	".writrun/conventions",
	".writrun/settings.json",
	"docs",
	"work",
}

// RemoveDirs and RemoveFiles are what `writrun uninstall` deletes. The
// commit-message hook is neither: it lives outside the worktree and is
// resolved through git, so the command carries it separately.
var RemoveDirs = []string{".writrun"}

// RemoveFiles are the kit's single files outside `.writrun/`.
func RemoveFiles() []string {
	return append([]string{"WRITRUN.md", "docs/writrun-instructions.md"}, Workflows...)
}

// Keep is what uninstall leaves standing: the queue is the project's
// record, not the kit's, and the docs the methodology helped write
// belong to the repository (docs/product/adoption/uninstall.md).
var Keep = []string{"work", "docs"}
