package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GitTickets/internal/ticket"
)

func newSearchCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Substring search across title, description, notes, and labels",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}

			ts, err := s.Load()
			if err != nil {
				return err
			}

			q := strings.ToLower(args[0])
			var hits []ticket.Ticket
			for _, t := range ts {
				if matchesQuery(t, q) {
					hits = append(hits, t)
				}
			}

			sort.Slice(hits, func(i, j int) bool { return hits[i].Slug < hits[j].Slug })
			out := cmd.OutOrStdout()
			for _, t := range hits {
				fmt.Fprintf(out, "%s — %s\n", t.Slug, t.Title)
			}

			return nil
		},
	}
}

func matchesQuery(t ticket.Ticket, q string) bool {
	if strings.Contains(strings.ToLower(t.Title), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Description), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Notes), q) {
		return true
	}
	for _, l := range t.Labels {
		if strings.Contains(strings.ToLower(l), q) {
			return true
		}
	}
	return false
}
