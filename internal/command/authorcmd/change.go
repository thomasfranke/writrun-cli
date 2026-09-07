package authorcmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/queue"
)

// The two trees an authoring change may touch: the permanent docs the
// rule is written into, and the queue the work is derived into. They
// are the methodology's own layout, not this command's choice.
const (
	docsTree  = "docs/"
	queueTree = queue.Root + "/"
)

// docsPrefix is what an authoring branch is named with
// (conventions/branches.md).
const docsPrefix = "docs/"

// originRemote is the forge's name in a clone, and the one remote this
// command reads: the branch it would push to, and the base it reads the
// change against, are both under it.
const originRemote = "origin"

// taskTag is the contract marker an implementing title leads with, and
// the one an authoring title never carries.
var taskTag = regexp.MustCompile(`\[TASK-[0-9]{4}\]`)

// change is what the diff says about itself: which branch it is on,
// which range it was read against, and what it touches.
type change struct {
	branch string
	rng    string
	files  []string
	docs   []string
}

// readChange reads the diff this command will open, and refuses
// anything that is not an authoring change. Everything here is
// structural — a path, a branch name, a working tree — so the refusals
// cost no script and no forge call (spec-0009, step 1).
func readChange(git gitx.Runner, root, override string) (change, error) {
	// A dirty tree is a rule half-written: the composition would be
	// read off a diff the branch is not going to carry.
	dirty, err := git(root, "status", "--porcelain")
	if err != nil {
		return change{}, fmt.Errorf("reading the working tree: %w", err)
	}
	if lines := trimmedLines(dirty); len(lines) > 0 {
		return change{}, fmt.Errorf("the working tree is dirty — commit or stash first:\n  %s", strings.Join(head(lines, 5), "\n  "))
	}

	out, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return change{}, fmt.Errorf("reading the current branch: %w", err)
	}
	branch := strings.TrimSpace(out)
	if branch == "" || branch == "HEAD" {
		return change{}, errors.New("this is a detached HEAD — author opens the pull request for a branch")
	}

	// The refs this command reads are a cache of the forge, and every
	// answer under them — the base the diff is read against, and
	// whether a branch is already public — is wrong by exactly however
	// long it has been since the last fetch. take_task.sh refreshes
	// before it reads for the same reason: a stale ref does not report
	// an unknown, it reports a confident wrong answer.
	if err := refresh(git, root); err != nil {
		return change{}, err
	}

	rng := override
	if rng == "" {
		if rng, err = baseRange(git, root); err != nil {
			return change{}, err
		}
	}

	out, err = git(root, "diff", "--name-only", rng)
	if err != nil {
		return change{}, fmt.Errorf("reading the diff: %w", err)
	}
	files := trimmedLines(out)
	if len(files) == 0 {
		return change{}, fmt.Errorf("the diff %s is empty — author opens the pull request for a rule that is already written", rng)
	}

	ch := change{branch: branch, rng: rng, files: files}
	var strays []string
	for _, f := range files {
		switch {
		case strings.HasPrefix(f, docsTree):
			ch.docs = append(ch.docs, f)
		case strings.HasPrefix(f, queueTree):
		default:
			strays = append(strays, f)
		}
	}
	// One kind per change: an authoring change writes a rule and
	// derives the work from it, and nothing else rides it
	// (spec-0009, edge cases).
	if len(strays) > 0 {
		return change{}, fmt.Errorf("the diff %s touches %s outside docs/ and work/ — one kind per change, so author opens a rule and nothing beside it",
			rng, strings.Join(head(strays, 5), ", "))
	}
	if len(ch.docs) == 0 {
		return change{}, fmt.Errorf("the diff %s touches no docs/ path — author opens the pull request that writes a rule", rng)
	}
	return ch, nil
}

// refresh brings the remote-tracking refs up to date before anything is
// read off them. A repository with no `origin` is not stale, it is
// local, and the local `main` below is the whole answer; a fetch that
// failed is neither, so it is a refusal rather than a read of whatever
// the cache still held.
func refresh(git gitx.Runner, root string) error {
	out, err := git(root, "remote")
	if err != nil {
		return fmt.Errorf("reading the remotes: %w", err)
	}
	found := false
	for _, r := range trimmedLines(out) {
		if r == originRemote {
			found = true
		}
	}
	if !found {
		return nil
	}
	if _, err := git(root, "fetch", "--quiet", originRemote); err != nil {
		return fmt.Errorf("%w\nThe base and the branch would be read against a stale %s, so nothing was done", err, originRemote)
	}
	return nil
}

// baseRange resolves the range the checks read the change against, the
// way the kit's own scripts resolve theirs: the pushed main, else the
// local one, else nothing to read against and the caller is asked to
// say.
func baseRange(git gitx.Runner, root string) (string, error) {
	bases := []struct{ ref, rng string }{
		{"refs/remotes/origin/main", "origin/main...HEAD"},
		{"refs/heads/main", "main...HEAD"},
	}
	for _, b := range bases {
		if _, err := git(root, "rev-parse", "--verify", "--quiet", b.ref); err == nil {
			return b.rng, nil
		}
	}
	return "", errors.New("no origin/main and no main to read this change against — name a range with --range")
}

// branchName is `docs/<short-name>` (conventions/branches.md), and the
// short name is not typed unless it has to be: a branch that is already
// an authoring one keeps its name, `--slug` overrides everything, and
// otherwise the subject is the doc the rule was written into.
func branchName(ch change, slug string) (string, error) {
	if slug != "" {
		s := normalize(slug)
		if s == "" {
			return "", fmt.Errorf("--slug %q leaves no subject words", slug)
		}
		return docsPrefix + s, nil
	}
	if strings.HasPrefix(ch.branch, docsPrefix) && len(ch.branch) > len(docsPrefix) {
		return ch.branch, nil
	}
	// `docs` is git's own listing, so this is the first `docs/` path in
	// path order — not, where a change writes into more than one, a
	// judgement about which of them the rule is. `--slug` makes that
	// judgement, and the composition is shown before it acts.
	if s := subjectOf(ch.docs[0]); s != "" {
		return docsPrefix + s, nil
	}
	return "", fmt.Errorf("no subject words could be read from %s — name them with --slug", ch.docs[0])
}

// subjectOf reads a branch's subject off the doc the rule was written
// into: the file's own name, or the folder's when the file is the
// folder's README.
func subjectOf(doc string) string {
	segs := strings.Split(strings.TrimSuffix(doc, ".md"), "/")
	name := segs[len(segs)-1]
	if strings.EqualFold(name, "README") && len(segs) > 1 {
		name = segs[len(segs)-2]
	}
	return normalize(name)
}

// normalize is the branch alphabet: lowercase words joined by single
// dashes, at most three of them — the same shape `new.sh --slug` writes.
var notWord = regexp.MustCompile(`[^a-z0-9]+`)

func normalize(s string) string {
	s = notWord.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return ""
	}
	words := strings.Split(s, "-")
	if len(words) > 3 {
		words = words[:3]
	}
	return strings.Join(words, "-")
}

// trimmedLines is one git listing, blank lines dropped.
func trimmedLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			lines = append(lines, t)
		}
	}
	return lines
}

func head(lines []string, n int) []string {
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}
