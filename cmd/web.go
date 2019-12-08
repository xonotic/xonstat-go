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

// Global log for the application.
var logger *log.Logger

func initLog() error {
	writer, err := syslog.New(syslog.LOG_DEBUG, "xonstat")
	if err != nil {
		return err
	}
	defer writer.Close()

	logger = log.New(writer, "", log.Ldate|log.Ltime|log.LUTC)

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetOutput(writer)

	return nil
}

func web(port string) {
	r := chi.NewRouter()

	// Save the real IP address from X-Forward-For and the like.
	r.Use(middleware.RealIP)

	// Log request metadata: the URI, the response, and how long it took.
	formatter := middleware.DefaultLogFormatter{Logger: logger, NoColor: true}
	middleware.DefaultLogger = middleware.RequestLogger(&formatter)
	r.Use(middleware.Logger)

	// Verify certain routes with do_blind_id.
	if viper.GetBool("VerifyRequests") {
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
