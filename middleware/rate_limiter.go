package middleware

import (
	"managedata/app_errors"
	"net/http"
	"time"

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
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "WARNING", Message: app_errors.ErrRateLimitExceeded, Endpoint: c.Request.URL.Path, Status_code: http.StatusTooManyRequests}
		c.AbortWithStatusJSON(http.StatusTooManyRequests, logDetails)

	}
}
