package domain

import "time"

type TaskStatus string

const (
	StatusPublished  TaskStatus = "published"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusCompleted  TaskStatus = "completed"
	StatusCanceled   TaskStatus = "canceled"
)

type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Reward      int        `json:"reward"`
	Status      TaskStatus `json:"status"`
	CustomerID  int        `json:"customer_id"`
	PerformerID *int       `json:"performer_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
