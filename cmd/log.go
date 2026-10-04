package cmd

import (
	"log"
	"os"
)

// Global log for the application, including all subcommands.
var logger *log.Logger

func initLog() error {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC)

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	return nil
}

