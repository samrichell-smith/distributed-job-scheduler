package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	testDBURL := "postgres://postgres:postgres@localhost:5432/job_scheduler_test?sslmode=disable"
	db, err := pgxpool.New(context.Background(), testDBURL)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	db.Exec(context.Background(), "TRUNCATE jobs")
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
