package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strings"
	"time"

	"picstash/internal/config"
	"picstash/internal/model"
	"picstash/internal/repository"
	"picstash/internal/storage"

	"github.com/google/uuid"
)

type ImageService struct {
	db         *sql.DB
	storage    storage.Storage
	cfg        *config.UploadConfig
	pathPrefix string
}

type SyncResult struct {
	CreatedCount int
	UpdatedCount int
	DeletedCount int
	SkippedCount int
	ErrorCount   int
	ErrorMessage string
	LogID        int64
}

func NewImageService(db *sql.DB, storage storage.Storage, cfg *config.UploadConfig, pathPrefix string) *ImageService {
	return &ImageService{
		db:         db,
		storage:    storage,
		cfg:        cfg,
		pathPrefix: pathPrefix,
	}
}

func (s *ImageService) BatchUploadWithMapping(
	ctx context.Context,
	originalFiles map[string][]byte,
	watermarkFiles map[string][]byte,
	thumbnailFiles map[string][]byte,
	mapping []FileMapping,
	tagIDs []int,
) ([]*model.Image, error) {
	images := make([]*model.Image, 0, len(mapping))
	slog.Info("开始批量上传（带映射）", "mapping_count", len(mapping), "original_count", len(originalFiles), "watermark_count", len(watermarkFiles), "thumbnail_count", len(thumbnailFiles))

	for _, m := range mapping {
		var mainContent []byte
		var mainFilename string
		var originalName string

		if content, ok := originalFiles[m.Original]; ok {
			mainContent = content
			mainFilename = m.Original
			originalName = m.OriginalName
		} else {
			slog.Error("未找到原图", "original", m.Original)
			continue
		}

		var watermarkContent []byte
		if m.Watermark != nil {
			if content, ok := watermarkFiles[*m.Watermark]; ok && len(content) > 0 {
				watermarkContent = content
				slog.Info("找到水印图", "watermark", *m.Watermark)
			}
		}

		var thumbnailContent []byte
		if m.Thumbnail != nil {
			if thumb, ok := thumbnailFiles[*m.Thumbnail]; ok && len(thumb) > 0 {
				thumbnailContent = thumb
				slog.Info("使用传入的缩略图", "thumbnail", *m.Thumbnail)
			}
		}

		image, err := s.UploadWithContent(ctx, mainFilename, mainContent, originalName, watermarkContent, tagIDs, thumbnailContent)
		if err != nil {
			slog.Error("上传失败", "filename", mainFilename, "error", err)
			continue
		}

		images = append(images, image)
	}

	slog.Info("批量上传完成", "success_count", len(images))
	return images, nil
}

type FileMapping struct {
	Original     string  `json:"original"`
	OriginalName string  `json:"original_name"`
	Watermark    *string `json:"watermark"`
	Thumbnail    *string `json:"thumbnail"`
}

