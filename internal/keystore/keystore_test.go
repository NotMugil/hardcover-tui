package keystore

import "testing"

func TestFormatToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{"mytoken123", "Bearer mytoken123"},
		{"Bearer mytoken123", "Bearer mytoken123"},
		{"bearer mytoken123", "Bearer mytoken123"},
		{"BEARER mytoken123", "Bearer mytoken123"},
		{"Bearer  Bearer mytoken123", "Bearer mytoken123"},
		{"Bearer", ""},
	}

	for _, tt := range tests {
		got := FormatToken(tt.input)
		if got != tt.expected {
			t.Errorf("FormatToken(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
