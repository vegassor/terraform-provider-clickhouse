package clickhouse_client

import (
	"testing"
)

func TestQuoteID(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No special characters",
			input:    "test",
			expected: `"test"`,
		},
		{
			name:     "With quote",
			input:    `test"quote`,
			expected: `"test\"quote"`,
		},
		{
			name:     "With backslash and quote",
			input:    `test\"quote`,
			expected: `"test\\\"quote"`,
		},
		{
			name:     "withNullTerminator",
			input:    "test\x00extra",
			expected: `"test"`,
		},
		{
			name:     "quoteAndNullTerminator",
			input:    "tes\"t\x00extra",
			expected: `"tes\"t"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := QuoteID(tc.input)

			if output != tc.expected {
				t.Errorf("Expected %q, but got %q", tc.expected, output)
			}
		})
	}
}

func TestQuoteValue(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No special characters",
			input:    "sampletext",
			expected: "'sampletext'",
		},
		{
			name:     "Backslash character",
			input:    `sample\text`,
			expected: `'sample\\text'`,
		},
		{
			name:     "Single quote character",
			input:    "sample'text",
			expected: `'sample\'text'`,
		},
		{
			name:     "Mixed special characters",
			input:    `sample\'te\xt`,
			expected: `'sample\\\'te\\xt'`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := QuoteValue(tc.input)
			if result != tc.expected {
				t.Errorf("QuoteValue(%s) = %s; expected %s", tc.input, result, tc.expected)
			}
		})
	}
}
