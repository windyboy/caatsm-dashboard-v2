package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultParser_Parse(t *testing.T) {
	p := NewDefaultParser()

	t.Run("valid JSON", func(t *testing.T) {
		tg := Telegram{
			MessageID:    "TEST-001",
			Type:         "AFTN",
			Time:         time.Now(),
			FlightNumber: "TEST123",
			Source:       "ZBAA",
			Destination:  "ZSPD",
			Priority:     2,
			Content:      "Test content",
			RawData:      "Raw data",
		}

		data, err := json.Marshal(tg)
		require.NoError(t, err)

		parsed, err := p.Parse(string(data))
		require.NoError(t, err)
		assert.Equal(t, tg.MessageID, parsed.MessageID)
		assert.Equal(t, tg.Type, parsed.Type)
		assert.Equal(t, tg.FlightNumber, parsed.FlightNumber)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := p.Parse("invalid json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestDefaultParser_Supports(t *testing.T) {
	p := NewDefaultParser()

	// Default parser supports all types
	assert.True(t, p.Supports("AFTN"))
	assert.True(t, p.Supports("SITA"))
	assert.True(t, p.Supports("ACARS"))
	assert.True(t, p.Supports("CPDLC"))
	assert.True(t, p.Supports("unknown"))
}
