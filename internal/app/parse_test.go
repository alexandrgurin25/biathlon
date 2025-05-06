package app

import (
	"testing"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      bool
	}{
		{"10:00:00.000", "10:00:00.000", false},
		{"23:59:59.999", "23:59:59.999", false},
		{"invalid", "", true},
		{"24:00:00.000", "", true}, // invalid hour
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tm, err := parseTime(tt.input)
			if tt.err && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.err {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					formatted := tm.Format("15:04:05.000")
					if formatted != tt.expected {
						t.Errorf("Expected %s, got %s", tt.expected, formatted)
					}
				}
			}
		})
	}
}
