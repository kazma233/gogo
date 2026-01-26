package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"picstash/internal/api"
	"picstash/internal/api/middleware"
	"picstash/internal/auth"
	"picstash/internal/config"
	"picstash/internal/database"
	"picstash/internal/service"
	"picstash/internal/storage"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	if cfg.Upload.ThumbnailFormat != "jpg" && cfg.Upload.ThumbnailFormat != "jpeg" {
		slog.Error("不支持的缩略图格式", "format", cfg.Upload.ThumbnailFormat, "supported", "jpg/jpeg")
		os.Exit(1)
	}

	if len(cfg.Auth.AllowedEmails) == 0 {
		slog.Error("未配置 allowed_emails，必须至少配置一个邮箱才能启动服务")
		os.Exit(1)
	}

	if err := config.InitLogger(cfg); err != nil {
		slog.Error("初始化日志失败", "error", err)
		os.Exit(1)
	}

	if err := database.Init(cfg.Database.Path); err != nil {
		slog.Error("初始化数据库失败", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.AutoMigrate(database.GetDB()); err != nil {
		slog.Error("数据库迁移失败", "error", err)
		os.Exit(1)
	}

	jwtService := auth.NewJWTService(cfg)
	emailService := auth.NewEmailService(cfg)
	verificationService := auth.NewVerificationCodeService(database.GetDB())

	slog.Info("GitHub配置", "owner", cfg.GitHub.Owner, "repo", cfg.GitHub.Repo, "branch", cfg.GitHub.Branch, "token_prefix", cfg.GitHub.Token[:10]+"...")
	storage := storage.NewGitHubStorage(cfg.GitHub.Token, cfg.GitHub.Owner, cfg.GitHub.Repo, cfg.GitHub.Branch)
	imageService := service.NewImageService(database.GetDB(), storage, &cfg.Upload, cfg.GitHub.PathPrefix)
	tagService := service.NewTagService(database.GetDB())

	app := fiber.New(fiber.Config{
		AppName: "Picstash Backend", ErrorHandler: defaultErrorHandler,
		BodyLimit: parseBodySize(cfg.Server.MaxBodySize),
	})

	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	app.Use(middleware.Logger())

	api.SetupRoutes(app, cfg, jwtService, emailService, verificationService, imageService, tagService, database.GetDB())

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		slog.Info("接收到关闭信号，正在关闭服务器...")
		app.Shutdown()
	}()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("服务器启动", "port", cfg.Server.Port, "mode", cfg.Server.Mode)

	if err := app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: false}); err != nil {
		slog.Error("服务器启动失败", "error", err)
		os.Exit(1)
	}
}

func defaultErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	slog.Error("请求错误", "path", c.Path(), "error", err)

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func parseBodySize(size string) int {
	size = strings.TrimSpace(size)
	if size == "" {
		return 100 * 1024 * 1024
	}

	unit := ""
	number := size
	for i := 0; i < len(size); i++ {
		if size[i] >= '0' && size[i] <= '9' {
			continue
		}
		unit = size[i:]
		number = size[:i]
		break
	}

	value, err := strconv.Atoi(number)
	if err != nil {
		return 100 * 1024 * 1024
	}

	switch strings.ToLower(unit) {
	case "kb":
		return value * 1024
	case "mb":
		return value * 1024 * 1024
	case "gb":
		return value * 1024 * 1024 * 1024
	default:
		return value
	}
}
