package main

import "testing"

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		message  string
		expected bool
	}{
		{
			message:  "More than twenty characters and first one is uppercase",
			expected: true,
		},
		{
			message:  "more than twenty characters and first one is uppercase",
			expected: false,
		},
		{
			message:  "false",
			expected: false,
		},
		{
			message:  "",
			expected: false,
		},
		{
			message:  "Dsadsa",
			expected: false,
		},
	}

	for i, test := range tests {
		if got := ValidateMessage(test.message); got != test.expected {
			t.Fatalf("test[%d]: for message=\"%s\" expected=%t, got=%t", i, test.message, test.expected, got)
		}
	}
}
