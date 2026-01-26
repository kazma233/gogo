package handler

import (
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"strconv"

	"picstash/internal/service"

	"github.com/gofiber/fiber/v3"
)

type ImageHandler struct {
	imageService *service.ImageService
}

func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

func (h *ImageHandler) GetList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	tagIDStr := c.Query("tag_id")

	var tagID *int
	if tagIDStr != "" {
		id, err := strconv.Atoi(tagIDStr)
		if err == nil {
			tagID = &id
		}
	}

	images, total, err := h.imageService.GetList(page, limit, tagID)
	if err != nil {
		slog.Error("获取图片列表失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取图片列表失败",
		})
	}

	return c.JSON(fiber.Map{
		"data":  images,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ImageHandler) GetByID(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")

	image, err := h.imageService.GetByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "图片不存在",
		})
	}

	return c.JSON(image)
}

func (h *ImageHandler) Upload(c fiber.Ctx) error {
	email := c.Locals("email").(string)
	slog.Info("开始处理上传请求", "user", email)

	form, err := c.MultipartForm()
	if err != nil {
		slog.Error("解析上传表单失败", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "解析上传文件失败",
		})
	}

	tagIDsStr := form.Value["tag_ids"]

	slog.Info("标签原始数据", "tag_ids_raw", tagIDsStr)

	var tagIDs []int
	for _, str := range tagIDsStr {
		id, err := strconv.Atoi(str)
		if err == nil {
			tagIDs = append(tagIDs, id)
		} else {
			slog.Warn("解析标签ID失败", "value", str, "error", err)
		}
	}
	slog.Info("解析后的标签ID", "tag_ids", tagIDs)

	ctx := c.RequestCtx()

	mappingStr := c.FormValue("file_mapping")
	if mappingStr == "" {
		slog.Error("缺少 file_mapping 参数")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "缺少文件映射参数",
		})
	}

	var mapping []service.FileMapping
	if err := json.Unmarshal([]byte(mappingStr), &mapping); err != nil {
		slog.Error("解析文件映射失败", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "文件映射格式错误",
		})
	}

	originalFiles := form.File["original_files"]
	watermarkFiles := form.File["watermark_files"]
	thumbnailFiles := form.File["thumbnail_files"]

	slog.Info("接收到的原图", "count", len(originalFiles))
	slog.Info("接收到的水印图", "count", len(watermarkFiles))
	slog.Info("接收到的缩略图", "count", len(thumbnailFiles))
	slog.Info("映射数量", "count", len(mapping))

	originalFileMap := readFilesToMap(originalFiles)
	watermarkFileMap := readFilesToMap(watermarkFiles)
	thumbnailFileMap := readFilesToMap(thumbnailFiles)

	images, err := h.imageService.BatchUploadWithMapping(
		ctx,
		originalFileMap,
		watermarkFileMap,
		thumbnailFileMap,
		mapping,
		tagIDs,
	)
	if err != nil {
		slog.Error("批量上传失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "上传图片失败",
		})
	}

	slog.Info("批量上传成功", "user", email, "count", len(images))

	return c.JSON(fiber.Map{
		"data":  images,
		"count": len(images),
	})
}

func readFilesToMap(files []*multipart.FileHeader) map[string][]byte {
	result := make(map[string][]byte)
	for _, f := range files {
		data := make([]byte, f.Size)
		file, _ := f.Open()
		file.Read(data)
		file.Close()
		result[f.Filename] = data
	}
	return result
}

func (h *ImageHandler) Delete(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")

	ctx := c.RequestCtx()
	if err := h.imageService.Delete(ctx, int64(id)); err != nil {
		slog.Error("删除图片失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "删除图片失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "删除成功",
	})
}

func (h *ImageHandler) UpdateTags(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")

	var req struct {
		TagIDs []int `json:"tag_ids"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请求参数错误",
		})
	}

	if err := h.imageService.UpdateTags(int64(id), req.TagIDs); err != nil {
		slog.Error("更新标签失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "更新标签失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "更新成功",
	})
}

func (h *ImageHandler) SyncFromStorage(c fiber.Ctx) error {
	email := c.Locals("email").(string)
	slog.Info("开始处理同步请求", "user", email)

	ctx := c.RequestCtx()
	result, err := h.imageService.SyncFromStorage(ctx, email)
	if err != nil {
		slog.Error("同步失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "同步失败",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "同步完成",
		"data": map[string]interface{}{
			"created_count": result.CreatedCount,
			"updated_count": result.UpdatedCount,
			"deleted_count": result.DeletedCount,
			"skipped_count": result.SkippedCount,
			"error_count":   result.ErrorCount,
			"log_id":        result.LogID,
		},
	})
}

// fiber:context-methods migrated
