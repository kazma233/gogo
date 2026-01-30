package cmd

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"deploygo/internal/config"
	"deploygo/internal/stage"

	"github.com/spf13/cobra"
)

var WriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Copy overlays to source directory",
	Long:  `Copy all files from overlays/ to source/ directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if projectName == "" {
			log.Fatal("Please specify a project using -P flag")
		}

		projectDir := filepath.Join(config.WorkspaceDir, projectName)
		overlaysDir := filepath.Join(projectDir, "overlays")
		sourceDir := filepath.Join(projectDir, "source")

		if _, err := os.Stat(overlaysDir); os.IsNotExist(err) {
			log.Printf("No overlays directory found for project '%s'", projectName)
			return
		}

		log.Printf("Project: %s", projectName)
		log.Printf("Overlays: %s", overlaysDir)
		log.Printf("Source: %s", sourceDir)

		if err := copyOverlays(overlaysDir, sourceDir); err != nil {
			log.Fatalf("Failed to copy overlays: %v", err)
		}

		log.Println("Write completed successfully!")
	},
}

func copyOverlays(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return stage.CopyFile(path, targetPath)
	})
}

func init() {
	WriteCmd.Flags().StringVarP(&projectName, "project", "P", "", "Project name")
}
