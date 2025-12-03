package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

var tracer = otel.Tracer("caatsm.repository")

const (
	// DefaultSearchLimit is the default number of results per page when limit is not specified or invalid.
	DefaultSearchLimit = 50

	// DefaultRouteStatsLimit is the default number of top routes to return when limit is not specified or invalid.
	DefaultRouteStatsLimit = 10
)

// Store implements ports.Repository using PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// NewPool initialises a pgx connection pool using the provided configuration.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	if cfg.MaxOpenConnections > 0 {
		poolCfg.MaxConns = int32(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		poolCfg.MinConns = int32(cfg.MaxIdleConnections)
	}
	if cfg.ConnectionMaxLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.ConnectionMaxLifetime
	} else {
		poolCfg.MaxConnLifetime = 30 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}

// New creates a new Store from the given pgx pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Save persists a single telegram.
func (s *Store) Save(ctx context.Context, telegram *domain.Telegram) error {
	ctx, span := tracer.Start(ctx, "repository.Save")
	defer span.End()

	span.SetAttributes(
		attribute.String("telegram.message_id", telegram.MessageID),
		attribute.String("telegram.type", telegram.Type),
	)

	query := `INSERT INTO telegrams (message_id, type, time, flight_number, source, destination, content, priority, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (message_id) DO NOTHING`

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
func (s *Store) BulkSave(ctx context.Context, telegrams []any) error {
	if len(telegrams) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction for bulk save (%d telegrams): %w", len(telegrams), err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `INSERT INTO telegrams (message_id, type, time, flight_number, source, destination, content, priority, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (message_id) DO NOTHING`

	batch := &pgx.Batch{}
	for _, item := range telegrams {
		telegram, ok := item.(*domain.Telegram)
		if !ok {
			return fmt.Errorf("bulk save: invalid type, expected *domain.Telegram, got %T", item)
		}
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

	for i := 0; i < len(telegrams); i++ {
		_, err := results.Exec()
		if err != nil {
			msgID := "unknown"
			if i < len(telegrams) && telegrams[i] != nil {
				if telegram, ok := telegrams[i].(*domain.Telegram); ok {
					msgID = telegram.MessageID
				}
			}
			_ = results.Close()
			return fmt.Errorf("bulk save failed at index %d (message_id: %s, total: %d): %w", i, msgID, len(telegrams), err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close batch results (processed %d/%d telegrams): %w", len(telegrams), len(telegrams), err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction after bulk save (%d telegrams): %w", len(telegrams), err)
	}
	return nil
}

// buildSearchConditions builds WHERE conditions and arguments from SearchFilters.
// Returns the WHERE clause (with "WHERE " prefix if conditions exist) and the args slice.
// Maintains proper argPos sequence for parameterized queries.
func (s *Store) buildSearchConditions(filter domain.SearchFilters) (whereClause string, args []any) {
	var conditions []string
	argPos := 1

	// If MessageIDs are provided, use them for filtering (from Meilisearch results)
	if len(filter.MessageIDs) > 0 {
		placeholders := make([]string, len(filter.MessageIDs))
		for i, msgID := range filter.MessageIDs {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, msgID)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("message_id IN (%s)", strings.Join(placeholders, ",")))
	} else if filter.Query != "" {
		// Only use ILIKE if no MessageIDs (Meilisearch handles full-text search)
		conditions = append(conditions, fmt.Sprintf("(content ILIKE $%d OR flight_number ILIKE $%d OR message_id ILIKE $%d)", argPos, argPos, argPos))
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}

	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, t)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Sources) > 0 {
		placeholders := make([]string, len(filter.Sources))
		for i, src := range filter.Sources {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, src)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("source IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Destinations) > 0 {
		placeholders := make([]string, len(filter.Destinations))
		for i, dst := range filter.Destinations {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, dst)
			argPos++
		}
		conditions = append(conditions, fmt.Sprintf("destination IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Priorities) > 0 {
		placeholders := make([]string, len(filter.Priorities))
		for i, p := range filter.Priorities {
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

	}

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	return whereClause, args
}

// buildSearchQuery builds the final SELECT query with pagination and sorting.
// Validates sort column and order, sets default limit/offset, and appends them to args.
// Returns the formatted query string and the final args slice (with limit/offset appended).
func (s *Store) buildSearchQuery(whereClause string, args []any, pagination domain.Pagination) (query string, finalArgs []any) {
	argPos := len(args) + 1

	limit := pagination.Limit
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	offset := pagination.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := pagination.SortBy
	if sortBy == "" {
		sortBy = "time"
	}
	// SQL Injection Safety: sortBy is validated against allowedSortColumns whitelist.
	// Only schema-defined column names can reach the query builder.
	// ORDER clause is safe to interpolate with fmt.Sprintf here.
	allowedSortColumns := map[string]bool{
		"time":          true,
		"priority":      true,
		"message_id":    true,
		"type":          true,
		"flight_number": true,
		"source":        true,
		"destination":   true,
	}
	if !allowedSortColumns[sortBy] {
		sortBy = "time"
	}

	order := strings.ToUpper(pagination.Order)
	if order == "" {
		order = "DESC"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	limitPlaceholder := fmt.Sprintf("$%d", argPos)
	offsetPlaceholder := fmt.Sprintf("$%d", argPos+1)
	finalArgs = append(args, limit, offset)

	query = fmt.Sprintf(`
		SELECT message_id, type, time, flight_number, source, destination, content, priority, raw_data
		FROM telegrams
		%s
		ORDER BY %s %s
		LIMIT %s OFFSET %s
	`, whereClause, sortBy, order, limitPlaceholder, offsetPlaceholder)

	return query, finalArgs
}

// Search performs a structured query over telegram records.
func (s *Store) Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error) {
	ctx, span := tracer.Start(ctx, "repository.Search")
	defer span.End()

	span.SetAttributes(
		attribute.String("filter.query", filter.Query),
		attribute.Int("filter.limit", filter.Pagination.Limit),
		attribute.Int("filter.offset", filter.Pagination.Offset),
	)

	whereClause, args := s.buildSearchConditions(filter)

	countQuery := "SELECT COUNT(*) FROM telegrams " + whereClause
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count telegrams: %w", err)
	}

	query, finalArgs := s.buildSearchQuery(whereClause, args, filter.Pagination)

	rows, err := s.pool.Query(ctx, query, finalArgs...)
	if err != nil {
		return nil, fmt.Errorf("search telegrams: %w", err)
	}
	defer rows.Close()

	var telegrams []domain.Telegram
	for rows.Next() {
		var t domain.Telegram
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
			return nil, fmt.Errorf("scan telegram row (limit: %d, offset: %d): %w", filter.Pagination.Limit, filter.Pagination.Offset, err)
		}
		telegrams = append(telegrams, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &domain.SearchResult{
		Telegrams: telegrams,
		Total:     total,
		Page:      filter.Pagination,
	}, nil
}

// StreamSearch streams search results for large exports.
// It returns two channels: one for telegrams and one for errors.
// The telegram channel will be closed when all results are sent or an error occurs.
func (s *Store) StreamSearch(ctx context.Context, filter domain.SearchFilters) (<-chan *domain.Telegram, <-chan error) {
	telegramCh := make(chan *domain.Telegram, 100) // Buffered channel for better performance
	errCh := make(chan error, 1)

	go func() {
		defer close(telegramCh)
		defer close(errCh)

		whereClause, args := s.buildSearchConditions(filter)
		query, finalArgs := s.buildSearchQuery(whereClause, args, filter.Pagination)

		rows, err := s.pool.Query(ctx, query, finalArgs...)
		if err != nil {
			errCh <- fmt.Errorf("search telegrams: %w", err)
			return
		}
		defer rows.Close()

		// Stream rows to channel
		for rows.Next() {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}

			var t domain.Telegram
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
				errCh <- fmt.Errorf("scan telegram row: %w", err)
				return
			}

			// Send telegram to channel
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case telegramCh <- &t:
			}
		}

		if err := rows.Err(); err != nil {
			errCh <- fmt.Errorf("rows error: %w", err)
			return
		}
	}()

	return telegramCh, errCh
}

