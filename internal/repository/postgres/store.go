package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/windy/caatsm-dashboard/internal/models"
)

// Store implements repository interfaces backed by PostgreSQL/TimescaleDB.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new Store from the given pgx pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Save persists a single telegram.
func (s *Store) Save(ctx context.Context, telegram *models.Telegram) error {
	query := `INSERT INTO telegrams (message_id, type, time, flight_number, source, destination, content, priority, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (message_id) DO UPDATE SET
			type = EXCLUDED.type,
			time = EXCLUDED.time,
			flight_number = EXCLUDED.flight_number,
			source = EXCLUDED.source,
			destination = EXCLUDED.destination,
			content = EXCLUDED.content,
			priority = EXCLUDED.priority,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW()`

	_, err := s.pool.Exec(ctx, query,
		telegram.MessageID,
		telegram.Type,
		telegram.Time,
		telegram.FlightNumber,
		telegram.Source,
		telegram.Destination,
		telegram.Content,
		telegram.Priority,
		telegram.RawData,
	)
	if err != nil {
		return fmt.Errorf("save telegram: %w", err)
	}
	return nil
}

// BulkSave persists multiple telegrams in a single transaction.
func (s *Store) BulkSave(ctx context.Context, telegrams []*models.Telegram) error {
	if len(telegrams) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO telegrams (message_id, type, time, flight_number, source, destination, content, priority, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (message_id) DO UPDATE SET
			type = EXCLUDED.type,
			time = EXCLUDED.time,
			flight_number = EXCLUDED.flight_number,
			source = EXCLUDED.source,
			destination = EXCLUDED.destination,
			content = EXCLUDED.content,
			priority = EXCLUDED.priority,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW()`

	batch := &pgx.Batch{}
	for _, telegram := range telegrams {
		batch.Queue(query,
			telegram.MessageID,
			telegram.Type,
			telegram.Time,
			telegram.FlightNumber,
			telegram.Source,
			telegram.Destination,
			telegram.Content,
			telegram.Priority,
			telegram.RawData,
		)
	}

	results := tx.SendBatch(ctx, batch)
	// Note: We'll close results explicitly before commit, defer is just a safety net
	defer results.Close()

	for i := 0; i < len(telegrams); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("bulk save telegram %d: %w", i, err)
		}
	}

	// Close batch results BEFORE committing - this is required by pgx
	if err := results.Close(); err != nil {
		return fmt.Errorf("close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// Search performs a structured query over telegram records.
func (s *Store) Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error) {
	var conditions []string
	var args []interface{}
	argPos := 1

	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(content ILIKE $%d OR flight_number ILIKE $%d OR message_id ILIKE $%d)", argPos, argPos, argPos))
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}

	if len(filter.Type) > 0 {
		placeholders := make([]string, len(filter.Type))
		for i, t := range filter.Type {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, t)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Source) > 0 {
		placeholders := make([]string, len(filter.Source))
		for i, src := range filter.Source {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, src)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("source IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Destination) > 0 {
		placeholders := make([]string, len(filter.Destination))
		for i, dst := range filter.Destination {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, dst)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("destination IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Priority) > 0 {
		placeholders := make([]string, len(filter.Priority))
		for i, p := range filter.Priority {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, p)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("priority IN (%s)", strings.Join(placeholders, ",")))
	}

	if !filter.TimeRange.Start.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time >= $%d", argPos))
		args = append(args, filter.TimeRange.Start)
		argPos++
	}

	if !filter.TimeRange.End.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time <= $%d", argPos))
		args = append(args, filter.TimeRange.End)
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := filter.Page.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Page.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := filter.Page.SortBy
	if sortBy == "" {
		sortBy = "time"
	}
	order := filter.Page.Order
	if order == "" {
		order = "DESC"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	countQuery := "SELECT COUNT(*) FROM telegrams " + whereClause
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count telegrams: %w", err)
	}

	limitPlaceholder := fmt.Sprintf("$%d", argPos)
	offsetPlaceholder := fmt.Sprintf("$%d", argPos+1)
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT message_id, type, time, flight_number, source, destination, content, priority, raw_data
		FROM telegrams
		%s
		ORDER BY %s %s
		LIMIT %s OFFSET %s
	`, whereClause, sortBy, order, limitPlaceholder, offsetPlaceholder)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search telegrams: %w", err)
	}
	defer rows.Close()

	var telegrams []models.Telegram
	for rows.Next() {
		var t models.Telegram
		err := rows.Scan(
			&t.MessageID,
			&t.Type,
			&t.Time,
			&t.FlightNumber,
			&t.Source,
			&t.Destination,
			&t.Content,
			&t.Priority,
			&t.RawData,
		)
		if err != nil {
			return nil, fmt.Errorf("scan telegram: %w", err)
		}
		telegrams = append(telegrams, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &models.SearchResult{
		Telegrams: telegrams,
		Total:     total,
		Page:      filter.Page,
	}, nil
}

// TrafficSummary returns aggregated data for dashboards.
func (s *Store) TrafficSummary(ctx context.Context, window models.TimeWindow) (*models.TrafficSummary, error) {
	var whereClause string
	var args []interface{}

	if !window.Start.IsZero() || !window.End.IsZero() {
		var conditions []string
		if !window.Start.IsZero() {
			conditions = append(conditions, "time >= $1")
			args = append(args, window.Start)
		}
		if !window.End.IsZero() {
			argPos := len(args) + 1
			conditions = append(conditions, fmt.Sprintf("time <= $%d", argPos))
			args = append(args, window.End)
		}
		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ")
		}
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM telegrams %s", whereClause)
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count total: %w", err)
	}

	typeQuery := fmt.Sprintf(`
		SELECT type, COUNT(*) as count
		FROM telegrams
		%s
		GROUP BY type
	`, whereClause)

	typeRows, err := s.pool.Query(ctx, typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query by type: %w", err)
	}
	defer typeRows.Close()

	byType := make(map[string]int64)
	for typeRows.Next() {
		var t string
		var count int64
		if err := typeRows.Scan(&t, &count); err != nil {
			return nil, fmt.Errorf("scan type: %w", err)
		}
		byType[t] = count
	}

	priorityQuery := fmt.Sprintf(`
		SELECT priority, COUNT(*) as count
		FROM telegrams
		%s
		GROUP BY priority
	`, whereClause)

	priorityRows, err := s.pool.Query(ctx, priorityQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query by priority: %w", err)
	}
	defer priorityRows.Close()

	byPriority := make(map[int]int64)
	for priorityRows.Next() {
		var p int
		var count int64
		if err := priorityRows.Scan(&p, &count); err != nil {
			return nil, fmt.Errorf("scan priority: %w", err)
		}
		byPriority[p] = count
	}

	return &models.TrafficSummary{
		TotalMessages: total,
		ByType:        byType,
		ByPriority:    byPriority,
	}, nil
}

// RouteStats returns top routes.
func (s *Store) RouteStats(ctx context.Context, limit int) ([]models.RouteStat, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT source, destination, COUNT(*) as count
		FROM telegrams
		WHERE source IS NOT NULL AND destination IS NOT NULL
		GROUP BY source, destination
		ORDER BY count DESC
		LIMIT $1
	`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query route stats: %w", err)
	}
	defer rows.Close()

	var stats []models.RouteStat
	for rows.Next() {
		var stat models.RouteStat
		if err := rows.Scan(&stat.Source, &stat.Destination, &stat.Count); err != nil {
			return nil, fmt.Errorf("scan route stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return stats, nil
}
