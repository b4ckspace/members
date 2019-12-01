package web

import (
	"log"
	"net/http"
	"time"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf(
			"%s %s",
			r.URL.String(),
			duration.String(),
		)
	})
}
