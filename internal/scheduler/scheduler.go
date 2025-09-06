package scheduler

import (
	"container/heap"
	"sync"
	"time"

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

// NewScheduler takes a list of worker pointers
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
	s.cond.Broadcast() // wake up all waiting worker loops
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

		// Wait while no jobs available
		for len(s.jobQ) == 0 {
			s.cond.Wait()
			select {
			case <-s.stopCh:
				s.mu.Unlock()
				return
			default:
			}
		}

		var selectedJob *job.Job
		var index int
		var fallbackSingleThread bool

		// Iterate jobs by priority
		for i, j := range s.jobQ {
			// Check if worker can execute immediately
			if j.ThreadDemand <= w.AvailableThreads() {
				selectedJob = j
				index = i
				break
			}
		}

		if selectedJob == nil && len(s.jobQ) > 0 {
			// No currently free worker can execute any job
			maxThreads := 0
			for _, other := range s.workers {
				if other.NumThreads > maxThreads {
					maxThreads = other.NumThreads
				}
			}

			// Pick highest priority job anyway if no worker can satisfy it
			if s.jobQ[0].ThreadDemand > maxThreads {
				selectedJob = s.jobQ[0]
				index = 0
				fallbackSingleThread = true
			} else {
				// Wait until threads become free
				s.cond.Wait()
				s.mu.Unlock()
				continue
			}
		}

		// Remove job from queue
		s.jobQ = append(s.jobQ[:index], s.jobQ[index+1:]...)
		// Set started_at timestamp if not already set
		if selectedJob.StartedAt.IsZero() {
			selectedJob.StartedAt = time.Now()
		}
		s.mu.Unlock()

		if fallbackSingleThread {
			// Run single-thread fallback
			selectedJob.ThreadDemand = 1
		}

		select {
		case w.JobQueue <- selectedJob:
		case <-s.stopCh:
			return
		}
	}
}

// Stop signals all worker loops to exit and stops workers
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.cond.Broadcast() // wake up all waiting worker loops
	s.wg.Wait()

	for _, w := range s.workers {
		w.Stop()
	}
}

// Optional: helper to wait for all jobs to complete (for testing)
func (s *Scheduler) WaitAllJobsDone(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		s.mu.Lock()
		if len(s.jobQ) == 0 {
			s.mu.Unlock()
			return true
		}
		s.mu.Unlock()
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(5 * time.Millisecond)
	}
}
