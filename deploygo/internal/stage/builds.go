package stage

import (
	"context"
	"deploygo/internal/config"
	"deploygo/internal/container"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func RunBuilds(runtime container.ContainerRuntime, cfg *config.Config, builds []config.StageConfig, projectDir string) error {
	for i, build := range builds {
		ctx := context.Background()

		log.Printf("Running build %d/%d: %s (runtime: %s, image: %s)", i+1, len(builds), build.Name, runtime.Name(), build.Image)

		if err := runtime.PullImage(ctx, build.Image); err != nil {
			return fmt.Errorf("failed to pull image %s: %w", build.Image, err)
		}

		containerCfg := &container.ContainerConfig{
			Image:      build.Image,
			Cmd:        []string{"sleep", "infinity"},
			WorkingDir: build.WorkingDir,
			Env:        build.Environment,
		}

		log.Printf("Creating container with image: %s", build.Image)
		containerID, err := runtime.CreateContainer(ctx, containerCfg)
		if err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}
		defer runtime.RemoveContainer(ctx, containerID)

		log.Printf("Starting container %s", containerID[:12])
		if err := runtime.StartContainer(ctx, containerID); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		if build.WorkingDir != "" {
			log.Printf("Creating working directory: %s", build.WorkingDir)
			if err := runtime.Exec(ctx, containerID, "mkdir", "-p", build.WorkingDir); err != nil {
				return fmt.Errorf("failed to create working directory: %w", err)
			}
		}

		for _, cp := range build.CopyToContainer {
			dir := path.Dir(cp.To)
			if dir != "" && dir != "." && dir != "/" {
				log.Printf("Creating directory: %s", dir)
				if err := runtime.Exec(ctx, containerID, "mkdir", "-p", dir); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
			}
		}

		for _, cp := range build.CopyToContainer {
			files := globFiles(cp.From, projectDir)
			for _, src := range files {
				srcAbs := filepath.Join(projectDir, src)
				dst := path.Join(cp.To, filepath.Base(src))
				log.Printf("Copying %s -> %s:%s", srcAbs, containerID[:12], dst)
				if err := runtime.CopyToContainer(ctx, containerID, srcAbs, dst); err != nil {
					return fmt.Errorf("failed to copy to container: %w", err)
				}
			}
		}

		log.Printf("Executing: %s", strings.Join(build.Commands, " && "))
		cmd := fmt.Sprintf("cd %s && %s && exit", build.WorkingDir, strings.Join(build.Commands, " && "))
		if err := runtime.Exec(ctx, containerID, "sh", "-c", cmd); err != nil {
			logs, _ := runtime.GetContainerLogs(ctx, containerID)
			if logs != nil {
				data, _ := io.ReadAll(logs)
				logs.Close()
				log.Printf("Container logs:\n%s", string(data))
			}
			return fmt.Errorf("command failed: %w", err)
		}

		for _, cp := range build.CopyToLocal {
			toAbs := filepath.Join(projectDir, cp.To)
			if err := os.MkdirAll(toAbs, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			log.Printf("Copying %s:%s -> %s", containerID[:12], cp.From, toAbs)
			if err := runtime.CopyFromContainer(ctx, containerID, cp.From, toAbs); err != nil {
				return fmt.Errorf("failed to copy from container: %w", err)
			}
		}

		logs, err := runtime.GetContainerLogs(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}
		defer logs.Close()

		logData, _ := io.ReadAll(logs)
		log.Printf("Output: %s", string(logData))
	}

	return nil
}

func globFiles(pattern, projectDir string) []string {
	normalizedPattern := filepath.ToSlash(pattern)

	if strings.HasSuffix(normalizedPattern, "/") {
		fullPath := filepath.Join(projectDir, normalizedPattern)
		if _, err := os.Stat(filepath.FromSlash(fullPath)); err == nil {
			return []string{normalizedPattern}
		}
	}

	g, err := glob.Compile(normalizedPattern)
	if err != nil {
		return []string{pattern}
	}

	var files []string
	baseDir := filepath.ToSlash(filepath.Dir(normalizedPattern))
	if baseDir == "." || baseDir == "" {
		baseDir = "."
	} else {
		baseDir = filepath.Join(projectDir, baseDir)
	}

	entries, err := os.ReadDir(filepath.FromSlash(baseDir))
	if err != nil {
		return []string{pattern}
	}

	for _, entry := range entries {
		fullPath := filepath.ToSlash(filepath.Join(filepath.FromSlash(baseDir), entry.Name()))

		if entry.IsDir() {
			files = append(files, globFilesRecursive(fullPath, g, projectDir)...)
		} else {
			if g.Match(fullPath) {
				relPath, _ := filepath.Rel(projectDir, fullPath)
				relPath = filepath.ToSlash(relPath)
				files = append(files, relPath)
			}
		}
	}

	if len(files) == 0 {
		return []string{pattern}
	}
	return files
}

func globFilesRecursive(dir string, g glob.Glob, projectDir string) []string {
	var files []string

	entries, err := os.ReadDir(filepath.FromSlash(dir))
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		fullPath := filepath.ToSlash(filepath.Join(filepath.FromSlash(dir), entry.Name()))
		if entry.IsDir() {
			files = append(files, globFilesRecursive(fullPath, g, projectDir)...)
		} else {
			if g.Match(fullPath) {
				relPath, _ := filepath.Rel(projectDir, fullPath)
				relPath = filepath.ToSlash(relPath)
				files = append(files, relPath)
			}
		}
	}

	return files
}
