package security

import (
	"strings"
	"testing"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Text with HTML tags",
			input:    "<script>alert('XSS')</script>",
			expected: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:     "Text with special characters",
			input:    `<img src="x" onerror="alert('XSS')">`,
			expected: "&lt;img src=&#34;x&#34; onerror=&#34;alert(&#39;XSS&#39;)&#34;&gt;",
		},
		{
			name:     "Text with quotes",
			input:    `He said "Hello"`,
			expected: "He said &#34;Hello&#34;",
		},
		{
			name:     "Text with ampersand",
			input:    "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		notContains []string
	}{
		{
			name:        "Remove script tags",
			input:       "Hello <script>alert('XSS')</script> World",
			notContains: []string{"script", "alert"},
		},
		{
			name:        "Remove javascript protocol",
			input:       `<a href="javascript:alert('XSS')">Click</a>`,
			notContains: []string{"javascript:"},
		},
		{
			name:        "Remove onclick handler",
			input:       `<div onclick="alert('XSS')">Click me</div>`,
			notContains: []string{"onclick"},
		},
		{
			name:     "Normal text unchanged",
			input:    "This is a normal post content",
			contains: []string{"This is a normal post content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeText(tt.input)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("SanitizeText(%q) should contain %q, got %q", tt.input, substr, result)
				}
			}

			for _, substr := range tt.notContains {
				if strings.Contains(result, substr) {
					t.Errorf("SanitizeText(%q) should not contain %q, got %q", tt.input, substr, result)
				}
			}
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal username",
			input:    "john_doe",
			expected: "john_doe",
		},
		{
			name:     "Username with dash",
			input:    "john-doe-123",
			expected: "john-doe-123",
		},
		{
			name:     "Username with special chars",
			input:    "john@doe.com",
			expected: "johndoecom",
		},
		{
			name:     "Username with spaces",
			input:    "john doe",
			expected: "johndoe",
		},
		{
			name:     "Username with HTML",
			input:    "<script>alert('XSS')</script>",
			expected: "scriptalertXSSscript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeUsername(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeUsername(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateLength(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  bool
	}{
		{
			name:      "Within limit",
			input:     "Hello",
			maxLength: 10,
			expected:  true,
		},
		{
			name:      "At limit",
			input:     "Hello",
			maxLength: 5,
			expected:  true,
		},
		{
			name:      "Exceeds limit",
			input:     "Hello World",
			maxLength: 5,
			expected:  false,
		},
		{
			name:      "Empty string",
			input:     "",
			maxLength: 10,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLength(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("ValidateLength(%q, %d) = %v, want %v", tt.input, tt.maxLength, result, tt.expected)
			}
		})
	}
}