func (s *ImageService) UploadWithContent(
	ctx context.Context,
	filename string,
	content []byte,
	originalFilename string,
	watermarkContent []byte,
	tagIDs []int,
	thumbnailContent []byte,
) (*model.Image, error) {
	slog.Info("开始上传图片", "filename", filename, "size", len(content))

	now := time.Now()
	year, month, day := now.Date()
	uuidStr := strings.ReplaceAll(uuid.New().String(), "-", "")

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	newFilename := fmt.Sprintf("%04d/%02d/%02d/%s%s", year, month, day, uuidStr, ext)
	storagePath := path.Join(s.pathPrefix, newFilename)

	slog.Info("生成文件路径", "filename", filename, "path", storagePath)

	width, height, mimeType, err := storage.GetImageInfo(content)
	if err != nil {
		slog.Warn("获取图片信息失败，使用默认值", "error", err)
		width, height, mimeType = 0, 0, "image/jpeg"
	}

	slog.Info("图片信息", "width", width, "height", height, "mime_type", mimeType)

	if mimeType == "" {
		mimeType = storage.GetMimeType(filename)
		slog.Warn("无法检测MIME类型，使用扩展名判断", "filename", filename, "mime_type", mimeType)
	}

	uploadFile := &storage.File{
		Path:        storagePath,
		Content:     content,
		ContentType: mimeType,
	}

	slog.Info("开始上传到存储", "path", storagePath)
	uploadResult, err := s.storage.Upload(ctx, uploadFile)
	if err != nil {
		return nil, fmt.Errorf("上传图片失败: %w", err)
	}
	slog.Info("上传成功", "path", storagePath, "sha", uploadResult.SHA)

	hasWatermark := len(watermarkContent) > 0

	image := &model.Image{
		Path:             uploadResult.Path,
		URL:              uploadResult.URL,
		SHA:              uploadResult.SHA,
		OriginalFilename: originalFilename,
		Filename:         filename,
		Size:             int64Ptr(int64(len(content))),
		MimeType:         mimeType,
		Width:            intPtr(width),
		Height:           intPtr(height),
		HasThumbnail:     false,
		HasWatermark:     hasWatermark,
		UploadedAt:       time.Now(),
	}

	if hasWatermark {
		now := time.Now()
		year, month, day := now.Date()
		watermarkFilename := fmt.Sprintf("%04d/%02d/%02d/watermark_%s%s", year, month, day, uuidStr, ext)
		watermarkStoragePath := path.Join(s.pathPrefix, watermarkFilename)

		slog.Info("开始上传水印图", "path", watermarkStoragePath, "size", len(watermarkContent))
		watermarkFile := &storage.File{
			Path:        watermarkStoragePath,
			Content:     watermarkContent,
			ContentType: mimeType,
		}

		watermarkResult, watermarkUploadErr := s.storage.Upload(ctx, watermarkFile)
		if watermarkUploadErr != nil {
			slog.Warn("上传水印图失败", "error", watermarkUploadErr)
		} else {
			slog.Info("水印图上传成功", "path", watermarkStoragePath)
			image.WatermarkPath = watermarkResult.Path
			image.WatermarkURL = watermarkResult.URL
			image.WatermarkSHA = watermarkResult.SHA
			image.WatermarkSize = int64Ptr(int64(len(watermarkContent)))
		}
	}

	slog.Info("开始处理缩略图", "original_width", width, "original_height", height)

	if thumbnailContent != nil && len(thumbnailContent) > 0 {
		slog.Info("使用前端传入的缩略图", "size", len(thumbnailContent))
		thumbData := thumbnailContent
		thumbWidth, thumbHeight, _, thumbErr := storage.GetImageInfo(thumbData)
		if thumbErr != nil {
			slog.Warn("获取缩略图信息失败", "error", thumbErr)
		} else {
			now := time.Now()
			year, month, day := now.Date()
			thumbFilename := fmt.Sprintf("%04d/%02d/%02d/thumb_%s.jpg", year, month, day, uuidStr)
			thumbStoragePath := path.Join(s.pathPrefix, thumbFilename)

			slog.Info("开始上传缩略图", "path", thumbStoragePath, "size", len(thumbData))
			thumbFile := &storage.File{
				Path:        thumbStoragePath,
				Content:     thumbData,
				ContentType: "image/jpeg",
			}

			thumbResult, thumbUploadErr := s.storage.Upload(ctx, thumbFile)
			if thumbUploadErr != nil {
				slog.Warn("上传缩略图失败", "error", thumbUploadErr)
			} else {
				slog.Info("缩略图上传成功", "path", thumbStoragePath)
				image.ThumbnailPath = thumbResult.Path
				image.ThumbnailURL = thumbResult.URL
				image.ThumbnailSHA = thumbResult.SHA
				image.ThumbnailSize = int64Ptr(int64(len(thumbData)))
				image.HasThumbnail = true
				image.Width = intPtr(thumbWidth)
				image.Height = intPtr(thumbHeight)
			}
		}
	} else if content != nil && len(content) > 0 {
		thumbData, thumbWidth, thumbHeight, hasThumb, thumbErr := storage.GenerateThumbnail(content, storage.ThumbnailConfig{
			Width:   s.cfg.ThumbnailWidth,
			Quality: s.cfg.ThumbnailQuality,
			Format:  s.cfg.ThumbnailFormat,
		})

		if thumbErr != nil {
			if !(&storage.ErrUnsupportedFormat{}).Is(thumbErr) {
				slog.Warn("生成缩略图失败", "error", thumbErr)
			} else {
				slog.Warn("不支持的格式，跳过缩略图生成", "mime_type", mimeType)
			}
		} else if hasThumb {
			now := time.Now()
			year, month, day := now.Date()
			thumbFilename := fmt.Sprintf("%04d/%02d/%02d/thumb_%s.webp", year, month, day, uuidStr)
			thumbStoragePath := path.Join(s.pathPrefix, thumbFilename)

			slog.Info("开始上传缩略图", "path", thumbStoragePath, "size", len(thumbData))
			thumbFile := &storage.File{
				Path:        thumbStoragePath,
				Content:     thumbData,
				ContentType: "image/webp",
			}

			thumbResult, thumbUploadErr := s.storage.Upload(ctx, thumbFile)
			if thumbUploadErr != nil {
				slog.Warn("上传缩略图失败", "error", thumbUploadErr)
			} else {
				slog.Info("缩略图上传成功", "path", thumbStoragePath)
				image.ThumbnailPath = thumbResult.Path
				image.ThumbnailURL = thumbResult.URL
				image.ThumbnailSHA = thumbResult.SHA
				image.ThumbnailSize = int64Ptr(int64(len(thumbData)))
				image.HasThumbnail = true
				image.Width = intPtr(thumbWidth)
				image.Height = intPtr(thumbHeight)
			}
		}
	}

	slog.Info("开始保存图片记录到数据库", "filename", filename)
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	imageRepo := repository.NewImageRepository(tx)

	id, err := imageRepo.CreateImage(image)
	if err != nil {
		return nil, err
	}
	image.ID = id

	slog.Info("准备关联标签", "image_id", id, "tag_count", len(tagIDs), "tag_ids", tagIDs)
	if len(tagIDs) > 0 {
		err = imageRepo.AddImageTags(id, tagIDs)
		if err != nil {
			return nil, err
		}
		slog.Info("标签关联完成", "image_id", id, "count", len(tagIDs))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("图片上传成功", "id", id, "filename", filename, "has_thumbnail", image.HasThumbnail, "has_watermark", image.HasWatermark)

	return image, nil
}

func (s *ImageService) Delete(ctx context.Context, id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	imageRepo := repository.NewImageRepository(tx)

	image, err := imageRepo.GetImageByID(id)
	if err != nil {
		return err
	}

	if image.Path != "" {
		if err := s.storage.Delete(ctx, image.Path); err != nil {
			slog.Warn("删除原图失败", "path", image.Path, "error", err)
		}
	}

	if image.ThumbnailPath != "" {
		if err := s.storage.Delete(ctx, image.ThumbnailPath); err != nil {
			slog.Warn("删除缩略图失败", "path", image.ThumbnailPath, "error", err)
		}
	}

	if image.WatermarkPath != "" {
		if err := s.storage.Delete(ctx, image.WatermarkPath); err != nil {
			slog.Warn("删除水印图失败", "path", image.WatermarkPath, "error", err)
		}
	}

	err = imageRepo.SoftDeleteImage(id)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("图片删除成功", "id", id, "path", image.Path, "thumbnail_path", image.ThumbnailPath, "watermark_path", image.WatermarkPath)

	return nil
}

func (s *ImageService) GetList(page, limit int, tagID *int) ([]*model.Image, int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, 0, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	imageRepo := repository.NewImageRepository(tx)

	images, total, err := imageRepo.GetImages(page, limit, tagID)
	if err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

func (s *ImageService) GetByID(id int64) (*model.Image, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	imageRepo := repository.NewImageRepository(tx)

	image, err := imageRepo.GetImageByID(id)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *ImageService) UpdateTags(id int64, tagIDs []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	imageRepo := repository.NewImageRepository(tx)

	err = imageRepo.UpdateImageTags(id, tagIDs)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("更新图片标签成功", "image_id", id, "tag_count", len(tagIDs))

	return nil
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func (s *ImageService) SyncFromStorage(ctx context.Context, triggeredBy string) (*SyncResult, error) {
	slog.Info("开始从存储同步", "triggered_by", triggeredBy)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	imageRepo := repository.NewImageRepository(tx)
	syncLogRepo := repository.NewSyncLogRepository(tx)

	now := time.Now()
	logID, err := syncLogRepo.CreateSyncLog(triggeredBy, now)
	if err != nil {
		return nil, fmt.Errorf("创建同步日志失败: %w", err)
	}
	slog.Info("创建同步日志", "log_id", logID)

	var storageFiles []*storage.RepositoryFile
	var dbImages []*model.Image

	result := &SyncResult{
		LogID: logID,
	}

	syncTag, err := imageRepo.GetOrCreateSyncTag()
	if err != nil {
		errMsg := "获取或创建同步标签失败"
		slog.Error(errMsg, "error", err)
		_ = syncLogRepo.UpdateSyncLog(logID, &now, 0, 0, 1, &errMsg, "failed")
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}
	slog.Info("同步标签", "tag_id", syncTag.ID, "tag_name", syncTag.Name)

	slog.Info("开始扫描 GitHub 仓库文件")
	storageFiles, err = s.storage.ListFiles(ctx, s.pathPrefix)
	if err != nil {
		errMsg := fmt.Sprintf("扫描仓库文件失败: %v", err)
		slog.Error(errMsg)
		_ = syncLogRepo.UpdateSyncLog(logID, &now, 0, 0, 1, &errMsg, "failed")
		return nil, fmt.Errorf("%s", errMsg)
	}
	slog.Info("扫描完成", "total_files", len(storageFiles))

	slog.Info("查询数据库所有图片")
	dbImages, err = imageRepo.GetAllImagesNotDeleted()
	if err != nil {
		errMsg := fmt.Sprintf("查询数据库图片失败: %v", err)
		slog.Error(errMsg)
		_ = syncLogRepo.UpdateSyncLog(logID, &now, 0, 0, 1, &errMsg, "failed")
		return nil, fmt.Errorf("%s", errMsg)
	}
	slog.Info("数据库图片", "total", len(dbImages))

	dbImageMap := make(map[string]*model.Image)
	for _, img := range dbImages {
		dbImageMap[img.Path] = img
	}

	storageFileMap := make(map[string]*storage.RepositoryFile)
	for _, file := range storageFiles {
		storageFileMap[file.Path] = file
	}

	imageExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}

	var originalFiles []*storage.RepositoryFile
	for _, file := range storageFiles {
		if file.Type != "file" {
			continue
		}

		baseName := filepath.Base(file.Path)
		if strings.HasPrefix(baseName, "thumb_") || strings.HasPrefix(baseName, "watermark_") {
			continue
		}

		ext := ""
		if idx := strings.LastIndex(file.Path, "."); idx > 0 {
			ext = strings.ToLower(file.Path[idx:])
		}
		if imageExtensions[ext] {
			originalFiles = append(originalFiles, file)
		}
	}

	slog.Info("找到的原图文件", "count", len(originalFiles))
	totalFiles := len(originalFiles)
	processedFiles := 0
	errorCount := 0

	for _, storageFile := range originalFiles {
		processedFiles++
		slog.Info("处理文件", "path", storageFile.Path, "progress", fmt.Sprintf("%d/%d", processedFiles, totalFiles))

		dbImage := dbImageMap[storageFile.Path]

		if dbImage == nil {
			slog.Info("文件不在数据库中，创建新记录", "path", storageFile.Path)
			image, err := s.createImageFromStorage(ctx, tx, storageFile, syncTag.ID)
			if err != nil {
				errMsg := fmt.Sprintf("创建图片记录失败: %v", err)
				slog.Error(errMsg, "path", storageFile.Path)
				_ = syncLogRepo.CreateSyncFileLog(logID, storageFile.Path, "created", "failed", &storageFile.SHA, nil, &storageFile.Size, nil, &errMsg)
				errorCount++
				continue
			}
			slog.Info("创建成功", "path", storageFile.Path, "image_id", image.ID)
			_ = syncLogRepo.CreateSyncFileLog(logID, storageFile.Path, "created", "success", &storageFile.SHA, nil, &storageFile.Size, nil, nil)
			result.CreatedCount++
			dbImageMap[storageFile.Path] = image
			s.updateDerivedImageMeta(ctx, image.ID, storageFile.Path, storageFileMap, imageRepo, syncLogRepo, logID, &errorCount)
		} else if dbImage.SHA != storageFile.SHA {
			slog.Info("文件SHA不一致，更新元数据", "path", storageFile.Path, "old_sha", dbImage.SHA, "new_sha", storageFile.SHA)
			oldSHA := dbImage.SHA
			oldSize := dbImage.Size

			width, height := 0, 0
			content, err := s.storage.GetRawFileContent(ctx, storageFile.Path)
			if err != nil {
				slog.Warn("获取文件内容失败，无法解析分辨率", "path", storageFile.Path, "error", err)
			} else {
				detectedWidth, detectedHeight, _, imgErr := storage.GetImageInfo(content)
				if imgErr != nil {
					slog.Warn("解析图片分辨率失败", "path", storageFile.Path, "error", imgErr)
				} else {
					width = detectedWidth
					height = detectedHeight
				}
			}

			err = imageRepo.UpdateImageMeta(dbImage.ID, storageFile.SHA, &storageFile.Size, &width, &height)
			if err != nil {
				errMsg := fmt.Sprintf("更新图片元数据失败: %v", err)
				slog.Error(errMsg, "path", storageFile.Path)
				_ = syncLogRepo.CreateSyncFileLog(logID, storageFile.Path, "updated", "failed", &storageFile.SHA, &oldSHA, &storageFile.Size, oldSize, &errMsg)
				errorCount++
				continue
			}
			slog.Info("更新成功", "path", storageFile.Path)
			_ = syncLogRepo.CreateSyncFileLog(logID, storageFile.Path, "updated", "success", &storageFile.SHA, &oldSHA, &storageFile.Size, oldSize, nil)
			result.UpdatedCount++
			s.updateDerivedImageMeta(ctx, dbImage.ID, storageFile.Path, storageFileMap, imageRepo, syncLogRepo, logID, &errorCount)
		} else {
			slog.Info("SHA一致，跳过", "path", storageFile.Path)
			_ = syncLogRepo.CreateSyncFileLog(logID, storageFile.Path, "skipped", "success", &storageFile.SHA, &storageFile.SHA, &storageFile.Size, &storageFile.Size, nil)
			result.SkippedCount++
			s.updateDerivedImageMeta(ctx, dbImage.ID, storageFile.Path, storageFileMap, imageRepo, syncLogRepo, logID, &errorCount)
		}
	}

	slog.Info("检查数据库中已删除的文件")
	for _, dbImage := range dbImages {
		storageFile := storageFileMap[dbImage.Path]
		if storageFile == nil {
			slog.Info("文件在存储中不存在，软删除", "path", dbImage.Path)
			err := imageRepo.SoftDeleteImage(dbImage.ID)
			if err != nil {
				errMsg := fmt.Sprintf("软删除图片失败: %v", err)
				slog.Error(errMsg, "path", dbImage.Path)
				_ = syncLogRepo.CreateSyncFileLog(logID, dbImage.Path, "deleted", "failed", nil, &dbImage.SHA, nil, dbImage.Size, &errMsg)
				errorCount++
			} else {
				slog.Info("软删除成功", "path", dbImage.Path)
				_ = syncLogRepo.CreateSyncFileLog(logID, dbImage.Path, "deleted", "success", nil, &dbImage.SHA, nil, dbImage.Size, nil)
				result.DeletedCount++
			}
		}
	}

	result.ErrorCount = errorCount

	completedAt := time.Now()
	status := "completed"
	errorMessage := (*string)(nil)
	if errorCount > 0 {
		status = "completed_with_errors"
	}

	err = syncLogRepo.UpdateSyncLog(logID, &completedAt, totalFiles, processedFiles, errorCount, errorMessage, status)
	if err != nil {
		slog.Error("更新同步日志失败", "error", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("同步完成", "created", result.CreatedCount, "updated", result.UpdatedCount, "deleted", result.DeletedCount, "skipped", result.SkippedCount, "errors", result.ErrorCount)

	return result, nil
}

func (s *ImageService) deriveThumbnailPath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, "thumb_"+name+".webp")
}

func (s *ImageService) deriveWatermarkPath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, "watermark_"+name+ext)
}

