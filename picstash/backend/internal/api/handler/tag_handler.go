package handler

import (
	"log/slog"

	"picstash/internal/service"

	"github.com/gofiber/fiber/v3"
)

type TagHandler struct {
	tagService *service.TagService
}

func NewTagHandler(tagService *service.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

func (h *TagHandler) GetAll(c fiber.Ctx) error {
	tags, err := h.tagService.GetAll()
	if err != nil {
		slog.Error("获取标签列表失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取标签列表失败",
		})
	}

	return c.JSON(tags)
}

func (h *TagHandler) Create(c fiber.Ctx) error {
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请求参数错误",
		})
	}

	if req.Name == "" || req.Color == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "名称和颜色不能为空",
		})
	}

	tag, err := h.tagService.Create(req.Name, req.Color)
	if err != nil {
		slog.Error("创建标签失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "创建标签失败",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tag)
}

func (h *TagHandler) Update(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请求参数错误",
		})
	}

	tag, err := h.tagService.Update(int64(id), req.Name, req.Color)
	if err != nil {
		slog.Error("更新标签失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "更新标签失败",
		})
	}

	return c.JSON(tag)
}

func (h *TagHandler) Delete(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")

	if err := h.tagService.Delete(int64(id)); err != nil {
		slog.Error("删除标签失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "删除标签失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "删除成功",
	})
}

func (h *TagHandler) GetByImageID(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")

	tags, err := h.tagService.GetByImageID(int64(id))
	if err != nil {
		slog.Error("获取图片标签失败", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取标签失败",
		})
	}

	return c.JSON(tags)
}
