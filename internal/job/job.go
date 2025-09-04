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
)

type Job struct {
	ID          string
	Name        string
	Type        JobType
	Status      Status
	Priority    int
	Payload     interface{}
	Result      interface{}
	CreatedAt   time.Time
	CompletedAt time.Time
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
	default:
		j.Status = Failed
		return
	}
	j.Status = Completed
	j.CompletedAt = time.Now()
}
