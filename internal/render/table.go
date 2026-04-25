// Package render formats tickets for human (table) and machine (JSON)
// output. Kept separate from the CLI package so commands are thin.
package render

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

// TicketsTable writes a column-aligned table of tickets to w. Columns:
// SLUG, STATUS, PRI, TITLE, LABELS.
func TicketsTable(w io.Writer, ts []ticket.Ticket) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SLUG\tSTATUS\tPRI\tTITLE\tLABELS")
	for _, t := range ts {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			t.Slug, t.Status, t.Priority, t.Title, strings.Join(t.Labels, ","))
	}
	_ = tw.Flush()
}
