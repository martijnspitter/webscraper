package logger

import (
	"context"
	"fmt"
	"log"
	"sync"

	"cloud.google.com/go/logging"
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
	logger *logging.Logger
	name   string
}

type GlobalLogger struct {
	client   *logging.Client
	loggers  map[string]*Logger
	mu       sync.Mutex
	localDev bool
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Severity  Severity               `json:"severity"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func NewGlobalLogger(isLocalDev bool) (*GlobalLogger, error) {
	ctx := context.Background()
	if !isLocalDev {
		client, err := logging.NewClient(ctx, "huurwoning")
		if err != nil {
			return nil, fmt.Errorf("failed to create logging client: %v", err)
		}

		return &GlobalLogger{
			client:  client,
			loggers: make(map[string]*Logger),
		}, nil
	} else {
		return &GlobalLogger{
			client:   nil,
			loggers:  make(map[string]*Logger),
			localDev: true,
		}, nil
	}
}

func (gl *GlobalLogger) Logger(moduleName string) *Logger {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if logger, exists := gl.loggers[moduleName]; exists {
		return logger
	}

	if !gl.localDev {
		logger := gl.client.Logger(moduleName)
		l := &Logger{
			logger: logger,
			name:   moduleName,
		}

		l.Debug = l.logFunc(logging.Debug)
		l.Info = l.logFunc(logging.Info)
		l.Warn = l.logFunc(logging.Warning)
		l.Error = l.logFunc(logging.Error)

		gl.loggers[moduleName] = l
		return l
	}

	return &Logger{
		Debug: func(msg string, args ...interface{}) {
			log.Printf("DEBUG: %s: %v", moduleName, fmt.Sprintf(msg, args...))
		},
		Info: func(msg string, args ...interface{}) {
			log.Printf("INFO: %s: %v", moduleName, fmt.Sprintf(msg, args...))
		},
		Warn: func(msg string, args ...interface{}) {
			log.Printf("WARN: %s: %v", moduleName, fmt.Sprintf(msg, args...))
		},
		Error: func(msg string, args ...interface{}) {
			log.Printf("ERROR: %s: %v", moduleName, fmt.Sprintf(msg, args...))
		},
	}
}

func (gl *GlobalLogger) Close() {
	if err := gl.client.Close(); err != nil {
		log.Printf("Failed to close logging client: %v", err)
	}
}

func (l *Logger) logFunc(severity logging.Severity) LogFunc {
	return func(msg string, args ...interface{}) {
		payload := make(map[string]interface{})
		payload["message"] = l.name + ": " + msg

		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				payload[args[i].(string)] = args[i+1]
			}
		}

		l.logger.Log(logging.Entry{
			Severity: severity,
			Payload:  payload,
		})
	}
}
