package ws

import (
	"net/http"
	"testing"
)

func TestPrepareAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice returns defaults",
			input:    []string{},
			expected: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:     "nil slice returns defaults",
			input:    nil,
			expected: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:     "trims whitespace",
			input:    []string{" http://example.com ", "  https://test.com  "},
			expected: []string{"http://example.com", "https://test.com"},
		},
		{
			name:     "filters empty entries",
			input:    []string{"http://example.com", "", "  ", "https://test.com"},
			expected: []string{"http://example.com", "https://test.com"},
		},
		{
			name:     "all empty entries returns defaults",
			input:    []string{"", "  ", "   "},
			expected: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:     "valid origins unchanged",
			input:    []string{"http://localhost:3000", "https://example.com"},
			expected: []string{"http://localhost:3000", "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareAllowedOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d origins, got %d", len(tt.expected), len(result))
				return
			}
			for i, origin := range result {
				if origin != tt.expected[i] {
					t.Errorf("origin[%d] = %s, want %s", i, origin, tt.expected[i])
				}
			}
		})
	}
}

func TestMakeCheckOriginFunc(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expected       bool
	}{
		{
			name:           "empty origin header returns true",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "",
			expected:       true,
		},
		{
			name:           "allowed origin returns true",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "http://localhost:3000",
			expected:       true,
		},
		{
			name:           "disallowed origin returns false",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://evil.com",
			expected:       false,
		},
		{
			name:           "multiple allowed origins",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com", "http://localhost:5173"},
			requestOrigin:  "https://example.com",
			expected:       true,
		},
		{
			name:           "case sensitive check",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "HTTP://LOCALHOST:3000",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkOrigin := makeCheckOriginFunc(tt.allowedOrigins)
			req := &http.Request{
				Header: http.Header{},
			}
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			
			result := checkOrigin(req)
			if result != tt.expected {
				t.Errorf("checkOrigin(%s) = %v, want %v", tt.requestOrigin, result, tt.expected)
			}
		})
	}
}

