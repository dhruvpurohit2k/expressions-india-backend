# Backend Codebase Review: Bugs and Improvements

This document lists the bugs, anti-patterns, and areas for improvement found during the review of the Go backend (`/cmd/api`, `/internal/storage`, `/internal/models`, `/internal/event`, `/internal/podcast`, `/internal/journal`, `/internal/article`).

## 🚨 Critical Bugs & Architecture Issues

### 1. Lack of Database Transactions
In most domain services (`article`, `event`, `journal`, `podcast`), operations that touch multiple tables or update records and perform external network calls (S3 uploads/deletions) lack atomic transactions. 
- **Effect**: If an S3 upload succeeds but the subsequent `db.Create()` fails, the file is orphaned in S3. Conversely, if `db.Delete()` for a record succeeds but the subsequent S3 deletion fails (or vice versa), the database is left in an inconsistent state.
- **Fix**: Wrap multi-step mutations in `gorm.Transaction(func(tx *gorm.DB) error { ... })`.

### 2. ID Generation Hook Collision
In `internal/models/event.go`, there is a `BeforeCreate` hook that generates a new UUID (`newID, err := uuid.NewV7()`). However, in `internal/event/service.go`, `CreateEvent` manually assigns an ID before calling `db.Create(&newEvent)`.
- **Effect**: The explicitly generated `uuid.NewV7()` in the service layer is overwritten by the GORM hook. This is mostly harmless but wastes CPU cycles and creates confusion about where the ID comes from. This pattern might exist in other models as well. 

## ⚠️ Performance & Scaling Bottlenecks

### 1. N+1 Query Problem in Relationships
In `article/service.go` (`resolveAudiences`), `event/service.go` (`getAudience`), and `podcast/service.go` (`CreatePodcast`), the code iterates over an array of strings (e.g., audience names) and executes a single `First()` query for each iteration.
- **Effect**: Causes an N+1 query performance bottleneck. Inserting a record with 10 audiences results in 10 separate SQL `SELECT` statements.
- **Fix**: Replace iterative queries with an `IN` clause: `db.Where("name IN ?", names).Find(&audiences)`.

### 2. Sequential File Uploads
In `event/service.go` (`appendPromotionalMedia`, `appendDocument`, `appendMedia`) and `article/service.go` (`uploadMediaFiles`), multiple files are iterated over and uploaded to S3 sequentially.
- **Effect**: Time to upload scales linearly. If a user uploads 5 files, they have to wait for the sum of all 5 network roundtrips.
- **Fix**: Use Go routines and a `sync.WaitGroup` (or an `errgroup`) to upload multiple files concurrently.

### 3. Inefficient Pagination Parsing
For functions returning lists (e.g., `GetEventList`, `GetUpcomingEvents`), the `total` count is calculated by running a `Count(&total)` query *before* executing the paginated `Find()`. While `Count()` is generally okay, applying complex joins or iterating through entire tables for large tables could impact performance.
- **Note**: Ensure indexes exist for `.Where("status = ?", "completed")` or `.Where("start_date >= ?", time.Now())` to prevent full table scans.

## 🛠️ Code Maintenance & Idiomatic Go

### 1. Hardcoded Configuration
- **`cmd/api/main.go`**: The server port is hardcoded to `:8000`. This should be parsed from `os.Getenv("PORT")` to allow deployment flexibility.
- **`cmd/api/server.go`**: The CORS middleware `r.Use(cors.Default())` is extremely permissive. For a production app, the allowed origins and headers should be restricted.

### 2. Error Handling Swallowed
- During file processing `f, err := file.Open()` loop iterations, if an error happens, the code logs it (`log.Println(err)`) and uses `continue`. The client caller receives no indication that one of their files failed to upload, masking potential partial failures. 

### 3. Server Initialization Logic
- The `AutoMigrate` function call in `initServer` takes a large hardcoded list of `&models.XYZ{}`. If a developer forgets to add the new model there after creating it, DB tables won't synchronize. Moving models to a centralized registry array could improve maintainability.
- The DB initialization connects directly without a solid retry mechanism. If the database container boots slower than the API container in a docker-compose setup, the backend might crash.

## Suggested Next Steps
1. Introduce `<model>_repository.go` patterns to encapsulate GORM usage and properly handle transactions to decouple the DB from `services`.
2. Standardize S3 error handling using a generic wrapper.
3. Update `GET` endpoints with simple N+1 replacements `Where("id IN ?", ids)`. 
