package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
)

// setupMetricsTestDB uses setupTestDB from db_integration_test.go
func setupMetricsTestDB(t *testing.T) *pgxpool.Pool {
	db := setupTestDB(t)
	db.Exec(context.Background(), "TRUNCATE job_metrics")
	return db
}

func TestInsertJobMetrics(t *testing.T) {
	db := setupMetricsTestDB(t)
	defer db.Close()

	j := &job.Job{
		ID:           uuid.New().String(),
		Type:         "add_numbers",
		Priority:     1,
		ThreadDemand: 2,
		Status:       "Completed",
		CreatedAt:    time.Now().Add(-10 * time.Second),
		StartedAt:    time.Now().Add(-8 * time.Second),
		CompletedAt:  time.Now(),
		Result:       map[string]interface{}{"Sum": 42},
	}

	// Insert job row (needed for FK)
	resultJSON := []byte(`{"Sum":42}`)
	_, err := db.Exec(context.Background(), `
		INSERT INTO jobs (id, type, priority, thread_demand, status, created_at, started_at, completed_at, result)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, j.ID, j.Type, j.Priority, j.ThreadDemand, j.Status, j.CreatedAt, j.StartedAt, j.CompletedAt, resultJSON)
	if err != nil {
		t.Fatalf("Insert job failed: %v", err)
	}

	// Call metric logging logic (simulate what insertJobToDB does)
	queueTime := j.StartedAt.Sub(j.CreatedAt).Seconds()
	execTime := j.CompletedAt.Sub(j.StartedAt).Seconds()
	totalTime := j.CompletedAt.Sub(j.CreatedAt).Seconds()
	_, err = db.Exec(context.Background(), `
		INSERT INTO job_metrics (job_id, metric_name, metric_value)
		VALUES ($1, $2, $3), ($1, $4, $5), ($1, $6, $7), ($1, $8, $9)
	`,
		j.ID, "queue_time", queueTime,
		"execution_time", execTime,
		"total_time", totalTime,
		"worker_threads", float64(j.ThreadDemand),
	)
	if err != nil {
		t.Fatalf("Insert metrics failed: %v", err)
	}

	// Query and check metrics
	rows, err := db.Query(context.Background(), "SELECT metric_name, metric_value FROM job_metrics WHERE job_id=$1", j.ID)
	if err != nil {
		t.Fatalf("Query metrics failed: %v", err)
	}
	defer rows.Close()

	metrics := map[string]float64{}
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		metrics[name] = value
	}

	if len(metrics) != 4 {
		t.Errorf("Expected 4 metrics, got %d", len(metrics))
	}
	if metrics["worker_threads"] != 2 {
		t.Errorf("Expected worker_threads 2, got %v", metrics["worker_threads"])
	}
	if metrics["queue_time"] <= 0 || metrics["execution_time"] <= 0 || metrics["total_time"] <= 0 {
		t.Errorf("Expected positive times, got queue=%v exec=%v total=%v", metrics["queue_time"], metrics["execution_time"], metrics["total_time"])
	}
}
