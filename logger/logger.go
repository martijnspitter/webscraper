package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Severity string

const (
	SeverityDefault   Severity = "DEFAULT"
	SeverityDebug     Severity = "DEBUG"
	SeverityInfo      Severity = "INFO"
	SeverityNotice    Severity = "NOTICE"
	SeverityWarning   Severity = "WARNING"
	SeverityError     Severity = "ERROR"
	SeverityCritical  Severity = "CRITICAL"
	SeverityAlert     Severity = "ALERT"
	SeverityEmergency Severity = "EMERGENCY"
)

type LogFunc func(msg string, args ...any)

type Logger struct {
	Debug  LogFunc
	Info   LogFunc
	Warn   LogFunc
	Error  LogFunc
	name   string
	stdout *log.Logger
}

type GlobalLogger struct {
	loggers map[string]*Logger
	mu      sync.Mutex
}

func NewGlobalLogger() (*GlobalLogger, error) {
	return &GlobalLogger{
		loggers: make(map[string]*Logger),
	}, nil
}

func (gl *GlobalLogger) Logger(moduleName string) *Logger {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if logger, exists := gl.loggers[moduleName]; exists {
		return logger
	}

	stdout := log.New(os.Stdout, "", 0)
	l := &Logger{
		name:   moduleName,
		stdout: stdout,
	}

	l.Debug = l.logFunc(SeverityDebug)
	l.Info = l.logFunc(SeverityInfo)
	l.Warn = l.logFunc(SeverityWarning)
	l.Error = l.logFunc(SeverityError)

	gl.loggers[moduleName] = l
	return l
}

func (gl *GlobalLogger) Close() {
	// No need to close anything
}

// logFunc creates a logging function that writes structured logs to stdout
func (l *Logger) logFunc(severity Severity) LogFunc {
	return func(msg string, args ...interface{}) {
		// Create the log entry structure
		entry := struct {
			Timestamp string                 `json:"timestamp"`
			Level     string                 `json:"level"`
			Module    string                 `json:"module"`
			Message   string                 `json:"message"`
			Fields    map[string]interface{} `json:"fields,omitempty"`
		}{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Level:     string(severity),
			Module:    l.name,
			Message:   fmt.Sprintf(msg, args...),
			Fields:    make(map[string]interface{}),
		}

		// Add additional fields from args
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				entry.Fields[args[i].(string)] = args[i+1]
			}
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(entry)
		if err != nil {
			l.stdout.Printf("Error marshaling log entry: %v", err)
			return
		}

		// Write to stdout
		l.stdout.Println(string(jsonData))
	}
}
