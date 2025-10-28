package security

import (
	"html"
	"regexp"
	"strings"
)

// SanitizeInput sanitizes user input to prevent XSS attacks
// It escapes HTML special characters and removes potentially dangerous content
func SanitizeInput(input string) string {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Escape HTML special characters to prevent XSS
	// This converts: < > & " ' to &lt; &gt; &amp; &quot; &#39;
	input = html.EscapeString(input)

	return input
}

// SanitizeText sanitizes text content (posts, comments, etc.)
// Allows basic formatting but prevents XSS
func SanitizeText(text string) string {
	// First escape all HTML
	text = html.EscapeString(strings.TrimSpace(text))

	// Remove any script tags or event handlers that might have been encoded
	scriptPattern := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	text = scriptPattern.ReplaceAllString(text, "")

	// Remove javascript: protocol in links
	jsPattern := regexp.MustCompile(`(?i)javascript:`)
	text = jsPattern.ReplaceAllString(text, "")

	// Remove on* event handlers (onclick, onerror, etc.)
	eventPattern := regexp.MustCompile(`(?i)\s*on\w+\s*=`)
	text = eventPattern.ReplaceAllString(text, "")

	return text
}

// SanitizeUsername sanitizes username input
// Only allows alphanumeric, underscore, dash
func SanitizeUsername(username string) string {
	username = strings.TrimSpace(username)

	// Remove any characters that aren't alphanumeric, underscore, or dash
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	username = re.ReplaceAllString(username, "")

	return username
}

// ValidateLength validates input length to prevent DoS attacks
func ValidateLength(input string, maxLength int) bool {
	return len(input) <= maxLength
}

// Constants for validation
const (
	MaxUsernameLength    = 100
	MaxNameLength        = 255
	MaxPostHeaderLength  = 500
	MaxPostContentLength = 5000
	MaxPasswordLength    = 128
)
