-- 同步任务日志表
CREATE TABLE IF NOT EXISTS sync_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    triggered_by TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    status TEXT DEFAULT 'running',
    total_files INTEGER DEFAULT 0,
    processed_files INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    error_message TEXT
);

-- 文件同步详情表
CREATE TABLE IF NOT EXISTS sync_file_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sync_log_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    action TEXT NOT NULL,
    status TEXT NOT NULL,
    sha TEXT,
    old_sha TEXT,
    size INTEGER,
    old_size INTEGER,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sync_log_id) REFERENCES sync_logs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sync_file_logs_sync_log_id ON sync_file_logs(sync_log_id);
CREATE INDEX IF NOT EXISTS idx_sync_file_logs_path ON sync_file_logs(path);
CREATE INDEX IF NOT EXISTS idx_sync_file_logs_created_at ON sync_file_logs(created_at DESC);
