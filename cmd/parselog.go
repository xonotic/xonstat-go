package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func writeMatchFile(gameTypeCd string, matchID string, body []string) {
	fn := fmt.Sprintf("%s_%s.txt", gameTypeCd, matchID)
	fRaw, err := os.Create(fn)
	if err != nil {
		fmt.Printf("Error creating match_id %s.\n", matchID)
	}
	defer fRaw.Close()

	f := bufio.NewWriter(fRaw)
	for _, v := range body {
		fmt.Fprint(f, v)
	}
	f.Flush()
}

func parseLog(file string, anonymize bool, matchID string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	var inreq bool
	var currMatchID string
	var gameTypeCd string
	lines := make([]string, 0, 100)
	hashkeys := make(map[string]int, 500)
	maxPID := 0

	line, err := r.ReadString('\n')
	for err == nil {
		switch {
		case strings.Contains(line, "BEGIN REQUEST BODY"):
			inreq = true
		case strings.Contains(line, "END REQUEST BODY"):
			if matchID == "" || currMatchID == matchID {
				// a simple heuristic to filter *most* empty games
				if len(lines) >= 20 {
					writeMatchFile(gameTypeCd, currMatchID, lines)
				}
			}
			inreq = false
			lines = make([]string, 0, 100)
		case inreq:
			if anonymize && line[0] == 'P' {
				if line[2:5] != "bot" && line[2:8] != "player" {
					hashkey := line[2 : len(line)-1]
					if _, ok := hashkeys[hashkey]; !ok {
						hashkeys[hashkey] = maxPID
						maxPID = maxPID + 1
					}
					line = fmt.Sprintf("P %d\n", hashkeys[hashkey])
				}
			}
			if line[0] == 'I' {
				currMatchID = line[2 : len(line)-1]
			}
			if line[0] == 'G' {
				gameTypeCd = line[2 : len(line)-1]
			}
			lines = append(lines, line)
		}

		line, err = r.ReadString('\n')
	}

	return nil
}

// parseLogCmd parses a given xonstat log file and creates individual match files from it.
var parseLogCmd = &cobra.Command{
	Use:   "parseLog",
	Short: "Parse XonStat submission POST request logs.",
	Long: `Parse XonStat submission POST request log files into individual match files
	for later inspection or even resubmission.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		anonymize, _ := cmd.Flags().GetBool("anonymize")
		matchID, _ := cmd.Flags().GetString("match")

		err := parseLog(file, anonymize, matchID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(parseLogCmd)
	parseLogCmd.Flags().StringP("file", "f", "", "log file")
	parseLogCmd.Flags().BoolP("anonymize", "a", false, "anonymize player hashkeys")
	parseLogCmd.Flags().StringP("match", "m", "", "extract only this match_id")
}
