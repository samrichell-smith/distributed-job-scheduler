package job

import (
	"sync"
	"testing"
)

func TestLargeArraySumJobSingleThread(t *testing.T) {
	array := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	payload := LargeArraySumPayload{Array: array}
	job := NewJob("1", "SumArray", LargeArraySumJob, 1, payload)

	job.ThreadDemand = 1 // single-thread test
	job.Execute()

	result := job.Result.(LargeArraySumResult)
	expected := 55

	if result.Sum != expected {
		t.Errorf("expected %d, got %d", expected, result.Sum)
	}
}

func TestLargeArraySumJobMultiThread(t *testing.T) {
	array := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		array[i] = i + 1
	}

	payload := LargeArraySumPayload{Array: array}
	job := NewJob("2", "SumArrayMulti", LargeArraySumJob, 1, payload)
	job.ThreadDemand = 4 // use 4 threads

	var wg sync.WaitGroup
	for i := 0; i < job.ThreadDemand; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			job.ExecuteChunk(threadID, job.ThreadDemand)
		}(i)
	}
	wg.Wait()

	result := job.Result.(LargeArraySumResult)
	expected := 1000 * (1000 + 1) / 2 // sum of 1..1000

	if result.Sum != expected {
		t.Errorf("expected %d, got %d", expected, result.Sum)
	}
}
