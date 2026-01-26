package model

import "time"

type Image struct {
	ID               int64      `json:"id"`
	Path             string     `json:"path"`
	URL              string     `json:"url"`
	SHA              string     `json:"sha,omitempty"`
	ThumbnailPath    string     `json:"thumbnail_path,omitempty"`
	ThumbnailURL     string     `json:"thumbnail_url,omitempty"`
	ThumbnailSHA     string     `json:"thumbnail_sha,omitempty"`
	ThumbnailSize    *int64     `json:"thumbnail_size,omitempty"`
	WatermarkPath    string     `json:"watermark_path,omitempty"`
	WatermarkURL     string     `json:"watermark_url,omitempty"`
	WatermarkSHA     string     `json:"watermark_sha,omitempty"`
	WatermarkSize    *int64     `json:"watermark_size,omitempty"`
	OriginalFilename string     `json:"original_filename"`
	Filename         string     `json:"filename"`
	Size             *int64     `json:"size"`
	MimeType         string     `json:"mime_type"`
	Width            *int       `json:"width,omitempty"`
	Height           *int       `json:"height,omitempty"`
	HasThumbnail     bool       `json:"has_thumbnail"`
	HasWatermark     bool       `json:"has_watermark"`
	UploadedAt       time.Time  `json:"uploaded_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	Tags             []Tag      `json:"tags,omitempty"`
}

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type ImageTag struct {
	ImageID   int64     `json:"image_id"`
	TagID     int64     `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

type VerificationCode struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Code      string     `json:"code"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type AppConfig struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Migration struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Checksum   string    `json:"checksum"`
	ExecutedAt time.Time `json:"executed_at"`
}

type SyncLog struct {
	ID             int64      `json:"id"`
	TriggeredBy    string     `json:"triggered_by"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Status         string     `json:"status"`
	TotalFiles     int        `json:"total_files"`
	ProcessedFiles int        `json:"processed_files"`
	ErrorCount     int        `json:"error_count"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
}

type SyncFileLog struct {
	ID           int64     `json:"id"`
	SyncLogID    int64     `json:"sync_log_id"`
	Path         string    `json:"path"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	SHA          *string   `json:"sha,omitempty"`
	OldSHA       *string   `json:"old_sha,omitempty"`
	Size         *int64    `json:"size,omitempty"`
	OldSize      *int64    `json:"old_size,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
