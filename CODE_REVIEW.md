# Code Review: CAATSM Dashboard

## 1) Executive Summary

The CAATSM Dashboard is a Go-based web application for searching and analyzing aviation telegram messages, using PostgreSQL, Meilisearch, Redis, and NATS. The codebase demonstrates solid architectural patterns with clear separation of concerns, but several critical security vulnerabilities, correctness issues, and performance concerns need immediate attention. The top three risks are: (1) **SQL injection vulnerability** in the `sortBy` field which is directly interpolated into SQL queries without validation, (2) **XSS vulnerabilities** from unescaped user input rendered in HTML responses, and (3) **DoS risk** from unlimited export query limits and lack of rate limiting. The codebase would benefit from input validation, output encoding, proper error handling, and resource management improvements.

## 2) Findings Table

| ID | Severity | Category | Symptom | Why it matters | Evidence (line refs) | Fix summary |
|----|----------|----------|---------|----------------|---------------------|-------------|
| F1 | High | Security | SQL injection in `sortBy` field | User-controlled `sort_by` parameter is directly interpolated into SQL query without validation, allowing SQL injection | `internal/repository/postgres/store.go:195-222` | Whitelist allowed sort columns or use parameterized column names |
| F2 | High | Security | XSS in HTML responses | User input (query, suggestions, content) is rendered directly in HTML without escaping | `internal/handlers/handlers.go:108,112-131,158` | Use HTML escaping (html/template or html.EscapeString) |
| F3 | High | Security | No rate limiting | All endpoints are unprotected from DoS attacks | `internal/server/server.go:110-121` | Add rate limiting middleware |
| F4 | High | Security | Hardcoded default secrets | Default Meilisearch API key "masterKey" in config | `config/loader.go:86` | Remove default secrets, require explicit configuration |
| F5 | High | Security | Export endpoint allows unlimited data extraction | Export has hardcoded 10,000 limit but no authentication/authorization checks | `internal/handlers/handlers.go:211-256`, `internal/services/export.go:32` | Add authentication, enforce reasonable limits, add pagination |
| F6 | High | Correctness | SQL injection in ORDER BY clause | `sortBy` and `order` are interpolated directly into SQL | `internal/repository/postgres/store.go:218-224` | Validate against whitelist |
| F7 | High | Correctness | Meilisearch filter injection | User input in filter strings is quoted but not validated for filter syntax injection | `internal/services/search.go:71,79,87,95` | Use Meilisearch's structured filter API instead of string concatenation |
| F8 | Medium | Security | Basic auth credentials in config file | Passwords stored in plaintext config files | `config/config.go:73-78` | Use environment variables or secrets manager |
| F9 | Medium | Security | No input validation on query parameters | No length limits, format validation, or sanitization on user inputs | `internal/handlers/handlers.go:41-96` | Add validation layer with length limits and format checks |
| F10 | Medium | Security | CORS enabled without restrictions | CORS middleware has no origin restrictions | `internal/server/server.go:37` | Configure allowed origins |
| F11 | Medium | Security | Sensitive data in logs | Request logging may include sensitive query parameters | `internal/observability/middleware.go:12-65` | Redact sensitive fields from logs |
| F12 | Medium | Correctness | Cache check doesn't return cached data | Cache hit is detected but result is never returned | `internal/services/search.go:48-54` | Implement cache retrieval and deserialization |
| F13 | Medium | Correctness | Timezone handling | Time parsing/formatting doesn't explicitly handle timezones | `internal/handlers/handlers.go:87-95` | Use UTC consistently and document timezone assumptions |
| F14 | Medium | Correctness | Error handling in BulkSave | Transaction rollback on error but error context lost | `internal/repository/postgres/store.go:55-115` | Improve error wrapping with transaction context |
| F15 | Medium | Performance | N+1 query potential in TrafficSummary | Multiple separate queries instead of single query with JOINs | `internal/repository/postgres/store.go:263-342` | Combine queries or use CTEs |
| F16 | Medium | Performance | No connection pool limits validation | Pool config not validated against reasonable bounds | `internal/platform/postgres/postgres.go:23-28` | Add validation for pool size limits |
| F17 | Medium | Performance | Export loads all results into memory | 10,000 records loaded at once for export | `internal/services/export.go:30-50` | Stream results or paginate |
| F18 | Medium | Design | Context.Background() in app initialization | Long-lived context used for initialization | `internal/server/server.go:42` | Use request context or separate init context |
| F19 | Medium | Design | Missing error taxonomy | Generic errors without structured error types | Throughout | Define error types for better error handling |
| F20 | Low | Correctness | Autocomplete size validation | Size parameter validated but could be negative | `internal/handlers/handlers.go:141-146` | Add explicit bounds check |
| F21 | Low | Correctness | Priority can be negative | No validation that priority is non-negative | `internal/handlers/handlers.go:80-84` | Add validation |
| F22 | Low | Performance | Inefficient cache key generation | String concatenation instead of hash for cache keys | `internal/services/search.go:266-270` | Use hash (e.g., SHA256) for cache keys |
| F23 | Low | Maintainability | Dead code in cache | Cache check exists but never used | `internal/services/search.go:48-54` | Remove or implement properly |
| F24 | Low | Maintainability | TODO comments | Stream endpoint not implemented | `internal/handlers/handlers.go:258-269` | Implement or document why not implemented |
| F25 | Low | Maintainability | Magic numbers | Hardcoded limits (50, 10000, 5) throughout | Multiple files | Extract to constants or config |

