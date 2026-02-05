package valueobject

import "testing"

func TestCleanJID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JID with device ID",
			input:    "201021347532:98@s.whatsapp.net",
			expected: "201021347532@s.whatsapp.net",
		},
		{
			name:     "JID without device ID",
			input:    "201021347532@s.whatsapp.net",
			expected: "201021347532@s.whatsapp.net",
		},
		{
			name:     "Group JID with device ID",
			input:    "120363123456789012:1@g.us",
			expected: "120363123456789012@g.us",
		},
		{
			name:     "Empty JID",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid format (no @)",
			input:    "201021347532",
			expected: "201021347532",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanJID(tt.input)
			if result != tt.expected {
				t.Errorf("CleanJID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JID with device ID",
			input:    "201021347532:98@s.whatsapp.net",
			expected: "201021347532",
		},
		{
			name:     "JID without device ID",
			input:    "201021347532@s.whatsapp.net",
			expected: "201021347532",
		},
		{
			name:     "Empty JID",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPhone(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractPhone(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsGroupJID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Group JID",
			input:    "120363123456789012@g.us",
			expected: true,
		},
		{
			name:     "User JID",
			input:    "201021347532@s.whatsapp.net",
			expected: false,
		},
		{
			name:     "Empty JID",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGroupJID(tt.input)
			if result != tt.expected {
				t.Errorf("IsGroupJID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidJID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid user JID",
			input:    "201021347532@s.whatsapp.net",
			expected: true,
		},
		{
			name:     "Valid group JID",
			input:    "120363123456789012@g.us",
			expected: true,
		},
		{
			name:     "Valid broadcast JID",
			input:    "status@broadcast",
			expected: true,
		},
		{
			name:     "Invalid JID (no @)",
			input:    "201021347532",
			expected: false,
		},
		{
			name:     "Invalid JID (wrong domain)",
			input:    "201021347532@invalid.com",
			expected: false,
		},
		{
			name:     "Empty JID",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidJID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidJID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
