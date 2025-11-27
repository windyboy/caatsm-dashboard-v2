package observability

import (
	"regexp"
	"strings"
)

// RedactionPolicy defines how to redact sensitive information from logs
type RedactionPolicy struct {
	MaxContentLength int
	RedactEmails     bool
	RedactIPs        bool
	RedactPhones     bool
}

// DefaultRedactionPolicy returns a sensible default redaction policy
func DefaultRedactionPolicy() RedactionPolicy {
	return RedactionPolicy{
		MaxContentLength: 100,
		RedactEmails:     true,
		RedactIPs:        true,
		RedactPhones:     true,
	}
}

var (
	// Email pattern for redaction
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	// IP address pattern (both IPv4 and IPv6)
	ipv4Regex = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	// Comprehensive IPv6 pattern (supports compression, link-local, etc.)
	ipv6Regex = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{1,4}:){7}[0-9a-f]{1,4}\b|(?:[0-9a-f]{1,4}:)*:(?:[0-9a-f]{1,4}:)*[0-9a-f]{1,4}\b`)

	// Phone number patterns (various formats)
	phoneRegex = regexp.MustCompile(`\+?1?[-.]?\(?\d{3}\)?[-.]?\d{3}[-.]?\d{4}\b`)
)

// RedactTelegramContent redacts sensitive information from telegram content
func RedactTelegramContent(content string, policy RedactionPolicy) string {
	if content == "" {
		return content
	}

	redacted := content

	// Redact emails
	if policy.RedactEmails {
		redacted = emailRegex.ReplaceAllString(redacted, "[EMAIL_REDACTED]")
	}

	// Redact IP addresses
	if policy.RedactIPs {
		redacted = ipv4Regex.ReplaceAllString(redacted, "[IP_REDACTED]")
		redacted = ipv6Regex.ReplaceAllString(redacted, "[IPv6_REDACTED]")
	}

	// Redact phone numbers
	if policy.RedactPhones {
		redacted = phoneRegex.ReplaceAllString(redacted, "[PHONE_REDACTED]")
	}

	// Truncate to max length after redaction (rune-aware to preserve UTF-8)
	if policy.MaxContentLength > 0 {
		runes := []rune(redacted)
		if len(runes) > policy.MaxContentLength {
			redacted = string(runes[:policy.MaxContentLength]) + "...[REDACTED]"
		}
	}

	return redacted
}

// RedactMessageID masks part of the message ID for privacy
func RedactMessageID(messageID string) string {
	runes := []rune(messageID)
	if len(runes) <= 8 {
		return messageID // Too short to redact meaningfully
	}

	// Keep first 4 and last 4 characters, mask the middle (rune-aware)
	return string(runes[:4]) + "***" + string(runes[len(runes)-4:])
}

// RedactFlightNumber keeps the airline code but masks the flight number
func RedactFlightNumber(flightNumber string) string {
	runes := []rune(flightNumber)
	if len(runes) <= 3 {
		return flightNumber // Too short to redact meaningfully
	}

	// Assuming format like "CA1234" - keep first 2 chars (airline), mask rest (rune-aware)
	airline := string(runes[:2])
	return airline + strings.Repeat("*", len(runes)-2)
}

// SanitizeForLogging sanitizes a telegram for safe logging
func SanitizeForLogging(messageID, flightNumber, content string, policy RedactionPolicy) map[string]string {
	return map[string]string{
		"message_id_redacted":    RedactMessageID(messageID),
		"flight_number_redacted": RedactFlightNumber(flightNumber),
		"content_snippet":        RedactTelegramContent(content, policy),
	}
}

// RedactUserIdentifiableInfo removes common PII patterns from any text
func RedactUserIdentifiableInfo(text string) string {
	policy := DefaultRedactionPolicy()
	policy.MaxContentLength = 0 // Don't truncate, just redact patterns
	return RedactTelegramContent(text, policy)
}
