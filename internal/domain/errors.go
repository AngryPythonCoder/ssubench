package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrInsufficientFunds  = errors.New("insufficient funds to complete operation")

	ErrUserNotFound                = errors.New("user not found")
	ErrUserAlreadyExists           = errors.New("user already exists")
	ErrUserInvalidStatusTransition = errors.New("invalid status change of a user")

	ErrTaskNotFound                = errors.New("task not found")
	ErrTaskAlreadyExists           = errors.New("task already exists")
	ErrTaskInvalidStatusTransition = errors.New("invalid status change of a task")

	ErrBidNotFound      = errors.New("bid not found")
	ErrBidAlreadyExists = errors.New("bid already exists")

	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists")
)
