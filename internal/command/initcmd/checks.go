package initcmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/chapter"
	"github.com/thomasfranke/writrun-cli/internal/kittag"
	"github.com/thomasfranke/writrun-cli/internal/pointer"
	"github.com/thomasfranke/writrun-cli/internal/queue"
	"github.com/thomasfranke/writrun-cli/internal/requirements"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// gap is one finding of the on-the-spot checks: which stage's
// assumption is unmet and what is expected. Gaps are named, never
// fixed, and never block the adoption (spec-0002); `writrun doctor`
// owns the full examination.
type gap struct {
	Stage int
	Text  string
}

// checkStages runs the chosen stage's doctor checks on the spot, from
// stage 0 up to the declared one — a project is never judged against
// machinery it did not enable (product/adoption/doctor.md).
func checkStages(root string, stage int, d Deps) []gap {
	var gaps []gap

	// Stage 0 — environment: the wrapped scripts' own requirements,
	// listed in internal/requirements. doctor probes the same list.
	for _, bin := range requirements.Missing(d.LookPath) {
		gaps = append(gaps, gap{0, bin + " is not on the PATH — the wrapped scripts require it"})
	}

	if stage >= 1 {
		gaps = append(gaps, checkFiles(d.Files, root)...)
	}
	if stage >= 2 {
		forgeGaps, reachable := checkForge(d)
		gaps = append(gaps, forgeGaps...)
		// Same reason checkForge stops at the first gap: an unusable gh
		// turns every further read into a restatement of the one fault
		// already named.
		if stage >= 3 && reachable {
			if out, err := d.Gh("api", "repos/{owner}/{repo}", "--jq", ".has_issues"); err != nil {
				gaps = append(gaps, gap{3, "whether Issues are enabled could not be read: " + firstLine(err.Error())})
			} else if strings.TrimSpace(out) != "true" {
				gaps = append(gaps, gap{3, "Issues are disabled — the mirror needs somewhere to land"})
			}
		}
	}
	return gaps
}

// checkFiles is stage 1: the three documents the methodology requires
// of the adopter, the docs/ and work/ split, and the gates answered in
// AGENTS.md. The kit's own files were just written, so what can gape
// here is the project's half.
//
// doctor checks stage 1 too, and what the two ask is deliberately not
// shared (task-0019, spec-0018). init reports on the adoption it has
// just run: the stage is the flag's, AGENTS.md is tested for the
// placeholders the kit wrote, and none of the repository's own scripts
// is run against a repository still being set up. doctor reports on a
// repository in use: the stage comes from `read_setting.sh`, each of
// the four gates is named, and `check_front_matter.sh` and
// `check_settings.sh` decide. Each is right for its command, so one
// shared stage-1 would give one of them the other's question.
//
// That reason covers the questions and the words, not the mechanics
// under them: what asks nothing is shared. The path of
// `.writrun/VERSION` and its reading are kittag's, the docs-chapter
// walk is internal/chapter, and the stage-0 list is
// internal/requirements. What stays here is which paths are probed,
// which of them is required, and how a gap is worded.
func checkFiles(disk vfs.FS, root string) []gap {
	var gaps []gap

	if _, err := disk.Stat(filepath.Join(root, "docs", "about.md")); err != nil {
		gaps = append(gaps, gap{1, "docs/about.md — an About file is required of the project, and none was found"})
	}
	for _, folder := range []string{"product", "technical"} {
		if !chapter.In(disk, filepath.Join(root, "docs", folder)) {
			gaps = append(gaps, gap{1, fmt.Sprintf("docs/%s/ — at least one real %s doc is required beyond the README", folder, folder)})
		}
	}
	for _, rel := range []string{queue.TasksDir, queue.SpecsDir, queue.ReportsDir} {
		if _, err := disk.Stat(filepath.Join(root, rel)); err != nil {
			gaps = append(gaps, gap{1, rel + "/ — the queue's folder is missing"})
		}
	}

	agents, err := disk.ReadFile(filepath.Join(root, "AGENTS.md"))
	switch {
	case err != nil:
		gaps = append(gaps, gap{1, "AGENTS.md — the agents' entry point is missing"})
	case !pointer.Has(agents):
		gaps = append(gaps, gap{1, "AGENTS.md — no section links " + pointer.Target + ", so an agent reading it never reaches the flow"})
	case strings.Contains(string(agents), todoPlaceholder):
		gaps = append(gaps, gap{1, "AGENTS.md — a TODO remains; the skeleton's paragraph is the project's to write"})
	}

	// The gates are the kit's own declaration, and the adoption just
	// copied it with every answer still a placeholder. Naming the file
	// is the whole of what init can say: which answers are right is the
	// project's, and doctor names them one by one once the repository
	// is in use (task-0019).
	gates, err := disk.ReadFile(filepath.Join(root, ".writrun", "gates.md"))
	switch {
	case err != nil:
		gaps = append(gaps, gap{1, ".writrun/gates.md — the project's gate answers are missing"})
	case strings.Contains(string(gates), todoPlaceholder):
		gaps = append(gaps, gap{1, ".writrun/gates.md — the gates are still the kit's TODOs; each must be answered"})
	}

	// The file and its reading are kittag's; what an unrecorded tag
	// means to an adoption just run is this command's, and it is a gap
	// rather than a graded finding.
	if tag, err := kittag.Read(disk, root); err != nil || tag == "" {
		gaps = append(gaps, gap{1, ".writrun/VERSION — the kit's tag is not recorded"})
	}

	var settings struct {
		Stage int `json:"stage"`
	}
	raw, err := disk.ReadFile(filepath.Join(root, ".writrun", "settings.json"))
	if err != nil || json.Unmarshal(raw, &settings) != nil || settings.Stage < 1 || settings.Stage > 3 {
		gaps = append(gaps, gap{1, ".writrun/settings.json — the settings are not canonical; a stage of 1, 2 or 3 is required"})
	}
	return gaps
}

// todoPlaceholder is the shape the kit's own placeholders take — an
// HTML comment, not the bare word. A project whose prose merely
// mentions TODO has answered all the same.
const todoPlaceholder = "<!-- TODO"

// checkForge is stage 2: the forge reachable and the settings the
// recording machinery depends on. An unauthenticated gh makes every
// further read noise, so it is the one gap named; the bool says whether
// the forge answered at all, so the stages above do not pile more
// unreadable reads on top of it.
func checkForge(d Deps) ([]gap, bool) {
	if _, err := d.LookPath("gh"); err != nil {
		return []gap{{2, "gh is not on the PATH — from stage 2 the flows ask the forge through it"}}, false
	}
	if _, err := d.Gh("auth", "status"); err != nil {
		return []gap{{2, "gh is not authenticated — run `gh auth login`"}}, false
	}
	var gaps []gap
	if out, err := d.Gh("api", "repos/{owner}/{repo}", "--jq", ".allow_squash_merge"); err != nil {
		gaps = append(gaps, gap{2, "whether squash merging is on could not be read: " + firstLine(err.Error())})
	} else if strings.TrimSpace(out) != "true" {
		gaps = append(gaps, gap{2, "squash merging is off — the methodology lands every pull request as one commit"})
	}
	if out, err := d.Gh("api", "repos/{owner}/{repo}/actions/permissions/workflow", "--jq", ".default_workflow_permissions"); err != nil {
		gaps = append(gaps, gap{2, "the Actions workflow permissions could not be read: " + firstLine(err.Error())})
	} else if strings.TrimSpace(out) != "write" {
		gaps = append(gaps, gap{2, "Actions workflow permissions are read-only — the recording bot needs read-and-write to push to main"})
	}
	return gaps, true
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
