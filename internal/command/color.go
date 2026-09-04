package command

// colorEnabled decides the reporting rule of product/rules.md: color
// appears only where stdout is a terminal, and NO_COLOR set or
// --no-color given disables it.
func colorEnabled(ttyOut bool, noColorFlag bool, getenv func(string) string) bool {
	if noColorFlag || !ttyOut {
		return false
	}
	return getenv("NO_COLOR") == ""
}