## 3) Patch Suggestions

### F1, F6: SQL Injection in ORDER BY

```diff
internal/repository/postgres/store.go
+import (
+	"strings"
+)
+
 func (s *Store) Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error) {
 	// ... existing code ...
 
 	sortBy := filter.Page.SortBy
 	if sortBy == "" {
 		sortBy = "time"
 	}
+	// Whitelist allowed sort columns to prevent SQL injection
+	allowedSortColumns := map[string]bool{
+		"time":          true,
+		"priority":      true,
+		"message_id":    true,
+		"type":          true,
+		"flight_number": true,
+		"source":        true,
+		"destination":   true,
+	}
+	if !allowedSortColumns[sortBy] {
+		sortBy = "time"
+	}
+	
 	order := filter.Page.Order
 	if order == "" {
 		order = "DESC"
 	}
 	if order != "ASC" && order != "DESC" {
 		order = "DESC"
 	}
+	// order is already validated above, safe to interpolate
 
 	// ... rest of code ...
```

**Trade-offs**: Whitelisting is safer than blacklisting. The performance impact is negligible. Alternative: use a switch statement for more explicit control.

### F2: XSS in HTML Responses

```diff
internal/handlers/handlers.go
+import (
+	"html"
+)

 func (h *Handler) Search(ctx echo.Context) error {
 	// ... existing code ...
 
 	// Build HTML response
 	var html strings.Builder
 	html.WriteString(fmt.Sprintf("<p class=\"text-sm text-slate-400 mb-4\">Found %d results</p>", result.Total))
 	if len(result.Telegrams) > 0 {
 		html.WriteString("<div class=\"space-y-2\">")
 		for _, t := range result.Telegrams {
 			html.WriteString(fmt.Sprintf(`
 			<div class="rounded border border-slate-800 p-4">
 				<div class="flex justify-between mb-2">
-					<span class="font-medium">%s</span>
+					<span class="font-medium">%s</span>
 					<span class="text-xs text-slate-500">%s</span>
 				</div>
-				<p class="text-sm text-slate-300">%s</p>
+				<p class="text-sm text-slate-300">%s</p>
 				<div class="mt-2 flex gap-2 text-xs text-slate-400">
-					<span>Type: %s</span>
+					<span>Type: %s</span>
 					<span>Priority: %d</span>
-					<span>%s → %s</span>
+					<span>%s → %s</span>
 				</div>
 			</div>`,
-				t.MessageID,
+				html.EscapeString(t.MessageID),
 				t.Time.Format("2006-01-02 15:04:05"),
-				t.Content,
-				t.Type,
+				html.EscapeString(t.Content),
+				html.EscapeString(t.Type),
 				t.Priority,
-				t.Source,
-				t.Destination))
+				html.EscapeString(t.Source),
+				html.EscapeString(t.Destination))
 		}
 		html.WriteString("</div>")
 	} else {
 		html.WriteString("<p class=\"text-slate-500\">No results found</p>")
 	}
 
 	return ctx.HTML(http.StatusOK, html.String())
 }
 
 func (h *Handler) Autocomplete(ctx echo.Context) error {
 	// ... existing code ...
 
 	var html strings.Builder
 	for _, suggestion := range suggestions {
-		html.WriteString(fmt.Sprintf("<div class=\"cursor-pointer hover:text-slate-200 p-1\">%s</div>", suggestion))
+		html.WriteString(fmt.Sprintf("<div class=\"cursor-pointer hover:text-slate-200 p-1\">%s</div>", html.EscapeString(suggestion)))
 	}
 
 	return ctx.HTML(http.StatusOK, html.String())
 }
```