func (s *ImageService) updateDerivedImageMeta(
	ctx context.Context,
	imageID int64,
	originalPath string,
	storageFileMap map[string]*storage.RepositoryFile,
	imageRepo *repository.ImageRepository,
	syncLogRepo *repository.SyncLogRepository,
	logID int64,
	errorCount *int,
) {
	thumbPath := s.deriveThumbnailPath(originalPath)
	if thumbFile, ok := storageFileMap[thumbPath]; ok {
		if err := imageRepo.UpdateThumbnailMeta(imageID, thumbFile.Path, s.storage.GetURL(ctx, thumbFile.Path), thumbFile.SHA, &thumbFile.Size); err != nil {
			errMsg := fmt.Sprintf("更新缩略图元数据失败: %v", err)
			slog.Error(errMsg, "path", thumbPath, "image_id", imageID)
			_ = syncLogRepo.CreateSyncFileLog(logID, thumbPath, "updated", "failed", &thumbFile.SHA, nil, &thumbFile.Size, nil, &errMsg)
			*errorCount++
		} else {
			slog.Info("缩略图元数据更新成功", "path", thumbPath, "image_id", imageID)
			_ = syncLogRepo.CreateSyncFileLog(logID, thumbPath, "updated", "success", &thumbFile.SHA, nil, &thumbFile.Size, nil, nil)
		}
	}

	watermarkPath := s.deriveWatermarkPath(originalPath)
	if watermarkFile, ok := storageFileMap[watermarkPath]; ok {
		if err := imageRepo.UpdateWatermarkMeta(imageID, watermarkFile.Path, s.storage.GetURL(ctx, watermarkFile.Path), watermarkFile.SHA, &watermarkFile.Size); err != nil {
			errMsg := fmt.Sprintf("更新水印图元数据失败: %v", err)
			slog.Error(errMsg, "path", watermarkPath, "image_id", imageID)
			_ = syncLogRepo.CreateSyncFileLog(logID, watermarkPath, "updated", "failed", &watermarkFile.SHA, nil, &watermarkFile.Size, nil, &errMsg)
			*errorCount++
		} else {
			slog.Info("水印图元数据更新成功", "path", watermarkPath, "image_id", imageID)
			_ = syncLogRepo.CreateSyncFileLog(logID, watermarkPath, "updated", "success", &watermarkFile.SHA, nil, &watermarkFile.Size, nil, nil)
		}
	}
}

