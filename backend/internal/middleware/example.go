package middleware

import (
	"net/http"
)

// ExampleMiddleware is a placeholder for your custom middleware.
func ExampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do something before
		next.ServeHTTP(w, r)
		// Do something after
	})
}
