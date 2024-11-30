package middleware

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/util"
)

func RateLimiter() gin.HandlerFunc {
	maxConnectionString := os.Getenv("LIMITER_MAX_REQUEST_PER_SECOND")
	maxConnection, err := strconv.Atoi(maxConnectionString)
	if err != nil {
		maxConnection = 10
	}

	return func(c *gin.Context) {

		limiter := rate.NewLimiter(rate.Limit(maxConnection/3600.0), 100)

		if limiter.Allow() {
			c.Next()
			return
		}

		util.Log("ERROR", "middleware", "RateLimiter", "limit exceeed")

		c.Set(constant.ERROR_KEY, constant.ErrorTooManyRequest)
		c.Abort()
	}
}

func RateLimitterPerClient() gin.HandlerFunc {
	type client struct {
		Limiter  *rate.Limiter
		LastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.LastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	maxConnectionString := os.Getenv("LIMITER_MAX_REQUEST_PER_SECOND")
	maxConnection, err := strconv.Atoi(maxConnectionString)
	if err != nil {
		maxConnection = 10
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Lock the mutex to protect this section from race conditions.
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{Limiter: rate.NewLimiter(rate.Limit(maxConnection), 100)}
		}

		clients[ip].LastSeen = time.Now()

		if !clients[ip].Limiter.Allow() {
			// check remaining time
			reservation := clients[ip].Limiter.Reserve()
			remainingTime := reservation.DelayFrom(time.Now())
			util.Log("INFO", "middleware", "RateLimiter", "remaining time: "+fmt.Sprintf("%d ms", remainingTime.Milliseconds()))

			util.Log("ERROR", "middleware", "RateLimiter", "limit exceeed")

			mu.Unlock()

			c.Set(constant.ERROR_KEY, constant.ErrorTooManyRequest)
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()

	}
}
