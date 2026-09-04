package initcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Stage 0 — environment: the wrapped scripts' own requirements
	// (technical/runtime/requirements.md).
	for _, bin := range []string{"git", "bash", "awk", "sed"} {
		if _, err := d.LookPath(bin); err != nil {
			gaps = append(gaps, gap{0, bin + " is not on the PATH — the wrapped scripts require it"})
		}
	}

	if stage >= 1 {
		gaps = append(gaps, checkFiles(root)...)
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
func checkFiles(root string) []gap {
	var gaps []gap

	if _, err := os.Stat(filepath.Join(root, "docs", "about.md")); err != nil {
		gaps = append(gaps, gap{1, "docs/about.md — an About file is required of the project, and none was found"})
	}
	for _, folder := range []string{"product", "technical"} {
		if !hasRealChapter(filepath.Join(root, "docs", folder)) {
			gaps = append(gaps, gap{1, fmt.Sprintf("docs/%s/ — at least one real %s doc is required beyond the README", folder, folder)})
		}
	}
	for _, rel := range []string{"work/tasks", "work/specs", "work/reports"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			gaps = append(gaps, gap{1, rel + "/ — the queue's folder is missing"})
		}
	}

	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	switch {
	case err != nil:
		gaps = append(gaps, gap{1, "AGENTS.md — the agents' entry point is missing"})
	case !strings.Contains(string(agents), markerBegin) || !strings.Contains(string(agents), markerEnd):
		gaps = append(gaps, gap{1, "AGENTS.md — the fenced writrun:begin/writrun:end markers are damaged"})
	case strings.Contains(string(agents), todoPlaceholder):
		gaps = append(gaps, gap{1, "AGENTS.md — TODOs remain; the four human gates must be answered, not left as placeholders"})
	}

	if v, err := os.ReadFile(filepath.Join(root, ".writrun", "VERSION")); err != nil || strings.TrimSpace(string(v)) == "" {
		gaps = append(gaps, gap{1, ".writrun/VERSION — the kit's tag is not recorded"})
	}

	var settings struct {
		Stage int `json:"stage"`
	}
	raw, err := os.ReadFile(filepath.Join(root, ".writrun", "settings.json"))
	if err != nil || json.Unmarshal(raw, &settings) != nil || settings.Stage < 1 || settings.Stage > 3 {
		gaps = append(gaps, gap{1, ".writrun/settings.json — the settings are not canonical; a stage of 1, 2 or 3 is required"})
	}
	return gaps
}

// todoPlaceholder is the shape the kit's own unanswered gates take —
// an HTML comment, not the bare word. A project whose AGENTS.md merely
// mentions TODO has answered its gates all the same.
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

// hasRealChapter reports whether a docs folder holds any markdown
// beyond its README — a real chapter, not a table of chapters to come.
func hasRealChapter(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".md") && !strings.EqualFold(entry.Name(), "README.md") {
			found = true
		}
		return nil
	})
	return found
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
