// Package kitpaths is what the binary must know about an adopted
// repository's layout that the kit itself does not declare: which paths
// are the adopter's, and how `uninstall` recognises the kit's files
// with no template to read.
//
// What the kit ships is not here. A refresh walks the fetched template
// and writes what it finds, so a WritRun tag that adds a file needs no
// change in this file (docs/technical/engineering/coupling.md, rule 1).
package kitpaths

import "strings"

// Untouchable are the paths a refresh never rewrites: the adopter's
// conventions, settings and gate answers, the entry points a project
// owns whole from v0.0.04 on, its docs and its queue.
var Untouchable = []string{
	".writrun/conventions",
	".writrun/settings.json",
	".writrun/gates.md",
	"AGENTS.md",
	"CLAUDE.md",
	"docs",
	"work",
}

// KitOwned are the kit's own files that sit under an Untouchable path.
// A refresh writes them although the path above them is the project's —
// `docs/` is the project's chapters and one file the kit installed.
var KitOwned = []string{"docs/writrun-instructions.md"}

// Seeded are the adopter-owned files a refresh writes where the tag
// ships one the repository does not have: a kit whose own files
// reference `gates.md` needs it present, and a project that has one
// keeps every word of it.
//
// `settings.json` is deliberately not among them. Its shipped default
// declares `stage: 1`, which is an answer a refresh may not give on the
// project's behalf.
var Seeded = []string{".writrun/gates.md"}

// namespaced are the directories where the kit prefixes its own files,
// which is how `uninstall` and a refresh's removals tell them from the
// project's without a template to read.
var namespaced = []struct{ dir, prefix string }{
	{".github/workflows", "writrun-"},
	{".github/ISSUE_TEMPLATE", "writrun-"},
}

// Untouched reports whether a refresh must leave rel alone. rel is
// slash-separated and relative to the repository root.
func Untouched(rel string) bool {
	for _, own := range KitOwned {
		if rel == own {
			return false
		}
	}
	for _, path := range Untouchable {
		if rel == path || strings.HasPrefix(rel, path+"/") {
			return true
		}
	}
	return false
}

// Seeds reports whether rel is written only where the repository does
// not already have it.
func Seeds(rel string) bool {
	for _, path := range Seeded {
		if rel == path {
			return true
		}
	}
	return false
}

// Removable reports whether a refresh may delete rel when the tag stops
// shipping it: everything under `.writrun/` that is not the adopter's,
// and the kit's namespaced files in the two `.github` folders. A
// workflow the project wrote is neither.
func Removable(rel string) bool {
	if Untouched(rel) {
		return false
	}
	if strings.HasPrefix(rel, ".writrun/") {
		return true
	}
	return Namespaced(rel)
}

// Namespaced reports whether rel is one of the kit's files recognised
// by the prefix it carries rather than by name.
func Namespaced(rel string) bool {
	for _, n := range namespaced {
		if strings.HasPrefix(rel, n.dir+"/") &&
			strings.HasPrefix(rel[len(n.dir)+1:], n.prefix) &&
			!strings.Contains(rel[len(n.dir)+1:], "/") {
			return true
		}
	}
	return false
}

// NamespacedDirs are the folders Namespaced governs, for the callers
// that must list what is there rather than judge one path.
func NamespacedDirs() []string {
	dirs := make([]string, 0, len(namespaced))
	for _, n := range namespaced {
		dirs = append(dirs, n.dir)
	}
	return dirs
}

// RemoveDirs is what `writrun uninstall` deletes whole. The
// commit-message hook is not among them: it lives outside the worktree
// and is resolved through git, so the command carries it separately.
var RemoveDirs = []string{".writrun"}

// RemoveFiles are the kit's single files outside `.writrun/` that carry
// no namespace to be recognised by.
func RemoveFiles() []string {
	return []string{"WRITRUN.md", "docs/writrun-instructions.md"}
}

// Keep is what uninstall leaves standing: the queue is the project's
// record, not the kit's, and the docs the methodology helped write
// belong to the repository (docs/product/adoption/uninstall.md).
var Keep = []string{"work", "docs"}
