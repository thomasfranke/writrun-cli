package workcmd

// Launch is one launch the fake was asked for: where it would have
// run, what it would have run, and the arguments it would have carried
// — the brief among them.
type Launch struct {
	Dir  string
	Name string
	Args []string
}

// FakeLauncher is the fake beside the Launcher port. Zero value: every
// launch succeeds and is recorded. A test that wants an agent exiting
// non-zero sets Err to an error carrying the code, instead of
// arranging a program that fails — which would be a real process, and
// the one thing this port exists to keep out of the suite.
type FakeLauncher struct {
	// Err is what every launch returns.
	Err error
	// Launched records every launch, in order. Empty is the assertion
	// a refusal has to pass: nothing was started.
	Launched []Launch
}

// Run is the Launcher: `Launcher(fake.Run)` wires it.
func (f *FakeLauncher) Run(dir, name string, args ...string) error {
	f.Launched = append(f.Launched, Launch{Dir: dir, Name: name, Args: args})
	return f.Err
}
