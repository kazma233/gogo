package repository

import (
	"database/sql"
	"fmt"
	"time"

	"picstash/internal/model"
)

type ImageRepositoryInterface interface {
	CreateImage(image *model.Image) (int64, error)
	AddImageTags(imageID int64, tagIDs []int) error
	DeleteImageTags(imageID int64) error
	GetImages(page, limit int, tagID *int) ([]*model.Image, int, error)
	GetImageByID(id int64) (*model.Image, error)
	SoftDeleteImage(id int64) error
	UpdateImageTags(imageID int64, tagIDs []int) error
	GetTagsByImageID(imageID int64) ([]*model.Tag, error)
	GetAllImagesNotDeleted() ([]*model.Image, error)
	FindByPath(path string) (*model.Image, error)
	UpdateImageMeta(id int64, sha string, size *int64, width *int, height *int) error
	UpdateThumbnailMeta(id int64, thumbnailPath, thumbnailURL, thumbnailSHA string, thumbnailSize *int64) error
	UpdateWatermarkMeta(id int64, watermarkPath, watermarkURL, watermarkSHA string, watermarkSize *int64) error
	GetOrCreateSyncTag() (*model.Tag, error)
}

type ImageRepository struct {
	tx *sql.Tx
}

func NewImageRepository(tx *sql.Tx) *ImageRepository {
	return &ImageRepository{tx: tx}
}

func (r *ImageRepository) CreateImage(image *model.Image) (int64, error) {
	result, err := r.tx.Exec(`
		INSERT INTO images (
			path, url, sha,
			thumbnail_path, thumbnail_url, thumbnail_sha,
			watermark_path, watermark_url, watermark_sha, watermark_size,
			original_filename, filename, size, thumbnail_size,
			mime_type, width, height, has_thumbnail, has_watermark
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		image.Path, image.URL, image.SHA,
		image.ThumbnailPath, image.ThumbnailURL, image.ThumbnailSHA,
		image.WatermarkPath, image.WatermarkURL, image.WatermarkSHA, image.WatermarkSize,
		image.OriginalFilename, image.Filename, image.Size, image.ThumbnailSize,
		image.MimeType, image.Width, image.Height, image.HasThumbnail, image.HasWatermark)
	if err != nil {
		return 0, fmt.Errorf("创建图片失败: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取图片ID失败: %w", err)
	}
	return id, nil
}

func (r *ImageRepository) AddImageTags(imageID int64, tagIDs []int) error {
	for _, tagID := range tagIDs {
		_, err := r.tx.Exec(`INSERT INTO image_tags (image_id, tag_id) VALUES (?, ?)`, imageID, tagID)
		if err != nil {
			return fmt.Errorf("关联标签失败: %w", err)
		}
	}
	return nil
}

func (r *ImageRepository) DeleteImageTags(imageID int64) error {
	_, err := r.tx.Exec(`DELETE FROM image_tags WHERE image_id = ?`, imageID)
	if err != nil {
		return fmt.Errorf("删除标签关联失败: %w", err)
	}
	return nil
}

func (r *ImageRepository) GetImages(page, limit int, tagID *int) ([]*model.Image, int, error) {
	offset := (page - 1) * limit

	var args []interface{}
	query := `
		SELECT id, path, url, sha,
			   thumbnail_path, thumbnail_url, thumbnail_sha,
			   watermark_path, watermark_url, watermark_sha, watermark_size,
			   original_filename, filename, size, thumbnail_size,
			   mime_type, width, height, has_thumbnail, has_watermark,
			   uploaded_at
		FROM images
		WHERE deleted_at IS NULL
	`

	if tagID != nil {
		query = `
			SELECT id, path, url, sha,
				   thumbnail_path, thumbnail_url, thumbnail_sha,
				   watermark_path, watermark_url, watermark_sha, watermark_size,
				   original_filename, filename, size, thumbnail_size,
				   mime_type, width, height, has_thumbnail, has_watermark,
				   uploaded_at
			FROM images
			WHERE deleted_at IS NULL
			AND id IN (SELECT image_id FROM image_tags WHERE tag_id = ?)
		`
		args = append(args, *tagID)
	}

	query += ` ORDER BY uploaded_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.tx.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询图片列表失败: %w", err)
	}
	defer rows.Close()

	var images []*model.Image
	for rows.Next() {
		img := &model.Image{}
		err := rows.Scan(
			&img.ID, &img.Path, &img.URL, &img.SHA,
			&img.ThumbnailPath, &img.ThumbnailURL, &img.ThumbnailSHA,
			&img.WatermarkPath, &img.WatermarkURL, &img.WatermarkSHA, &img.WatermarkSize,
			&img.OriginalFilename, &img.Filename, &img.Size, &img.ThumbnailSize,
			&img.MimeType, &img.Width, &img.Height, &img.HasThumbnail, &img.HasWatermark,
			&img.UploadedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描图片数据失败: %w", err)
		}
		images = append(images, img)
	}

	if len(images) > 0 {
		imageIDs := make([]int64, len(images))
		for i, img := range images {
			imageIDs[i] = img.ID
		}

		tagRows, err := r.tx.Query(`
			SELECT it.image_id, t.id, t.name, t.color, t.created_at
			FROM tags t
			INNER JOIN image_tags it ON t.id = it.tag_id
			WHERE it.image_id IN (`+placeholders(len(imageIDs))+`)
		`, int64sToInterfaces(imageIDs...)...)
		if err != nil {
			return nil, 0, fmt.Errorf("查询图片标签失败: %w", err)
		}
		defer tagRows.Close()

		tagMap := make(map[int64][]model.Tag)
		for tagRows.Next() {
			var imageID int64
			var tag model.Tag
			if err := tagRows.Scan(&imageID, &tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
				continue
			}
			tagMap[imageID] = append(tagMap[imageID], tag)
		}
		for _, img := range images {
			img.Tags = tagMap[img.ID]
		}
	}

	var countQuery string
	var countArgs []interface{}
	if tagID != nil {
		countQuery = `SELECT COUNT(*) FROM images WHERE deleted_at IS NULL AND id IN (SELECT image_id FROM image_tags WHERE tag_id = ?)`
		countArgs = append(countArgs, *tagID)
	} else {
		countQuery = `SELECT COUNT(*) FROM images WHERE deleted_at IS NULL`
	}

	var total int
	err = r.tx.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询图片总数失败: %w", err)
	}

	return images, total, nil
}

