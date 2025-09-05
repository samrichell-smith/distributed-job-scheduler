package worker

import (
	"sync"

	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
)

type Worker struct {
	ID          string
	JobQueue    chan *job.Job
	NumThreads  int
	FreeThreads chan struct{}
	WaitGroup   sync.WaitGroup
}

func NewWorker(id string, numThreads int) *Worker {
	w := &Worker{
		ID:          id,
		JobQueue:    make(chan *job.Job, 100),
		NumThreads:  numThreads,
		FreeThreads: make(chan struct{}, numThreads),
	}

	// Initialize all threads as free
	for i := 0; i < numThreads; i++ {
		w.FreeThreads <- struct{}{}
	}

	return w
}

func (w *Worker) Start() {
	for i := 0; i < w.NumThreads; i++ {
		w.WaitGroup.Add(1)
		go func(threadID int) {
			defer w.WaitGroup.Done()

			for job := range w.JobQueue {
				w.processJob(job)
			}
		}(i)
	}
}

func (w *Worker) processJob(j *job.Job) {
	j.Status = job.Running

	threadsToUse := j.ThreadDemand
	if threadsToUse <= 1 {
		// Single-threaded job
		j.Execute()
	} else {
		// Acquire the requested number of threads (blocks until available)
		for i := 0; i < threadsToUse; i++ {
			<-w.FreeThreads
		}

		// Execute in multiple goroutines
		var wg sync.WaitGroup
		for i := 0; i < threadsToUse; i++ {
			wg.Add(1)
			go func(threadID int) {
				defer wg.Done()
				j.ExecuteChunk(threadID, threadsToUse)
			}(i)
		}

		wg.Wait() // wait for all threads to finish

		// Release threads back to the pool
		for i := 0; i < threadsToUse; i++ {
			w.FreeThreads <- struct{}{}
		}
	}

	j.Status = job.Completed
}

// Helper to see how many threads are currently free
func (w *Worker) AvailableThreads() int {
	return len(w.FreeThreads)
}

func (w *Worker) Stop() {
	close(w.JobQueue)
	w.WaitGroup.Wait()
}
