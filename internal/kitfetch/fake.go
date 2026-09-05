package kitfetch

// Fake is the fake beside the port: a template tree it hands over, and
// a way to say *this tag fails*. A test that wants a fetch that cannot
// work names the tag and the refusal instead of arranging a clone that
// cannot run — which is what the two removed partial-state tests
// needed, and why they proved nothing once they were gone.
type Fake struct {
	// Template is the directory every successful fetch hands back.
	Template string
	// Cleaned counts the cleanups the caller ran. A fake handing back
	// a directory and no cleanup lets a test pass while the real fetch
	// leaks, so the cleanup is handed over and counted.
	Cleaned int
	// Asked records every fetch, in order.
	Asked []Ask

	fails map[string]func(tag, source string) error
}

// Ask is one fetch the fake was asked for.
type Ask struct {
	Tag    string
	Source string
}

// NewFake returns a fake handing template over for every tag.
func NewFake(template string) *Fake {
	return &Fake{Template: template, fails: map[string]func(tag, source string) error{}}
}

// Fail makes tag fail the way a clone that could not run does: the
// refusal names the tag and the source and says nothing was written.
func (f *Fake) Fail(tag string, cause error) {
	f.fails[tag] = func(tag, source string) error { return errClone(tag, source, cause) }
}

// FailNoTemplate makes tag fail the way a clone carrying no
// `template/` does — a repository, but not a WritRun one, answered
// without a clone.
func (f *Fake) FailNoTemplate(tag string) {
	f.fails[tag] = errNoTemplate
}

// Heal undoes a Fail.
func (f *Fake) Heal(tag string) { delete(f.fails, tag) }

// Fetch hands the template over, or the refusal the tag was given.
func (f *Fake) Fetch(tag, source string) (*Fetched, error) {
	f.Asked = append(f.Asked, Ask{Tag: tag, Source: source})
	if fail, there := f.fails[tag]; there {
		return nil, fail(tag, source)
	}
	return &Fetched{Template: f.Template, Cleanup: func() { f.Cleaned++ }}, nil
}
