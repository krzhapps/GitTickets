// Package git is a thin wrapper around the system `git` binary.
//
// It is deliberately small — only the operations the tickets CLI needs
// (mv, branch, checkout, rev-parse) — and is structured around a Runner
// interface so unit tests can swap in a fake without exec'ing anything.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes a git subcommand and returns its stdout, or an error
// that includes stderr context. Tests inject a fake to assert on the
// argument list without touching the filesystem.
type Runner interface {
	Run(args ...string) ([]byte, error)
}

// Git wraps a git binary invocation rooted at a working directory.
type Git struct {
	// Dir is the working directory git commands run in (the repository
	// root, typically). Empty means inherit the parent's CWD.
	Dir string

	// Runner overrides the default exec-based runner. Leave nil in
	// production; tests set this to a fake.
	Runner Runner
}

// New returns a Git wrapper rooted at dir.
func New(dir string) *Git { return &Git{Dir: dir} }

// run dispatches to the injected Runner if one is set, otherwise shells
// out to the real `git` binary with cmd.Dir = g.Dir.
func (g *Git) run(args ...string) ([]byte, error) {
	if g.Runner != nil {
		return g.Runner.Run(args...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("git %s: %w (%s)",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return out, nil
}

// IsRepo reports whether g.Dir is inside a git work tree. Non-zero exit
// (including "not a git repo") becomes false rather than an error so
// callers can branch cleanly on the boolean.
func (g *Git) IsRepo() bool {
	out, err := g.run("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) == "true"
}
