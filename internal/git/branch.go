package git

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
