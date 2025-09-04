package job

import (
	"sync"
	"testing"
	"time"
)

func generateLargeArray(size int) []int {
	array := make([]int, size)
	for i := 0; i < size; i++ {
		array[i] = i + 1
	}
	return array
}

func TestLargeArraySumJobPerformanceComparison(t *testing.T) {
	size := 100_000_000 // 10 million for faster test runs
	array := generateLargeArray(size)
	expected := size * (size + 1) / 2

	// Single-threaded job
	singleJob := NewJob("1", "LargeArraySumSingle", LargeArraySumJob, 1, LargeArraySumPayload{Array: array})
	singleJob.ThreadDemand = 1

	start := time.Now()
	singleJob.ExecuteChunk(0, 1)
	durationSingle := time.Since(start)

	resultSingle := singleJob.Result.(LargeArraySumResult)
	if resultSingle.Sum != expected {
		t.Errorf("single-threaded: expected %d, got %d", expected, resultSingle.Sum)
	}

	// Multi-threaded job
	numThreads := 8
	multiJob := NewJob("2", "LargeArraySumMulti", LargeArraySumJob, 1, LargeArraySumPayload{Array: array})
	multiJob.ThreadDemand = numThreads

	var wg sync.WaitGroup
	start = time.Now()
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			multiJob.ExecuteChunk(threadID, numThreads)
		}(i)
	}
	wg.Wait()
	durationMulti := time.Since(start)

	resultMulti := multiJob.Result.(LargeArraySumResult)
	if resultMulti.Sum != expected {
		t.Errorf("multi-threaded: expected %d, got %d", expected, resultMulti.Sum)
	}

	// Print timing comparison
	t.Logf("Array size: %d", size)
	t.Logf("Single-threaded took: %s", durationSingle)
	t.Logf("Multi-threaded with %d threads took: %s", numThreads, durationMulti)
	t.Logf("Speedup: %.2fx", float64(durationSingle)/float64(durationMulti))

}

func TestAddNumbersJob(t *testing.T) {
	job := NewJob("add1", "AddNumbers", AddNumbersJob, 1, AddNumbersPayload{X: 3, Y: 4})
	job.Execute()
	result := job.Result.(AddNumbersResult)
	expected := 7
	if result.Sum != expected {
		t.Errorf("AddNumbersJob: expected %d, got %d", expected, result.Sum)
	}
}

func TestAddNumbersJobZeroValues(t *testing.T) {
	job := NewJob("add2", "AddNumbersZero", AddNumbersJob, 1, AddNumbersPayload{X: 0, Y: 0})
	job.Execute()
	result := job.Result.(AddNumbersResult)
	expected := 0
	if result.Sum != expected {
		t.Errorf("AddNumbersJobZero: expected %d, got %d", expected, result.Sum)
	}
}

func TestReverseStringJob(t *testing.T) {
	job := NewJob("rev1", "ReverseString", ReverseStringJob, 1, ReverseStringPayload{Text: "hello"})
	job.Execute()
	result := job.Result.(ReverseStringResult)
	expected := "olleh"
	if result.Reversed != expected {
		t.Errorf("ReverseStringJob: expected %s, got %s", expected, result.Reversed)
	}
}

func TestReverseStringJobEmpty(t *testing.T) {
	job := NewJob("rev2", "ReverseEmpty", ReverseStringJob, 1, ReverseStringPayload{Text: ""})
	job.Execute()
	result := job.Result.(ReverseStringResult)
	expected := ""
	if result.Reversed != expected {
		t.Errorf("ReverseStringJobEmpty: expected empty string, got %s", result.Reversed)
	}
}

func TestResizeImageJobDummy(t *testing.T) {
	payload := ResizeImagePayload{URL: "http://example.com/image.png", Width: 100, Height: 200}
	job := NewJob("img1", "ResizeImage", ResizeImageJob, 1, payload)
	job.Execute()
	// Since we haven’t implemented actual resizing, just check job completes
	if job.Status != Completed {
		t.Errorf("ResizeImageJob: expected status Completed, got %s", job.Status)
	}
}

func TestLargeArraySumJobEmptyArray(t *testing.T) {
	payload := LargeArraySumPayload{Array: []int{}}
	job := NewJob("arr1", "EmptyArraySum", LargeArraySumJob, 1, payload)
	job.ThreadDemand = 1
	job.Execute()
	result := job.Result.(LargeArraySumResult)
	if result.Sum != 0 {
		t.Errorf("Empty array sum: expected 0, got %d", result.Sum)
	}
}

func TestLargeArraySumJobSingleElement(t *testing.T) {
	payload := LargeArraySumPayload{Array: []int{42}}
	job := NewJob("arr2", "SingleElementSum", LargeArraySumJob, 1, payload)
	job.ThreadDemand = 1
	job.Execute()
	result := job.Result.(LargeArraySumResult)
	if result.Sum != 42 {
		t.Errorf("Single element sum: expected 42, got %d", result.Sum)
	}
}

func TestLargeArraySumJobMultiThreadEdge(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	payload := LargeArraySumPayload{Array: array}
	job := NewJob("arr3", "MultiThreadEdge", LargeArraySumJob, 1, payload)
	job.ThreadDemand = 10 // more threads than array length

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
	expected := 15
	if result.Sum != expected {
		t.Errorf("Multi-thread edge sum: expected %d, got %d", expected, result.Sum)
	}
}
