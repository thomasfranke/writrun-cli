package initcmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// agentsAction is what the plan decided about AGENTS.md.
type agentsAction int

const (
	agentsSkeleton agentsAction = iota // absent — the template's skeleton is written
	agentsGraft                        // present — the fenced section is appended
	agentsKept                         // present and already carrying the markers
)

// copyStep is one file the adoption writes: where it comes from in the
// clone, where it lands, and the mode it keeps.
type copyStep struct {
	src  string
	rel  string
	mode fs.FileMode
}

// adoption is the whole plan, computed before anything is written and
// shown before the confirmation (spec-0002).
type adoption struct {
	root     string
	template string
	tag      string
	stage    int

	copies   []copyStep
	kept     []string // rel paths the project already owns, left alone
	agents   agentsAction
	vocab    vocabulary
	hookPath string
}

// ownedSkeletons are template files that are skeletons for documents
// the methodology requires of the project — where a real one exists it
// is the project's and stays (spec-0002).
var ownedSkeletons = map[string]bool{
	filepath.Join("docs", "product", "README.md"):   true,
	filepath.Join("docs", "technical", "README.md"): true,
}

// plan walks the fetched template and decides every write without
// performing one.
func plan(root, template, tag string, stage int, hookAt string, git gitRunner) (*adoption, error) {
	a := &adoption{root: root, template: template, tag: tag, stage: stage, hookPath: hookAt}

	err := filepath.WalkDir(template, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(template, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if rel == "AGENTS.md" {
			a.agents = agentsDecision(filepath.Join(root, "AGENTS.md"))
			return nil
		}
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			// The project's file always wins — an existing one is
			// never overwritten (product/rules.md), and the two named
			// skeletons are the expected case of it.
			a.kept = append(a.kept, rel)
			return nil
		}
		a.copies = append(a.copies, copyStep{src: path, rel: rel, mode: info.Mode()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading the template: %w", err)
	}

	a.vocab = extractVocabulary(root, git)
	return a, nil
}

func agentsDecision(path string) agentsAction {
	content, err := os.ReadFile(path)
	if err != nil {
		return agentsSkeleton
	}
	if strings.Contains(string(content), markerBegin) {
		return agentsKept
	}
	return agentsGraft
}

// render prints the plan as plain text, one decision per line — the
// whole of what the confirmation is about.
func (a *adoption) render(w io.Writer) {
	fmt.Fprintln(w, "writrun init — the plan; nothing is written before the confirmation:")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  source       WritRun %s\n", a.tag)
	fmt.Fprintf(w, "  stage        %d — %s\n", a.stage, stageNames[a.stage-1])
	fmt.Fprintf(w, "  copy         %s\n", summarizeCopies(a.copies))
	for _, rel := range a.kept {
		fmt.Fprintf(w, "  kept         %s — the project's own file stays\n", filepath.ToSlash(rel))
	}
	switch a.agents {
	case agentsSkeleton:
		fmt.Fprintln(w, "  AGENTS.md    absent — the skeleton is written; its TODOs are yours to answer")
	case agentsGraft:
		fmt.Fprintln(w, "  AGENTS.md    graft — the fenced WritRun section is appended; every byte outside it stays")
	case agentsKept:
		fmt.Fprintln(w, "  AGENTS.md    already carries the fenced markers — left alone")
	}
	if len(a.vocab.Types) == 0 {
		fmt.Fprintln(w, "  conventions  shipped defaults — no Conventional history and no contributing guide to extract from")
	} else {
		line := "  conventions  extracted from " + a.vocab.Source + ": types " + strings.Join(a.vocab.Types, ", ")
		if len(a.vocab.Scopes) > 0 {
			line += "; scopes " + strings.Join(a.vocab.Scopes, ", ")
		} else {
			line += "; scopes stay shipped — the history never used one"
		}
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w, "  hook         .git/hooks/commit-msg validates the Conventional subject; it never writes one")
	fmt.Fprintf(w, "  settings     .writrun/settings.json records stage %d\n", a.stage)
	fmt.Fprintf(w, "  version      .writrun/VERSION records %s\n", a.tag)
	fmt.Fprintln(w)
}

var stageNames = []string{"files", "pull requests", "GitHub issues"}

// summarizeCopies condenses the copy list into counts per top-level
// entry — sixty file names are a wall, not a plan.
func summarizeCopies(copies []copyStep) string {
	counts := map[string]int{}
	for _, c := range copies {
		top := filepath.ToSlash(c.rel)
		if i := strings.IndexByte(top, '/'); i >= 0 {
			top = top[:i] + "/"
		}
		counts[top]++
	}
	tops := make([]string, 0, len(counts))
	for t := range counts {
		tops = append(tops, t)
	}
	sort.Strings(tops)
	parts := make([]string, 0, len(tops))
	for _, t := range tops {
		if strings.HasSuffix(t, "/") {
			parts = append(parts, fmt.Sprintf("%s (%d files)", t, counts[t]))
		} else {
			parts = append(parts, t)
		}
	}
	return fmt.Sprintf("%d files: %s", len(copies), strings.Join(parts, ", "))
}

var stageLineRE = regexp.MustCompile(`"stage":\s*\d+`)

// apply performs exactly the rendered plan, in an order where the
// generic copies land first and the targeted writes follow.
func (a *adoption) apply() error {
	for _, c := range a.copies {
		dst := filepath.Join(a.root, c.rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("copying %s: %w", c.rel, err)
		}
		content, err := os.ReadFile(c.src)
		if err != nil {
			return fmt.Errorf("copying %s: %w", c.rel, err)
		}
		if err := os.WriteFile(dst, content, c.mode.Perm()); err != nil {
			return fmt.Errorf("copying %s: %w", c.rel, err)
		}
	}

	// The chosen stage lands in the copied settings by targeted
	// replacement — the rest of the file stays byte-for-byte the
	// shipped default, which is the adopter's to edit next.
	if err := rewriteFile(filepath.Join(a.root, ".writrun", "settings.json"), func(s string) string {
		return stageLineRE.ReplaceAllString(s, fmt.Sprintf(`"stage": %d`, a.stage))
	}); err != nil {
		return err
	}

	// The tag is recorded from what was actually fetched, never
	// trusted from the clone's own file (spec-0002).
	if err := os.WriteFile(filepath.Join(a.root, ".writrun", "VERSION"), []byte(a.tag+"\n"), 0o644); err != nil {
		return fmt.Errorf("recording the tag: %w", err)
	}

	if err := a.applyAgents(); err != nil {
		return err
	}
	if err := applyVocabulary(a.root, a.vocab); err != nil {
		return err
	}
	return installHook(a.hookPath)
}

func (a *adoption) applyAgents() error {
	templateAgents, err := os.ReadFile(filepath.Join(a.template, "AGENTS.md"))
	if err != nil {
		return fmt.Errorf("reading the template's AGENTS.md: %w", err)
	}
	dst := filepath.Join(a.root, "AGENTS.md")
	switch a.agents {
	case agentsSkeleton:
		return os.WriteFile(dst, templateAgents, 0o644)
	case agentsGraft:
		section, err := graftSection(templateAgents)
		if err != nil {
			return err
		}
		existing, err := os.ReadFile(dst)
		if err != nil {
			return fmt.Errorf("grafting AGENTS.md: %w", err)
		}
		return os.WriteFile(dst, graft(existing, section), 0o644)
	}
	return nil
}
