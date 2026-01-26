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

type DockerRuntime struct {
	command string
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cmd, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}
	return &DockerRuntime{command: cmd}, nil
}

func (d *DockerRuntime) Name() string {
	return "docker"
}

func (d *DockerRuntime) runCommand(ctx context.Context, args ...string) (string, error) {
	log.Printf("Executing: %s %s", d.command, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, d.command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker %s failed: %s", strings.Join(args, " "), string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DockerRuntime) PullImage(ctx context.Context, image string) error {
	_, err := d.runCommand(ctx, "pull", "--policy", "newer", image)
	return err
}

func (d *DockerRuntime) CreateContainer(ctx context.Context, cfg *ContainerConfig) (string, error) {
	args := []string{"create"}

	for _, env := range cfg.Env {
		args = append(args, "-e", env)
	}

	args = append(args, cfg.Image)
	args = append(args, cfg.Cmd...)

	return d.runCommand(ctx, args...)
}

func (d *DockerRuntime) StartContainer(ctx context.Context, id string) error {
	_, err := d.runCommand(ctx, "start", id)
	return err
}

func (d *DockerRuntime) Exec(ctx context.Context, containerID string, cmd ...string) error {
	args := []string{"exec", containerID}
	args = append(args, cmd...)
	_, err := d.runCommand(ctx, args...)
	return err
}

func (d *DockerRuntime) WaitContainer(ctx context.Context, id string) error {
	_, err := d.runCommand(ctx, "wait", id)
	return err
}

func (d *DockerRuntime) RemoveContainer(ctx context.Context, id string) error {
	_, err := d.runCommand(ctx, "rm", "-f", id)
	return err
}

func (d *DockerRuntime) CopyToContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	_, err := d.runCommand(ctx, "cp", srcPath, fmt.Sprintf("%s:%s", containerID, dstPath))
	return err
}

func (d *DockerRuntime) CopyFromContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	os.MkdirAll(dstPath, 0755)
	_, err := d.runCommand(ctx, "cp", fmt.Sprintf("%s:%s", containerID, srcPath), dstPath)
	return err
}

func (d *DockerRuntime) GetContainerLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	output, err := os.CreateTemp("", "deploygo-docker-logs-*")
	if err != nil {
		return nil, err
	}

	args := []string{"logs", id}
	cmd := exec.CommandContext(ctx, d.command, args...)
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Run(); err != nil {
		output.Close()
		return nil, err
	}

	output.Seek(0, 0)
	return output, nil
}

func (d *DockerRuntime) Close() error {
	return nil
}