// TrafficSummary returns aggregated data for dashboards.
func (s *Store) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	var whereClause string
	var args []any

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
	if err := typeRows.Err(); err != nil {
		return nil, fmt.Errorf("type rows error: %w", err)
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
	if err := priorityRows.Err(); err != nil {
		return nil, fmt.Errorf("priority rows error: %w", err)
	}

	return &domain.TrafficSummary{
		TotalMessages: total,
		ByType:        byType,
		ByPriority:    byPriority,
	}, nil
}

// RouteStats returns top routes.
func (s *Store) RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	if limit <= 0 {
		limit = DefaultRouteStatsLimit
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

	var stats []domain.RouteStat
	for rows.Next() {
		var stat domain.RouteStat
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

// FindByID retrieves a single telegram by message ID.
func (s *Store) FindByID(ctx context.Context, messageID string) (*domain.Telegram, error) {
	query := `SELECT message_id, type, time, flight_number, source, destination, content, priority, raw_data
		FROM telegrams WHERE message_id = $1`

	var t domain.Telegram
	err := s.pool.QueryRow(ctx, query, messageID).Scan(
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
		return nil, fmt.Errorf("find telegram by ID: %w", err)
	}

	return &t, nil
}

// Delete removes a telegram by message ID.
func (s *Store) Delete(ctx context.Context, messageID string) error {
	query := `DELETE FROM telegrams WHERE message_id = $1`

	_, err := s.pool.Exec(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("delete telegram: %w", err)
	}

	return nil
}
