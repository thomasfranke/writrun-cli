// Package chapter answers one question about a docs folder: does it
// hold a real chapter, a `.md` file that is not the README the kit
// wrote. `init` and `doctor` both ask it of `docs/product` and
// `docs/technical`; the walk is one implementation, and the words each
// command puts on the answer stay with that command.
package chapter

import (
	"io/fs"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// In reports whether dir holds a real chapter. A folder that cannot be
// walked holds none, which is the same answer as an empty one: the
// caller is asking whether the project wrote a doc, not why it did not.
func In(files vfs.FS, dir string) bool {
	found := false
	_ = files.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil || entry == nil || entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".md") && !strings.EqualFold(entry.Name(), "README.md") {
			found = true
		}
		return nil
	})
	return found
}
