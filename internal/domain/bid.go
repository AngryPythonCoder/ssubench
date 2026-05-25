package domain

import "time"

type Bid struct {
	ID          int       `json:"id"`
	TaskID      int       `json:"task_id"`
	PerformerID int       `json:"performer_id"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}
