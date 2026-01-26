package container

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type PodmanRuntime struct {
	command string
}

func NewPodmanRuntime() (*PodmanRuntime, error) {
	cmd, err := exec.LookPath("podman")
	if err != nil {
		return nil, fmt.Errorf("podman not found in PATH: %w", err)
	}
	return &PodmanRuntime{command: cmd}, nil
}

func (p *PodmanRuntime) Name() string {
	return "podman"
}

func (p *PodmanRuntime) runCommand(ctx context.Context, args ...string) (string, error) {
	log.Printf("Executing: %s %s", p.command, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, p.command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("podman %s failed: %s", strings.Join(args, " "), string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (p *PodmanRuntime) PullImage(ctx context.Context, image string) error {
	_, err := p.runCommand(ctx, "pull", "--policy", "newer", image)
	return err
}

func (p *PodmanRuntime) CreateContainer(ctx context.Context, cfg *ContainerConfig) (string, error) {
	args := []string{"create"}

	for _, env := range cfg.Env {
		args = append(args, "-e", env)
	}

	args = append(args, cfg.Image)
	args = append(args, cfg.Cmd...)

	return p.runCommand(ctx, args...)
}

func (p *PodmanRuntime) StartContainer(ctx context.Context, id string) error {
	_, err := p.runCommand(ctx, "start", id)
	return err
}

func (p *PodmanRuntime) Exec(ctx context.Context, containerID string, cmd ...string) error {
	args := []string{"exec", containerID}
	args = append(args, cmd...)
	_, err := p.runCommand(ctx, args...)
	return err
}

func (p *PodmanRuntime) WaitContainer(ctx context.Context, id string) error {
	_, err := p.runCommand(ctx, "wait", id)
	return err
}

func (p *PodmanRuntime) RemoveContainer(ctx context.Context, id string) error {
	_, err := p.runCommand(ctx, "rm", "-f", id)
	return err
}

func (p *PodmanRuntime) CopyToContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	_, err := p.runCommand(ctx, "cp", srcPath, fmt.Sprintf("%s:%s", containerID, dstPath))
	return err
}

func (p *PodmanRuntime) CopyFromContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	os.MkdirAll(dstPath, 0755)
	_, err := p.runCommand(ctx, "cp", fmt.Sprintf("%s:%s", containerID, srcPath), dstPath)
	return err
}

func (p *PodmanRuntime) GetContainerLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	output, err := os.CreateTemp("", "deploygo-podman-logs-*")
	if err != nil {
		return nil, err
	}

	args := []string{"logs", id}
	cmd := exec.CommandContext(ctx, p.command, args...)
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Run(); err != nil {
		output.Close()
		return nil, err
	}

	output.Seek(0, 0)
	return output, nil
}

func (p *PodmanRuntime) Close() error {
	return nil
}
