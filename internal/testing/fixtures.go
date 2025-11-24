package testing

import (
	"fmt"
	"time"

	"github.com/windy/caatsm-dashboard/internal/models"
)

// NewTelegram creates a test telegram with default values
func NewTelegram(messageID string) *models.Telegram {
	return &models.Telegram{
		MessageID:    messageID,
		Type:         "aftn",
		Time:         time.Now(),
		FlightNumber: "TEST123",
		Source:       "ZBAA",
		Destination:  "ZSPD",
		Priority:     3,
		Content:      "Test telegram content",
		RawData:      "TEST RAW DATA",
	}
}

// NewTelegramWithType creates a test telegram with specific type
func NewTelegramWithType(messageID, telegramType string) *models.Telegram {
	tg := NewTelegram(messageID)
	tg.Type = telegramType
	return tg
}

// NewTelegramWithPriority creates a test telegram with specific priority
func NewTelegramWithPriority(messageID string, priority int) *models.Telegram {
	tg := NewTelegram(messageID)
	tg.Priority = priority
	return tg
}

// NewTelegramWithRoute creates a test telegram with specific route
func NewTelegramWithRoute(messageID, source, destination string) *models.Telegram {
	tg := NewTelegram(messageID)
	tg.Source = source
	tg.Destination = destination
	return tg
}

// NewTelegramWithTime creates a test telegram with specific time
func NewTelegramWithTime(messageID string, t time.Time) *models.Telegram {
	tg := NewTelegram(messageID)
	tg.Time = t
	return tg
}

// NewTelegrams creates multiple test telegrams
func NewTelegrams(count int) []*models.Telegram {
	telegrams := make([]*models.Telegram, count)
	baseTime := time.Now()
	for i := 0; i < count; i++ {
		telegrams[i] = NewTelegramWithTime(
			fmt.Sprintf("TEST-%d", i+1),
			baseTime.Add(time.Duration(i)*time.Minute),
		)
	}
	return telegrams
}

// NewSearchFilter creates a test search filter
func NewSearchFilter() models.SearchFilter {
	return models.SearchFilter{
		Query: "",
		Page: models.Pagination{
			Limit:  10,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}
}

// NewSearchFilterWithQuery creates a test search filter with query
func NewSearchFilterWithQuery(query string) models.SearchFilter {
	filter := NewSearchFilter()
	filter.Query = query
	return filter
}

// NewSearchFilterWithType creates a test search filter with type filter
func NewSearchFilterWithType(telegramType string) models.SearchFilter {
	filter := NewSearchFilter()
	filter.Type = []string{telegramType}
	return filter
}

// NewTimeWindow creates a test time window
func NewTimeWindow(start, end time.Time) models.TimeWindow {
	return models.TimeWindow{
		Start: start,
		End:   end,
	}
}

// NewTimeWindowLast24h creates a time window for last 24 hours
func NewTimeWindowLast24h() models.TimeWindow {
	now := time.Now()
	return models.TimeWindow{
		Start: now.Add(-24 * time.Hour),
		End:   now,
	}
}
