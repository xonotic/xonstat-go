package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd prints out the build version and exits
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the build version",
	Long: `Show the build version in the form of the git commit hash from which the binary was built`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("XonStat version %s, built on %s\n", Commit, BuildDt)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