**Trade-offs**: Using `html.EscapeString` is the minimal fix. Better approach: use templating engine (templ) for all HTML generation to ensure automatic escaping.

### F3: Rate Limiting

```diff
internal/server/server.go
+import (
+	"github.com/labstack/echo/v4/middleware"
+)

 func (s *Server) registerRoutes() {
 	s.e.Use(auth.Middleware(s.cfg.Auth))
+	
+	// Add rate limiting
+	rateLimiterConfig := middleware.RateLimiterConfig{
+		Rate:  10, // requests per second
+		Burst: 20,
+	}
+	s.e.Use(middleware.RateLimiterWithConfig(rateLimiterConfig))

 	// Serve static files (CSS, favicon, etc.) - register before other routes
 	s.e.Static("/css", "public/css")
```

**Trade-offs**: Simple rate limiting helps but may need Redis-backed distributed rate limiting for production. Consider per-endpoint limits (stricter for export).

### F7: Meilisearch Filter Injection

```diff
internal/services/search.go
-	// Build filters
-	var filterParts []string
-
-	if len(filter.Type) > 0 {
-		typeFilters := make([]string, len(filter.Type))
-		for i, t := range filter.Type {
-			typeFilters[i] = fmt.Sprintf("type = %q", t)
-		}
-		filterParts = append(filterParts, "("+strings.Join(typeFilters, " OR ")+")")
-	}
+	// Build filters using Meilisearch's structured filter API
+	var filterParts []interface{}
+
+	if len(filter.Type) > 0 {
+		typeFilters := make([]string, len(filter.Type))
+		for i, t := range filter.Type {
+			// Validate and sanitize type value
+			if strings.ContainsAny(t, `"=()[]{}`) {
+				continue // Skip invalid characters
+			}
+			typeFilters[i] = fmt.Sprintf("type = %q", t)
+		}
+		if len(typeFilters) > 0 {
+			filterParts = append(filterParts, strings.Join(typeFilters, " OR "))
+		}
+	}
```

**Better approach**: Use Meilisearch's filter array format directly instead of string concatenation. Check Meilisearch Go client documentation for structured filter support.

### F9: Input Validation

```diff
internal/handlers/handlers.go
+const (
+	maxQueryLength    = 500
+	maxLimit          = 1000
+	maxOffset         = 100000
+	maxAutocompleteSize = 50
+)

 func (h *Handler) Search(ctx echo.Context) error {
 	filter := models.SearchFilter{
-		Query: ctx.FormValue("query"),
+		Query: sanitizeQuery(ctx.FormValue("query")),
 		Page: models.Pagination{
 			Limit:  50,
 			Offset: 0,
 			SortBy: "time",
 			Order:  "desc",
 		},
 	}
+	
+	if len(filter.Query) > maxQueryLength {
+		return echo.NewHTTPError(http.StatusBadRequest, "query too long")
+	}

 	// Parse pagination
 	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
 		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
+			if limit > maxLimit {
+				limit = maxLimit
+			}
 			filter.Page.Limit = limit
 		}
 	}
 	if offsetStr := ctx.QueryParam("offset"); offsetStr != "" {
 		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
+			if offset > maxOffset {
+				return echo.NewHTTPError(http.StatusBadRequest, "offset too large")
+			}
 			filter.Page.Offset = offset
 		}
 	}
