# Distributed Job Scheduler — Project Information

This document provides a complete, machine-readable overview of the repository so another LLM (or developer) can obtain a perfect understanding of the system: what exists, how it works, assumptions made, known issues, and recommended next steps.

---

## High-level Description

The Distributed Job Scheduler is a small, self-contained Go-based system with a React/Vite frontend. It implements a simple job model, a priority-based scheduler, and a worker runtime that supports per-worker thread pools and multi-threaded job execution (via splitting work into chunks). It persists completed jobs and metrics to PostgreSQL and caches job state in Redis. The API uses Gin to expose endpoints for submitting jobs, querying active jobs, and fetching historical jobs from the DB.

Primary goals accomplished so far:
- Submit jobs via HTTP API
- In-memory job queue with priority ordering
- Scheduler assigns jobs to workers based on thread availability and priority
- Worker-side thread pool and multi-threaded execution for jobs that support chunking
- Persist completed jobs and simple metrics to PostgreSQL
- Cache job state in Redis for faster job lookups
- Frontend can submit and fetch jobs, and merge in-memory + historical results

---

## Repo layout (concise)

- `cmd/`
	- `api.go` — main HTTP server, env loading, DB/Redis init, job registry, worker creation, scheduler start, endpoints: `POST /jobs`, `GET /jobs`, `GET /jobs/:id`, `GET /db/jobs`, `GET /db/jobs/:id`.
	- `distributed-job-scheduler/main.go` — placeholder CLI that prints startup.
- `internal/job/` — job model and job-specific logic
	- `job.go` — `Job` struct, types (`Status`, `JobType`), `NewJob`, `Execute`, and `ExecuteChunk` implementations; uses interface{} for `Payload`/`Result`.
	- `payloads.go` — payload structs (`AddNumbersPayload`, `LargeArraySumPayload`, etc.).
	- `results.go` — result structs.
	- `utils.go` — helpers: `reverse()`, `ResizeImage()` (simulated), global `resultMutex`, `addPartialSum()`.
- `internal/scheduler/` — scheduler implementation
	- `scheduler.go` — priority queue (heap) and `Scheduler` that runs worker loops, picks jobs by priority and thread availability, supports single-thread fallback when no worker can satisfy `ThreadDemand`.
- `internal/worker/` — worker runtime
	- `worker.go` — `Worker` struct with `JobQueue`, `FreeThreads` channel representing thread pool, `Start()`, `processJob()`, `AvailableThreads()`, `Stop()`.
- `db/schema.sql` — PostgreSQL schema for `jobs`, `workers`, `job_logs`, `job_metrics`.
- `frontend/` — React + Vite frontend
	- `src/services/api.ts` — client that calls backend endpoints, merges current and historical jobs, performs dedupe and sorting; exposes `submitJob()` and `fetchJobs()`.
	- UI components under `src/components`, pages under `src/pages` (not exhaustively listed here but present).
- Tests: unit tests for `internal/job`, `internal/scheduler`, `internal/worker` and some integration-style tests under `cmd/`.

---

## Key files and responsibilities (detailed)

- `cmd/api.go`:
	- Loads `.env` via `godotenv` (non-fatal if missing).
	- Connects to Postgres using `pgxpool` and to Redis using `redis/go-redis`.
	- Registers job factories in `jobRegistry` for incoming job `type` values (currently `add_numbers` and `large_array_sum`). Factories convert JSON payload maps to typed payload structs and call `job.NewJob`.
	- Creates worker instances from env vars and `NewWorkerWithQueueSize`, starts them, creates `Scheduler`, and calls `sched.Run()`.
	- `POST /jobs` flow: bind JSON to `SubmitJobRequest`, create job via registry, store pointer in `jobs` map, cache job JSON in Redis under `job:<id>`, submit to scheduler, spin a goroutine that polls the job until it's `Completed` and then calls `insertJobToDB` to persist to Postgres.
	- `GET /jobs` returns in-memory jobs; `GET /db/jobs` returns rows from DB (historical). `GET /jobs/:id` attempts Redis first, then in-memory.
	- `insertJobToDB` marshals `Result` JSON and upserts into `jobs` table, and inserts three job metric rows (`queue_time`, `execution_time`, `total_time`) into `job_metrics`.

