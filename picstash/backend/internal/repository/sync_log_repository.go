package repository

import (
	"database/sql"
	"fmt"
	"time"

	"picstash/internal/model"
)

type SyncLogRepository struct {
	tx *sql.Tx
}

func NewSyncLogRepository(tx *sql.Tx) *SyncLogRepository {
	return &SyncLogRepository{tx: tx}
}

func (r *SyncLogRepository) CreateSyncLog(triggeredBy string, startedAt time.Time) (int64, error) {
	result, err := r.tx.Exec(`
		INSERT INTO sync_logs (triggered_by, started_at, status, total_files, processed_files, error_count)
		VALUES (?, ?, 'running', 0, 0, 0)
	`, triggeredBy, startedAt)
	if err != nil {
		return 0, fmt.Errorf("创建同步日志失败: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取同步日志ID失败: %w", err)
	}
	return id, nil
}

func (r *SyncLogRepository) UpdateSyncLog(id int64, completedAt *time.Time, totalFiles, processedFiles, errorCount int, errorMessage *string, status string) error {
	query := `
		UPDATE sync_logs
		SET completed_at = ?, total_files = ?, processed_files = ?, error_count = ?, error_message = ?, status = ?
		WHERE id = ?
	`
	_, err := r.tx.Exec(query, completedAt, totalFiles, processedFiles, errorCount, errorMessage, status, id)
	if err != nil {
		return fmt.Errorf("更新同步日志失败: %w", err)
	}
	return nil
}

func (r *SyncLogRepository) CreateSyncFileLog(syncLogID int64, path, action, status string, sha, oldSha *string, size, oldSize *int64, errorMessage *string) error {
	query := `
		INSERT INTO sync_file_logs (sync_log_id, path, action, status, sha, old_sha, size, old_size, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.tx.Exec(query, syncLogID, path, action, status, sha, oldSha, size, oldSize, errorMessage)
	if err != nil {
		return fmt.Errorf("创建文件同步日志失败: %w", err)
	}
	return nil
}

func (r *SyncLogRepository) GetSyncFilesByLogID(syncLogID int64) ([]*model.SyncFileLog, error) {
	query := `
		SELECT id, sync_log_id, path, action, status, sha, old_sha, size, old_size, error_message, created_at
		FROM sync_file_logs
		WHERE sync_log_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.tx.Query(query, syncLogID)
	if err != nil {
		return nil, fmt.Errorf("查询文件同步日志失败: %w", err)
	}
	defer rows.Close()

	var logs []*model.SyncFileLog
	for rows.Next() {
		log := &model.SyncFileLog{}
		err := rows.Scan(
			&log.ID, &log.SyncLogID, &log.Path, &log.Action, &log.Status,
			&log.SHA, &log.OldSHA, &log.Size, &log.OldSize, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描文件同步日志失败: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *SyncLogRepository) GetAllSyncLogs(limit, offset int) ([]*model.SyncLog, error) {
	query := `
		SELECT id, triggered_by, started_at, completed_at, status, total_files, processed_files, error_count, error_message
		FROM sync_logs
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.tx.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询同步日志失败: %w", err)
	}
	defer rows.Close()

	var logs []*model.SyncLog
	for rows.Next() {
		log := &model.SyncLog{}
		err := rows.Scan(
			&log.ID, &log.TriggeredBy, &log.StartedAt, &log.CompletedAt,
			&log.Status, &log.TotalFiles, &log.ProcessedFiles, &log.ErrorCount, &log.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描同步日志失败: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *SyncLogRepository) GetSyncLogCount() (int, error) {
	var count int
	err := r.tx.QueryRow("SELECT COUNT(*) FROM sync_logs").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询同步日志总数失败: %w", err)
	}
	return count, nil
}
