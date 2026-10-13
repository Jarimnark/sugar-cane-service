package middleware

import "net/http"

// RateLimit is a pass-through placeholder. TODO: implement per-IP token-bucket limiting.
func RateLimit(next http.Handler) http.Handler {
	return next
}
