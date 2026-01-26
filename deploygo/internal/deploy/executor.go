package deploy

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"deploygo/internal/config"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type FileUploader interface {
	UploadDir(source, dest string, excludes []string) error
	Upload(source, dest string, excludes []string) error
}

type SSHConfig struct {
	Host     string
	User     string
	Port     int
	KeyPath  string
	Password string
	Timeout  time.Duration
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

type SSHExecutor struct {
	config SSHConfig
	client *ssh.Client
}

func NewSSHExecutor(server *config.ServerConfig) (*SSHExecutor, error) {
	return &SSHExecutor{
		config: SSHConfig{
			Host:     server.Host,
			User:     server.User,
			Port:     server.Port,
			KeyPath:  server.KeyPath,
			Password: server.Password,
			Timeout:  30 * time.Second,
		},
	}, nil
}

func (s *SSHExecutor) Connect() error {
	var authMethods []ssh.AuthMethod

	if s.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.config.Password))
	}

	if s.config.KeyPath != "" {
		keyPath := expandHome(s.config.KeyPath)
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method provided")
	}

	host := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	client, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            s.config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.config.Timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}

	s.client = client
	return nil
}

func (s *SSHExecutor) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *SSHExecutor) Execute(command string) ([]byte, error) {
	if s.client == nil {
		if err := s.Connect(); err != nil {
			return nil, err
		}
	}

	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return output, fmt.Errorf("ssh command failed: %w, output: %s", err, string(output))
	}

	return output, nil
}

func (s *SSHExecutor) ExecuteBatch(commands []string) ([]byte, error) {
	script := strings.Join(commands, " && ")
	return s.Execute(script)
}

type SFTPConfig struct {
	Host     string
	User     string
	Port     int
	KeyPath  string
	Password string
	Timeout  time.Duration
}

type SFTPUploader struct {
	config SFTPConfig
	client *ssh.Client
}

func NewSFTPUploader(server *config.ServerConfig) (*SFTPUploader, error) {
	return &SFTPUploader{
		config: SFTPConfig{
			Host:     server.Host,
			User:     server.User,
			Port:     server.Port,
			KeyPath:  server.KeyPath,
			Password: server.Password,
			Timeout:  30 * time.Second,
		},
	}, nil
}

func (s *SFTPUploader) Connect() error {
	var authMethods []ssh.AuthMethod

	if s.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.config.Password))
	}

	if s.config.KeyPath != "" {
		keyPath := expandHome(s.config.KeyPath)
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method provided")
	}

	host := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	client, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            s.config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.config.Timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}

	s.client = client
	return nil
}

func (s *SFTPUploader) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *SFTPUploader) Upload(source, dest string, excludes []string) error {
	if s.client == nil {
		if err := s.Connect(); err != nil {
			return err
		}
	}

	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	dest = filepath.ToSlash(dest)

	srcInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if srcInfo.IsDir() {
		return s.uploadDir(sftpClient, source, dest, excludes)
	}

	return s.uploadFile(sftpClient, source, dest)
}

func (s *SFTPUploader) uploadFile(sftp *sftp.Client, source, dest string) error {
	log.Printf("SFTP upload file: %s -> %s", source, dest)

	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 确保目标路径使用 Unix 风格的斜杠（Linux 服务器）
	dest = filepath.ToSlash(dest)
	dstDir := filepath.Dir(dest)
	// 再次确保目录路径是 Unix 风格
	dstDir = filepath.ToSlash(dstDir)

	log.Printf("SFTP creating directory: %s", dstDir)
	if err := sftp.MkdirAll(dstDir); err != nil {
		return fmt.Errorf("failed to create destination directory '%s': %w", dstDir, err)
	}

	log.Printf("SFTP creating file: %s", dest)
	dstFile, err := sftp.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %w", dest, err)
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("SFTP upload complete: %s -> %s", source, dest)
	return nil
}

func (s *SFTPUploader) uploadDir(sftp *sftp.Client, source, dest string, excludes []string) error {
	// 确保目标路径使用 Unix 风格的斜杠（Linux 服务器）
	dest = filepath.ToSlash(dest)

	return filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径（本地格式）
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// 转换为 Unix 风格的斜杠用于远程路径
		relPathUnix := filepath.ToSlash(relPath)

		if relPathUnix == "." {
			return nil
		}

		// 检查排除规则（使用 Unix 风格路径）
		for _, exclude := range excludes {
			if strings.HasPrefix(relPathUnix, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// 构建目标路径（Unix 风格）
		dstPath := dest + "/" + relPathUnix
		log.Printf("DEBUG uploadDir dstPath=%s", dstPath)

		if info.IsDir() {
			return sftp.MkdirAll(dstPath)
		}

		return s.uploadFile(sftp, path, dstPath)
	})
}

func (s *SFTPUploader) UploadDir(source, dest string, excludes []string) error {
	return s.Upload(source+"/", dest, excludes)
}
