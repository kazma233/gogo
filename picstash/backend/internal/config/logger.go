package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func InitLogger(cfg *Config) error {
	level := parseLogLevel(cfg.Log.Level)

	if cfg.Server.Mode == "debug" {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}))
		slog.SetDefault(logger)
		slog.Info("日志已初始化", "level", cfg.Log.Level, "output", "stdout")
		return nil
	}

	if err := os.MkdirAll(cfg.Log.Path, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(cfg.Log.Path, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	var handler slog.Handler
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("日志已初始化", "level", cfg.Log.Level, "format", cfg.Log.Format, "path", cfg.Log.Path)

	return nil
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
