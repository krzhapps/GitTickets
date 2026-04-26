package ticket

import (
	"fmt"
	"regexp"
)

// slugRe enforces kebab-case: lowercase letters and digits, single hyphens
// between segments, no leading/trailing/double hyphens.
var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateSlug returns nil if s is a valid kebab-case slug.
func ValidateSlug(s string) error {
	if s == "" {
		return fmt.Errorf("slug cannot be empty")
	}

	if !slugRe.MatchString(s) {
		return fmt.Errorf("invalid slug %q: must be kebab-case (lowercase letters, digits, single hyphens between segments)", s)
	}

	return nil
}
