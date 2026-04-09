# Code Review: Expressions India Backend

## Critical Bugs (Will Crash in Production)

### 1. Nil pointer dereference on `RegistrationURL` - CRASH ✅ FIXED
**Files:** `internal/event/service.go:188`, `internal/event/service.go:313`

**Fix Applied:** Added nil checks before dereferencing `*data.RegistrationURL` and `*newData.RegistrationURL`.

---

### 2. Division by zero in pagination - CRASH ✅ FIXED
**Files:** `internal/event/controller.go:61`, `internal/article/controller.go:36,54,72`

**Fix Applied:** Added `utils.SafeTotalPages(total, limit)` helper in `utils/response.go` that returns 0 if limit ≤ 0. Updated all controllers to use this helper instead of direct `math.Ceil` calls.

---

### 3. `DB_FRESH_START` drops all tables on every restart
**File:** `cmd/api/server.go:47-61`

If `DB_FRESH_START=true` is left in the production `.env`, every server restart wipes the entire database. There is no environment guard - this runs regardless of `APP_ENV`.

**Fix:** Either remove this flag entirely, or gate it behind `APP_ENV != "production"`.

---

### 4. Database connection string logged in production ✅ FIXED
**File:** `internal/storage/db.go:36`

**Fix Applied:** Removed the `log.Println(dns)` line entirely.

---

### 5. `UploadLocal` ignores upload error ✅ FIXED
**File:** `internal/storage/s3.go:75`

**Fix Applied:** Added error check: `if err != nil { return "", "", err }` after the upload call.

---

## Major Bugs (Wrong Behavior, Data Loss)

### 6. No authentication on admin endpoints
**File:** `cmd/api/server.go:138`

The entire `/admin` route group has no authentication middleware. Anyone can create, update, and delete events, articles, podcasts, journals, enquiries, and audience data.

**Fix:** Add JWT or session-based auth middleware to the `/admin` group.

---

### 7. `DeletedDocumentIds` and `DeletedMediaIds` are ignored during event update ✅ FIXED
**File:** `internal/event/service.go:279-424`

**Fix Applied:** Added delete-from-DB-then-S3 loops for both `newData.DeletedMediaIds` and `newData.DeletedDocumentIds`, matching the existing `DeletedPromotionalMediaIds` pattern.

---

### 8. Event list query applies filters twice (including LIMIT/OFFSET on count) ✅ FIXED
**File:** `internal/event/service.go:240-253`

**Fix Applied:** Created `ApplyEventListBaseFilters` in `utils/filter.go` (excludes LIMIT/OFFSET). Count query now uses base filters, data query uses full filters with pagination.

---

### 9. Enquiry model has no `BeforeCreate` hook - relies on manual UUID ✅ FIXED
**File:** `internal/models/enquiry.go`

**Fix Applied:** Added UUID `BeforeCreate` hook. Removed manual UUID generation from `enquiry/service.go:51`.

---

### 10. Podcast has no `BeforeCreate` hook either ✅ FIXED
**File:** `internal/models/podcast.go`

**Fix Applied:** Added UUID `BeforeCreate` hook. Removed manual UUID generation from `podcast/service.go:18`.

---

### 11. No input validation on public enquiry endpoint ✅ FIXED
**File:** `internal/dto/enquiry.go:14-20`

**Fix Applied:** Added `binding:"required"` on all fields and `binding:"required,email"` on the email field.

---

### 12. Podcast search uses `LIKE` instead of `ILIKE` in production ✅ FIXED
**File:** `internal/podcast/service.go:64`

**Fix Applied:** Replaced conditional `LIKE`/`ILIKE` with cross-db compatible `LOWER(title) LIKE LOWER(?)` in:
- `utils/filter.go` (BySearch scope)
- `podcast/service.go` 
- `article/service.go` (both title and category)

---

### 13. Article search also uses `LIKE` instead of `ILIKE` in production
**File:** `internal/article/service.go:32,34`

Same issue as podcast search.

---

### 14. Latest activity only returns events, not other content types
**File:** `internal/latest-activity/service.go:17`

