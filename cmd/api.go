package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/scheduler"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/worker"
)

var (
	redisClient *redis.Client
	redisCtx    = context.Background()
)

var (
	sched  *scheduler.Scheduler
	jobsMu sync.RWMutex
	jobs   = make(map[string]*job.Job)
)

var (
	db *pgxpool.Pool
)

type SubmitJobRequest struct {
	Type         string      `json:"type" binding:"required"`
	Priority     int         `json:"priority" binding:"required"`
	ThreadDemand int         `json:"thread_demand" binding:"required"`
	Payload      interface{} `json:"payload" binding:"required"`
}

type JobResponse struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Priority     int         `json:"priority"`
	ThreadDemand int         `json:"thread_demand"`
	Status       string      `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	StartedAt    *time.Time  `json:"started_at,omitempty"`
	CompletedAt  *time.Time  `json:"completed_at,omitempty"`
	Result       interface{} `json:"result,omitempty"`
}

func jobToResponse(j *job.Job) JobResponse {
	return JobResponse{
		ID:           j.ID,
		Type:         string(j.Type),
		Priority:     j.Priority,
		ThreadDemand: j.ThreadDemand,
		Status:       string(j.Status),
		CreatedAt:    j.CreatedAt,
		StartedAt: func() *time.Time {
			if !j.StartedAt.IsZero() {
				return &j.StartedAt
			}
			return nil
		}(),
		CompletedAt: func() *time.Time {
			if !j.CompletedAt.IsZero() {
				return &j.CompletedAt
			}
			return nil
		}(),
		Result: j.Result,
	}
}

// Registry for job types
type JobFactory func(id string, req SubmitJobRequest) (*job.Job, error)

var jobRegistry = map[string]JobFactory{}

// Helper: convert map[string]interface{} to struct
func mapToStruct(m map[string]interface{}, out interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func main() {
	// Create Gin router
	r := gin.Default()

	// API endpoint: GET /db/jobs - fetch all jobs from PostgreSQL
	r.GET("/db/jobs", func(c *gin.Context) {
		rows, err := db.Query(context.Background(), "SELECT id, type, priority, thread_demand, status, created_at, completed_at, result FROM jobs ORDER BY created_at DESC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var jobs []JobResponse
		for rows.Next() {
			var j JobResponse
			var resultRaw []byte
			var startedAt time.Time
			err := rows.Scan(&j.ID, &j.Type, &j.Priority, &j.ThreadDemand, &j.Status, &j.CreatedAt, &startedAt, &j.CompletedAt, &resultRaw)
			if !startedAt.IsZero() {
				j.StartedAt = &startedAt
			} else {
				j.StartedAt = nil
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if len(resultRaw) > 0 {
				json.Unmarshal(resultRaw, &j.Result)
			}
			jobs = append(jobs, j)
		}
		c.JSON(http.StatusOK, jobs)
	})

	// API endpoint: GET /db/jobs/:id - fetch a single job from PostgreSQL by ID
	r.GET("/db/jobs/:id", func(c *gin.Context) {
		id := c.Param("id")
		var j JobResponse
		var resultRaw []byte
		var startedAt time.Time
		err := db.QueryRow(context.Background(), "SELECT id, type, priority, thread_demand, status, created_at, started_at, completed_at, result FROM jobs WHERE id=$1", id).Scan(
			&j.ID, &j.Type, &j.Priority, &j.ThreadDemand, &j.Status, &j.CreatedAt, &startedAt, &j.CompletedAt, &resultRaw)
		if !startedAt.IsZero() {
			j.StartedAt = &startedAt
		} else {
			j.StartedAt = nil
		}
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		if len(resultRaw) > 0 {
			json.Unmarshal(resultRaw, &j.Result)
		}
		c.JSON(http.StatusOK, j)
	})

	// Initialize PostgreSQL connection
	dbURL := "postgres://postgres:postgres@localhost:5432/job_scheduler"
	var err error
	db, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping PostgreSQL: %v", err)
	}

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		panic("Could not connect to Redis: " + err.Error())
	}
	// Register job types
	jobRegistry["add_numbers"] = func(id string, req SubmitJobRequest) (*job.Job, error) {
		var payload job.AddNumbersPayload
		m, ok := req.Payload.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid payload for add_numbers")
		}
		if err := mapToStruct(m, &payload); err != nil {
			return nil, err
		}
		return job.NewJob(id, "add_numbers", job.AddNumbersJob, req.Priority, payload), nil
	}

	jobRegistry["large_array_sum"] = func(id string, req SubmitJobRequest) (*job.Job, error) {
		var payload job.LargeArraySumPayload
		m, ok := req.Payload.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid payload for large_array_sum")
		}
		if err := mapToStruct(m, &payload); err != nil {
			return nil, err
		}
		return job.NewJob(id, "large_array_sum", job.LargeArraySumJob, req.Priority, payload), nil
	}
	// Create workers
	workers := []*worker.Worker{
		worker.NewWorker("w1", 8),
		worker.NewWorker("w2", 2),
	}
	for _, w := range workers {
		w.Start()
	}

	// Create scheduler
	sched = scheduler.NewScheduler(workers)
	sched.Run()
	defer sched.Stop()

	r.POST("/jobs", func(c *gin.Context) {
		var req SubmitJobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		factory, ok := jobRegistry[req.Type]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported job type"})
			return
		}

		id := uuid.New().String()
		created := time.Now()
		j, err := factory(id, req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		j.ThreadDemand = req.ThreadDemand
		j.CreatedAt = created

		jobsMu.Lock()
		jobs[j.ID] = j
		jobsMu.Unlock()

		// Write job state to Redis
		jobJSON, _ := json.Marshal(j)
		redisClient.Set(redisCtx, "job:"+j.ID, jobJSON, 0)

		sched.Submit(j)

		// Wait for job completion and insert into DB
		go func(jobPtr *job.Job) {
			for {
				time.Sleep(50 * time.Millisecond)
				if jobPtr.Status == "Completed" {
					if err := insertJobToDB(jobPtr); err != nil {
						log.Printf("Failed to insert job %s to DB: %v", jobPtr.ID, err)
					}
					break
				}
			}
		}(j)

		c.JSON(http.StatusAccepted, jobToResponse(j))
	})

	r.GET("/jobs", func(c *gin.Context) {
		jobsMu.RLock()
		resp := make([]JobResponse, 0, len(jobs))
		for _, j := range jobs {
			resp = append(resp, jobToResponse(j))
		}
		jobsMu.RUnlock()
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/jobs/:id", func(c *gin.Context) {
		id := c.Param("id")
		// Try Redis first
		val, err := redisClient.Get(redisCtx, "job:"+id).Result()
		if err == nil {
			var jobObj job.Job
			if err := json.Unmarshal([]byte(val), &jobObj); err == nil {
				c.JSON(http.StatusOK, jobToResponse(&jobObj))
				return
			}
		}
		// Fallback to in-memory
		jobsMu.RLock()
		j, ok := jobs[id]
		jobsMu.RUnlock()
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		c.JSON(http.StatusOK, jobToResponse(j))
	})

	r.Run(":8080")
}

// insertJobToDB inserts a completed job into the jobs table
func insertJobToDB(j *job.Job) error {
	resultJSON, err := json.Marshal(j.Result)
	if err != nil {
		return err
	}
	_, err = db.Exec(context.Background(), `
	       INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, started_at, completed_at, result, worker_id)
	       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	       ON CONFLICT (id) DO UPDATE SET
		       status = EXCLUDED.status,
		       started_at = EXCLUDED.started_at,
		       completed_at = EXCLUDED.completed_at,
		       result = EXCLUDED.result,
		       worker_id = EXCLUDED.worker_id
	       `,
		j.ID,
		j.Type,
		j.Priority,
		j.ThreadDemand,
		j.Status,
		j.CreatedAt,
		j.StartedAt,
		j.CompletedAt,
		resultJSON,
		nil, // worker_id
	)

	// Log performance metrics if job is completed
	if !j.StartedAt.IsZero() && !j.CompletedAt.IsZero() {
		queueTime := j.StartedAt.Sub(j.CreatedAt).Seconds()
		execTime := j.CompletedAt.Sub(j.StartedAt).Seconds()
		totalTime := j.CompletedAt.Sub(j.CreatedAt).Seconds()
		_, _ = db.Exec(context.Background(), `
		       INSERT INTO job_metrics (job_id, metric_name, metric_value)
		       VALUES ($1, $2, $3), ($1, $4, $5), ($1, $6, $7), ($1, $8, $9)
	       `,
			j.ID, "queue_time", queueTime,
			"execution_time", execTime,
			"total_time", totalTime,
			"worker_threads", float64(j.ThreadDemand),
		)
	}
	return err
}
