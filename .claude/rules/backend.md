# Backend & Database Standards

> Applies to: `backend/**` (Go code, `backend/sql/queries/`, `backend/sql/migrations/`).

## SQL & Migrations (Goose/SQLC)
- **Migrations:** New tables MUST include standard columns: `id` (UUID), `created_at`, and `updated_at`.
- **Queries:** When writing `.sql` files in `backend/sql/queries/`, use the `:one` or `:many` or `:exec` annotations correctly. 
- **Type Sync:** Ensure that any new SQL columns are reflected in the `backend/sql/sqlc.yaml` overrides and that `emit_pointers_for_null_types: true` is respected.
- **Sequential Migrations:** Always check backend/sql/migrations for the highest prefix (e.g., 003) and ensure the new migration uses the next increment (004).
- **Context Passing:** Always pass the Gin request context (c.Request.Context()) to SQLC querier methods to ensure proper cancellation and timeout propagation.

## Type Management
- **Native Types:** Always map DB types to native Go types (e.g., `uuid.UUID`, `*time.Time`)
- **API Layer:** Do not use `sql.NullString` or `pgtype` wrappers in the API layer; stick to native Go pointers for clean JSON

## Model Architecture (internal/model)
- **Strict Isolation:** Handlers MUST NOT return SQLC-generated structs; they must only return structs from `backend/internal/model`.
- **Entity Isolation:** For every `sqlc` struct, create or extend a dedicated `[entity].go` in `backend/internal/model/`.
- **Mapping Implementation:**
    - Implement `func MapDb[Entity]To[Entity](db[Entity] db.[Entity]) [Entity]` within the entity file (e.g. `MapDbGameModeToGameMode`). For slices, add `MapDb[Entities]To[Entities]`; for sqlc row types, name after the row (e.g. `MapDbGetGamesForUserRowToUserGame`).
    - Mapping tests go in `model_mapping_test.go` (`package model_test`). Entities with non-mapping logic (e.g. `auth.go`) get their own `[entity]_test.go`.
    - Explicitly handle `nil` checks for pointers to ensure the API model receives safe values.
    - Ensure JSON tags match the `openapi.yaml` snake_case naming convention.

## Store Error Contracts
- **Sentinel Errors for Domain Outcomes:** Domain, authorization, and validation outcomes that map to non-500 HTTP responses MUST be represented with exported sentinel errors in the store layer (e.g., `ErrForbidden`, `ErrInvalidCursor`) rather than only string-matching wrapped DB errors.
- **Handler-Owned HTTP Mapping:** Handlers map sentinel errors to HTTP status codes and client-facing messages. Store methods should stay transport-agnostic.
- **Wrapped Operational Errors:** Unexpected DB/query failures MUST be wrapped with operation context using `%w` (e.g., `fmt.Errorf("create resource: %w", err)`) to preserve root cause and call-site context.
    
## Authorization
- Authentication (is the user logged in?) is handled by JWT middleware -- do not re-check token validity inside handlers.
- Authorization (can this user access this resource?) MUST be checked explicitly in the handler before any DB write or sensitive read.
- To retrieve the authenticated user's ID, call the unexported helper `userIDFromContext(c)` (lives in `backend/internal/handler/`):
```go
  userID, ok := userIDFromContext(c)
  if !ok {
      return // response and log already written by the helper
  }
```
- Never read a user ID from request params and trust it as the actor identity without first obtaining the verified ID from `userIDFromContext`.
- If a resource does not belong to the requesting user or another authorized actor, return 403.
- Do NOT log or write an additional response after a `!ok` return -- `userIDFromContext` has already done both.

