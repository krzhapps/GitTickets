package ticket

import "testing"

func TestValidateSlug(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"auth-google-oauth-errors", false},
		{"a", false},
		{"a1", false},
		{"a-1-b", false},
		{"abc123", false},

		{"", true},
		{"-leading", true},
		{"trailing-", true},
		{"double--hyphen", true},
		{"UPPER", true},
		{"under_score", true},
		{"has space", true},
		{"slash/path", true},
	}
	for _, c := range cases {
		err := ValidateSlug(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateSlug(%q) error=%v, wantErr=%v", c.in, err, c.wantErr)
		}
	}
}
