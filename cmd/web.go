package cmd

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/handlers"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global log for the application.
var logger *log.Logger

func initLog() error {
	syslogWriter, err := syslog.New(syslog.LOG_DEBUG, "xonstat")
	if err != nil {
		return err
	}
	defer syslogWriter.Close()

	// Multiplex log messages to syslog and standard out.
	multiwriter := io.MultiWriter(syslogWriter, os.Stdout)

	logger = log.New(multiwriter, "", log.Ldate|log.Ltime|log.LUTC)

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetOutput(multiwriter)

	return nil
}

func web(port string) {
	dsn := viper.GetString("ConnStr")
	db, err := models.NewPGDatastore(dsn)
	if err != nil {
		log.Fatal("Unable to initialize database connection.")
	}

	requestLogger := lumberjack.Logger{
		Filename:   viper.GetString("RequestsLogFile"),
		MaxSize:    viper.GetInt("RequestsMaxSize"),
		MaxBackups: viper.GetInt("RequestsMaxBackups"),
		MaxAge:     viper.GetInt("RequestsMaxAge"),
		Compress:   true,
	}

	env := handlers.NewAppEnv(db, &requestLogger)

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
	r.Post("/stats/submit", env.SubmissionHandler)
	r.Get("/summary", env.SummaryStatsHandler)

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
