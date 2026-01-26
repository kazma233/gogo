package cmd

import (
	"log"
	"os"
	"path/filepath"

	"deploygo/internal/config"
	"deploygo/internal/stage"

	"github.com/spf13/cobra"
)

var deployStep string

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the application",
	Long:  `Execute deployment steps defined in workspace/<project>/config.yaml`,
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

		log.Printf("Project: %s", projectName)
		log.Printf("Project directory: %s", basicPath)

		if deployStep != "" {
			step := config.FindDeploymentStep(cfg.Deploys, deployStep)
			if step == nil {
				log.Fatalf("Deploy step '%s' not found", deployStep)
			}
			if err := stage.RunDeploys(cfg, []config.DeploymentStep{*step}, basicPath); err != nil {
				log.Fatalf("Failed to execute '%s': %v", deployStep, err)
			}
		} else {
			if err := stage.RunDeploys(cfg, cfg.Deploys, basicPath); err != nil {
				log.Fatalf("Failed to execute deploy steps: %v", err)
			}
		}

		log.Println("Deployment completed successfully!")
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&projectName, "project", "P", "", "Project name")
	DeployCmd.Flags().StringVarP(&deployStep, "step", "s", "", "Specific deployment step to execute")
}
