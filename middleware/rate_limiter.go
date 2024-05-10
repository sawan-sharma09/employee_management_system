package middleware

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

// NewRateLimiter returns a new rate limiter middleware with the given rate and burst limit.
func NewRateLimiter(rateLimit rate.Limit, burstLimit int) fiber.Handler {
	limiter := rate.NewLimiter(rateLimit, burstLimit)

	return func(c *fiber.Ctx) error {
		if limiter.Allow() {
			return c.Next()
		}
		return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded")
	}
}
