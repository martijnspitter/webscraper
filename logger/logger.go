package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"huurwoning/config"
	"log"
	"net/http"
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
	client *http.Client
	url    string
	stdout *log.Logger
}

type GlobalLogger struct {
	loggers  map[string]*Logger
	mu       sync.Mutex
	localDev bool
	url      string
	client   *http.Client
}

type BasicAuthTransport struct {
	Username string
	Password string
	Base     http.RoundTripper
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	return t.Base.RoundTrip(req)
}

func NewGlobalLogger(isLocalDev bool) (*GlobalLogger, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to load config: %v", err)
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := ""
	if !isLocalDev {
		url = "https://logs-prod-012.grafana.net/loki/api/v1/push"
	}

	client.Transport = &BasicAuthTransport{
		Username: config.GRAFANA_USERNAME,
		Password: config.GRAFANA_PASSWORD,
		Base:     http.DefaultTransport,
	}

	return &GlobalLogger{
		loggers:  make(map[string]*Logger),
		localDev: isLocalDev,
		url:      url,
		client:   client,
	}, nil
}

func (gl *GlobalLogger) Logger(moduleName string) *Logger {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if logger, exists := gl.loggers[moduleName]; exists {
		return logger
	}

	if gl.localDev {
		// Create stdout logger for local development
		stdout := log.New(os.Stdout, "", 0)
		l := &Logger{
			name:   moduleName,
			stdout: stdout,
		}

		l.Debug = l.localLogFunc(SeverityDebug)
		l.Info = l.localLogFunc(SeverityInfo)
		l.Warn = l.localLogFunc(SeverityWarning)
		l.Error = l.localLogFunc(SeverityError)

		gl.loggers[moduleName] = l
		return l
	}

	// Production logger
	l := &Logger{
		name:   moduleName,
		client: gl.client,
		url:    gl.url,
	}

	l.Debug = l.prodLogFunc(SeverityDebug)
	l.Info = l.prodLogFunc(SeverityInfo)
	l.Warn = l.prodLogFunc(SeverityWarning)
	l.Error = l.prodLogFunc(SeverityError)

	gl.loggers[moduleName] = l
	return l
}

func (gl *GlobalLogger) Close() {
	// No need to close anything
}

// localLogFunc creates a logging function for local development that writes to stdout in JSON format
func (l *Logger) localLogFunc(severity Severity) LogFunc {
	return func(msg string, args ...interface{}) {
		entry := struct {
			Timestamp string                 `json:"timestamp"`
			Level     string                 `json:"level"`
			Module    string                 `json:"module"`
			Message   string                 `json:"message"`
			Fields    map[string]interface{} `json:"fields,omitempty"`
		}{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
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

		jsonData, err := json.Marshal(entry)
		if err != nil {
			l.stdout.Printf("Error marshaling log entry: %v", err)
			return
		}

		l.stdout.Println(string(jsonData))
	}
}

// prodLogFunc creates a logging function for production that sends logs to Loki
func (l *Logger) prodLogFunc(severity Severity) LogFunc {
	return func(msg string, args ...interface{}) {
		// Create labels and fields maps
		labels := map[string]string{
			"module": l.name,
			"level":  string(severity),
		}
		fields := make(map[string]interface{})

		// Add additional fields from args
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				fields[args[i].(string)] = args[i+1]
			}
		}

		// Format the message with fields if present
		formattedMsg := msg
		if len(fields) > 0 {
			fieldsJSON, _ := json.Marshal(fields)
			formattedMsg = fmt.Sprintf("%s %s", msg, string(fieldsJSON))
		}

		// Format the log entry for Loki
		logEntry := struct {
			Streams []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			} `json:"streams"`
		}{
			Streams: []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			}{
				{
					Stream: labels,
					Values: [][]string{
						{
							time.Now().UTC().Format(time.RFC3339Nano),
							formattedMsg,
						},
					},
				},
			},
		}

		jsonData, err := json.Marshal(logEntry)
		if err != nil {
			log.Printf("Error marshaling log entry: %v", err)
			return
		}

		// Debug logging (consider making this conditional)
		if l.stdout != nil {
			l.stdout.Printf("Sending to Loki: %s", string(jsonData))
		}

		// Send to Loki
		resp, err := l.client.Post(l.url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error sending log to Loki: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			log.Printf("Unexpected status code from Loki: %d", resp.StatusCode)
		}
	}
}
