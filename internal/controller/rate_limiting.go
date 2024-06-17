package controller

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*client),
	}

	go rl.cleanupClients()
	return rl
}

func (rl *rateLimiter) getClient(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if c, exists := rl.clients[ip]; exists {
		c.lastSeen = time.Now()
		return c.limiter
	}

	limiter := rate.NewLimiter(rate.Limit(1), 5) // 1 request per second, burst of 5
	rl.clients[ip] = &client{limiter, time.Now()}
	return limiter
}

func (rl *rateLimiter) cleanupClients() {
	for {
		time.Sleep(time.Minute)

		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
