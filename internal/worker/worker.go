package worker

import (
	"sync"

	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
)

type Worker struct {
	ID         string
	JobQueue   chan *job.Job
	NumThreads int
	WaitGroup  sync.WaitGroup
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
	j.Execute()

}

func (w *Worker) Stop() {
	close(w.JobQueue)
	w.WaitGroup.Wait()
}

func NewWorker(id string, numThreads int) *Worker {
	return &Worker{
		ID:         id,
		JobQueue:   make(chan *job.Job, 100),
		NumThreads: numThreads,
	}
}