## Middleware
- All middleware lives in `backend/internal/middleware/`.
- The standard chain (applied in `main.go`) is: `RequestID → Recovery → RequestLogger → OriginVerify → CORS → Auth`.
- `OriginVerify` is a no-op when `ORIGIN_VERIFY_SECRET` is unset (local Compose). When set, it requires header `X-Origin-Verify` and returns 403 on mismatch; `GET /health` is always exempt.
- Auth middleware (`middleware.Auth(jwtSecretBytes)`) is applied only to the `protected` route group, not globally — public routes (`/auth/*`, `/health`) are intentionally unauthenticated (Auth-wise; origin-verify may still apply in production).
- Route-level middleware (e.g. role checks) is applied at the route group, not inside handlers.
- Do NOT add `gin.Logger()` or `gin.Recovery()` — the project uses custom replacements for both. `gin.New()` is used intentionally over `gin.Default()` for this reason.

## Structured Logging (slog)

### Envelope
All log entries use the standard `log/slog` JSON envelope: `time` (RFC3339), `level`, `msg`, plus structured fields. Never construct log messages by string concatenation — always use slog key-value pairs.

### Context-Based Logging
- Always use context-aware variants (`slog.InfoContext`, `slog.WarnContext`, `slog.ErrorContext`) with `c.Request.Context()` -- never call the plain `slog.Info`/`slog.Warn`/`slog.Error` variants inside handlers or middleware, as they lose the `request_id` and other context-propagated fields.
- Startup and fatal logs (before a request context exists) are the only valid use of the plain `slog` variants.

### Request Completion (middleware-owned)
- Request completion logging is handled entirely by the logging middleware -- do NOT add per-endpoint completion logs in handlers.
- The middleware emits `msg: "request completed"` with fields: `method`, `path`, `status`, `duration_ms`, `client_ip`, `route` (omitted for unmatched paths), `request_id`.
- Level is derived from status: `INFO` for <400, `WARN` for 400–499, `ERROR` for >=500.

### Panic Recovery (middleware-owned)
- Panic recovery logging is handled by the recovery middleware — do NOT wrap handlers in manual `recover()` calls.
- Emits `msg: "panic recovered"` at `ERROR` with fields: `panic`, `method`, `path`, `client_ip`, `stack`, `route`, `request_id`.

### Handler Logging (what handlers own)
Handlers are only responsible for logging events the middleware cannot observe — i.e. business logic outcomes, not request lifecycle.

| Situation | Level | Required fields |
|---|---|---|
| External service / DB error causing 5xx | `ERROR` | `error`, plus relevant entity IDs |
| Expected client error (e.g. not found, conflict) | `WARN` | `error`, plus relevant entity IDs |
| Significant business event (e.g. user created) | `INFO` | relevant entity IDs |

- Do NOT log the status code or duration in handlers; the middleware owns those.

## API Error Responses
- **Standard JSON Error Envelope:** For JSON error responses, use a consistent client-facing shape:
```json
{
  "status": "error",
  "message": "Human-readable explanation"
}
```
- **Scope:** Apply this envelope to 4xx/5xx JSON responses. Empty-body responses that intentionally use status-only semantics (such as some 401/204 cases) are allowed when appropriate.
- **Separation of concerns:** Logging standards define observability behavior; error envelope standards define API response contracts. Keep both consistent.

### Startup / Fatal Logs
- Startup events (e.g. `msg: "API listening"`) use `INFO` with relevant config fields (e.g. `port`).
- Fatal errors before server start (e.g. migration failure) use `ERROR` with an `error` field, followed by `os.Exit(1)` via `fatalExit`.
- Never use `log.Fatal` or `fmt.Println` in the server or `internal/` application code — always go through `slog`.
- **Exceptions:** Test helpers (e.g. `backend/internal/test_util`) may use `log.Fatal` for setup failures, and standalone CLI scripts under `backend/cmd/scripts/` may print to stdout with `fmt`.

## Request Validation
- Bind request bodies using `c.ShouldBindJSON(&req)` into a private request struct defined at the top of the handler file.
- On bind failure (malformed JSON or wrong value types), return 400 and abort:
```go
  c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "Improper json or json value types",
  })
  return
```
- `c.ShouldBindJSON` handles structural validation only — it does NOT validate business rules or field constraints. Any additional validation (e.g. required fields, value ranges) must be done manually after binding, before any DB call.

