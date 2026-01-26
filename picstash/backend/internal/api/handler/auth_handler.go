package handler

import (
	"fmt"
	"log/slog"

	"picstash/internal/auth"
	"picstash/internal/config"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	emailService        *auth.EmailService
	verificationService *auth.VerificationCodeService
	jwtService          *auth.JWTService
	authConfig          *config.AuthConfig
}

func NewAuthHandler(emailService *auth.EmailService, verificationService *auth.VerificationCodeService, jwtService *auth.JWTService, authConfig *config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		emailService:        emailService,
		verificationService: verificationService,
		jwtService:          jwtService,
		authConfig:          authConfig,
	}
}

func (h *AuthHandler) isEmailAllowed(email string) bool {
	for _, allowed := range h.authConfig.AllowedEmails {
		if email == allowed {
			return true
		}
	}
	return false
}

func (h *AuthHandler) SendCode(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请求参数错误",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "邮箱不能为空",
		})
	}

	if !h.isEmailAllowed(req.Email) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "此邮箱无权登录",
		})
	}

	canSend, err := h.verificationService.CanSendCode(req.Email)
	if err != nil {
		slog.Error("检查发送限制失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "系统错误",
		})
	}

	if !canSend {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "发送频率过快，请稍后再试",
		})
	}

	code, err := h.verificationService.GenerateAndSave(req.Email, c.IP(), 5)
	if err != nil {
		slog.Error("生成验证码失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "系统错误",
		})
	}

	if err := h.emailService.SendVerificationCode(req.Email, code, 5); err != nil {
		slog.Error("发送验证码失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "发送验证码失败",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "验证码已发送",
		"expires_in": 5,
	})
}

func (h *AuthHandler) VerifyCode(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请求参数错误",
		})
	}

	if !h.isEmailAllowed(req.Email) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "此邮箱无权登录",
		})
	}

	valid, err := h.verificationService.Verify(req.Email, req.Code)
	if err != nil {
		slog.Error("验证验证码失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "系统错误",
		})
	}

	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "验证码无效或已过期",
		})
	}

	token, err := h.jwtService.GenerateToken(req.Email)
	if err != nil {
		slog.Error("生成token失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "系统错误",
		})
	}

	return c.JSON(fiber.Map{
		"token":      token,
		"expires_at": fmt.Sprintf("%d", 24*60*60),
	})
}
