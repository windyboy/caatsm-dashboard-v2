package tracing

import "go.opentelemetry.io/otel/attribute"

// Custom attribute keys for CAATSM Dashboard
const (
	// Telegram attributes
	AttrTelegramID       = attribute.Key("telegram.id")
	AttrTelegramType     = attribute.Key("telegram.type")
	AttrFlightNumber     = attribute.Key("telegram.flight_number")
	AttrSource           = attribute.Key("telegram.source")
	AttrDestination      = attribute.Key("telegram.destination")
	AttrPriority         = attribute.Key("telegram.priority")

	// Search attributes
	AttrSearchQuery      = attribute.Key("search.query")
	AttrSearchLimit      = attribute.Key("search.limit")
	AttrSearchOffset     = attribute.Key("search.offset")
	AttrSearchTotal      = attribute.Key("search.total")
	AttrSearchDuration   = attribute.Key("search.duration_ms")

	// Export attributes
	AttrExportFormat     = attribute.Key("export.format")
	AttrExportCount      = attribute.Key("export.count")
	AttrExportSize       = attribute.Key("export.size_bytes")

	// Database attributes
	AttrDBQuery          = attribute.Key("db.query")
	AttrDBTable          = attribute.Key("db.table")
	AttrDBOperation      = attribute.Key("db.operation")

	// Cache attributes
	AttrCacheKey         = attribute.Key("cache.key")
	AttrCacheHit         = attribute.Key("cache.hit")

	// Event attributes
	AttrEventType        = attribute.Key("event.type")
	AttrEventID          = attribute.Key("event.id")
)

// TelegramAttributes returns common telegram attributes
func TelegramAttributes(messageID, msgType, flightNumber string, priority int) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrTelegramID.String(messageID),
		AttrTelegramType.String(msgType),
		AttrFlightNumber.String(flightNumber),
		AttrPriority.Int(priority),
	}
}

// SearchAttributes returns common search attributes
func SearchAttributes(query string, limit, offset, total int) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrSearchQuery.String(query),
		AttrSearchLimit.Int(limit),
		AttrSearchOffset.Int(offset),
		AttrSearchTotal.Int(total),
	}
}

// ExportAttributes returns common export attributes
func ExportAttributes(format string, count int) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrExportFormat.String(format),
		AttrExportCount.Int(count),
	}
}

// DBAttributes returns common database attributes
func DBAttributes(table, operation string) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrDBTable.String(table),
		AttrDBOperation.String(operation),
	}
}

// CacheAttributes returns common cache attributes
func CacheAttributes(key string, hit bool) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrCacheKey.String(key),
		AttrCacheHit.Bool(hit),
	}
}

// EventAttributes returns common event attributes
func EventAttributes(eventType, eventID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrEventType.String(eventType),
		AttrEventID.String(eventID),
	}
}

