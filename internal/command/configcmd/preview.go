package configcmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/command"
)

// Preview is what a stage raise shows before it writes: one stage's
// requirements, examined and marked, written to w, and the three counts
// the sentence under them quotes.
//
// It is a port because the answer is `doctor`'s. `config` runs it and
// shows it; it judges none of it, and holds no second copy of a check
// (docs/product/config.md, spec-0041).
type Preview func(root string, stage int, w io.Writer) (met, unmet, unread int, err error)

// stageKey is the one key whose new value is a claim about the
// repository rather than about this project's conduct.
const stageKey = "stage"

// stageNames are the rungs as `init` offers them, which is the only
// vocabulary this binary has for them.
var stageNames = map[int]string{1: "files", 2: "pull requests", 3: "GitHub issues"}

// preview shows what the target stage would require of this repository,
// where the change is a raise and a preview can be run at all.
//
// **It never refuses.** The stage is a declaration: a check that could
// not be read is `unread`, not unmet, and an adopter with no network can
// still declare a stage (spec-0041).
func preview(ctx *command.Ctx, d Deps, k key, was, value string) {
	if d.Preview == nil || k.dotted() != stageKey {
		return
	}
	from, err := strconv.Atoi(strings.TrimSpace(was))
	if err != nil {
		return
	}
	to, err := strconv.Atoi(strings.TrimSpace(value))
	// Nothing new is required of the repository when the stage is
	// lowered or unchanged, so nothing is previewed.
	if err != nil || to <= from {
		return
	}

	fmt.Fprintf(ctx.Stdout, " stage %d — %s. doctor previews what the stage would\n        require of this repository:\n\n",
		to, stageNames[to])
	met, unmet, unread, err := d.Preview(ctx.Root, to, ctx.Stdout)
	if err != nil {
		fmt.Fprintf(ctx.Stdout, "        the requirements could not be examined: %v\n\n", err)
		return
	}
	fmt.Fprintf(ctx.Stdout, "\n        %d of %d met, %d unmet, %d unread. Declaring it writes\n", met, met+unmet+unread, unmet, unread)
	fmt.Fprintf(ctx.Stdout, "        `stage: %d` and installs nothing; the flows that read the\n", to)
	fmt.Fprintf(ctx.Stdout, "        forge will stop at the unmet one until it is answered.\n\n")
}
