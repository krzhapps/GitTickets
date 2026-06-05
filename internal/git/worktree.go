package git

import (
	"bufio"
	"bytes"
	"strings"
)

// Worktree describes one entry from `git worktree list`.
type Worktree struct {
	// Path is the absolute path of the worktree's working directory.
	Path string
	// Branch is the short branch name checked out there (e.g. "ticket/foo"),
	// or "" for a detached HEAD.
	Branch string
	// Head is the commit hash currently checked out.
	Head string
}

// WorktreeAdd creates a new worktree at path. When createBranch is true the
// branch is created from HEAD as part of the add (`git worktree add -b
// <branch> <path>`); otherwise an existing branch is attached (`git worktree
// add <path> <branch>`). A branch can only be checked out in one worktree at
// a time, so the caller must not also check it out in the main tree.
func (g *Git) WorktreeAdd(path, branch string, createBranch bool) error {
	if createBranch {
		_, err := g.run("worktree", "add", "-b", branch, path)
		return err
	}

	_, err := g.run("worktree", "add", path, branch)
	return err
}

// WorktreeRemove removes the worktree at path. Without force, git refuses if
// the worktree has uncommitted changes; callers surface that as a warning
// rather than a hard failure so the lifecycle move still succeeds.
func (g *Git) WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	_, err := g.run(args...)
	return err
}

// WorktreeList parses `git worktree list --porcelain` into a slice of
// Worktree. The first entry is always the main working tree.
func (g *Git) WorktreeList() ([]Worktree, error) {
	out, err := g.run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeList(out), nil
}

// WorktreeExists reports whether a worktree is registered at path.
func (g *Git) WorktreeExists(path string) bool {
	list, err := g.WorktreeList()
	if err != nil {
		return false
	}
	for _, w := range list {
		if w.Path == path {
			return true
		}
	}
	return false
}

// parseWorktreeList turns porcelain output into Worktree records. Records are
// separated by blank lines; each begins with a "worktree <path>" line, then
// optionally "HEAD <sha>" and "branch refs/heads/<name>" (absent for a
// detached or bare entry).
func parseWorktreeList(out []byte) []Worktree {
	var list []Worktree
	var cur Worktree
	have := false

	flush := func() {
		if have {
			list = append(list, cur)
		}
		cur = Worktree{}
		have = false
	}

	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			cur.Path = strings.TrimPrefix(line, "worktree ")
			have = true
		case strings.HasPrefix(line, "HEAD "):
			cur.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		}
	}
	flush()
	return list
}
