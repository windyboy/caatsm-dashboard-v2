package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/models"
)

func TestToDomain(t *testing.T) {
	t.Run("valid conversion", func(t *testing.T) {
		model := &models.Telegram{
			MessageID:    "TEST-001",
			Type:         "aftn",
			Time:         time.Now(),
			FlightNumber: "TEST123",
			Source:       "ZBAA",
			Destination:  "ZSPD",
			Priority:     2,
			Content:      "Test content",
			RawData:      "Raw data",
		}

		domain := ToDomain(model)

		require.NotNil(t, domain)
		assert.Equal(t, model.MessageID, domain.MessageID)
		assert.Equal(t, model.Type, domain.Type)
		assert.Equal(t, model.Time, domain.Time)
		assert.Equal(t, model.FlightNumber, domain.FlightNumber)
		assert.Equal(t, model.Source, domain.Source)
		assert.Equal(t, model.Destination, domain.Destination)
		assert.Equal(t, model.Priority, domain.Priority)
		assert.Equal(t, model.Content, domain.Content)
		assert.Equal(t, model.RawData, domain.RawData)
	})

	t.Run("nil input", func(t *testing.T) {
		domain := ToDomain(nil)
		assert.Nil(t, domain)
	})
}

func TestFromDomain(t *testing.T) {
	t.Run("valid conversion", func(t *testing.T) {
		domain := &Telegram{
			MessageID:    "TEST-001",
			Type:         "aftn",
			Time:         time.Now(),
			FlightNumber: "TEST123",
			Source:       "ZBAA",
			Destination:  "ZSPD",
			Priority:     2,
			Content:      "Test content",
			RawData:      "Raw data",
		}

		model := FromDomain(domain)

		require.NotNil(t, model)
		assert.Equal(t, domain.MessageID, model.MessageID)
		assert.Equal(t, domain.Type, model.Type)
		assert.Equal(t, domain.Time, model.Time)
		assert.Equal(t, domain.FlightNumber, model.FlightNumber)
		assert.Equal(t, domain.Source, model.Source)
		assert.Equal(t, domain.Destination, model.Destination)
		assert.Equal(t, domain.Priority, model.Priority)
		assert.Equal(t, domain.Content, model.Content)
		assert.Equal(t, domain.RawData, model.RawData)
	})

	t.Run("nil input", func(t *testing.T) {
		model := FromDomain(nil)
		assert.Nil(t, model)
	})

	t.Run("round trip conversion", func(t *testing.T) {
		original := &models.Telegram{
			MessageID:    "TEST-001",
			Type:         "aftn",
			Time:         time.Now(),
			FlightNumber: "TEST123",
			Source:       "ZBAA",
			Destination:  "ZSPD",
			Priority:     2,
			Content:      "Test content",
			RawData:      "Raw data",
		}

		domain := ToDomain(original)
		converted := FromDomain(domain)

		assert.Equal(t, original.MessageID, converted.MessageID)
		assert.Equal(t, original.Type, converted.Type)
		assert.Equal(t, original.Time, converted.Time)
		assert.Equal(t, original.FlightNumber, converted.FlightNumber)
		assert.Equal(t, original.Source, converted.Source)
		assert.Equal(t, original.Destination, converted.Destination)
		assert.Equal(t, original.Priority, converted.Priority)
		assert.Equal(t, original.Content, converted.Content)
		assert.Equal(t, original.RawData, converted.RawData)
	})
}

