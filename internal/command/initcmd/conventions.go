package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// vocabulary is what extraction found: the commit types and scopes the
// repository already uses. Empty Types means nothing was found and the
// shipped defaults stand.
type vocabulary struct {
	Types  []string
	Scopes []string
	// Source names where the vocabulary came from, for the plan and
	// the adopter reading it: "the commit history", "the contributing
	// guide", or both. Empty with empty Types: shipped defaults.
	Source string
}

// subjectRE is the Conventional subject grammar, the same shape the
// installed hook enforces: type, optional scope, an imperative summary.
var subjectRE = regexp.MustCompile(`^([a-z]+)(?:\(([a-z0-9-]+)\))?!?: .+$`)

// exampleRE finds Conventional subjects quoted in a contributing
// guide — backticked examples like "feat(api): add the thing".
var exampleRE = regexp.MustCompile("`([a-z]+)(?:\\(([a-z0-9-]+)\\))?!?: [^`]+`")

// guidePaths is where a contributing guide conventionally lives, in
// the order they are looked for; the first that exists is the guide.
var guidePaths = []string{"CONTRIBUTING.md", filepath.Join(".github", "CONTRIBUTING.md"), filepath.Join("docs", "CONTRIBUTING.md")}

// extractVocabulary reads the repository's own conventions — its
// commit history and its contributing guide — rather than imposing the
// shipped defaults (product/adoption/init.md). A repository with
// neither returns the zero vocabulary, and the plan says so.
func extractVocabulary(root string, git gitRunner) vocabulary {
	types := map[string]int{}
	scopes := map[string]int{}
	var sources []string

	// The history: every Conventional subject votes. A repository with
	// no commits has git refuse the log, which is simply no history.
	if out, err := git(root, "log", "--format=%s"); err == nil {
		found := false
		for _, s := range strings.Split(out, "\n") {
			m := subjectRE.FindStringSubmatch(strings.TrimSpace(s))
			if m == nil {
				continue
			}
			found = true
			types[m[1]]++
			if m[2] != "" {
				scopes[m[2]]++
			}
		}
		if found {
			sources = append(sources, "the commit history")
		}
	}

	// The guide: backticked Conventional examples are its declaration
	// of intent, and each counts once.
	for _, rel := range guidePaths {
		content, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		found := false
		for _, m := range exampleRE.FindAllStringSubmatch(string(content), -1) {
			found = true
			types[m[1]]++
			if m[2] != "" {
				scopes[m[2]]++
			}
		}
		if found {
			sources = append(sources, "the contributing guide")
		}
		break
	}

	if len(types) == 0 {
		return vocabulary{}
	}
	return vocabulary{
		Types:  rankVocabulary(types),
		Scopes: rankVocabulary(scopes),
		Source: strings.Join(sources, " and "),
	}
}

// rankVocabulary orders by frequency, most used first, ties
// alphabetical — deterministic, and the list reads most-common-first
// the way a person would write it.
func rankVocabulary(counts map[string]int) []string {
	words := make([]string, 0, len(counts))
	for w := range counts {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		if counts[words[i]] != counts[words[j]] {
			return counts[words[i]] > counts[words[j]]
		}
		return words[i] < words[j]
	})
	return words
}

// applyVocabulary writes the extracted vocabulary onto both halves of
// the copied kit's statement — the prose lists in
// conventions/commits.md and the TYPES/SCOPES lines in
// check_observance.sh — because conventions/commits.md's own rule is
// "change one and change the other". An empty vocabulary changes
// nothing: the shipped defaults stand. Scopes stay shipped when the
// project's history never used one — scopes are optional, so absence
// is no vote against the shipped list.
func applyVocabulary(root string, v vocabulary) error {
	if len(v.Types) == 0 {
		return nil
	}
	commitsPath := filepath.Join(root, ".writrun", "conventions", "commits.md")
	if err := rewriteFile(commitsPath, func(s string) (string, error) {
		s = replaceBullet(s, "- **Types**", "- **Types**: "+backtickList(v.Types)+".")
		if len(v.Scopes) > 0 {
			s = replaceBullet(s, "- **Scopes**", "- **Scopes** (optional — omit when a change genuinely spans the repository): "+backtickList(v.Scopes)+".")
		}
		return rewriteExample(s, v), nil
	}); err != nil {
		return err
	}
	observancePath := filepath.Join(root, ".writrun", "scripts", "stage-2-pull-requests", "check_observance.sh")
	return rewriteFile(observancePath, func(s string) (string, error) {
		s = replaceLinePrefix(s, `TYPES="`, `TYPES="`+strings.Join(v.Types, " ")+`"`)
		if len(v.Scopes) > 0 {
			s = replaceLinePrefix(s, `SCOPES="`, `SCOPES="`+strings.Join(v.Scopes, " ")+`"`)
		}
		return s, nil
	})
}

// exampleBulletRE is the `- Example:` bullet of conventions/commits.md:
// a backticked Conventional subject demonstrating the two lists above
// it.
var exampleBulletRE = regexp.MustCompile("(?m)^(- Example: `)([a-z]+)(?:\\(([a-z0-9-]+)\\))?(!?: [^`]+`.*)$")

// rewriteExample respells the example's type and scope in the extracted
// vocabulary, keeping its summary — the prose is the kit's to teach
// with, the vocabulary is the project's. Left as shipped, the example
// would exhibit a subject the hook installed by the same run refuses.
// The scope follows the same rule as the bullet above it: shipped
// scopes stand where the project's history never used one.
func rewriteExample(content string, v vocabulary) string {
	return exampleBulletRE.ReplaceAllStringFunc(content, func(bullet string) string {
		m := exampleBulletRE.FindStringSubmatch(bullet)
		scope := m[3]
		if len(v.Scopes) > 0 {
			scope = v.Scopes[0]
		}
		if scope != "" {
			scope = "(" + scope + ")"
		}
		return m[1] + v.Types[0] + scope + m[4]
	})
}

func backtickList(words []string) string {
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = "`" + w + "`"
	}
	return strings.Join(quoted, ", ")
}

// replaceBullet swaps one markdown bullet — the line starting with
// prefix and every indented continuation line under it — for the one
// replacement line.
func replaceBullet(content, prefix, replacement string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		if !strings.HasPrefix(lines[i], prefix) {
			out = append(out, lines[i])
			i++
			continue
		}
		out = append(out, replacement)
		i++
		for i < len(lines) && strings.HasPrefix(lines[i], "  ") {
			i++
		}
	}
	return strings.Join(out, "\n")
}

// replaceLinePrefix swaps every line starting with prefix for the
// replacement line.
func replaceLinePrefix(content, prefix, replacement string) string {
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, prefix) {
			lines[i] = replacement
		}
	}
	return strings.Join(lines, "\n")
}

// rewriteFile applies fn to a file's content in place, keeping its
// mode. fn returns an error where the content is not the shape the
// rewrite needs: a rewrite that silently matched nothing would leave
// the kit saying something other than what the plan promised.
func rewriteFile(path string, fn func(string) (string, error)) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("rewriting %s: %w", path, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("rewriting %s: %w", path, err)
	}
	next, err := fn(string(content))
	if err != nil {
		return fmt.Errorf("rewriting %s: %w", path, err)
	}
	return os.WriteFile(path, []byte(next), info.Mode())
}