- `internal/job/job.go`:
	- `Job` struct fields: `ID`, `Name`, `Type` (`JobType`), `Status`, `Priority`, `Payload` (interface{}), `Result` (interface{}), `CreatedAt`, `StartedAt`, `CompletedAt`, `ThreadDemand`.
	- `Execute()` handles multiple job types (AddNumbers, ReverseString, ResizeImage, LargeArraySum). It sets `Status = Pending` at top (bug — see Known Issues) then computes result and sets `Status = Completed` and `CompletedAt = time.Now()`.
	- `ExecuteChunk(threadID, totalThreads)` only implemented for `LargeArraySumJob`: computes local sum for chunk and calls `addPartialSum` to combine partial results.

- `internal/job/utils.go`:
	- `reverse` helper.
	- `ResizeImage` simulates processing by sleeping and returning a formatted URL.
	- `resultMutex` is a package-level `sync.Mutex` used by `addPartialSum` to protect updates to `j.Result`.

- `internal/scheduler/scheduler.go`:
	- Implements `JobQueue` as heap-based priority queue; `Less` favors higher `Priority` and earlier `CreatedAt` timestamps.
	- `Scheduler` maintains `jobQ` protected by mutex + cond var; `Run()` spawns `workerLoop` per worker.
	- `workerLoop` wakes when jobs are present, scans the queue for a job the worker can satisfy (based on `AvailableThreads()`), supports fallback to single-thread execution if no worker has enough threads for the top job.

- `internal/worker/worker.go`:
	- `Worker` exposes `JobQueue` channel of job pointers and `FreeThreads` channel used as a counting semaphore for available threads.
	- `processJob` sets job `Status = Running`, if `ThreadDemand <= 1` calls `Execute()`, otherwise consumes `FreeThreads` slots, launches `threadCount` goroutines that call `ExecuteChunk`, waits, then returns tokens to `FreeThreads`.

---

## Runtime configuration and environment variables

The server reads environment variables (commonly supplied via `.env` or Docker Compose):
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_SSL_MODE` — Postgres connection.
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` — Redis connection.
- `API_PORT` — port for the Gin server (default `8080`).
- `WORKER_1_ID`, `WORKER_1_THREADS`, `WORKER_2_ID`, `WORKER_2_THREADS` — worker identifiers and thread counts used by `cmd/api.go` to create workers.
- `WORKER_QUEUE_SIZE` — optional queue size for worker job channels (default 100 in code if missing).

Docker Compose and `docker/` contain relevant Dockerfiles for the API and frontend; a `docker-compose.yml` exists at repo root (not detailed here) and can be used to run Postgres, Redis, API, and frontend together.

---

## Tests and how to run them

Unit tests exist for jobs, scheduler, and workers. Typical commands (Go + Node) — run from repo root:

```bash
go test ./...    # runs Go unit tests

# Frontend (from frontend/):
cd frontend
npm install
npm run test     # runs frontend tests
```

Notes:
- Some tests include large arrays or performance comparisons; they may be slow by default and should be tuned (reduce sizes) for CI.
- There are integration-style test files under `cmd/` that assume DB/Redis availability; use Docker Compose to stand up services before running them.

---

## Known issues, inconsistencies, and bugs (actionable)

1. Job lifecycle/state bug
	 - In `internal/job/job.go`, `Execute()` sets `j.Status = Pending` at the start. This is incorrect: it should set `j.Status = Running` when starting execution. This inconsistency can cause incorrect lifecycle observations and metric calculations.

2. Global result mutex and concurrency model
	 - `internal/job/utils.go` defines a package-level `resultMutex` used by `addPartialSum`. This serializes result aggregation across all jobs in the process and can become a bottleneck; it also mixes per-job state with a global lock. Replace with a per-`Job` mutex (e.g., embed `sync.Mutex` or `sync/atomic` usage) to isolate concurrency.

3. Job type naming / representation inconsistencies
 	- At runtime the API registers factories under snake_case keys (`add_numbers`, `large_array_sum`) and the UI submission form (`frontend/src/components/JobSubmitForm.tsx`) sends snake_case job type values (e.g., `add_numbers`, `large_array_sum`) — this is compatible with the backend factories.
 	- However, the frontend TypeScript service (`frontend/src/services/api.ts`) defines a `JobType` type using PascalCase (`AddNumbers`, `LargeArraySum`), which is inconsistent with the submission form's runtime values and may confuse developers or static checks. Recommend normalizing to a single canonical representation (preferably snake_case for runtime payloads) or adding a clear mapping layer in the API.

