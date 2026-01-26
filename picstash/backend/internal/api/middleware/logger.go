package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()

		slog.Info("HTTP请求",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
		)

		return err
	}
}
