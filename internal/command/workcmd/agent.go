package workcmd

import (
	"fmt"
	"strings"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
)

// configKey is where the adopter's agent command lives. git config
// layers repository over user, so a contributor's own agent needs no
// file and nothing is committed by accident (decision 0003).
const configKey = "writrun.agent"

// unset is what `git config --get` exits with when the key is not
// there — an answer, not a failure.
const unset = 1

// configured is step 1: the agent command, or the abort that shows how
// to set one. `writrun` never guesses which agent is installed
// (docs/product/rules.md), so an unset key launches nothing and names
// the line that would fix it.
func configured(git gitx.Runner, root string) (string, error) {
	out, err := git(root, "config", "--get", configKey)
	if err != nil && exitCode(err) != unset {
		return "", fmt.Errorf("reading %s: %w", configKey, err)
	}
	agent := strings.TrimSpace(out)
	if agent == "" {
		return "", fmt.Errorf(
			"no agent is configured, and writrun never guesses which agent is installed — set the command to launch:\n\n    git config %s '<your agent command>'",
			configKey)
	}
	return agent, nil
}

// words splits the configured command into the program and its
// arguments. Quoting is honoured because a real agent command carries
// flags with values — `claude --append-system-prompt "be terse"` is one
// program and two arguments, and splitting on spaces alone would make
// it four. Nothing else about the string is interpreted: no shell runs
// it, so a pipeline or a variable in the value is a mistake this
// reports rather than a feature it grants.
func words(cmd string) (string, []string, error) {
	var parts []string
	var cur strings.Builder
	open := false // a word is open, even where it is empty: "" is a word
	quote := rune(0)

	for _, r := range cmd {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
			open = true
		case r == ' ' || r == '\t' || r == '\n':
			if open {
				parts = append(parts, cur.String())
				cur.Reset()
				open = false
			}
		default:
			cur.WriteRune(r)
			open = true
		}
	}
	if quote != 0 {
		return "", nil, fmt.Errorf("%s = %q has an unclosed %c quote", configKey, cmd, quote)
	}
	if open {
		parts = append(parts, cur.String())
	}
	if len(parts) == 0 || parts[0] == "" {
		return "", nil, fmt.Errorf("%s = %q names no command to launch", configKey, cmd)
	}
	return parts[0], parts[1:], nil
}
