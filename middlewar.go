package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func getLimiters(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		host = ip
	}

	if limiter, ok := limiters[host]; ok {
		return limiter
	}

	limiter := rate.NewLimiter(1, 1)
	limiters[host] = limiter

	return limiter

}
func rateLimiterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limiter := getLimiters(r.RemoteAddr)

		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, "Too Many Requests")
			return
		}
		next(w, r)
	}
}
