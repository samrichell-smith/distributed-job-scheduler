//go:build integration
// +build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	// integration tests gated by env var
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration tests disabled; set RUN_INTEGRATION=1 to enable")
	}

	if err := godotenv.Load("../.env.test"); err != nil {
		t.Logf("Warning: .env.test file not found: %v", err)
	}

	// Start a postgres container for tests using dockertest
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %v", err)
	}

	// Pull and run postgres
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16",
		Env: []string{
			"POSTGRES_USER=your_username",
			"POSTGRES_PASSWORD=your_password",
			"POSTGRES_DB=job_scheduler",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start postgres container: %v", err)
	}

	// Ensure cleanup
	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	// Exponential backoff to wait for Postgres
	var db *pgxpool.Pool
	if err := pool.Retry(func() error {
		hostPort := resource.GetHostPort("5432/tcp")
		testDBURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			"your_username", "your_password", hostPort, "job_scheduler")
		var err error
		db, err = pgxpool.New(context.Background(), testDBURL)
		if err != nil {
			return err
		}
		return db.Ping(context.Background())
	}); err != nil {
		t.Fatalf("Could not connect to database: %v", err)
	}

	// Create the schema if it doesn't exist
	schemaSQL := `
		CREATE TABLE IF NOT EXISTS jobs (
			id UUID PRIMARY KEY,
			type TEXT NOT NULL,
			priority INT NOT NULL,
			thread_demand INT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			result JSONB,
			worker_id TEXT
		);
	`

	if _, err := db.Exec(context.Background(), schemaSQL); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Clean up any existing test data
	if _, err := db.Exec(context.Background(), "TRUNCATE jobs CASCADE"); err != nil {
		t.Logf("Warning: Failed to truncate jobs table: %v", err)
	}

	return db
}

func TestInsertAndQueryJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a fake job
	j := &job.Job{
		ID:           uuid.New().String(),
		Type:         "add_numbers",
		Priority:     1,
		ThreadDemand: 1,
		Status:       "Completed",
		CreatedAt:    time.Now(),
		CompletedAt:  time.Now(),
		Result:       map[string]interface{}{"Sum": 42},
	}

	// Insert
	resultJSON, _ := json.Marshal(j.Result)
	_, err := db.Exec(context.Background(), `
		INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, completed_at, result)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, j.ID, j.Type, j.Priority, j.ThreadDemand, j.Status, j.CreatedAt, j.CompletedAt, resultJSON)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query
	var status string
	var resultRaw []byte
	err = db.QueryRow(context.Background(), "SELECT status, result FROM jobs WHERE id=$1", j.ID).Scan(&status, &resultRaw)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if status != "Completed" {
		t.Errorf("Expected status Completed, got %s", status)
	}
	var result map[string]interface{}
	json.Unmarshal(resultRaw, &result)
	if result["Sum"] != float64(42) {
		t.Errorf("Expected Sum 42, got %v", result["Sum"])
	}
}

func TestUpdateJobStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	j := &job.Job{
		ID:           uuid.New().String(),
		Type:         "add_numbers",
		Priority:     1,
		ThreadDemand: 1,
		Status:       "Pending",
		CreatedAt:    time.Now(),
		Result:       map[string]interface{}{"Sum": 0},
	}
	resultJSON, _ := json.Marshal(j.Result)
	_, err := db.Exec(context.Background(), `
		INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, result)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, j.ID, j.Type, j.Priority, j.ThreadDemand, j.Status, j.CreatedAt, resultJSON)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	// Update status
	_, err = db.Exec(context.Background(), `UPDATE jobs SET status=$1 WHERE id=$2`, "Completed", j.ID)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	var status string
	err = db.QueryRow(context.Background(), "SELECT status FROM jobs WHERE id=$1", j.ID).Scan(&status)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if status != "Completed" {
		t.Errorf("Expected status Completed, got %s", status)
	}
}

func TestDeleteJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	j := &job.Job{
		ID:           uuid.New().String(),
		Type:         "add_numbers",
		Priority:     1,
		ThreadDemand: 1,
		Status:       "Completed",
		CreatedAt:    time.Now(),
		CompletedAt:  time.Now(),
		Result:       map[string]interface{}{"Sum": 99},
	}
	resultJSON, _ := json.Marshal(j.Result)
	_, err := db.Exec(context.Background(), `
		INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, completed_at, result)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, j.ID, j.Type, j.Priority, j.ThreadDemand, j.Status, j.CreatedAt, j.CompletedAt, resultJSON)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	// Delete
	_, err = db.Exec(context.Background(), `DELETE FROM jobs WHERE id=$1`, j.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	var count int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM jobs WHERE id=$1", j.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Count query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 jobs, got %d", count)
	}
}

func TestInsertDuplicateJobID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	j := &job.Job{
		ID:           uuid.New().String(),
		Type:         "add_numbers",
		Priority:     1,
		ThreadDemand: 1,
		Status:       "Completed",
		CreatedAt:    time.Now(),
		CompletedAt:  time.Now(),
		Result:       map[string]interface{}{"Sum": 1},
	}
	resultJSON, _ := json.Marshal(j.Result)
	_, err := db.Exec(context.Background(), `
		INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, completed_at, result)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, j.ID, j.Type, j.Priority, j.ThreadDemand, j.Status, j.CreatedAt, j.CompletedAt, resultJSON)
	if err != nil {
		t.Fatalf("First insert failed: %v", err)
	}
	// Try to insert duplicate
	_, err = db.Exec(context.Background(), `
		INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, completed_at, result)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, j.ID, j.Type, j.Priority, j.ThreadDemand, j.Status, j.CreatedAt, j.CompletedAt, resultJSON)
	if err == nil {
		t.Errorf("Expected error on duplicate insert, got nil")
	}
}