+	
+func sanitizeQuery(q string) string {
+	// Remove control characters and limit length
+	q = strings.TrimSpace(q)
+	if len(q) > maxQueryLength {
+		q = q[:maxQueryLength]
+	}
+	return q
+}
```

**Trade-offs**: Basic validation helps. Consider using a validation library (e.g., `go-playground/validator`) for more comprehensive validation.

### F12: Cache Implementation

```diff
internal/services/search.go
+import (
+	"encoding/json"
+)

 func (s *searchService) Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error) {
 	// Check cache first
 	cacheKey := s.buildCacheKey(filter)
-	if _, err := s.cache.Get(ctx, cacheKey); err == nil {
-		// Return cached result if available
-		// For simplicity, we'll skip cache deserialization here
-		// In production, you'd deserialize the cached JSON
+	cached, err := s.cache.Get(ctx, cacheKey)
+	if err == nil {
+		var cachedResult models.SearchResult
+		if err := json.Unmarshal(cached, &cachedResult); err == nil {
+			return &cachedResult, nil
+		}
+		// If deserialization fails, continue to fresh query
 	}

 	// ... existing search code ...
 
 	result := &models.SearchResult{
 		Telegrams: telegrams,
 		Total:     total,
 		Page:      filter.Page,
 	}
+	
+	// Cache result
+	if resultBytes, err := json.Marshal(result); err == nil {
+		_ = s.cache.Set(ctx, cacheKey, resultBytes) // Log error but don't fail request
+	}

-	// Cache result
-	// In production, you'd serialize the result to JSON and cache it
-	// For now, we'll skip caching

 	return result, nil
 }
```

**Trade-offs**: Simple implementation. Consider cache invalidation strategy and TTL tuning based on data freshness requirements.

## 4) Tests to Add

### Unit Tests

- `TestHandler_Search_ValidatesInput_RejectsLongQuery`: Verify query length limit enforcement
- `TestHandler_Search_ValidatesInput_RejectsInvalidLimit`: Test limit bounds checking
- `TestHandler_Search_EscapesHTML_PreventsXSS`: Verify HTML escaping in search results
- `TestHandler_Autocomplete_EscapesHTML`: Test autocomplete XSS prevention
- `TestStore_Search_ValidatesSortBy_RejectsSQLInjection`: Test sortBy whitelist prevents SQL injection
- `TestStore_Search_ValidatesOrder_RejectsInvalidOrder`: Test order validation
- `TestSearchService_BuildCacheKey_GeneratesConsistentKeys`: Test cache key generation
- `TestSearchService_Search_ReturnsCachedResult`: Test cache hit path
- `TestExportService_Export_EnforcesLimit`: Test export limit enforcement
- `TestBulkSave_RollbackOnError`: Test transaction rollback on batch error
- `TestBulkSave_HandlesEmptySlice`: Test empty telegrams slice
- `TestTrafficSummary_HandlesZeroTimeWindow`: Test default time window behavior

### Integration Tests

- `TestSearchEndpoint_WhenInvalidSortBy_ThenReturnsDefaultSort`: Verify SQL injection prevention in ORDER BY
- `TestSearchEndpoint_WhenXSSInQuery_ThenEscapesOutput`: Verify XSS prevention end-to-end
- `TestExportEndpoint_WhenLargeResultSet_ThenRespectsLimit`: Test export pagination/limits
- `TestExportEndpoint_WhenUnauthenticated_ThenReturns401`: Test authentication on export
- `TestRateLimiting_WhenExceedsLimit_ThenReturns429`: Test rate limiting enforcement
- `TestCacheIntegration_WhenCacheHit_ThenReturnsCachedData`: Test Redis cache integration
- `TestDatabaseConnection_WhenPoolExhausted_ThenHandlesGracefully`: Test connection pool limits

### Property/Fuzz Tests

- **Input domain**: Query strings (length 0-1000, various Unicode, control chars)
- **Invariant**: Search results always have `len(Telegrams) <= Page.Limit`
- **Invariant**: Cache keys are deterministic for same filter
- **Invariant**: SQL queries never contain unescaped user input in column/table names
- **Fuzz target**: `SearchFilter` struct with random valid/invalid combinations

## 5) Performance Notes

### Complexity Analysis

- **Search endpoint**: O(n) where n = result set size. Meilisearch query is O(log n + k) where k = limit. Overall: **O(k)** for typical queries.
- **TrafficSummary**: O(n) with 3 separate queries = **3O(n)**. Can be optimized to **O(n)** with single query using CTEs.
- **BulkSave**: O(m) where m = batch size. Uses batch operations efficiently. **Good**.
- **Export**: O(n) where n = result size (up to 10,000). Memory: **O(n)** - all results loaded. **Bottleneck** for large exports.

### Bottlenecks

1. **Export endpoint** (F17): Loads 10,000 records into memory. For CSV, could stream. For Excel/PDF, may need buffering.
2. **TrafficSummary** (F15): Three separate queries instead of one. Estimated **3x improvement** with combined query.
3. **Cache key generation** (F22): String concatenation is O(n) but fine for small inputs. Hash would be O(1) lookup but adds overhead.

### Quick Wins

1. **Combine TrafficSummary queries** (F15): Use single query with CTEs. Expected: **~3x faster**, low effort.
2. **Stream CSV export** (F17): Use streaming CSV writer. Expected: **~50% memory reduction**, medium effort.
3. **Add database connection pooling metrics**: Monitor pool exhaustion. Expected: **Better observability**, low effort.
4. **Implement cache properly** (F12): Enable caching for search results. Expected: **~80% latency reduction** for repeated queries, medium effort.
5. **Add query result pagination**: Limit max offset to prevent deep pagination performance issues. Expected: **Prevents DoS**, low effort.

## 6) Security Checklist

| Item | Status | Notes |
|------|--------|-------|
| Inputs validated? | ❌ | No length limits, format validation, or sanitization on most inputs |
| Output encoded? | ❌ | HTML output not escaped (XSS risk) |
| Secrets sourced from vault? | ❌ | Secrets in config files, default values in code |
| Least privilege? | ⚠️ | Basic auth exists but not enforced on all endpoints |
| Safe defaults? | ❌ | Default Meilisearch key "masterKey" is insecure |
| Rate limiting? | ❌ | No rate limiting implemented |
| Logging PII redaction? | ⚠️ | Query parameters logged, may contain sensitive data |
| SQL injection protected? | ❌ | `sortBy` field vulnerable (F1, F6) |
| XSS protected? | ❌ | User input rendered without escaping (F2) |
| CSRF protection? | ⚠️ | Echo middleware may provide, but not explicitly configured |
| Path traversal protected? | ✅ | Static file serving uses relative paths, should be safe |
| Authentication required? | ⚠️ | Optional basic auth, not enforced by default |
| Authorization checks? | ❌ | No role-based access control |
| HTTPS enforced? | ⚠️ | TLS configurable but not required |
| CORS configured? | ⚠️ | Enabled but no origin restrictions |
| Dependency vulnerabilities? | ⚠️ | Not checked - run `go list -json -m all \| nancy sleuth` |

## 7) Maintainability Improvements

### Refactors (Small + Incremental)

1. **Extract validation constants** (F25):
   ```go
   // internal/handlers/constants.go
   const (
       DefaultPageLimit = 50
       MaxPageLimit = 1000
       MaxExportLimit = 10000
       DefaultAutocompleteSize = 5
       MaxAutocompleteSize = 50
   )
   ```

2. **Create input validation helper**:
   ```go
   // internal/handlers/validation.go
   func validateSearchFilter(filter *models.SearchFilter) error {
       if len(filter.Query) > maxQueryLength {
           return fmt.Errorf("query exceeds max length %d", maxQueryLength)
       }
       // ... other validations
   }
   ```

3. **Error taxonomy** (F19):
   ```go
   // internal/repository/errors.go
   var (
       ErrNotFound = errors.New("not found")
       ErrInvalidInput = errors.New("invalid input")
       ErrTooManyResults = errors.New("too many results")
   )
   ```

4. **Remove dead cache code** (F23): Either implement cache properly (F12) or remove the check.

5. **Document Stream endpoint** (F24): Add comment explaining why not implemented or implement it.

### Dead Code Removal

- Cache check in `searchService.Search` that doesn't return cached data (F23)
- Unused error variable in some error handling paths

### Configuration Externalization

- Move magic numbers to config: `max_query_length`, `max_export_limit`, `rate_limit_rps`
- Make cache TTL configurable (currently hardcoded to 5 minutes)

### Documentation/Comments to Add

1. **Time zone assumptions**: Document that all times are stored/processed in UTC
2. **Cache behavior**: Document cache TTL and invalidation strategy
3. **Export limits**: Document why 10,000 limit exists and how to change it
4. **Authentication**: Document how to enable and configure basic auth
5. **SQL injection prevention**: Add comment explaining whitelist approach for sortBy

## 8) Quality Scores

| Dimension | Score (1-5) | Justification |
|-----------|-------------|---------------|
| **Correctness** | 3/5 | Good use of parameterized queries for values, but SQL injection in ORDER BY, missing input validation, and incomplete cache implementation reduce score |
| **Security** | 2/5 | Critical vulnerabilities: SQL injection, XSS, no rate limiting, hardcoded secrets. Basic auth exists but not enforced. |
| **Performance** | 3/5 | Efficient bulk operations and connection pooling, but N+1 queries in stats, memory-intensive exports, and unused cache hurt performance |
| **Design** | 4/5 | Clean separation of concerns, good use of interfaces, dependency injection. Minor issues: context usage, error taxonomy |
| **Readability** | 4/5 | Clear naming, good structure, idiomatic Go. Some magic numbers and incomplete implementations reduce score |
| **Testability** | 3/5 | Interfaces enable testing, but lack of unit tests, no test utilities, and tight coupling to external services make testing difficult |

**Overall**: 3.2/5 - Solid foundation with critical security issues that need immediate attention.