func (s *ImageService) createImageFromStorage(ctx context.Context, tx *sql.Tx, storageFile *storage.RepositoryFile, syncTagID int64) (*model.Image, error) {
	mimeType := "image/jpeg"
	if idx := strings.LastIndex(storageFile.Path, "."); idx > 0 {
		ext := strings.ToLower(storageFile.Path[idx:])
		mimeType = storage.GetMimeType(storageFile.Path)
		_ = ext
	}

	width, height := 0, 0
	content, err := s.storage.GetRawFileContent(ctx, storageFile.Path)
	if err != nil {
		slog.Warn("获取文件内容失败，无法解析分辨率", "path", storageFile.Path, "error", err)
	} else {
		detectedWidth, detectedHeight, detectedMimeType, imgErr := storage.GetImageInfo(content)
		if imgErr != nil {
			slog.Warn("解析图片分辨率失败", "path", storageFile.Path, "error", imgErr)
		} else {
			width = detectedWidth
			height = detectedHeight
			if mimeType == "image/jpeg" && detectedMimeType != "" {
				mimeType = detectedMimeType
			}
		}
	}

	image := &model.Image{
		Path:             storageFile.Path,
		URL:              s.storage.GetURL(ctx, storageFile.Path),
		SHA:              storageFile.SHA,
		OriginalFilename: filepath.Base(storageFile.Path),
		Filename:         filepath.Base(storageFile.Path),
		Size:             &storageFile.Size,
		MimeType:         mimeType,
		Width:            &width,
		Height:           &height,
		HasThumbnail:     false,
		HasWatermark:     false,
		UploadedAt:       time.Now(),
	}

	imageRepo := repository.NewImageRepository(tx)
	id, err := imageRepo.CreateImage(image)
	if err != nil {
		return nil, fmt.Errorf("创建图片记录失败: %w", err)
	}
	image.ID = id

	err = imageRepo.AddImageTags(id, []int{int(syncTagID)})
	if err != nil {
		return nil, fmt.Errorf("关联同步标签失败: %w", err)
	}

	return image, nil
}
