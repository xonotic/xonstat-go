package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// submitCmd takes a stats submission as a file and POSTs it to a XonStat server.
var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit stats requests from files",
	Long:  `Submit stats requests from files.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		url, _ := cmd.Flags().GetString("url")
		verbose, _ := cmd.Flags().GetBool("verbose")

		f, err := os.Open(file)
		if err != nil {
			fmt.Printf("Issue opening file %s\n", file)
			os.Exit(1)
		}
		defer f.Close()

		r := bufio.NewReader(f)

		var buf bytes.Buffer
		var sig string
		line, err := r.ReadBytes('\n')
		for err == nil {
			// The 'SIG' line is special. It's written to the POST file but it's used as an HTTP header.
			// For this reason we'll not include it in the body, else the d0 verification will not return
			// the expected result.
			if strings.HasPrefix(string(line), "SIG") {
				pieces := strings.Split(string(line), " ")
				if len(pieces[1]) > 0 {
					sig = strings.TrimSuffix(pieces[1], "\n")
				}
			} else {
				buf.Write(line)
			}

			line, err = r.ReadBytes('\n')
		}

		req, _ := http.NewRequest("POST", url, &buf)
		if sig != "" {
			fmt.Println("adding d0 header")
			req.Header.Add("X-D0-Blind-Id-Detached-Signature", sig)
		}
		req.ContentLength = int64(buf.Len())

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Issue with the response %s\n", err)
			os.Exit(1)
		}
		defer res.Body.Close()

		fmt.Printf("%s: %s\n", file, res.Status)
		if verbose {
			output, _ := ioutil.ReadAll(res.Body)
			fmt.Print(string(output))
		}

		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().StringP("file", "f", "", "submission request file")
	submitCmd.Flags().StringP("url", "u", "http://localhost:8080/stats/submit", "XonStat server submission URL")
	submitCmd.Flags().BoolP("verbose", "v", false, "Show response body")
}
