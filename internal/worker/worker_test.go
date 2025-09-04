package worker

import (
	"testing"

	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
)

func TestWorkerProcessesJob(t *testing.T) {
	worker := NewWorker("w1", 4)
	worker.Start()

	array := make([]int, 100)
	for i := 0; i < 100; i++ {
		array[i] = i + 1
	}
	payload := job.LargeArraySumPayload{Array: array}
	j := job.NewJob("1", "TestJob", job.LargeArraySumJob, 1, payload)
	j.ThreadDemand = 4

	worker.JobQueue <- j
	worker.Stop()

	result := j.Result.(job.LargeArraySumResult)
	expected := 100 * (100 + 1) / 2

	if result.Sum != expected {
		t.Errorf("expected %d, got %d", expected, result.Sum)
	}
}
