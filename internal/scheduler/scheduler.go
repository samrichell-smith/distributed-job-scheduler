package scheduler

import (
	"container/heap"
	"sync"

	"github.com/samrichell-smith/distributed-job-scheduler/internal/job"
	"github.com/samrichell-smith/distributed-job-scheduler/internal/worker"
)

// ---------------------
// Job Priority Queue
// ---------------------

type JobQueue []*job.Job

func (jq JobQueue) Len() int { return len(jq) }

func (jq JobQueue) Less(i, j int) bool {
	// Higher priority first, fallback to earlier creation time
	if jq[i].Priority == jq[j].Priority {
		return jq[i].CreatedAt.Before(jq[j].CreatedAt)
	}
	return jq[i].Priority > jq[j].Priority
}

func (jq JobQueue) Swap(i, j int) {
	jq[i], jq[j] = jq[j], jq[i]
}

func (jq *JobQueue) Push(x interface{}) {
	*jq = append(*jq, x.(*job.Job))
}

func (jq *JobQueue) Pop() interface{} {
	old := *jq
	n := len(old)
	item := old[n-1]
	*jq = old[0 : n-1]
	return item
}

// ---------------------
// Scheduler
// ---------------------

type Scheduler struct {
	mu      sync.Mutex
	cond    *sync.Cond
	jobQ    JobQueue
	workers []*worker.Worker
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

// NewScheduler takes a list of worker pointers (manual setup for now)
func NewScheduler(workers []*worker.Worker) *Scheduler {
	s := &Scheduler{
		jobQ:    make(JobQueue, 0),
		workers: workers,
		stopCh:  make(chan struct{}),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// Submit adds a job to the priority queue
func (s *Scheduler) Submit(j *job.Job) {
	s.mu.Lock()
	heap.Push(&s.jobQ, j)
	s.cond.Signal() // wake up a worker loop
	s.mu.Unlock()
}

// Run starts one goroutine per worker
func (s *Scheduler) Run() {
	for _, w := range s.workers {
		s.wg.Add(1)
		go s.workerLoop(w)
	}
}

// workerLoop continuously tries to get jobs and assign them to this worker
func (s *Scheduler) workerLoop(w *worker.Worker) {
	defer s.wg.Done()
	for {
		s.mu.Lock()
		// wait while no jobs available
		for len(s.jobQ) == 0 {
			s.cond.Wait()
			select {
			case <-s.stopCh:
				s.mu.Unlock()
				return
			default:
			}
		}

		// Find a job that this worker has enough threads for
		var j *job.Job
		for i, candidate := range s.jobQ {
			if candidate.ThreadDemand <= w.NumThreads {
				j = candidate
				// Remove from queue
				s.jobQ = append(s.jobQ[:i], s.jobQ[i+1:]...)
				break
			}
		}

		// If no suitable job, wait and retry
		if j == nil {
			s.cond.Wait()
			s.mu.Unlock()
			continue
		}
		s.mu.Unlock()

		// Push the job into the worker's queue
		w.JobQueue <- j
	}
}

// Stop signals all worker loops to exit
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.cond.Broadcast() // wake up all waiting worker loops
	s.wg.Wait()

	// Also stop all workers
	for _, w := range s.workers {
		w.Stop()
	}
}
