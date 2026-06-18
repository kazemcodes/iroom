package sanitize

import "github.com/microcosm-cc/bluemonday"

var sanitizer = bluemonday.UGCPolicy()

// Sanitize strips HTML tags from a string to prevent XSS attacks.
// Used for user-provided display names and other text fields.
func Sanitize(input string) string {
	return sanitizer.Sanitize(input)
}
