# Implementation Notes

Changes made in response to `BACKEND_REVIEW.md`, plus newly discovered issues.

---

## Changes Implemented

### 1. Critical Bug: Media ID vs S3 Key Mismatch (NEW — not in review)

**File:** `internal/models/media.go`

`Media.BeforeCreate` unconditionally overwrote `m.ID` with a new UUID even when an ID was already set. Since all upload helpers set `Media{ID: s3Key, ...}` (where `s3Key` is the S3 object key), the hook replaced the S3 key with a random UUID. This caused:
- DB stores a UUID that has no matching object in S3
- Delete operations call `s3.Delete(media.ID)` which tries to delete a non-existent key → S3 error
- S3 objects are permanently orphaned and unbillable

**Fix:** Added `if m.ID == ""` guard, mirroring how `Article.BeforeCreate` and `Journal.BeforeCreate` already handled it.

Same guard added to `Event.BeforeCreate` (though less critical since event IDs aren't S3 keys).

---

### 2. N+1 Queries → IN Clause

**Files:** `internal/event/service.go`, `internal/article/service.go`, `internal/podcast/service.go`

Replaced per-row `SELECT ... WHERE name = ?` loops with a single `WHERE name IN ?` query.

| Before | After |
|--------|-------|
| N queries for N audiences | 1 query for N audiences |

Affected functions: `getAudience()`, `resolveAudiences()`, `CreatePodcast()`.

---

### 3. Sequential → Concurrent File Uploads

**Files:** `internal/event/service.go`, `internal/article/service.go`

Replaced sequential upload loops with `errgroup`-based concurrent uploads. Files in a single request now upload in parallel.

New helper: `appendUploadedMedia()` (event) and `uploadMediaFiles()` (article).

If any upload fails, the helper cleans up already-uploaded S3 files before returning the error.

---

### 4. Swallowed Errors on File Open

**Files:** `internal/event/service.go`, `internal/article/service.go`

Removed `log.Println(err); continue` patterns that silently skipped files that failed to open. These are now propagated as errors — the request fails fast and the client is notified.

---

### 5. S3/DB Consistency on CreateEvent Failure

**File:** `internal/event/service.go`

Added deferred cleanup in `CreateEvent`. If `db.Create` fails after S3 uploads have succeeded:
- The thumbnail Media record is deleted from the DB
- All uploaded S3 keys (thumbnail + promotional media + documents + media) are deleted

Uses a named return value + `defer` pattern to detect failure.

Same pattern applied to `CreateArticle` and `UpdateArticle` in `internal/article/service.go`.

**UpdateEvent** is also covered: new thumbnail and new media S3 keys are tracked; if `db.Save` fails, the deferred cleanup deletes the new thumbnail from DB+S3 and deletes all new media S3 files. For *deletions* of old media (old thumbnail, deleted promotional media IDs), the DB delete is hard-fail but the S3 delete is best-effort (logged, not fatal) — DB is the source of truth.

**DeleteEvent** S3 deletes are now best-effort (logged, not fatal). Previously a failed S3 delete would return an error to the client while the DB record was already gone, causing inconsistency on retry.

---

### 6. ID Generation Hook Collision

**File:** `internal/event/service.go`

Removed manual `EventID := uuid.Must(uuid.NewV7()).String()` from `CreateEvent`. The `BeforeCreate` hook (now guarded) handles ID generation. No behavior change — this just removes the redundant UUID that was silently discarded.

---

### 7. `CreatePodcast` Return Type `any` → `error`

**File:** `internal/podcast/service.go`

Changed return type from `any` to `error`. The old type was non-idiomatic Go and could cause subtle interface-nil bugs if the code path changed in future.

---

### 8. Proper 404 vs 500 in GetById Controllers

**Files:** `internal/event/controller.go`, `internal/journal/controller.go`, `internal/podcast/controller.go`, `internal/article/controller.go`, `internal/enquiry/controller.go`

All `GetById` handlers now check `errors.Is(err, gorm.ErrRecordNotFound)`:
- Record not found → 404 `NOT_FOUND`
- Any other DB error → 500 `FETCH_ERROR`

Previously all errors returned 500.

Also fixed `podcast/controller.go::GetById` which had a dead `if podcast == nil` check after the error path.

---

### 9. Hardcoded Port

**File:** `cmd/api/main.go`

Port now reads from `PORT` environment variable, defaulting to `8000` if unset.

---

### 10. CORS Restricted in Production

**File:** `cmd/api/server.go`

If `ALLOWED_ORIGINS` env var is set (comma-separated list), uses a restricted CORS config. Falls back to `cors.Default()` (allow all) when unset, preserving dev behavior.

Example: `ALLOWED_ORIGINS=https://expressionsindia.com,https://admin.expressionsindia.com`

---

### 11. DB Connection Retry

**File:** `internal/storage/db.go`

Added 5-attempt retry loop with exponential backoff (2s, 4s, 6s, 8s) before calling `log.Fatal`. Prevents startup crashes when the API container starts before the DB container in docker-compose.

---

## Frontend Changes Required

### A. Handle 404 on Detail Pages (important)

Before these fixes, all `GetById` API calls returned `500` for missing records. Now they return `404`. The frontend's `parseApiResponse` in `src/lib/api.ts` throws the server's error message regardless of HTTP status, so basic error display still works.

However, the frontend detail pages (event/$id, journal/$id, podcast/$id, article/$id, enquiry/$id) currently show a generic error state for both "record not found" and "server error". Consider differentiating:

```typescript
// In fetch functions, surface the HTTP status so the UI can show a proper 404 page
export async function parseApiResponse<T>(response: Response, schema: z.ZodType<T>) {
  const json = await response.json();
  if (!json.success) {
    const err = new Error(json.error?.message ?? "Request failed");
    (err as any).status = response.status; // attach status for callers to check
    throw err;
  }
  // ...
}
```

Then in detail page components, check `error.status === 404` to show a "Not Found" page vs a generic error.

**Files to update:**
- `src/lib/api.ts` — attach HTTP status to thrown errors
- `src/routes/admin/event/$id.tsx`
- `src/routes/admin/journal/$id.tsx`
- `src/routes/admin/podcast/$id.tsx`
- `src/routes/admin/article/$id.tsx`
- `src/routes/admin/enquiry/$id.tsx`

### B. No Changes Required For

- Concurrent uploads: same API contract, just faster
- N+1 query fix: same API contract, just faster
- Port / CORS: server-side config only
- DB retry: server-side only
- `CreatePodcast` type fix: same behavior from client perspective
