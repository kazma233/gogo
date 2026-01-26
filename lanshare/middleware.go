package main

import (
	"fmt"
	"mime/multipart"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// 文件验证中间件
func FileValidationMiddleware(config *Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() != "POST" || !strings.HasPrefix(c.Path(), "/fs") {
			return c.Next()
		}

		// 检查文件大小
		if int64(c.Request().Header.ContentLength()) > config.MaxFileSize {
			maxSizeMB := config.MaxFileSize / (1024 * 1024)
			return c.Status(413).JSON(fiber.Map{
				"error": fmt.Sprintf("文件太大，最大支持 %d MB", maxSizeMB),
			})
		}

		return c.Next()
	}
}

// 文件类型验证
func ValidateFileType(fh *multipart.FileHeader, allowedTypes []string) error {
	if slices.Contains(allowedTypes, "*") {
		return nil
	}

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("无法确定文件类型，请确保文件格式正确")
	}

	if !slices.Contains(allowedTypes, contentType) {
		return fmt.Errorf("不支持的文件类型: %s", contentType)
	}

	return nil
}

// CORS 中间件
func CORSMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}
