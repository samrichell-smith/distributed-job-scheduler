package job

import "time"

type Status string

const (
	Pending   Status = "Pending"
	Running   Status = "Running"
	Completed Status = "Completed"
	Failed    Status = "Failed"
)

type Job struct {
	ID          string
	Name        string
	Status      Status
	Priority    int
	Payload     interface{}
	Result      interface{}
	CreatedAt   time.Time
	CompletedAt time.Time
}
