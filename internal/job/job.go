package job

import (
	"sync"
	"time"
)

type Status string

const (
	Pending   Status = "Pending"
	Running   Status = "Running"
	Completed Status = "Completed"
	Failed    Status = "Failed"
)

type JobType string

const (
	AddNumbersJob    JobType = "AddNumbers"
	ReverseStringJob JobType = "ReverseString"
	ResizeImageJob   JobType = "ResizeImage"
	LargeArraySumJob JobType = "LargeArraySum"
)

type Job struct {
	ID           string
	Name         string
	Type         JobType
	Status       Status
	Priority     int
	Payload      interface{}
	Result       interface{}
	resultMu     sync.Mutex
	CreatedAt    time.Time
	StartedAt    time.Time
	CompletedAt  time.Time
	ThreadDemand int
}

func NewJob(id, name string, jobType JobType, priority int, payload interface{}) *Job {
	return &Job{
		ID:        id,
		Name:      name,
		Type:      jobType,
		Status:    Pending,
		Priority:  priority,
		Payload:   payload,
		CreatedAt: time.Now(),
	}
}

// probably want to abstract this more in the fututre, so we don't have to hard code job definition cases into here
func (j *Job) Execute() {
	// mark as running and set StartedAt if not set
	j.Status = Running
	if j.StartedAt.IsZero() {
		j.StartedAt = time.Now()
	}
	switch j.Type {
	case AddNumbersJob:
		payload, ok := j.Payload.(AddNumbersPayload)
		if !ok {
			j.Status = Failed
			j.Result = AddNumbersResult{Sum: 0}
			j.CompletedAt = time.Now()
			return
		}
		j.Result = AddNumbersResult{Sum: payload.X + payload.Y}

	case ReverseStringJob:
		payload, ok := j.Payload.(ReverseStringPayload)
		if !ok {
			j.Status = Failed
			j.Result = ReverseStringResult{Reversed: ""}
			j.CompletedAt = time.Now()
			return
		}
		j.Result = ReverseStringResult{Reversed: reverse(payload.Text)}

	case ResizeImageJob:
		payload, ok := j.Payload.(ResizeImagePayload)
		if !ok {
			j.Status = Failed
			j.Result = ResizeImageResult{ResizedURL: ""}
			j.CompletedAt = time.Now()
			return
		}
		resized := ResizeImage(payload.URL, payload.Width, payload.Height) // call helper
		j.Result = ResizeImageResult{ResizedURL: resized}

	case LargeArraySumJob:
		// fallback if called single threaded
		payload, ok := j.Payload.(LargeArraySumPayload)
		if !ok {
			j.Status = Failed
			j.Result = LargeArraySumResult{Sum: 0}
			j.CompletedAt = time.Now()
			return
		}
		sum := 0
		for _, v := range payload.Array {
			sum += v
		}
		j.Result = LargeArraySumResult{Sum: sum}

	default:
		j.Status = Failed
		return
	}
	j.Status = Completed
	j.CompletedAt = time.Now()
}

func (j *Job) ExecuteChunk(threadID, totalThreads int) {
	if j.Type != LargeArraySumJob {
		return // other jobs do nothing
	}

	payload, ok := j.Payload.(LargeArraySumPayload)
	if !ok {
		return
	}
	n := len(payload.Array)
	if n == 0 || totalThreads <= 0 || threadID < 0 || threadID >= totalThreads {
		return
	}
	// Use integer-math partitioning that works when totalThreads > n
	start := threadID * n / totalThreads
	end := (threadID + 1) * n / totalThreads
	if start >= end {
		return
	}
	localSum := 0
	for i := start; i < end; i++ {
		localSum += payload.Array[i]
	}
	j.addPartialSum(localSum)
}
