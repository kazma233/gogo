package storage

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"strings"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
)

type ThumbnailConfig struct {
	Width   int
	Quality int
	Format  string
}

func GenerateThumbnail(imgData []byte, config ThumbnailConfig) ([]byte, int, int, bool, error) {
	src, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		slog.Error("解码图片失败", "error", err)
		return nil, 0, 0, false, err
	}

	bounds := src.Bounds()
	width := bounds.Dx()

	if width <= config.Width {
		slog.Debug("图片宽度小于配置，跳过缩略图", "width", width, "config_width", config.Width)
		return nil, 0, 0, false, nil
	}

	thumbnail := imaging.Resize(src, config.Width, 0, imaging.Lanczos)

	var buf bytes.Buffer

	switch config.Format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: config.Quality})
	default:
		err = png.Encode(&buf, thumbnail)
	}

	if err != nil {
		slog.Error("编码缩略图失败", "error", err)
		return nil, 0, 0, false, err
	}

	thumbBounds := thumbnail.Bounds()
	return buf.Bytes(), thumbBounds.Dx(), thumbBounds.Dy(), true, nil
}

func GetImageInfo(imgData []byte) (int, int, string, error) {
	config, format, err := image.DecodeConfig(bytes.NewReader(imgData))
	if err != nil {
		return 0, 0, "", err
	}
	return config.Width, config.Height, format, nil
}

func GetMimeType(filename string) string {
	ext := filename[strings.LastIndex(filename, "."):]
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

type ErrUnsupportedFormat struct {
	Format string
}

func (e *ErrUnsupportedFormat) Error() string {
	return "不支持的图片格式: " + e.Format
}

func (e *ErrUnsupportedFormat) Is(target error) bool {
	_, ok := target.(*ErrUnsupportedFormat)
	return ok
}
