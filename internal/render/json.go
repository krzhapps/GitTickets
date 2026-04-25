package render

import (
	"encoding/json"
	"io"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

// ticketJSON is the wire format. It mirrors ticket.Ticket but uses
// snake_case field names and omits filesystem metadata (Dir).
type ticketJSON struct {
	Slug               string              `json:"slug"`
	Title              string              `json:"title"`
	Status             ticket.Status       `json:"status"`
	Priority           ticket.Priority     `json:"priority"`
	Created            string              `json:"created"`
	Labels             []string            `json:"labels,omitempty"`
	Assignee           string              `json:"assignee,omitempty"`
	Description        string              `json:"description,omitempty"`
	AcceptanceCriteria []ticket.ACItem     `json:"acceptance_criteria,omitempty"`
	Notes              string              `json:"notes,omitempty"`
	Dependencies       []ticket.Dependency `json:"dependencies,omitempty"`
}

func toJSON(t ticket.Ticket) ticketJSON {
	return ticketJSON{
		Slug:               t.Slug,
		Title:              t.Title,
		Status:             t.Status,
		Priority:           t.Priority,
		Created:            t.Created,
		Labels:             t.Labels,
		Assignee:           t.Assignee,
		Description:        t.Description,
		AcceptanceCriteria: t.AcceptanceCriteria,
		Notes:              t.Notes,
		Dependencies:       t.Dependencies,
	}
}

// TicketsJSON writes ts as a pretty-printed JSON array.
func TicketsJSON(w io.Writer, ts []ticket.Ticket) error {
	out := make([]ticketJSON, len(ts))
	for i, t := range ts {
		out[i] = toJSON(t)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// TicketJSON writes a single ticket as a pretty-printed JSON object.
func TicketJSON(w io.Writer, t ticket.Ticket) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(toJSON(t))
}
