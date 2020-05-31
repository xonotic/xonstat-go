package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"gitlab.com/xonotic/xonstat/pkg/submission"
	"github.com/spf13/cobra"
)

// inspectCmd parses a given submission file and prints out its parse form in JSON format to STDOUT
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect submission files in JSON format",
	Long: `Inspect XonStat submission files using the JSON format. This is mostly used for 
debugging purposes when you want to see the raw data that results from the parsing phase
of the stats submission process.`,
	Run: func(cmd *cobra.Command, args []string) {
		filename := cmd.Flag("file").Value.String()
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Unable to open the submission file %s for inspection\n", filename)
			os.Exit(1)
		}
		defer f.Close()

		body := bufio.NewReader(f)
		rawSubmission, err := submission.NewRawSubmission(body)
		if err != nil {
			fmt.Printf("Unable to parse %s into a RawSubmission: %s\n", filename, err)
			os.Exit(1)
		}

		s, err := submission.NewSubmission(rawSubmission)
		if err != nil {
			fmt.Printf("Could not parse %s into an actual Submission: %s\n", filename, err)
			os.Exit(1)
		}

		submissionBytes, err := json.MarshalIndent(s, "", "  ")
		fmt.Println(string(submissionBytes))
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringP("file", "f", "", "submission file")
}
