package cmd

import (
	"io"
	"log"
	"log/syslog"
	"os"
)

// Global log for the application, including all subcommands.
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