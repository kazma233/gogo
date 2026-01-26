package handler

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"

	"picstash/internal/model"
	"picstash/internal/repository"

	"github.com/gofiber/fiber/v3"
)

type SyncLogHandler struct {
	db *sql.DB
}

func NewSyncLogHandler(db *sql.DB) *SyncLogHandler {
	return &SyncLogHandler{db: db}
}

func (h *SyncLogHandler) GetList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	tx, err := h.db.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "开始事务失败",
		})
	}
	defer tx.Rollback()

	syncLogRepo := repository.NewSyncLogRepository(tx)
	offset := (page - 1) * limit

	logs, err := syncLogRepo.GetAllSyncLogs(limit, offset)
	if err != nil {
		slog.Error("获取同步日志失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取同步日志失败",
		})
	}

	total, err := syncLogRepo.GetSyncLogCount()
	if err != nil {
		slog.Error("获取同步日志总数失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取同步日志总数失败",
		})
	}

	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *SyncLogHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的同步日志ID",
		})
	}

	tx, err := h.db.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "开始事务失败",
		})
	}
	defer tx.Rollback()

	syncLogRepo := repository.NewSyncLogRepository(tx)

	log, err := h.getSyncLog(tx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "同步日志不存在",
		})
	}

	fileLogs, err := syncLogRepo.GetSyncFilesByLogID(id)
	if err != nil {
		slog.Error("获取文件同步日志失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取文件同步日志失败",
		})
	}

	return c.JSON(fiber.Map{
		"log":       log,
		"file_logs": fileLogs,
	})
}

func (h *SyncLogHandler) GetFileLogs(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的同步日志ID",
		})
	}

	tx, err := h.db.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "开始事务失败",
		})
	}
	defer tx.Rollback()

	syncLogRepo := repository.NewSyncLogRepository(tx)

	_, err = h.getSyncLog(tx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "同步日志不存在",
		})
	}

	fileLogs, err := syncLogRepo.GetSyncFilesByLogID(id)
	if err != nil {
		slog.Error("获取文件同步日志失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取文件同步日志失败",
		})
	}

	return c.JSON(fiber.Map{
		"data": fileLogs,
	})
}

func (h *SyncLogHandler) getSyncLog(tx *sql.Tx, id int64) (*model.SyncLog, error) {
	log := &model.SyncLog{}
	err := tx.QueryRow(`
		SELECT id, triggered_by, started_at, completed_at, status, total_files, processed_files, error_count, error_message
		FROM sync_logs
		WHERE id = ?
	`, id).Scan(
		&log.ID, &log.TriggeredBy, &log.StartedAt, &log.CompletedAt,
		&log.Status, &log.TotalFiles, &log.ProcessedFiles, &log.ErrorCount, &log.ErrorMessage,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("同步日志不存在")
		}
		return nil, fmt.Errorf("查询同步日志失败: %w", err)
	}
	return log, nil
}
