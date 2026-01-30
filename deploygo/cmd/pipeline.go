package cmd

import (
	"log"
	"os"
	"path/filepath"

	"deploygo/internal/config"
	"deploygo/internal/container"
	"deploygo/internal/git"
	"deploygo/internal/stage"

	"github.com/spf13/cobra"
)

var PipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Run build and deploy pipeline",
	Long:  `Execute all build stages followed by deployment steps`,
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

		containerMgr, err := container.NewManager(&container.ManagerConfig{
			Type: cfg.Container.Type,
		})
		if err != nil {
			log.Fatalf("Failed to initialize container runtime: %v", err)
		}
		defer containerMgr.Close()

		log.Printf("Using container runtime: %s", containerMgr.Name())
		log.Printf("Project: %s", projectName)
		log.Printf("Project directory: %s", basicPath)

		projectDir := filepath.Join(config.WorkspaceDir, projectName)
		overlaysDir := filepath.Join(projectDir, "overlays")
		sourceDir := filepath.Join(projectDir, "source")

		// 执行 Git 克隆（如果配置了）
		if cfg.Clone != nil && cfg.Clone.URL != "" {
			log.Println("=== Cloning Git Repository ===")
			log.Printf("Git URL: %s", cfg.Clone.URL)
			if cfg.Clone.Branch != "" {
				log.Printf("Branch: %s", cfg.Clone.Branch)
			}
			opts := git.CloneOptions{
				URL:       cfg.Clone.URL,
				Branch:    cfg.Clone.Branch,
				TargetDir: sourceDir,
			}
			if err := git.Clone(opts); err != nil {
				log.Fatalf("Failed to clone repository: %v", err)
			}
		}

		if _, err := os.Stat(overlaysDir); err == nil {
			log.Println("=== Copying Overlays ===")
			if err := copyOverlays(overlaysDir, sourceDir); err != nil {
				log.Fatalf("Failed to copy overlays: %v", err)
			}
		}

		if len(cfg.Builds) > 0 {
			log.Println("=== Building ===")
			for i, build := range cfg.Builds {
				log.Printf("Building %d/%d: %s", i+1, len(cfg.Builds), build.Name)
				if err := stage.RunBuilds(containerMgr, cfg, []config.StageConfig{build}, basicPath); err != nil {
					log.Fatalf("Failed to build '%s': %v", build.Name, err)
				}
			}
		}

		if len(cfg.Deploys) > 0 {
			log.Println("=== Deploying ===")
			if err := stage.RunDeploys(cfg, cfg.Deploys, basicPath); err != nil {
				log.Fatalf("Failed to deploy: %v", err)
			}
		}

		log.Println("Pipeline completed successfully!")
	},
}

func init() {
	PipelineCmd.Flags().StringVarP(&projectName, "project", "P", "", "Project name")
}
