package ticket

import "testing"

func TestStatus_Valid(t *testing.T) {
	t.Parallel()
	cases := map[Status]bool{
		StatusPending:    true,
		StatusInProgress: true,
		StatusBlocked:    true,
		StatusDone:       true,
		StatusArchived:   true,
		"":               false,
		"unknown":        false,
		"PENDING":        false,
	}
	for s, want := range cases {
		if got := s.Valid(); got != want {
			t.Errorf("Status(%q).Valid() = %v, want %v", s, got, want)
		}
	}
}

func TestPriority_Valid(t *testing.T) {
	t.Parallel()
	cases := map[Priority]bool{
		PriorityLow:    true,
		PriorityMedium: true,
		PriorityHigh:   true,
		"":             false,
		"urgent":       false,
	}
	for p, want := range cases {
		if got := p.Valid(); got != want {
			t.Errorf("Priority(%q).Valid() = %v, want %v", p, got, want)
		}
	}
}
