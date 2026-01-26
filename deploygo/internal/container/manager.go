package container

import (
	"os"
)

type ManagerConfig struct {
	Type string
}

func NewManager(cfg *ManagerConfig) (ContainerRuntime, error) {
	if cfg == nil {
		cfg = &ManagerConfig{}
	}

	containerType := cfg.Type
	if containerType == "" {
		containerType = os.Getenv("CONTAINER_RUNTIME")
	}

	switch containerType {
	case "podman":
		return NewPodmanRuntime()
	case "docker":
		fallthrough
	default:
		return NewDockerRuntime()
	}
}

func NewManagerFromEnv() (ContainerRuntime, error) {
	return NewManager(nil)
}
