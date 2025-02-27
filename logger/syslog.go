package logger

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger provides a structured logging interface with syslog support
type Logger struct {
	syslogWriter *syslog.Writer
	stdLogger    *log.Logger
	minLevel     LogLevel
	facility     syslog.Priority
	useSyslog    bool
}

// LoggerOptions allows for configuring the logger
type LoggerOptions struct {
	MinLevel   LogLevel
	Facility   syslog.Priority
	UseSyslog  bool
	SyslogTag  string
	SyslogIP   string
	SyslogPort int
}

// DefaultOptions provides sensible defaults
func DefaultOptions() LoggerOptions {
	return LoggerOptions{
		MinLevel:   INFO,
		Facility:   syslog.LOG_LOCAL0,
		UseSyslog:  true,
		SyslogTag:  filepath.Base(os.Args[0]),
		SyslogIP:   "127.0.0.1",
		SyslogPort: 514,
	}
}

// NewLogger creates a new logger with the given options
func NewLogger(opts LoggerOptions) (*Logger, error) {
	logger := &Logger{
		stdLogger: log.New(os.Stderr, "", 0),
		minLevel:  opts.MinLevel,
		facility:  opts.Facility,
		useSyslog: opts.UseSyslog,
	}

	if opts.UseSyslog {
		syslogAddr := fmt.Sprintf("%s:%d", opts.SyslogIP, opts.SyslogPort)
		writer, err := syslog.Dial("udp", syslogAddr, opts.Facility, opts.SyslogTag)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to syslog: %v", err)
		}
		logger.syslogWriter = writer
	}

	return logger, nil
}

// Close properly closes the logger
func (l *Logger) Close() error {
	if l.useSyslog && l.syslogWriter != nil {
		return l.syslogWriter.Close()
	}
	return nil
}

// levelToString converts a LogLevel to its string representation
func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// log logs a message with the given level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format(time.RFC3339)
	caller := getCallerInfo(3) // skip through our logging stack
	logLine := fmt.Sprintf("%s [%s] %s - %s", timestamp, levelToString(level), caller, message)

	// Always log to stdout/stderr
	l.stdLogger.Print(logLine)

	// Log to syslog if enabled
	if l.useSyslog && l.syslogWriter != nil {
		var err error
		switch level {
		case DEBUG:
			err = l.syslogWriter.Debug(message)
		case INFO:
			err = l.syslogWriter.Info(message)
		case WARN:
			err = l.syslogWriter.Warning(message)
		case ERROR:
			err = l.syslogWriter.Err(message)
		case FATAL:
			err = l.syslogWriter.Crit(message)
		}

		if err != nil {
			l.stdLogger.Printf("Failed to write to syslog: %v", err)
		}
	}

	// Exit on fatal errors
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}
