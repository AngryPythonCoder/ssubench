package domain

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	UserRoleKey    contextKey = "role"
	TransactionKey contextKey = "transaction"
)