func (r *ImageRepository) GetImageByID(id int64) (*model.Image, error) {
	img := &model.Image{}

	err := r.tx.QueryRow(`
		SELECT id, path, url, sha,
			   thumbnail_path, thumbnail_url, thumbnail_sha,
			   watermark_path, watermark_url, watermark_sha, watermark_size,
			   original_filename, filename, size, thumbnail_size,
			   mime_type, width, height, has_thumbnail, has_watermark,
			   uploaded_at
		FROM images
		WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(
		&img.ID, &img.Path, &img.URL, &img.SHA,
		&img.ThumbnailPath, &img.ThumbnailURL, &img.ThumbnailSHA,
		&img.WatermarkPath, &img.WatermarkURL, &img.WatermarkSHA, &img.WatermarkSize,
		&img.OriginalFilename, &img.Filename, &img.Size, &img.ThumbnailSize,
		&img.MimeType, &img.Width, &img.Height, &img.HasThumbnail, &img.HasWatermark,
		&img.UploadedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("图片不存在")
		}
		return nil, fmt.Errorf("查询图片失败: %w", err)
	}

	rows, err := r.tx.Query(`
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		INNER JOIN image_tags it ON t.id = it.tag_id
		WHERE it.image_id = ?
		ORDER BY it.created_at DESC
	`, id)
	if err != nil {
		return img, nil
	}
	defer rows.Close()

	for rows.Next() {
		tag := &model.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			continue
		}
		img.Tags = append(img.Tags, *tag)
	}

	return img, nil
}

func (r *ImageRepository) SoftDeleteImage(id int64) error {
	_, err := r.tx.Exec(`UPDATE images SET deleted_at = ? WHERE id = ?`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("删除图片失败: %w", err)
	}
	return nil
}

func (r *ImageRepository) UpdateImageTags(imageID int64, tagIDs []int) error {
	_, err := r.tx.Exec(`DELETE FROM image_tags WHERE image_id = ?`, imageID)
	if err != nil {
		return fmt.Errorf("删除旧标签失败: %w", err)
	}

	for _, tagID := range tagIDs {
		_, err = r.tx.Exec(`INSERT INTO image_tags (image_id, tag_id) VALUES (?, ?)`, imageID, tagID)
		if err != nil {
			return fmt.Errorf("添加新标签失败: %w", err)
		}
	}
	return nil
}

func (r *ImageRepository) GetTagsByImageID(imageID int64) ([]*model.Tag, error) {
	rows, err := r.tx.Query(`
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		INNER JOIN image_tags it ON t.id = it.tag_id
		WHERE it.image_id = ?
		ORDER BY it.created_at DESC
	`, imageID)
	if err != nil {
		return nil, fmt.Errorf("查询图片标签失败: %w", err)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		tag := &model.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	result := "?"
	for i := 1; i < n; i++ {
		result += ", ?"
	}
	return result
}

func int64sToInterfaces(nums ...int64) []interface{} {
	interfaces := make([]interface{}, len(nums))
	for i, n := range nums {
		interfaces[i] = n
	}
	return interfaces
}

func (r *ImageRepository) GetAllImagesNotDeleted() ([]*model.Image, error) {
	query := `
		SELECT id, path, url, sha,
			   thumbnail_path, thumbnail_url, thumbnail_sha,
			   watermark_path, watermark_url, watermark_sha, watermark_size,
			   original_filename, filename, size, thumbnail_size,
			   mime_type, width, height, has_thumbnail, has_watermark,
			   uploaded_at
		FROM images
		WHERE deleted_at IS NULL
	`
	rows, err := r.tx.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询所有图片失败: %w", err)
	}
	defer rows.Close()

	var images []*model.Image
	for rows.Next() {
		img := &model.Image{}
		err := rows.Scan(
			&img.ID, &img.Path, &img.URL, &img.SHA,
			&img.ThumbnailPath, &img.ThumbnailURL, &img.ThumbnailSHA,
			&img.WatermarkPath, &img.WatermarkURL, &img.WatermarkSHA, &img.WatermarkSize,
			&img.OriginalFilename, &img.Filename, &img.Size, &img.ThumbnailSize,
			&img.MimeType, &img.Width, &img.Height, &img.HasThumbnail, &img.HasWatermark,
			&img.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描图片数据失败: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *ImageRepository) FindByPath(path string) (*model.Image, error) {
	img := &model.Image{}
	err := r.tx.QueryRow(`
		SELECT id, path, url, sha,
			   thumbnail_path, thumbnail_url, thumbnail_sha,
			   watermark_path, watermark_url, watermark_sha, watermark_size,
			   original_filename, filename, size, thumbnail_size,
			   mime_type, width, height, has_thumbnail, has_watermark,
			   uploaded_at
		FROM images
		WHERE path = ? AND deleted_at IS NULL
	`, path).Scan(
		&img.ID, &img.Path, &img.URL, &img.SHA,
		&img.ThumbnailPath, &img.ThumbnailURL, &img.ThumbnailSHA,
		&img.WatermarkPath, &img.WatermarkURL, &img.WatermarkSHA, &img.WatermarkSize,
		&img.OriginalFilename, &img.Filename, &img.Size, &img.ThumbnailSize,
		&img.MimeType, &img.Width, &img.Height, &img.HasThumbnail, &img.HasWatermark,
		&img.UploadedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询图片失败: %w", err)
	}
	return img, nil
}

func (r *ImageRepository) UpdateImageMeta(id int64, sha string, size *int64, width *int, height *int) error {
	_, err := r.tx.Exec(`UPDATE images SET sha = ?, size = ?, width = ?, height = ?, updated_at = NOW() WHERE id = ?`, sha, size, width, height, id)
	if err != nil {
		return fmt.Errorf("更新图片元数据失败: %w", err)
	}
	return nil
}

func (r *ImageRepository) UpdateThumbnailMeta(id int64, thumbnailPath, thumbnailURL, thumbnailSHA string, thumbnailSize *int64) error {
	_, err := r.tx.Exec(`
		UPDATE images
		SET thumbnail_path = ?, thumbnail_url = ?, thumbnail_sha = ?, thumbnail_size = ?, has_thumbnail = 1
		WHERE id = ?
	`, thumbnailPath, thumbnailURL, thumbnailSHA, thumbnailSize, id)
	if err != nil {
		return fmt.Errorf("更新缩略图元数据失败: %w", err)
	}
	return nil
}

func (r *ImageRepository) UpdateWatermarkMeta(id int64, watermarkPath, watermarkURL, watermarkSHA string, watermarkSize *int64) error {
	_, err := r.tx.Exec(`
		UPDATE images
		SET watermark_path = ?, watermark_url = ?, watermark_sha = ?, watermark_size = ?, has_watermark = 1
		WHERE id = ?
	`, watermarkPath, watermarkURL, watermarkSHA, watermarkSize, id)
	if err != nil {
		return fmt.Errorf("更新水印图元数据失败: %w", err)
	}
	return nil
}

func (r *ImageRepository) GetOrCreateSyncTag() (*model.Tag, error) {
	tag := &model.Tag{}
	err := r.tx.QueryRow(`SELECT id, name, color, created_at FROM tags WHERE name = '同步'`).Scan(
		&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt,
	)
	if err == nil {
		return tag, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("查询同步标签失败: %w", err)
	}

	result, err := r.tx.Exec(`INSERT INTO tags (name, color) VALUES ('同步', '#9CA3AF')`)
	if err != nil {
		return nil, fmt.Errorf("创建同步标签失败: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取同步标签ID失败: %w", err)
	}

	tag.ID = id
	tag.Name = "同步"
	tag.Color = "#9CA3AF"
	tag.CreatedAt = time.Now()

	return tag, nil
}
