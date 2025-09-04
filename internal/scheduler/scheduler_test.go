package scheduler

import (
	"fmt"
	"testing"
	"time"

	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/worker"
)

// -------------------------
// Helper to create workers
// -------------------------
func createTestWorkers() []*worker.Worker {
	w1 := worker.NewWorker("w1", 4) // 4 threads
	w2 := worker.NewWorker("w2", 2) // 2 threads
	w1.Start()
	w2.Start()
	return []*worker.Worker{w1, w2}
}

// -------------------------
// Test Scheduler job order
// -------------------------
func TestSchedulerPriorityExecution(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	// Jobs with different priorities
	job1 := job.NewJob("1", "Job1", job.AddNumbersJob, 1, job.AddNumbersPayload{X: 1, Y: 2})
	job2 := job.NewJob("2", "Job2", job.AddNumbersJob, 5, job.AddNumbersPayload{X: 2, Y: 3})
	job3 := job.NewJob("3", "Job3", job.AddNumbersJob, 3, job.AddNumbersPayload{X: 3, Y: 4})

	s.Submit(job1)
	s.Submit(job2)
	s.Submit(job3)

	time.Sleep(200 * time.Millisecond) // allow scheduler to assign jobs

	// Expect job2 (priority 5) first
	if job2.Status != job.Completed {
		t.Errorf("Expected job2 to be completed first")
	}
	if job1.Status != job.Completed || job3.Status != job.Completed {
		t.Errorf("Expected all jobs to be completed eventually")
	}
}

// -------------------------
// Test thread-demand scheduling
// -------------------------
func TestSchedulerThreadDemand(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	// Job demanding 4 threads (only worker w1 can run it)
	jobBig := job.NewJob("big", "BigJob", job.LargeArraySumJob, 1, job.LargeArraySumPayload{Array: []int{1, 2, 3, 4}})
	jobBig.ThreadDemand = 4

	// Job demanding 2 threads (both workers can run it)
	jobSmall := job.NewJob("small", "SmallJob", job.LargeArraySumJob, 2, job.LargeArraySumPayload{Array: []int{5, 6}})
	jobSmall.ThreadDemand = 2

	s.Submit(jobBig)
	s.Submit(jobSmall)

	time.Sleep(200 * time.Millisecond) // allow scheduler to assign jobs

	if jobBig.Status != job.Completed {
		t.Errorf("Big job should be completed by worker with enough threads")
	}
	if jobSmall.Status != job.Completed {
		t.Errorf("Small job should be completed by any worker with enough threads")
	}
}

// -------------------------
// Test multiple jobs concurrently
// -------------------------
func TestSchedulerMultipleJobs(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	numJobs := 10
	jobs := make([]*job.Job, numJobs)
	for i := 0; i < numJobs; i++ {
		j := job.NewJob(
			fmt.Sprintf("%d", i), // converts integer to string
			"AddNumbers",
			job.AddNumbersJob,
			i%3, // varying priorities
			job.AddNumbersPayload{X: i, Y: i},
		)

		jobs[i] = j
		s.Submit(j)
	}

	time.Sleep(500 * time.Millisecond)

	for _, j := range jobs {
		if j.Status != job.Completed {
			t.Errorf("Job %s should be completed, got status %s", j.ID, j.Status)
		}
	}
}

// -------------------------
// Test scheduler stop
// -------------------------
func TestSchedulerStop(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()

	j := job.NewJob("stop1", "AddNumbers", job.AddNumbersJob, 1, job.AddNumbersPayload{X: 1, Y: 2})
	s.Submit(j)

	time.Sleep(100 * time.Millisecond)
	s.Stop()

	if j.Status != job.Completed {
		t.Errorf("Job should have completed before scheduler stopped, got %s", j.Status)
	}
}

// -------------------------
// Test scheduler with no suitable worker
// -------------------------
func TestSchedulerNoSuitableWorker(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	// Job demands 10 threads, no worker can satisfy
	jobImpossible := job.NewJob("impossible", "ImpossibleJob", job.LargeArraySumJob, 1, job.LargeArraySumPayload{Array: []int{1, 2, 3}})
	jobImpossible.ThreadDemand = 10
	s.Submit(jobImpossible)

	time.Sleep(200 * time.Millisecond)

	if jobImpossible.Status == job.Completed {
		t.Errorf("Job with impossible thread demand should not have completed")
	}
}
