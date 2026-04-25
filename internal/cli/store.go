package cli

import (
	"os"
	"path/filepath"

	"github.com/krzhapps/GithubTickets/internal/git"
	"github.com/krzhapps/GithubTickets/internal/store"
)

// storeFromGlobals resolves a *store.Store from --root, then $TICKETS_ROOT,
// then by walking up from CWD. The Store is wired with git.Mv so directory
// moves preserve rename history when the tree is inside a git repo (Mv
// itself falls back to os.Rename outside one).
func storeFromGlobals(g *Globals) (*store.Store, error) {
	root := g.Root
	if root == "" {
		root = os.Getenv("TICKETS_ROOT")
	}

	var s *store.Store
	var err error
	if root != "" {
		s, err = store.Open(root)
	} else {
		wd, werr := os.Getwd()
		if werr != nil {
			return nil, werr
		}
		s, err = store.Discover(wd)
	}
	if err != nil {
		return nil, err
	}

	s.Rename = git.New(repoRoot(s.Root)).Mv
	return s, nil
}

// repoRoot returns the assumed git repository root from a tickets/ root
// path: tickets/ is conventionally <repo>/tickets, so repo root is its
// parent.
func repoRoot(ticketsRoot string) string {
	return filepath.Dir(ticketsRoot)
}
