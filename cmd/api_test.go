//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/scheduler"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/worker"
)

// helper to initialize router and scheduler for testing
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	if os.Getenv("RUN_INTEGRATION") != "1" {
		// make tests explicit to run; default skipped
		// Use `RUN_INTEGRATION=1 go test ./cmd` to run them
		// We use t.Skip in callers, but this check helps fast-fail if invoked directly without t
	}

	// Load test environment variables
	if err := godotenv.Load("../.env.test"); err != nil {
		log.Printf("Warning: .env.test file not found")
	}

	// Reset globals for each test
	jobs = make(map[string]*job.Job)

	queueSize := getEnvInt("WORKER_QUEUE_SIZE", 10)
	workers := []*worker.Worker{
		worker.NewWorkerWithQueueSize(
			os.Getenv("WORKER_1_ID"),
			getEnvInt("WORKER_1_THREADS", 4),
			queueSize,
		),
		worker.NewWorkerWithQueueSize(
			os.Getenv("WORKER_2_ID"),
			getEnvInt("WORKER_2_THREADS", 2),
			queueSize,
		),
	}
	for _, w := range workers {
		w.Start()
	}

	sched = scheduler.NewScheduler(workers)
	sched.Run()

	// register job types
	jobRegistry = map[string]JobFactory{}
	jobRegistry["add_numbers"] = func(id string, req SubmitJobRequest) (*job.Job, error) {
		var payload job.AddNumbersPayload
		m, ok := req.Payload.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid payload for add_numbers")
		}
		if err := mapToStruct(m, &payload); err != nil {
			return nil, err
		}
		return job.NewJob(id, "AddNumbers", job.AddNumbersJob, req.Priority, payload), nil
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
		return job.NewJob(id, "LargeArraySum", job.LargeArraySumJob, req.Priority, payload), nil
	}

	r := gin.Default()

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

		sched.Submit(j)
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
		jobsMu.RLock()
		j, ok := jobs[id]
		jobsMu.RUnlock()
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		c.JSON(http.StatusOK, jobToResponse(j))
	})

	return r
}

// wait until a job is completed, or timeout
func waitForJobCompletion(id string, timeout time.Duration) *job.Job {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		jobsMu.RLock()
		j, ok := jobs[id]
		jobsMu.RUnlock()
		if ok && j.Status == job.Completed {
			return j
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func TestPostJobAndGetJob(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobData := `{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":5,"y":7}}`
	req, _ := http.NewRequest(http.MethodPost, "/jobs", strings.NewReader(jobData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", w.Code)
	}

	var resp JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	j := waitForJobCompletion(resp.ID, 1*time.Second)
	if j == nil {
		t.Fatalf("job did not complete within timeout")
	}

	addResult, ok := j.Result.(job.AddNumbersResult)
	if !ok {
		t.Fatalf("expected result to be AddNumbersResult, got %T", j.Result)
	}
	if addResult.Sum != 12 {
		t.Errorf("expected sum 12, got %d", addResult.Sum)
	}

	req2, _ := http.NewRequest(http.MethodGet, "/jobs/"+resp.ID, nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w2.Code)
	}

	var resp2 JobResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
}

func TestGetJobsList(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobData := `{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":1,"y":2}}`
	req, _ := http.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(jobData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	j := waitForJobCompletion(resp.ID, 1*time.Second)
	if j == nil {
		t.Fatalf("job did not complete within timeout")
	}

	reqGet, _ := http.NewRequest(http.MethodGet, "/jobs", nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", wGet.Code)
	}

	var jobsList []JobResponse
	if err := json.Unmarshal(wGet.Body.Bytes(), &jobsList); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(jobsList) != 1 {
		t.Errorf("expected 1 job in list, got %d", len(jobsList))
	}

	// Assert result
	addResult, ok := j.Result.(job.AddNumbersResult)
	if !ok {
		t.Fatalf("expected result to be AddNumbersResult, got %T", j.Result)
	}
	if addResult.Sum != 3 {
		t.Errorf("expected sum 3, got %d", addResult.Sum)
	}
}

func TestLargeArraySumJob(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobData := `{"type":"large_array_sum","priority":1,"thread_demand":2,"payload":{"array":[1,2,3,4,5]}}`
	req, _ := http.NewRequest(http.MethodPost, "/jobs", strings.NewReader(jobData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	j := waitForJobCompletion(resp.ID, 1*time.Second)
	if j == nil {
		t.Fatalf("job did not complete within timeout")
	}

	sumResult, ok := j.Result.(job.LargeArraySumResult)
	if !ok {
		t.Fatalf("expected result to be LargeArraySumResult, got %T", j.Result)
	}
	if sumResult.Sum != 15 {
		t.Errorf("expected sum 15, got %d", sumResult.Sum)
	}
}

func TestUnsupportedJobType(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobData := `{"type":"nonexistent","priority":1,"thread_demand":1,"payload":{}}`
	req, _ := http.NewRequest(http.MethodPost, "/jobs", strings.NewReader(jobData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unsupported job type, got %d", w.Code)
	}
}

func TestInvalidPayload(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobData := `{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":"notanumber"}}`
	req, _ := http.NewRequest(http.MethodPost, "/jobs", strings.NewReader(jobData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid payload, got %d", w.Code)
	}
}

func TestJobThreadDemandTooHigh(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobData := `{"type":"add_numbers","priority":1,"thread_demand":100,"payload":{"x":1,"y":2}}`
	req, _ := http.NewRequest(http.MethodPost, "/jobs", strings.NewReader(jobData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	j := waitForJobCompletion(resp.ID, 1*time.Second)
	if j == nil {
		t.Fatalf("job did not complete within timeout")
	}

	addResult, ok := j.Result.(job.AddNumbersResult)
	if !ok {
		t.Fatalf("expected result to be AddNumbersResult, got %T", j.Result)
	}
	if addResult.Sum != 3 {
		t.Errorf("expected sum 3, got %d", addResult.Sum)
	}
}

func TestMultipleJobsConcurrency(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}
	r := setupRouter()
	defer sched.Stop()

	jobRequests := []string{
		`{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":1,"y":2}}`,
		`{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":3,"y":4}}`,
		`{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":5,"y":6}}`,
		`{"type":"add_numbers","priority":1,"thread_demand":1,"payload":{"x":7,"y":8}}`,
	}

	jobIDs := make([]string, 0, len(jobRequests))

	// POST all jobs
	for _, jobData := range jobRequests {
		req, _ := http.NewRequest(http.MethodPost, "/jobs", strings.NewReader(jobData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Fatalf("expected status 202, got %d", w.Code)
		}

		var resp JobResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}
		jobIDs = append(jobIDs, resp.ID)
	}

	// Wait for all jobs to complete
	for i, id := range jobIDs {
		j := waitForJobCompletion(id, 5*time.Second) // longer timeout
		if j == nil {
			t.Fatalf("job %d did not complete within timeout", i+1)
		}
		if addRes, ok := j.Result.(job.AddNumbersResult); !ok {
			t.Errorf("job %d expected AddNumbersResult, got %T", i+1, j.Result)
		} else {
			// Optional: verify the sum
			expected := 0
			switch i {
			case 0:
				expected = 3
			case 1:
				expected = 7
			case 2:
				expected = 11
			case 3:
				expected = 15
			}
			if addRes.Sum != expected {
				t.Errorf("job %d expected sum %d, got %d", i+1, expected, addRes.Sum)
			}
		}
	}
}
