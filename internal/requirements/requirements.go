// Package requirements is what the wrapped scripts need on the PATH,
// listed once. `init` names what is missing as a gap and `doctor` as a
// finding; the list they read is the same
// (docs/technical/runtime/requirements.md).
package requirements

// Binaries are the wrapped scripts' own requirements, and this binary
// adds none. `gh` is not among them: it is asked for at stage 2, where
// the flows already reach the forge.
var Binaries = []string{"git", "bash", "awk", "sed"}

// Missing returns the binaries lookPath cannot find, in Binaries'
// order. Every one is named, so a reader installs all of them in one
// pass rather than one per run. What a missing binary means is the
// caller's: this reports the fact and grades nothing.
func Missing(lookPath func(name string) (string, error)) []string {
	var missing []string
	for _, bin := range Binaries {
		if _, err := lookPath(bin); err != nil {
			missing = append(missing, bin)
		}
	}
	return missing
}
