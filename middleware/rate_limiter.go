package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// NewRateLimiter returns a new rate limiter middleware with the given rate and burst limit.
func NewRateLimiter(rateLimit rate.Limit, burstLimit int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rateLimit, burstLimit)

	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
			return
		}
		// return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded")
		c.JSON(http.StatusTooManyRequests, gin.H{"Message": "Rate Limit Exceeded"})
		c.Abort()
	}
}