```go
query := `SELECT id, title, start_date, end_date, 'event' AS type, created_at FROM events ORDER BY created_at DESC LIMIT 5`
```
This hardcodes `'event'` type and only queries the events table. Articles, podcasts, and journals are never included despite the endpoint being called "latest activity".

**Fix:** Use a `UNION ALL` query across content types, or query each table and merge in Go.

---

### 15. `godotenv.Load()` failure kills the server even in production ✅ FIXED
**File:** `cmd/api/main.go:11-13`

**Fix Applied:** Now checks if `.env` file exists before attempting to load. Only errors if the file is present but unparseable. In production (no `.env` file), startup is unaffected.

---

## Inconsistencies & Bad Practices

### 16. CORS allows everything when `ALLOWED_ORIGINS` is unset
**File:** `cmd/api/server.go:114-116`

```go
} else {
    r.Use(cors.Default())  // allows ALL origins
}
```
If `ALLOWED_ORIGINS` isn't configured in production, the server accepts requests from any origin.

---

### 17. Error messages leak internal details ✅ FIXED
**Files:** Multiple controllers

**Fix Applied:** Added `utils.FailInternal(c, code, clientMsg, err)` helper that logs full error server-side via `log.Printf` and returns only generic client message. Updated:
- `event/controller.go`
- `article/controller.go`
- `podcast/controller.go`
- `course/controller.go`

---

### 18. Service passed by value instead of pointer ✅ FIXED
**File:** `cmd/api/server.go:82`

**Fix Applied:**
- `event/service.go`: Changed `(s Service)` → `(s *Service)` on `GetUpcomingEvents` and `GetPastEvents`
- `event/controller.go`: Changed `service Service` → `service *Service`, updated `NewController` signature
- `cmd/api/server.go`: Removed pointer dereference — `event.NewService(db, s3)` already returns `*Service`

---

### 19. No request body size limit
**File:** `cmd/api/server.go`

Gin's default `MaxMultipartMemory` is 32MB but there's no explicit limit set. File upload endpoints can be abused to exhaust server memory.

**Fix:** Set `r.MaxMultipartMemory` to a reasonable value and add middleware to limit request body size.

---

### 20. No rate limiting on public endpoints
**File:** `cmd/api/server.go:174`

The public `/api/enquiry` POST endpoint has no rate limiting. Bots can flood the database with spam enquiries.

---

### 21. `SEEDING_DB` and `CLEAR_WHILE_SEEDING` env vars are unused
**File:** `.env`

These are defined in `.env` but never checked in the code. `DB_FRESH_START` handles dropping, but these are dead config that could confuse operators.

---

### 22. No graceful shutdown
**File:** `cmd/api/main.go:28`

```go
server.r.Run(":" + port)
```
`gin.Run()` doesn't handle OS signals. In-flight requests and S3 uploads will be terminated abruptly on deploy/restart.

**Fix:** Use `http.Server` with `Shutdown()` and signal handling.

---

### 23. S3 bucket name is hardcoded
**File:** `internal/storage/s3.go:69,92,109`

`"expressions-india"` is hardcoded in three places. Should be an environment variable for flexibility across environments.

---

### 24. Deprecated AWS endpoint resolver API
**File:** `internal/storage/s3.go:27-32`

`aws.EndpointResolverWithOptionsFunc` is deprecated in the AWS SDK v2. Use `BaseEndpoint` on service options instead.

---

### 25. `sqlDB` error ignored
**File:** `internal/storage/db.go:55`

```go
sqlDB, _ := db.DB()  // error silently ignored
```
If this fails, subsequent `Set*` calls will panic on a nil `sqlDB`.

---

## Summary

| Severity | Count |
|----------|-------|
| Critical (crash/data loss) | 5 |
| Major (wrong behavior) | 10 |
| Bad practices | 10 |

**Priority order for fixes:**
1. Remove `DB_FRESH_START` production guard (data loss risk)
2. Fix `UploadLocal` nil pointer crash
3. Add authentication to admin routes
4. Fix pagination division by zero
5. Fix count query applying LIMIT
6. Handle `DeletedMediaIds`/`DeletedDocumentIds` in event update
7. Stop logging connection strings
8. Make `.env` loading optional
9. Add input validation on enquiry endpoint
10. Fix case-sensitive search in production
