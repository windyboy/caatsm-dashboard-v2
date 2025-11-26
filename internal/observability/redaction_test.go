package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactTelegramContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		policy   RedactionPolicy
		expected string
	}{
		{
			name:    "truncate long content",
			content: "This is a very long message that should be truncated because it exceeds the maximum allowed length for logging purposes",
			policy: RedactionPolicy{
				MaxContentLength: 50,
			},
			expected: "This is a very long message that should be truncat...[REDACTED]",
		},
		{
			name:    "redact email addresses",
			content: "Contact pilot at john.doe@example.com for details",
			policy: RedactionPolicy{
				MaxContentLength: 0,
				RedactEmails:     true,
			},
			expected: "Contact pilot at [EMAIL_REDACTED] for details",
		},
		{
			name:    "redact IPv4 addresses",
			content: "Server located at 192.168.1.100 for backup",
			policy: RedactionPolicy{
				MaxContentLength: 0,
				RedactIPs:        true,
			},
			expected: "Server located at [IP_REDACTED] for backup",
		},
		{
			name:    "redact phone numbers",
			content: "Call +1-555-123-4567 for emergency contact",
			policy: RedactionPolicy{
				MaxContentLength: 0,
				RedactPhones:     true,
			},
			expected: "Call [PHONE_REDACTED] for emergency contact",
		},
		{
			name:    "redact multiple patterns",
			content: "Contact john@example.com at 192.168.1.1 or call 555-123-4567",
			policy: RedactionPolicy{
				MaxContentLength: 0,
				RedactEmails:     true,
				RedactIPs:        true,
				RedactPhones:     true,
			},
			expected: "Contact [EMAIL_REDACTED] at [IP_REDACTED] or call [PHONE_REDACTED]",
		},
		{
			name:    "truncate and redact",
			content: "Long message with email test@example.com and more text that exceeds limit",
			policy: RedactionPolicy{
				MaxContentLength: 40,
				RedactEmails:     true,
			},
			expected: "Long message with email [EMAIL_REDACTED]...[REDACTED]",
		},
		{
			name:     "empty content",
			content:  "",
			policy:   DefaultRedactionPolicy(),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactTelegramContent(tt.content, tt.policy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactMessageID(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		expected  string
	}{
		{
			name:      "standard message ID",
			messageID: "MSG-12345678-ABCD",
			expected:  "MSG-***ABCD",
		},
		{
			name:      "short message ID",
			messageID: "MSG123",
			expected:  "MSG123", // Too short to redact
		},
		{
			name:      "long message ID",
			messageID: "TELEGRAM-2024-01-15-123456789",
			expected:  "TELE***6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactMessageID(tt.messageID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactFlightNumber(t *testing.T) {
	tests := []struct {
		name         string
		flightNumber string
		expected     string
	}{
		{
			name:         "standard flight number",
			flightNumber: "CA1234",
			expected:     "CA****",
		},
		{
			name:         "long flight number",
			flightNumber: "BA9876543",
			expected:     "BA*******",
		},
		{
			name:         "short code",
			flightNumber: "AB",
			expected:     "AB",
		},
		{
			name:         "three char code",
			flightNumber: "ABC",
			expected:     "ABC", // Too short to redact
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactFlightNumber(tt.flightNumber)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	messageID := "MSG-12345678-TEST"
	flightNumber := "CA1234"
	content := "Flight delayed. Contact pilot@example.com at 192.168.1.1"
	policy := DefaultRedactionPolicy()

	result := SanitizeForLogging(messageID, flightNumber, content, policy)

	assert.Contains(t, result["message_id_redacted"], "MSG-")
	assert.Contains(t, result["message_id_redacted"], "***")
	assert.Equal(t, "CA****", result["flight_number_redacted"])
	assert.Contains(t, result["content_snippet"], "[EMAIL_REDACTED]")
	assert.Contains(t, result["content_snippet"], "[IP_REDACTED]")
}

func TestRedactUserIdentifiableInfo(t *testing.T) {
	text := "User email is admin@company.com, IP: 10.0.0.1, phone: 555-123-4567"
	result := RedactUserIdentifiableInfo(text)

	assert.NotContains(t, result, "admin@company.com")
	assert.NotContains(t, result, "10.0.0.1")
	assert.NotContains(t, result, "555-123-4567")
	assert.Contains(t, result, "[EMAIL_REDACTED]")
	assert.Contains(t, result, "[IP_REDACTED]")
	assert.Contains(t, result, "[PHONE_REDACTED]")
}

func TestDefaultRedactionPolicy(t *testing.T) {
	policy := DefaultRedactionPolicy()

	assert.Equal(t, 100, policy.MaxContentLength)
	assert.True(t, policy.RedactEmails)
	assert.True(t, policy.RedactIPs)
	assert.True(t, policy.RedactPhones)
}

