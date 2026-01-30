package cmd

import (
	"log"
	"os"
	"path/filepath"

	"deploygo/internal/config"
	"deploygo/internal/git"

	"github.com/spf13/cobra"
)

var CloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone Git repository to source directory",
	Long:  `Clone a Git repository to the project's source directory, replacing any existing content.`,
	Run: func(cmd *cobra.Command, args []string) {
		if projectName == "" {
			log.Fatal("Please specify a project using -P flag")
		}

		configPath := filepath.Join(config.WorkspaceDir, projectName, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("Project '%s' not found (file: %s)", projectName, configPath)
		}

		cfg, basicPath, err := config.Load(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}

		// 检查是否配置了 clone
		if cfg.Clone == nil || cfg.Clone.URL == "" {
			log.Fatalf("No clone configuration found for project '%s'. Please configure 'clone.url' in config.yaml", projectName)
		}

		sourceDir := filepath.Join(basicPath, "source")

		log.Printf("Project: %s", projectName)
		log.Printf("Git URL: %s", cfg.Clone.URL)
		if cfg.Clone.Branch != "" {
			log.Printf("Branch: %s", cfg.Clone.Branch)
		} else {
			log.Printf("Branch: master (default)")
		}
		log.Printf("Target: %s", sourceDir)

		// 执行克隆
		opts := git.CloneOptions{
			URL:       cfg.Clone.URL,
			Branch:    cfg.Clone.Branch,
			TargetDir: sourceDir,
		}

		if err := git.Clone(opts); err != nil {
			log.Fatalf("Failed to clone repository: %v", err)
		}

		log.Println("Clone completed successfully!")
	},
}

func init() {
	CloneCmd.Flags().StringVarP(&projectName, "project", "P", "", "Project name")
}
