package job

import "time"

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

func (j *Job) Execute() {
	j.Status = Pending
	switch j.Type {
	case AddNumbersJob:
		payload := j.Payload.(AddNumbersPayload)
		j.Result = AddNumbersResult{Sum: payload.X + payload.Y}

	case ReverseStringJob:
		payload := j.Payload.(ReverseStringPayload)
		j.Result = ReverseStringResult{Reversed: reverse(payload.Text)}

	case ResizeImageJob:
		payload := j.Payload.(ResizeImagePayload)
		resized := ResizeImage(payload.URL, payload.Width, payload.Height) // call helper
		j.Result = ResizeImageResult{ResizedURL: resized}

	case LargeArraySumJob:
		// fallback if called single threaded
		payload := j.Payload.(LargeArraySumPayload)
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

	payload := j.Payload.(LargeArraySumPayload)
	n := len(payload.Array)
	chunkSize := n / totalThreads
	start := threadID * chunkSize
	end := start + chunkSize
	if threadID == totalThreads-1 {
		end = n // last thread handles remainder
	}

	localSum := 0
	for _, v := range payload.Array[start:end] {
		localSum += v
	}

	j.addPartialSum(localSum)
}
