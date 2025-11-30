package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegram_Validate(t *testing.T) {
	tests := []struct {
		name     string
		telegram Telegram
		wantErr  bool
		errField string
	}{
		{
			name: "valid telegram",
			telegram: Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  2,
			},
			wantErr: false,
		},
		{
			name: "empty message_id",
			telegram: Telegram{
				MessageID: "",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  2,
			},
			wantErr:  true,
			errField: "message_id",
		},
		{
			name: "zero time",
			telegram: Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Time{},
				Priority:  2,
			},
			wantErr:  true,
			errField: "time",
		},
		{
			name: "empty type",
			telegram: Telegram{
				MessageID: "TEST-001",
				Type:      "",
				Time:      time.Now(),
				Priority:  2,
			},
			wantErr:  true,
			errField: "type",
		},
		{
			name: "priority too low",
			telegram: Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  0,
			},
			wantErr:  true,
			errField: "priority",
		},
		{
			name: "priority too high",
			telegram: Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  4,
			},
			wantErr:  true,
			errField: "priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.telegram.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errField != "" {
					assert.Contains(t, err.Error(), tt.errField)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTelegram_IsHighPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		want     bool
	}{
		{
			name:     "priority 1 is high",
			priority: 1,
			want:     true,
		},
		{
			name:     "priority 2 is not high",
			priority: 2,
			want:     false,
		},
		{
			name:     "priority 3 is not high",
			priority: 3,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  tt.priority,
			}
			assert.Equal(t, tt.want, tg.IsHighPriority())
		})
	}
}

func TestTelegram_ExtractRoute(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		destination string
		want        string
	}{
		{
			name:        "both source and destination",
			source:      "ZBAA",
			destination: "ZSPD",
			want:        "ZBAA → ZSPD",
		},
		{
			name:        "only source",
			source:      "ZBAA",
			destination: "",
			want:        "ZBAA",
		},
		{
			name:        "only destination",
			source:      "",
			destination: "ZSPD",
			want:        "ZSPD",
		},
		{
			name:        "neither source nor destination",
			source:      "",
			destination: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := Telegram{
				MessageID:   "TEST-001",
				Type:        "aftn",
				Time:        time.Now(),
				Priority:    2,
				Source:      tt.source,
				Destination: tt.destination,
			}
			assert.Equal(t, tt.want, tg.ExtractRoute())
		})
	}
}

func TestTelegram_Normalize(t *testing.T) {
	tg := Telegram{
		MessageID:    "TEST-001",
		Type:         "aftn",
		Time:         time.Now(),
		Priority:     2,
		FlightNumber: "test123",
		Source:       "zbaa",
		Destination:  "zspd",
		Content:      "  test content  ",
	}

	// Normalize should not panic
	tg.Normalize()

	// Verify the telegram has been normalized (uppercase fields)
	assert.Equal(t, "TEST-001", tg.MessageID)
	assert.Equal(t, "TEST123", tg.FlightNumber) // Normalize() uppercases flight numbers
	assert.Equal(t, "ZBAA", tg.Source)          // Normalize() uppercases ICAO codes
	assert.Equal(t, "ZSPD", tg.Destination)     // Normalize() uppercases ICAO codes
	assert.Equal(t, "test content", tg.Content) // Normalize() trims whitespace
}
