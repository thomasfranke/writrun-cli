package workcmd

// Launcher is the boundary this command exists to cross: one
// configured program, started from the repository root, inheriting the
// terminal, and waited for. It is declared here because here is where
// it is consumed; internal/agentx carries the production
// implementation and FakeLauncher, beside this file, is what the tests
// inject — no test starts a real agent
// (docs/technical/engineering/boundaries.md).
//
// The streams are absent from the signature on purpose. An agent draws
// on the terminal and reads keys from it, so what it inherits is the
// process's own stdin, stdout and stderr rather than the writers the
// frame hands a command; a port that took writers would promise a
// redirection the launch cannot honour.
//
// The error is the launched command's own verdict: an *exec.ExitError
// carrying its exit code, which `work` passes up unedited.
type Launcher func(dir, name string, args ...string) error
