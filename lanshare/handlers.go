package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handlers struct {
	storage Storage
	config  *Config
}

func NewHandlers(storage Storage, config *Config) *Handlers {
	return &Handlers{
		storage: storage,
		config:  config,
	}
}

func (h *Handlers) GetMessages(c fiber.Ctx) error {
	return c.JSON(h.storage.List())
}

func (h *Handlers) AddMessage(c fiber.Ctx) error {
	var item Item
	if err := c.Bind().Body(&item); err != nil {
		log.Printf("Bind JSON error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// 验证内容长度
	if len(item.Content) > 10000 { // 10KB limit for text
		return c.Status(400).JSON(fiber.Map{"error": "Text content too long"})
	}

	newItem := NewItem(TEXT, item.Content)
	h.storage.Add(newItem)

	return c.JSON(fiber.Map{"id": newItem.ID, "message": "Message added successfully"})
}

func (h *Handlers) DeleteMessage(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	// 检查项目是否存在
	item := h.storage.Get(id)
	if item == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Item not found"})
	}

	// 如果是文件，删除物理文件
	if item.Type != TEXT {
		if saveName, ok := item.Meta["saveName"].(string); ok {
			filePath := filepath.Join(h.config.DownloadPath, saveName)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to delete file %s: %v", filePath, err)
			}
		}
	}

	h.storage.Remove(id)
	return c.JSON(fiber.Map{"message": "Item deleted successfully"})
}

func (h *Handlers) UploadFile(c fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "请选择要上传的文件"})
	}

	// 验证文件类型
	if err := ValidateFileType(fh, h.config.AllowedTypes); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 验证文件大小
	if fh.Size > h.config.MaxFileSize {
		maxSizeMB := h.config.MaxFileSize / (1024 * 1024)
		return c.Status(413).JSON(fiber.Map{
			"error": fmt.Sprintf("文件太大，最大支持 %d MB", maxSizeMB),
		})
	}

	f, err := fh.Open()
	if err != nil {
		log.Printf("Failed to open uploaded file: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "文件处理失败，请重试"})
	}
	defer f.Close()

	// 生成安全的文件名
	ext := filepath.Ext(fh.Filename)
	randomFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	savePath := filepath.Join(h.config.DownloadPath, randomFileName)

	out, err := os.Create(savePath)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "文件保存失败，请检查服务器存储空间"})
	}
	defer out.Close()

	// 限制复制的数据量
	_, err = io.CopyN(out, f, h.config.MaxFileSize)
	if err != nil && err != io.EOF {
		log.Printf("Failed to copy file: %v", err)
		os.Remove(savePath) // 清理失败的文件
		return c.Status(500).JSON(fiber.Map{"error": "文件保存失败，请重试"})
	}

	var itemType Type
	if isImageFileHeader(fh) {
		itemType = IMAGE
	} else {
		itemType = OTHER
	}

	item := NewItem(itemType, fh.Filename)
	item.Meta = map[string]interface{}{
		"saveName":     randomFileName,
		"path":         h.config.DownloadPath,
		"originalName": fh.Filename,
		"size":         fh.Size,
		"contentType":  fh.Header.Get("Content-Type"),
	}

	h.storage.Add(item)

	return c.JSON(fiber.Map{
		"message":  "File uploaded successfully",
		"id":       item.ID,
		"filename": fh.Filename,
	})
}

func (h *Handlers) DownloadFile(c fiber.Ctx) error {
	id := c.Params("*")
	item := h.storage.Get(id)
	if item == nil {
		return c.SendStatus(404)
	}

	path, ok := item.Meta["path"].(string)
	if !ok {
		return c.SendStatus(404)
	}

	saveName, ok := item.Meta["saveName"].(string)
	if !ok {
		return c.SendStatus(404)
	}

	realFilePath := filepath.Join(path, saveName)

	// 检查文件是否存在
	if _, err := os.Stat(realFilePath); os.IsNotExist(err) {
		return c.SendStatus(404)
	}

	// 设置正确的 Content-Type
	if contentType, ok := item.Meta["contentType"].(string); ok {
		c.Set("Content-Type", contentType)
	}

	// 对于图片，让浏览器直接显示；对于其他文件，设置为下载
	if item.Type == IMAGE {
		// 图片文件：让浏览器直接显示
		if originalName, ok := item.Meta["originalName"].(string); ok {
			c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", originalName))
		}
	} else {
		// 其他文件：设置为下载
		if originalName, ok := item.Meta["originalName"].(string); ok {
			c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", originalName))
		}
	}

	return c.SendFile(realFilePath)
}
