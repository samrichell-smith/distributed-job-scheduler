package scheduler

import (
	"testing"
	"time"

	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/worker"
)

// Helper to create test workers
func createTestWorkers() []*worker.Worker {
	w1 := worker.NewWorker("w1", 4)
	w2 := worker.NewWorker("w2", 2)
	w1.Start()
	w2.Start()
	return []*worker.Worker{w1, w2}
}

// Helper to wait for a job to complete (polling)
func waitJobCompletion(j *job.Job, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if j.Status == job.Completed {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestSchedulerBasicExecution(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	j1 := job.NewJob("j1", "Add", job.AddNumbersJob, 1, job.AddNumbersPayload{X: 1, Y: 2})
	s.Submit(j1)

	done := waitJobCompletion(j1, 200*time.Millisecond)
	if !done {
		t.Fatal("Job j1 did not complete in time")
	}

	res := j1.Result.(job.AddNumbersResult)
	if res.Sum != 3 {
		t.Errorf("Expected sum 3, got %d", res.Sum)
	}
}

func TestSchedulerThreadDemand(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	jobBig := job.NewJob("big", "BigJob", job.LargeArraySumJob, 1, job.LargeArraySumPayload{Array: []int{1, 2, 3, 4}})
	jobBig.ThreadDemand = 4

	jobSmall := job.NewJob("small", "SmallJob", job.LargeArraySumJob, 2, job.LargeArraySumPayload{Array: []int{5, 6}})
	jobSmall.ThreadDemand = 2

	s.Submit(jobBig)
	s.Submit(jobSmall)

	doneBig := waitJobCompletion(jobBig, 500*time.Millisecond)
	doneSmall := waitJobCompletion(jobSmall, 500*time.Millisecond)

	if !doneBig {
		t.Errorf("Big job did not complete")
	}
	if !doneSmall {
		t.Errorf("Small job did not complete")
	}
}

func TestSchedulerSingleThreadFallback(t *testing.T) {
	workers := createTestWorkers()
	s := NewScheduler(workers)
	s.Run()
	defer s.Stop()

	jobImpossible := job.NewJob("impossible", "ImpossibleJob", job.LargeArraySumJob, 1, job.LargeArraySumPayload{Array: []int{1, 2, 3}})
	jobImpossible.ThreadDemand = 10
	s.Submit(jobImpossible)

	done := waitJobCompletion(jobImpossible, 500*time.Millisecond)
	if !done {
		t.Errorf("Job with impossible thread demand did not complete with single-thread fallback")
	}
}