## Pagination
- Cursor-based keyset pagination is the project standard. Do not implement page/offset pagination for any new list endpoint.
- New list endpoints MUST follow the same contract: keyset cursor, `pageSize + 1` fetch, opaque `next_cursor`, stable composite ordering with `id` as tie-breaker, and `{ <collection>, has_more, next_cursor }` response shape.

### Page Size
- Fixed at 20 items per page, defined as a named constant in the store layer (e.g. `dashboardEventsPageSize = 20`). Do not hardcode `20` inline.
- The DB query MUST fetch `pageSize + 1` rows to detect whether another page exists--never use a `COUNT` query for this.

### Cursor Format
- Cursors are opaque base64url-encoded JSON objects containing the ordering key fields of the last item on the current page.
- For `(created_at ASC, id ASC)` or `(created_at DESC, id DESC)` style ordering: `{ "created_at": <timestamp>, "id": <uuid> }` (the cursor contents are the same either way; only the query's comparison operator changes)
- Other endpoints must encode whichever fields constitute their stable ordering key using the same base64url-encoded JSON pattern.

### Ordering & Filtering
- Results MUST be sorted by a composite key that includes `id` as a tiebreaker to guarantee stable ordering when multiple rows share the primary sort value.
- Cursor filtering uses a row-value comparison: `WHERE (sort_col, id) > (cursorVal, cursorID)`
- **Direction:** The comparison operator must match the sort direction. Use `>` for ascending (`ORDER BY sort_col ASC, id ASC`, oldest-first) and `<` for descending (`ORDER BY sort_col DESC, id DESC`, newest-first). Both columns must sort in the same direction for the row-value comparison to be correct; mixed directions (e.g. `sort_col DESC, id ASC`) require an expanded `sort_col < x OR (sort_col = x AND id > y)` predicate instead.
- Additional filters (date ranges, timezone, etc.) are applied in the same query--cursor pagination slices the already-filtered result set, it does not replace filtering.

### has_more / next_cursor Logic
- If `len(rows) > pageSize`, set `has_more = true`, trim to `rows[:pageSize]` and encode `next_cursor` from the last row in the trimmed slice.
- If `len(rows) <= pageSize`, set `has_more = false` and `next_cursor = ""`.
- Cursor encoding: base64url-encode a JSON object containing the ordering key fields of the last row (e.g. `{"created_at": <timestamp>, "id": <uuid>}`).
- Canonical store logic example:
```go
rows, err := q.ListResources(ctx, params) // params include LimitCount = pageSize + 1
hasMore := len(rows) > pageSize
if hasMore {
    rows = rows[:pageSize]
}
```

### Response Envelope
All paginated endpoints MUST use this shape:
```json
{
  "<entity_collection_key>": [...],
  "has_more": true,
  "next_cursor": ""
}
```
- Canonical handler response assembly:
```go
response := Response{
    Items:      items,
    HasMore:    hasMore,
    NextCursor: nextCursor,
}
c.JSON(http.StatusOK, response)
```

## Database Transactions
- For operations requiring multiple queries to succeed atomically, use:
```go
  err := store.WithTx(ctx, func(tx Store) error {
      // use tx for all queries in this operation
      return nil
  })
```
- `WithTx` handles `Begin`, deferred `Rollback` on error, and `Commit` on success--never manage transactions manually in handlers.
- The `tx Store` passed to the callback is a transaction-scoped store--use it exclusively for all queries inside the callback. Do not close over the outer store and mix scopes within the same operation.
- Return any error immediately from the callback to trigger rollback; do not suppress errors to "continue anyway".

### Testing
- Handler tests use `NewPostgresStoreFromTx`, where `pool == nil` and `WithTx` runs inline — the outer test transaction controls rollback, so there is no real commit. Do not add manual transaction logic in tests to compensate for this; it is intentional.

## Soft Deletes
- This project does NOT use soft deletes. Records are hard-deleted.
- If a future entity requires audit history, use an append-only audit log/table instead of adding `deleted_at`.

## OpenAPI Sync
- Whenever an endpoint's logic, parameters, or responses change, you MUST update `backend/openapi.yaml`.
