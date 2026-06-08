package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		log.Printf("[%s] %s - %v | %d", c.Method(), c.Path(), duration, c.Response().StatusCode())

		return err
	}
}
