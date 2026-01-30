package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CloneOptions 定义 Git 克隆的选项
type CloneOptions struct {
	URL       string // Git 仓库地址
	Branch    string // 分支名称，为空则使用默认分支
	TargetDir string // 目标目录（source 目录）
}

// Clone 执行 Git 克隆操作
// 1. 清空目标目录
// 2. 执行 git clone 命令
func Clone(opts CloneOptions) error {
	// 验证必要参数
	if opts.URL == "" {
		return fmt.Errorf("git URL is required")
	}
	if opts.TargetDir == "" {
		return fmt.Errorf("target directory is required")
	}

	// 设置默认分支
	branch := opts.Branch
	if branch == "" {
		branch = "master"
	}

	// 检查 git 命令是否可用
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found in PATH: %w", err)
	}

	// 获取目标目录的父目录（用于克隆）
	parentDir := filepath.Dir(opts.TargetDir)
	dirName := filepath.Base(opts.TargetDir)

	// 确保父目录存在
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// 清空目标目录（如果存在）
	if err := cleanDirectory(opts.TargetDir); err != nil {
		return fmt.Errorf("failed to clean target directory: %w", err)
	}

	// 执行 git clone 命令
	// 使用 -b 指定分支，--single-branch 只克隆指定分支
	args := []string{"clone", "-b", branch, "--single-branch", opts.URL, dirName}
	cmd := exec.Command("git", args...)
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// cleanDirectory 清空目录内容，但保留目录本身
func cleanDirectory(dir string) error {
	// 检查目录是否存在
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// 目录不存在，无需清理
		return nil
	}
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// 读取目录内容
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 删除所有子项
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}
