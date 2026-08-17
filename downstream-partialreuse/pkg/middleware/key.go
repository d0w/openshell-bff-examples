package middleware

import "net/http"

// requireDownstreamKey is a downstream-only auth scheme -- a different
// header, a different check, zero relation to upstream's Bearer-token
// middleware.RequireAuth. It exists purely to show that an Option's
// middleware doesn't have to come from, or agree with, upstream at all.
func RequireDownstreamKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Downstream-Key") != "downstream-secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
