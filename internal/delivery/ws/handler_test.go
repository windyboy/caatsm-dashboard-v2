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
			name:           "case insensitive scheme and host per RFC 6454",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "HTTP://LOCALHOST:3000",
			expected:       true,
		},
		{
			name:           "mixed case domain matches",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://Example.Com",
			expected:       true,
		},
		{
			name:           "uppercase scheme matches",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "HTTPS://example.com",
			expected:       true,
		},
		{
			name:           "default port 80 removed for http",
			allowedOrigins: []string{"http://localhost"},
			requestOrigin:  "http://localhost:80",
			expected:       true,
		},
		{
			name:           "default port 443 removed for https",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com:443",
			expected:       true,
		},
		{
			name:           "non-default port must match exactly",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:8080",
			expected:       false,
		},
		{
			name:           "non-default port preserved",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expected:       true,
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

func TestNormalizeOrigin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "lowercase scheme and host",
			input:    "HTTP://LOCALHOST:3000",
			expected: "http://localhost:3000",
		},
		{
			name:     "mixed case domain",
			input:    "https://Example.Com",
			expected: "https://example.com",
		},
		{
			name:     "already normalized",
			input:    "http://localhost:3000",
			expected: "http://localhost:3000",
		},
		{
			name:     "remove default http port 80",
			input:    "http://localhost:80",
			expected: "http://localhost",
		},
		{
			name:     "remove default https port 443",
			input:    "https://example.com:443",
			expected: "https://example.com",
		},
		{
			name:     "preserve non-default http port",
			input:    "http://localhost:3000",
			expected: "http://localhost:3000",
		},
		{
			name:     "preserve non-default https port",
			input:    "https://example.com:8443",
			expected: "https://example.com:8443",
		},
		{
			name:     "complex case with uppercase scheme and domain",
			input:    "HTTPS://API.EXAMPLE.COM:443",
			expected: "https://api.example.com",
		},
		{
			name:     "invalid URL returns original",
			input:    "not a valid url",
			expected: "not a valid url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeOrigin(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeOrigin(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

