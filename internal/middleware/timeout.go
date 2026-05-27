package middleware

import (
	"net/http"
	"time"
)

func StandardTimeout(durationTime time.Duration, message string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, durationTime, message)
	}
}
