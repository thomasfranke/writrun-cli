package command

import "testing"

func TestColorEnabled(t *testing.T) {
	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "NO_COLOR" {
				return v
			}
			return ""
		}
	}
	cases := []struct {
		name    string
		ttyOut  bool
		noColor bool
		noEnv   string
		want    bool
	}{
		{"terminal, nothing disabling", true, false, "", true},
		{"not a terminal", false, false, "", false},
		{"--no-color given", true, true, "", false},
		{"NO_COLOR set", true, false, "1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := colorEnabled(tc.ttyOut, tc.noColor, env(tc.noEnv)); got != tc.want {
				t.Fatalf("colorEnabled = %v, want %v", got, tc.want)
			}
		})
	}
}