4. Marshaling `interface{}` fields
	 - `Job.Payload` and `Job.Result` are `interface{}`; serialized forms (in Redis/DB) will be JSON blobs or `map[string]interface{}`, which complicates rehydration into typed structs. Implement a registry-based unmarshal: store job `Type` alongside payload, and unmarshal into the expected payload struct on read.

5. Worker attribution
	 - `insertJobToDB` currently writes `worker_id` as `NULL`. If worker attribution is valuable, modify workers to set `j.WorkerID` (add to `Job` struct) and populate DB accordingly.

6. Graceful shutdown & retries
	 - `cmd/api.go` currently fails fast on Redis connection. Consider retry/backoff and a graceful shutdown path for the HTTP server and scheduler/worker goroutines.

7. Tests and resource sizes
	 - Some tests use very large arrays by default (100_000_000). Reduce sizes for CI or mark as benchmarks.

8. `addPartialSum` type assertions
	 - `addPartialSum` assumes `j.Result` is `LargeArraySumResult` when not `nil`. If `Result` has another type or is unexpectedly set, this will panic. Ensure `Result` is initialized and/or use type-safe aggregation.

---

## Security and operational notes

- No authentication/authorization is present on `/jobs` endpoints; add auth (JWT or API tokens) before exposing externally.
- Inputs are minimally validated. Use stricter validation for `thread_demand`, `priority`, and payload shapes.
- Persisted `result` is stored as JSONB; if sensitive data could be stored, consider encryption or access controls.
- Add health endpoints (`/healthz`) and readiness checks used by orchestrators.

---

## Recommended immediate fixes (prioritized)

1. Fix `Execute()` status bug (set `Running` at start), run unit tests.
2. Replace package-level `resultMutex` with a per-Job mutex (embed `sync.Mutex` in `Job`) and update `addPartialSum` accordingly.
3. Normalize job type names between frontend and `cmd/api.go` (accept multiple canonical forms or update frontend to send snake_case).
4. Implement registry-based (de)serialization for `Payload`/`Result` so DB and Redis entries can be rehydrated into typed structs.
5. Populate `worker_id` on job completion and persist it.

---

## Recommended medium/long-term improvements

- Add Prometheus metrics and instrument scheduler/worker performance.
- Implement worker registration and heartbeat to persist worker capacity and status into `workers` table.
- Add preemption and job cancellation support, plus priority changes at runtime.
- Replace `interface{}` payloads with explicit typed messages or a schema format (protobuf/JSON schema) for stability.
- Add integration tests that run full stack via Docker Compose (Postgres + Redis + API + frontend smoke tests).
- Add rate limiting, authentication, and RBAC for API access.

---

## How to reason about the code (notes for an LLM)

- `Job` instances are passed by pointer through `Scheduler` -> `Worker` -> worker goroutines. The scheduler removes jobs from the central heap and enqueues them to a worker's `JobQueue`.
- Concurrency model summary:
	- Scheduler protects the priority queue with a mutex + condition variable.
	- Workers use a buffered `JobQueue` channel for inbound jobs and a `FreeThreads` buffered channel as a counting semaphore.
	- Jobs that are multi-threaded call `ExecuteChunk` concurrently; aggregation happens via `addPartialSum`.
- Persistence lifecycle:
	- Jobs are created and kept in an in-memory map and cached to Redis on submission.
	- Once `j.Status == Completed`, a background goroutine inserts/upserts the job into Postgres and emits metrics.

---

## Next actions I can take (pick any)

- Implement the high-priority fixes (status bug, per-job mutex, registry normalization) and run `go test ./...`.
- Implement robust payload/result (de)serialization for Redis/DB reads.
- Add health endpoints and graceful shutdown logic.
- Create a small integration `docker-compose` workflow that exercises the API with Postgres and Redis and runs tests.

If you want, I will now implement the three immediate fixes (status bug + per-job mutex + registry normalization) and run the Go tests. Reply with `yes` to proceed or specify a different next action.

---

Appendix: quick file links

- API entry: [cmd/api.go](cmd/api.go)
- Job model: [internal/job/job.go](internal/job/job.go)
- Job utils: [internal/job/utils.go](internal/job/utils.go)
- Scheduler: [internal/scheduler/scheduler.go](internal/scheduler/scheduler.go)
- Worker: [internal/worker/worker.go](internal/worker/worker.go)
- DB schema: [db/schema.sql](db/schema.sql)
- Frontend API client: [frontend/src/services/api.ts](frontend/src/services/api.ts)

