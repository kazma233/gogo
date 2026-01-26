package storage

import "context"

type File struct {
	Path        string
	Content     []byte
	ContentType string
}

type UploadResult struct {
	Path string
	URL  string
	SHA  string
}

type Storage interface {
	Upload(ctx context.Context, file *File) (*UploadResult, error)
	BatchUpload(ctx context.Context, files []*File) ([]*UploadResult, error)
	Delete(ctx context.Context, path string) error
	GetURL(ctx context.Context, path string) string
	Exists(ctx context.Context, path string) (bool, error)
	ListFiles(ctx context.Context, path string) ([]*RepositoryFile, error)
	GetRawFileContent(ctx context.Context, path string) ([]byte, error)
}
