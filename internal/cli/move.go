package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GitTickets/internal/git"
	"github.com/krzhapps/GitTickets/internal/ticket"
)

// branchPrefix is prepended to a ticket slug to form its git branch name.
const branchPrefix = "ticket/"

// worktreeSuffix names the sibling directory that holds per-ticket
// worktrees: for a repo at /path/to/repo, worktrees live under
// /path/to/repo-worktrees/<slug>. Keeping them outside the repo avoids
// nested-worktree confusion and keeps the main working tree clean.
const worktreeSuffix = "-worktrees"

// worktreePath returns the sibling-directory path that holds the worktree
// for slug, derived from the git repository root.
func worktreePath(repoRoot, slug string) string {
	return filepath.Join(filepath.Dir(repoRoot), filepath.Base(repoRoot)+worktreeSuffix, slug)
}

// runMove is the shared implementation behind `move`, `start`, `done`,
// `archive`. Prints the resulting bucket path on success.
func runMove(cmd *cobra.Command, g *Globals, slug string, target ticket.Status) error {
	s, err := storeFromGlobals(g)
	if err != nil {
		return err
	}

	moved, err := s.Move(slug, target)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s (%s)\n", moved.Slug, moved.Status, moved.Dir)
	return nil
}

func newMoveCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "move <slug> <status>",
		Short: "Move a ticket to a new status (pending|in-progress|blocked|done|archived)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ticket.Status(args[1])
			if !target.Valid() {
				return fmt.Errorf("invalid status %q", args[1])
			}

			return runMove(cmd, g, args[0], target)
		},
	}
}

func newStartCmd(g *Globals) *cobra.Command {
	var worktree bool

	cmd := &cobra.Command{
		Use:   "start <slug>",
		Short: "Move a ticket to in-progress and check out a ticket/<slug> branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			if err := runMove(cmd, g, slug, ticket.StatusInProgress); err != nil {
				return err
			}

			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}

			root := repoRoot(s.Root)
			gw := git.New(root)
			if !gw.IsRepo() {
				return nil
			}

			branch := branchPrefix + slug

			if worktree {
				return startWorktree(cmd, gw, root, slug, branch)
			}

			created, err := gw.EnsureBranch(branch)
			if err != nil {
				return err
			}

			verb := "switched to existing branch"
			if created {
				verb = "created and switched to branch"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", verb, branch)
			return nil
		},
	}

	cmd.Flags().BoolVar(&worktree, "worktree", false,
		"check the branch out in a dedicated sibling worktree directory instead of in place")

	return cmd
}

// startWorktree creates (or reuses) a sibling worktree for slug and prints
// its path. A branch can only be checked out in one worktree, so the branch
// is attached to the worktree rather than checked out in the main tree.
func startWorktree(cmd *cobra.Command, gw *git.Git, repoRoot, slug, branch string) error {
	path := worktreePath(repoRoot, slug)

	if gw.WorktreeExists(path) {
		fmt.Fprintf(cmd.OutOrStdout(), "worktree already present at %s\n", path)
		return nil
	}

	createBranch := !gw.BranchExists(branch)
	if err := gw.WorktreeAdd(path, branch, createBranch); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "created worktree %s on branch %s\n", path, branch)
	return nil
}

// cleanupWorktree removes the worktree for slug if one is present. It is
// best-effort: a dirty worktree makes `git worktree remove` fail, which we
// surface as a warning rather than failing the lifecycle move.
func cleanupWorktree(cmd *cobra.Command, g *Globals, slug string) {
	s, err := storeFromGlobals(g)
	if err != nil {
		return
	}

	root := repoRoot(s.Root)
	gw := git.New(root)
	if !gw.IsRepo() {
		return
	}

	path := worktreePath(root, slug)
	if !gw.WorktreeExists(path) {
		return
	}

	if err := gw.WorktreeRemove(path, false); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"warning: worktree not removed (uncommitted changes?): %s\n", path)
		return
	}

	fmt.Fprintf(cmd.OutOrStdout(), "removed worktree %s\n", path)
}

func newDoneCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "done <slug>",
		Short: "Mark a ticket done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runMove(cmd, g, args[0], ticket.StatusDone); err != nil {
				return err
			}
			cleanupWorktree(cmd, g, args[0])
			return nil
		},
	}
}

func newArchiveCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <slug>",
		Short: "Archive a ticket (closed without implementation)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runMove(cmd, g, args[0], ticket.StatusArchived); err != nil {
				return err
			}
			cleanupWorktree(cmd, g, args[0])
			return nil
		},
	}
}
