package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const WorkspaceDir = "workspace"

type Config struct {
	Builds    []StageConfig           `yaml:"builds"`
	Deploys   []DeploymentStep        `yaml:"deploys"`
	Container ContainerConfig         `yaml:"container"`
	Servers   map[string]ServerConfig `yaml:"servers"`
	Clone     *CloneConfig            `yaml:"clone,omitempty"`
}

type CloneConfig struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
}

type ContainerConfig struct {
	Type string `yaml:"type"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Port     int    `yaml:"port"`
	KeyPath  string `yaml:"key_path"`
	Password string `yaml:"password"`
}

type StageConfig struct {
	Name            string     `yaml:"name"`
	Image           string     `yaml:"image"`
	WorkingDir      string     `yaml:"working_dir"`
	Environment     []string   `yaml:"environment"`
	CopyToContainer []CopyPath `yaml:"copy_to_container"`
	CopyToLocal     []CopyPath `yaml:"copy_to_local"`
	Commands        []string   `yaml:"commands"`
}

type CopyPath struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type VolumeConfig struct {
	Local string `yaml:"local"`
	To    string `yaml:"to"`
}

type CopyConfig struct {
	Local   string   `yaml:"local"`
	To      string   `yaml:"to"`
	Exclude []string `yaml:"exclude"`
}

type DeploymentStep struct {
	Name     string   `yaml:"name"`
	Server   string   `yaml:"server"`
	Commands []string `yaml:"commands"`
	From     string   `yaml:"from"`
	To       string   `yaml:"to"`
}

type ConfigInfo struct {
	DirName  string
	FilePath string
}

func LoadConfigInfo(workspaceDir string) ([]ConfigInfo, error) {
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	var configs []ConfigInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		configPath := filepath.Join(workspaceDir, dirName, "config.yaml")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		configs = append(configs, ConfigInfo{
			DirName:  dirName,
			FilePath: configPath,
		})
	}

	return configs, nil
}

func Load(configPath string) (*Config, string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, "", fmt.Errorf("failed to parse config file: %w", err)
	}

	basicPath := filepath.Dir(configPath)

	applyEnvSubstitution(&cfg)

	return &cfg, basicPath, nil
}

func applyEnvSubstitution(cfg *Config) {
	for i := range cfg.Builds {
		for j := range cfg.Builds[i].Commands {
			cfg.Builds[i].Commands[j] = os.ExpandEnv(cfg.Builds[i].Commands[j])
		}
		for j := range cfg.Builds[i].Environment {
			cfg.Builds[i].Environment[j] = os.ExpandEnv(cfg.Builds[i].Environment[j])
		}
	}

	for i := range cfg.Deploys {
		for j := range cfg.Deploys[i].Commands {
			cfg.Deploys[i].Commands[j] = os.ExpandEnv(cfg.Deploys[i].Commands[j])
		}
	}
}

func FindStage(stages []StageConfig, name string) *StageConfig {
	for i := range stages {
		if stages[i].Name == name {
			return &stages[i]
		}
	}
	return nil
}

func FindDeploymentStep(steps []DeploymentStep, name string) *DeploymentStep {
	for i := range steps {
		if steps[i].Name == name {
			return &steps[i]
		}
	}
	return nil
}

func GetServer(cfg *Config, name string) *ServerConfig {
	if server, ok := cfg.Servers[name]; ok {
		return &server
	}
	return nil
}
