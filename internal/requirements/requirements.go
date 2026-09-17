// Package requirements is what the wrapped scripts need on the PATH,
// listed once. `init` names what is missing as a gap and `doctor` as a
// finding; the list they read is the same
// (docs/technical/runtime/requirements.md).
package requirements

// binaries are the wrapped scripts' own requirements, and this binary
// adds none. `gh` is not among them: it is asked for at stage 2, where
// the flows already reach the forge.
//
// It is unexported and handed out only as a copy: a package-level
// variable another package could write is state this project does not
// keep (technical/engineering/boundaries.md).
var binaries = []string{"git", "bash", "awk", "sed"}

// reasons say what each binary is needed for, in one clause. They sit
// beside the list because they are the same fact: why a name is on it.
//
// `doctor` does not read them — its sentences are
// `docs/product/adoption/doctor.md`'s own table, held to the document
// by a test. This is what the first-run screen says, where there is no
// adopted repository and so no document to read
// (docs/product/screens/first-run.excalidraw).
var reasons = map[string]string{
	"git":  "every flow the kit installs reads and writes history through git",
	"bash": "every kit script is written for bash, and the scripts are the authority",
	"awk":  "the kit's scripts read the queue's front matter with POSIX awk",
	"sed":  "the kit's scripts write the queue's edits with sed",
}

// Reason is the clause for one binary, empty for a name this list does
// not carry.
func Reason(name string) string { return reasons[name] }

// All returns the required binaries, in the order a reader should be
// told about them. The slice is the caller's own — reordering it
// reorders nothing here.
func All() []string {
	out := make([]string, len(binaries))
	copy(out, binaries)
	return out
}

// Missing returns the binaries lookPath cannot find, in All's order.
// Every one is named, so a reader installs all of them in one pass
// rather than one per run. What a missing binary means is the caller's:
// this reports the fact and grades nothing.
func Missing(lookPath func(name string) (string, error)) []string {
	var missing []string
	for _, bin := range binaries {
		if _, err := lookPath(bin); err != nil {
			missing = append(missing, bin)
		}
	}
	return missing
}
