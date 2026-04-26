// Package ticket is the pure domain layer: a Ticket struct, the parser
// that turns DESCRIPTION.md bytes into one, and the renderer that turns
// one back into canonical Markdown. No filesystem or git access lives
// here so every operation is trivially testable.
package ticket

// Status is the lifecycle state of a ticket. The frontmatter `status`
// must always be one of these values.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in-progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
	StatusArchived   Status = "archived"
)

// Valid reports whether s is a known Status value.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusBlocked, StatusDone, StatusArchived:
		return true
	}

	return false
}

// Priority is the urgency rank of a ticket.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Valid reports whether p is a known Priority value.
func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}

	return false
}

// Ticket is the in-memory representation of a DESCRIPTION.md file.
//
// Frontmatter fields carry yaml tags and are written/read via gopkg.in/yaml.v3.
// Body fields and filesystem-derived metadata are tagged `yaml:"-"` so they
// stay out of the marshaled frontmatter while remaining ordinary Go fields.
type Ticket struct {
	// Frontmatter — order here is the order written to disk.
	Title    string   `yaml:"title"`
	Status   Status   `yaml:"status"`
	Priority Priority `yaml:"priority"`
	Created  string   `yaml:"created"` // YYYY-MM-DD; kept as string for fidelity
	Labels   []string `yaml:"labels,omitempty"`
	Assignee string   `yaml:"assignee,omitempty"`

	// Body sections — populated from the Markdown body, never marshaled.
	Description        string       `yaml:"-"`
	AcceptanceCriteria []ACItem     `yaml:"-"`
	Notes              string       `yaml:"-"`
	Dependencies       []Dependency `yaml:"-"`

	// Filesystem-derived; set by the store layer, never serialized.
	Slug string `yaml:"-"`
	Dir  string `yaml:"-"`
}

// ACItem is a single Acceptance Criteria checkbox.
type ACItem struct {
	Done bool
	Text string
}

// Dependency is a pointer to another ticket plus a free-form reason.
type Dependency struct {
	Slug   string
	Reason string
}
