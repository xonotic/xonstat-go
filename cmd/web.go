package cmd

import (
	"fmt"
	"log"
	"log/syslog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/antibody/xonstat/pkg/handlers"
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

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // TODO: emit to the already-configured logger here

	if viper.GetBool("VerifyRequests") {
		log.Println("Verifying requests.")
		r.Use(handlers.D0Verify)
	}

	// Register all routes and handlers.
	r.Post("/stats/submit", handlers.SubmissionHandler)

	// Start the web application server on the specified port.
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
