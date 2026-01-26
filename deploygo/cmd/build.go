package cmd

import (
	"log"
	"os"
	"path/filepath"

	"deploygo/internal/config"
	"deploygo/internal/container"
	"deploygo/internal/stage"

	"github.com/spf13/cobra"
)

var buildStage string

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all or specific stages",
	Long:  `Build all stages defined in workspace/<project>/config.yaml or a specific stage`,
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

		if buildStage != "" {
			build := config.FindStage(cfg.Builds, buildStage)
			if build == nil {
				log.Fatalf("Build '%s' not found", buildStage)
			}
			if err := stage.RunBuilds(containerMgr, cfg, []config.StageConfig{*build}, basicPath); err != nil {
				log.Fatalf("Failed to build '%s': %v", buildStage, err)
			}
		} else {
			for i, build := range cfg.Builds {
				log.Printf("Building %d/%d: %s", i+1, len(cfg.Builds), build.Name)
				if err := stage.RunBuilds(containerMgr, cfg, []config.StageConfig{build}, basicPath); err != nil {
					log.Fatalf("Failed to build '%s': %v", build.Name, err)
				}
			}
		}

		log.Println("Build completed successfully!")
	},
}

func init() {
	BuildCmd.Flags().StringVarP(&projectName, "project", "P", "", "Project name")
	BuildCmd.Flags().StringVarP(&buildStage, "stage", "s", "", "Specific stage to build")
}
