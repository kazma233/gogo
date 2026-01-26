package api

import (
	"database/sql"

	"picstash/internal/api/handler"
	"picstash/internal/api/middleware"
	"picstash/internal/auth"
	"picstash/internal/config"
	"picstash/internal/service"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtService *auth.JWTService,
	emailService *auth.EmailService,
	verificationService *auth.VerificationCodeService,
	imageService *service.ImageService,
	tagService *service.TagService,
	db *sql.DB,
) {
	authHandler := handler.NewAuthHandler(emailService, verificationService, jwtService, &cfg.Auth)
	imageHandler := handler.NewImageHandler(imageService)
	tagHandler := handler.NewTagHandler(tagService)
	syncLogHandler := handler.NewSyncLogHandler(db)

	api := app.Group("/api")

	public := api.Group("")
	public.Get("/images", imageHandler.GetList)
	public.Get("/images/:id", imageHandler.GetByID)
	public.Get("/tags", tagHandler.GetAll)
	public.Get("/tags/:id/images", tagHandler.GetByImageID)
	public.Get("/sync/logs", syncLogHandler.GetList)
	public.Get("/sync/logs/:id", syncLogHandler.GetByID)
	public.Get("/sync/logs/:id/files", syncLogHandler.GetFileLogs)

	authGroup := api.Group("/auth")
	authGroup.Post("/send-code", authHandler.SendCode)
	authGroup.Post("/verify", authHandler.VerifyCode)

	protected := api.Group("")
	protected.Use(middleware.JWTAuth(jwtService))
	protected.Post("/images/upload", imageHandler.Upload)
	protected.Delete("/images/:id", imageHandler.Delete)
	protected.Put("/images/:id/tags", imageHandler.UpdateTags)
	protected.Post("/tags", tagHandler.Create)
	protected.Put("/tags/:id", tagHandler.Update)
	protected.Delete("/tags/:id", tagHandler.Delete)
	protected.Post("/images/sync", imageHandler.SyncFromStorage)
}
