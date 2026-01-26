package container

import (
	"context"
	"io"
)

type ContainerConfig struct {
	Image      string
	Cmd        []string
	WorkingDir string
	Env        []string
}

type ContainerRuntime interface {
	Name() string

	PullImage(ctx context.Context, image string) error

	CreateContainer(ctx context.Context, cfg *ContainerConfig) (string, error)

	StartContainer(ctx context.Context, id string) error

	Exec(ctx context.Context, containerID string, cmd ...string) error

	WaitContainer(ctx context.Context, id string) error

	RemoveContainer(ctx context.Context, id string) error

	CopyToContainer(ctx context.Context, containerID, srcPath, dstPath string) error

	CopyFromContainer(ctx context.Context, containerID, srcPath, dstPath string) error

	GetContainerLogs(ctx context.Context, id string) (io.ReadCloser, error)

	Close() error
}
