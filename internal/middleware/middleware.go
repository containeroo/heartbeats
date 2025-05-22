package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler with additional functionality.
type Middleware func(http.Handler) http.Handler

// Chain applies a list of middleware in order to an http.Handler.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
