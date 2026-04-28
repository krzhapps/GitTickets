package git

import "strings"

// Branch creates a branch named name. If checkout is true, the branch
// is created and checked out in one step (`git checkout -b`); the new
// branch is created from HEAD either way.
func (g *Git) Branch(name string, checkout bool) error {
	if checkout {
		_, err := g.run("checkout", "-b", name)
		return err
	}

	_, err := g.run("branch", name)
	return err
}

// BranchExists reports whether a local branch with the given name exists.
func (g *Git) BranchExists(name string) bool {
	_, err := g.run("rev-parse", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil
}

// Checkout switches to an existing branch.
func (g *Git) Checkout(name string) error {
	_, err := g.run("checkout", name)
	return err
}

// EnsureBranch checks out a branch named name, creating it from HEAD if
// it does not already exist. Returns true if the branch was newly created.
func (g *Git) EnsureBranch(name string) (created bool, err error) {
	if g.BranchExists(name) {
		return false, g.Checkout(name)
	}
	if err := g.Branch(name, true); err != nil {
		// If a parallel process raced us, fall back to checkout.
		if strings.Contains(err.Error(), "already exists") {
			return false, g.Checkout(name)
		}
		return false, err
	}
	return true, nil
}
