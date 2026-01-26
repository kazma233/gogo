package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig `mapstructure:"server"`
	Database DBConfig     `mapstructure:"database"`
	JWT      JWTConfig    `mapstructure:"jwt"`
	SMTP     SMTPConfig   `mapstructure:"smtp"`
	GitHub   GitHubConfig `mapstructure:"github"`
	Upload   UploadConfig `mapstructure:"upload"`
	Log      LogConfig    `mapstructure:"log"`
	Auth     AuthConfig   `mapstructure:"auth"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Mode        string `mapstructure:"mode"`
	MaxBodySize string `mapstructure:"max_body_size"`
}

type DBConfig struct {
	Path string `mapstructure:"path"`
}

type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn string `mapstructure:"expires_in"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
}

type GitHubConfig struct {
	Token      string `mapstructure:"token"`
	Owner      string `mapstructure:"owner"`
	Repo       string `mapstructure:"repo"`
	Branch     string `mapstructure:"branch"`
	PathPrefix string `mapstructure:"path_prefix"`
}

type UploadConfig struct {
	ThumbnailWidth   int    `mapstructure:"thumbnail_width"`
	ThumbnailQuality int    `mapstructure:"thumbnail_quality"`
	ThumbnailFormat  string `mapstructure:"thumbnail_format"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Path   string `mapstructure:"path"`
}

type AuthConfig struct {
	AllowedEmails []string `mapstructure:"allowed_emails"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.port", 6100)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.max_body_size", "100MB")
	v.SetDefault("database.path", "./data/picstash.db")
	v.SetDefault("jwt.expires_in", "24h")
	v.SetDefault("smtp.port", "587")
	v.SetDefault("github.branch", "main")
	v.SetDefault("github.path_prefix", "images")
	v.SetDefault("upload.thumbnail_width", 1920)
	v.SetDefault("upload.thumbnail_quality", 80)
	v.SetDefault("upload.thumbnail_format", "jpeg")
	v.SetDefault("log.level", "debug")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.path", "./logs")

	v.BindEnv("server.mode", "SERVER_MODE")

	if configPath == "" {
		configPath = "config.yaml"
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("获取配置文件绝对路径失败: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", absPath)
	}

	v.SetConfigFile(absPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}
