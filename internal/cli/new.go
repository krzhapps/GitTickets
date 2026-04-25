package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GithubTickets/internal/editor"
	"github.com/krzhapps/GithubTickets/internal/ticket"
)

func newNewCmd(g *Globals) *cobra.Command {
	var (
		title    string
		priority string
		labels   []string
		assignee string
		created  string
		noEdit   bool
	)
	cmd := &cobra.Command{
		Use:   "new <slug>",
		Short: "Create a new ticket under to-do/<slug>/DESCRIPTION.md",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			if err := ticket.ValidateSlug(slug); err != nil {
				return err
			}

			pri := ticket.Priority(priority)
			if !pri.Valid() {
				return fmt.Errorf("invalid --priority %q (low|medium|high)", priority)
			}
			if title == "" {
				title = slug
			}
			if created == "" {
				created = time.Now().UTC().Format("2006-01-02")
			}

			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}
			if err := s.Init(); err != nil {
				return err
			}

			t := &ticket.Ticket{
				Title:    title,
				Status:   ticket.StatusPending,
				Priority: pri,
				Created:  created,
				Labels:   labels,
				Assignee: assignee,
				Slug:     slug,
			}
			if err := s.Create(t); err != nil {
				return err
			}

			descPath := filepath.Join(t.Dir, "DESCRIPTION.md")
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", descPath)

			if noEdit || !stdinIsTTY() {
				return nil
			}
			return editor.Open(descPath)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "ticket title (defaults to slug)")
	cmd.Flags().StringVar(&priority, "priority", "medium", "low|medium|high")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "label (repeatable)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "assignee")
	cmd.Flags().StringVar(&created, "created", "", "YYYY-MM-DD (defaults to today UTC)")
	cmd.Flags().BoolVar(&noEdit, "no-edit", false, "do not open $EDITOR after scaffolding")
	return cmd
}

func stdinIsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
