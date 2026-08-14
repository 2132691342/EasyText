package utils

import (
	"log"
	"os"
	"path/filepath"
)

// Logger is a simple logger for the application
type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// Global logger instance
var Log *Logger

// InitLogger initializes the global logger
func InitLogger(logPath string) error {
	// Create log directory if it doesn't exist
	if logPath != "" {
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		Log = &Logger{
			infoLog:  log.New(logFile, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
			errorLog: log.New(logFile, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
			debugLog: log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		}
	} else {
		Log = &Logger{
			infoLog:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
			errorLog: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
			debugLog: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		}
	}
	return nil
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLog.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLog.Printf(format, v...)
}
