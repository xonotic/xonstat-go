package cmd

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/spf13/cobra"
	"gitlab.com/antibody/xonstat/pkg/submission"
)

func initLog() error {
	writer, err := syslog.New(syslog.LOG_DEBUG, "xonstat")
	if err != nil {
		return err
	}
	defer writer.Close()

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetOutput(writer)

	return nil
}

func web(port string) {
	r := chi.NewRouter()
	r.Post("/stats/submit", func(w http.ResponseWriter, r *http.Request) {
		body := bufio.NewReader(r.Body)
		rawSubmission, err := submission.NewRawSubmission(body)
		if err != nil {
			http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
			return
		}

		_, err = submission.NewSubmission(rawSubmission)
		if err != nil {
			http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	})

	log.Printf("Starting XonStat web application server on port %s...", port)

	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, r)
}

// webCmd starts up the web application server
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Run the web application server",
	Long:  `Run the XonStat web application server.`,
	Run: func(cmd *cobra.Command, args []string) {
		port := cmd.Flag("port").Value.String()
		web(port)
	},
}

func init() {
	// set up logging
	err := initLog()
	if err != nil {
		log.Fatal("Unable to initialize logging.")
	}

	rootCmd.AddCommand(webCmd)
	webCmd.Flags().StringP("port", "p", "8080", "port number")
}
