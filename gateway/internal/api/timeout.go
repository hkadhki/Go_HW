package api

import (
	"net/http"
	"time"
)

func TimeoutMiddleware() func(http.Handler) http.Handler {
	return TimeoutMiddlewareWithDuration(2 * time.Second)
}

func TimeoutMiddlewareWithDuration(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, `{"error":"Request timeout"}`)
	}
}
