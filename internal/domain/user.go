package domain

import "time"

type UserRole string

const (
	RoleCustomer  UserRole = "customer"
	RolePerformer UserRole = "performer"
	RoleAdmin     UserRole = "admin"
)

type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusBlocked UserStatus = "blocked"
)

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Role         UserRole   `json:"role"`
	Status       UserStatus `json:"status"`
	Balance      int        `json:"balance"`
	CreatedAt    time.Time  `json:"created_at"`
}
