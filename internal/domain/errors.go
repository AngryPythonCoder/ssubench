package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrNotAuthorized = errors.New("not authorized")

	ErrBidAlreadyExists = errors.New("bid already exists")
	ErrTaskNotFound     = errors.New("task not found")
)
