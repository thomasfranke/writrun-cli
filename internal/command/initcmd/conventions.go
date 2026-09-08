package initcmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
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
func extractVocabulary(disk vfs.FS, root string, git gitx.Runner) vocabulary {
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
		content, err := disk.ReadFile(filepath.Join(root, rel))
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

// applyVocabulary writes the extracted vocabulary where the checks read
// it: `stage_2.commit_types` and `stage_2.commit_scopes` in the
// adopter's settings. An empty vocabulary changes nothing — the shipped
// defaults stand. Scopes stay shipped when the project's history never
// used one, because scopes are optional and absence is no vote against
// the shipped list.
//
// It used to write both halves of a statement kept in two places: the
// prose lists in `conventions/commits.md` and the TYPES/SCOPES lines in
// `check_observance.sh`. The second was a kit file, and a refresh
// replaces every file in the kit's home — so an extracted vocabulary
// survived until the first update and then silently reverted, while the
// prose half went on claiming it. From WritRun v0.0.07 the value has one
// home and the convention only explains it
// (docs/technical/engineering/coupling.md, rule 3).
func applyVocabulary(disk vfs.FS, root string, v vocabulary) error {
	if len(v.Types) == 0 {
		return nil
	}
	return rewriteFile(disk, filepath.Join(root, filepath.FromSlash(kit.Settings)), func(s string) (string, error) {
		out, err := replaceSetting(s, "commit_types", strings.Join(v.Types, " "))
		if err != nil {
			return "", err
		}
		if len(v.Scopes) > 0 {
			if out, err = replaceSetting(out, "commit_scopes", strings.Join(v.Scopes, " ")); err != nil {
				return "", err
			}
		}
		return out, nil
	})
}

// replaceSetting rewrites one string-valued key in the settings file.
// A miss is an error rather than a silent no-op: the run would otherwise
// report a vocabulary the file does not record.
func replaceSetting(s, key, value string) (string, error) {
	re := regexp.MustCompile(`"` + key + `":\s*"[^"]*"`)
	if !re.MatchString(s) {
		return "", fmt.Errorf("no %q key in %s to write the extracted vocabulary into", key, kit.Settings)
	}
	return re.ReplaceAllString(s, `"`+key+`": "`+value+`"`), nil
}

// rewriteFile applies fn to a file's content in place, keeping its
// mode. fn returns an error where the content is not the shape the
// rewrite needs: a rewrite that silently matched nothing would leave
// the kit saying something other than what the plan promised.
func rewriteFile(disk vfs.FS, path string, fn func(string) (string, error)) error {
	info, err := disk.Stat(path)
	if err != nil {
		return fmt.Errorf("rewriting %s: %w", path, err)
	}
	content, err := disk.ReadFile(path)
	if err != nil {
		return fmt.Errorf("rewriting %s: %w", path, err)
	}
	next, err := fn(string(content))
	if err != nil {
		return fmt.Errorf("rewriting %s: %w", path, err)
	}
	return disk.WriteFile(path, []byte(next), info.Mode())
}
