package competition

import (
	"testing"
)

func TestExtractFlags(t *testing.T) {
	monitor := &FlowMonitor{}
	
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple flag uppercase",
			input:    "Found the flag: FLAG{abc123}",
			expected: []string{"FLAG{abc123}"},
		},
		{
			name:     "simple flag lowercase",
			input:    "The flag is flag{xyz789}",
			expected: []string{"FLAG{xyz789}"},
		},
		{
			name:     "mixed case flag",
			input:    "Flag{test} should be normalized",
			expected: []string{"FLAG{test}"},
		},
		{
			name:     "multiple flags",
			input:    "FLAG{abc} and flag{xyz} were found",
			expected: []string{"FLAG{abc}", "FLAG{xyz}"},
		},
		{
			name:     "duplicate flags",
			input:    "FLAG{abc} and FLAG{abc} again",
			expected: []string{"FLAG{abc}"},
		},
		{
			name:     "no flag",
			input:    "No flag found in this text",
			expected: []string{},
		},
		{
			name:     "uuid flag",
			input:    "Successfully got FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}",
			expected: []string{"FLAG{e82fba66-3230-4818-8960-800ab47b0ba3}"},
		},
		{
			name:     "flag with special chars",
			input:    "FLAG{test-123_abc@xyz}",
			expected: []string{"FLAG{test-123_abc@xyz}"},
		},
		{
			name:     "flag in multiline text",
			input:    "Line 1\nLine 2 with FLAG{multiline}\nLine 3",
			expected: []string{"FLAG{multiline}"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.extractFlags(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d flags, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected flag %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestFlagPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"FLAG{abc}", true},
		{"flag{abc}", true},
		{"Flag{abc}", true},
		{"fLaG{abc}", true},
		{"FLAG{}", true},
		{"FLAG{a}", true},
		{"FLAG{abc-123_xyz}", true},
		{"FLAG{abc def}", true},
		{"FLAG", false},
		{"{abc}", false},
		{"FLAG}", false},
		{"FLAGabc}", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matched := FlagPattern.MatchString(tt.input)
			if matched != tt.expected {
				t.Errorf("pattern match for %q: expected %v, got %v", tt.input, tt.expected, matched)
			}
		})
	}
}

