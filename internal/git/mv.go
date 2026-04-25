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
