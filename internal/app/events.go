package app

import "time"

// Event represents a domain event in the system.
// Domain events capture something meaningful that happened in the app.
type Event interface {
	EventType() string
	OccurredAt() time.Time
}

// TelegramReceived is published when a telegram is received from an external source.
type TelegramReceived struct {
	Telegram *Telegram
	Source   string
	At       time.Time
}

func (e TelegramReceived) EventType() string {
	return "telegram.received"
}

func (e TelegramReceived) OccurredAt() time.Time {
	return e.At
}

// TelegramValidated is published when a telegram passes domain validation.
type TelegramValidated struct {
	Telegram *Telegram
	At       time.Time
}

func (e TelegramValidated) EventType() string {
	return "telegram.validated"
}

func (e TelegramValidated) OccurredAt() time.Time {
	return e.At
}

// TelegramPersisted is published when a telegram is successfully saved to the database.
type TelegramPersisted struct {
	Telegram *Telegram
	At       time.Time
}

func (e TelegramPersisted) EventType() string {
	return "telegram.persisted"
}

func (e TelegramPersisted) OccurredAt() time.Time {
	return e.At
}

// TelegramIndexed is published when a telegram is successfully indexed in the search engine.
type TelegramIndexed struct {
	Telegram *Telegram
	At       time.Time
}

func (e TelegramIndexed) EventType() string {
	return "telegram.indexed"
}

func (e TelegramIndexed) OccurredAt() time.Time {
	return e.At
}

// TelegramProcessed is published when a telegram completes the full processing pipeline.
// This is a higher-level event that indicates all steps (validation, persistence, indexing) are complete.
type TelegramProcessed struct {
	Telegram *Telegram
	At       time.Time
}

func (e TelegramProcessed) EventType() string {
	return "telegram.processed"
}

func (e TelegramProcessed) OccurredAt() time.Time {
	return e.At
}

// SearchPerformed is published when a search operation is completed.
type SearchPerformed struct {
	Query       string
	ResultCount int64
	Duration    time.Duration
	Timestamp   time.Time
}

func (e SearchPerformed) EventType() string {
	return "search.performed"
}

func (e SearchPerformed) OccurredAt() time.Time {
	return e.Timestamp
}
