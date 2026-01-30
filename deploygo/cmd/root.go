package cmd

import (
	"github.com/spf13/cobra"
)

var projectName string

var RootCmd = &cobra.Command{
	Use:   "deploygo",
	Short: "DeployGo - Container-based CI/CD tool",
	Long: `DeployGo is a lightweight CI/CD tool that uses container technology
for building and deploying applications.`,
}

func init() {
	RootCmd.AddCommand(BuildCmd)
	RootCmd.AddCommand(DeployCmd)
	RootCmd.AddCommand(PipelineCmd)
	RootCmd.AddCommand(WriteCmd)
	RootCmd.AddCommand(ListCmd)
	RootCmd.AddCommand(CloneCmd)
}
