package middleware

import (
	"fmt"
	"strings"

	"picstash/internal/auth"

	"github.com/gofiber/fiber/v3"
)

func JWTAuth(jwtService *auth.JWTService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "缺少认证token",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "无效的认证格式",
			})
		}

		token := parts[1]
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("无效的token: %v", err),
			})
		}

		c.Locals("email", claims.Email)
		return c.Next()
	}
}
