package git

import "os"

// Mv renames a path, preferring `git mv` so the rename is recorded as a
// rename in history (rather than delete + add). Falls back to os.Rename
// when g.Dir is not inside a git work tree, so the tickets CLI keeps
// working in plain directories.
//
// Signature is identical to store.RenameFunc — the CLI plugs this in via
//
//	s.Rename = git.New(repoRoot).Mv
func (g *Git) Mv(oldPath, newPath string) error {
	if !g.IsRepo() {
		return os.Rename(oldPath, newPath)
	}

	_, err := g.run("mv", oldPath, newPath)
	return err
}

// Add stages path with `git add`, so a freshly written file is tracked.
// No-ops when g.Dir is not inside a git work tree, mirroring Mv's fallback
// so the tickets CLI keeps working in plain directories.
//
// Signature is identical to store.TrackFunc — the CLI plugs this in via
//
//	s.Track = git.New(repoRoot).Add
func (g *Git) Add(path string) error {
	if !g.IsRepo() {
		return nil
	}

	_, err := g.run("add", path)
	return err
}
