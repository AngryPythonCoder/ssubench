package domain

import "time"

type Payment struct {
	ID          int       `json:"id"`
	TaskID      int       `json:"task_id"`
	CustomerID  int       `json:"customer_id"`
	PerformerID int       `json:"performer_id"`
	Amount      int       `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
