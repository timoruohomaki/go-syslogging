package main

import (
	"log"
	"log/syslog"

	"logger"
)

func main() {
	// Create a logger with default options
	opts := logger.DefaultOptions()

	// Or customize the options
	opts = logger.LoggerOptions{
		MinLevel:   logger.INFO,
		Facility:   syslog.LOG_LOCAL0,
		UseSyslog:  true,
		SyslogTag:  "go-syslog",
		SyslogIP:   "10.0.2.62", // Your syslog server IP
		SyslogPort: 514,
	}

	l, err := logger.NewLogger(opts)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer l.Close()

	// levels according to RFC something
	// Use the logger
	l.Info("Application started")
	l.Debug("This is a debug message")
	l.Warn("Something might be wrong: %s", "connection timeout")
	l.Error("Failed to connect to database: %v", err)

	// This will log and exit with status code 1
	l.Fatal("Unrecoverable error occurred")
}
