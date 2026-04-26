package cli

import (
	"slices"
	"sort"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GitTickets/internal/render"
	"github.com/krzhapps/GitTickets/internal/ticket"
)

func newListCmd(g *Globals) *cobra.Command {
	var (
		statuses   []string
		priorities []string
		labels     []string
		assignee   string
		asJSON     bool
		all        bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tickets (default: open work — pending, in-progress, blocked)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}

			ts, err := s.Load()
			if err != nil {
				return err
			}

			if !all && len(statuses) == 0 {
				statuses = []string{
					string(ticket.StatusPending),
					string(ticket.StatusInProgress),
					string(ticket.StatusBlocked),
				}
			}

			ts = filterTickets(ts, statuses, priorities, labels, assignee)
			sort.Slice(ts, func(i, j int) bool { return ts[i].Slug < ts[j].Slug })

			if asJSON {
				return render.TicketsJSON(cmd.OutOrStdout(), ts)
			}

			render.TicketsTable(cmd.OutOrStdout(), ts)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&statuses, "status", nil, "filter by status (repeatable)")
	cmd.Flags().StringSliceVar(&priorities, "priority", nil, "filter by priority (repeatable)")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "filter by label — ticket must have ALL given labels")
	cmd.Flags().StringVar(&assignee, "assignee", "", "filter by assignee")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	cmd.Flags().BoolVar(&all, "all", false, "include done & archived tickets")
	return cmd
}

func filterTickets(ts []ticket.Ticket, statuses, priorities, labels []string, assignee string) []ticket.Ticket {
	out := ts[:0:0]
	for _, t := range ts {
		if len(statuses) > 0 && !slices.Contains(statuses, string(t.Status)) {
			continue
		}
		if len(priorities) > 0 && !slices.Contains(priorities, string(t.Priority)) {
			continue
		}
		if assignee != "" && t.Assignee != assignee {
			continue
		}
		if !hasAllLabels(t.Labels, labels) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func hasAllLabels(have, need []string) bool {
	for _, n := range need {
		if !slices.Contains(have, n) {
			return false
		}
	}
	return true
}
