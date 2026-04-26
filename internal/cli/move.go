package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GitTickets/internal/ticket"
)

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
	return &cobra.Command{
		Use:   "start <slug>",
		Short: "Move a ticket to in-progress",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, g, args[0], ticket.StatusInProgress)
		},
	}
}

func newDoneCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "done <slug>",
		Short: "Mark a ticket done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, g, args[0], ticket.StatusDone)
		},
	}
}

func newArchiveCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <slug>",
		Short: "Archive a ticket (closed without implementation)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, g, args[0], ticket.StatusArchived)
		},
	}
}
