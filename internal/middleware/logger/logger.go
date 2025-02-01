package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type Logger struct {
	env         string
	component   string
	logLevel    LogLevel
	requestID   string
	requestPath string
}

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Path      string `json:"path,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Trace     string `json:"trace,omitempty"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
}

type ctxKeyLogger struct{}

var levelMap = map[string]LogLevel{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
}

// NewLogger initializes the logger with the provided environment and component.
func NewLogger(env, component string) *Logger {
	logLevelStr := os.Getenv("LOG_LEVEL")
	logLevel := parseLogLevel(logLevelStr)

	return &Logger{
		env:       env,
		component: component,
		logLevel:  logLevel,
	}
}

// parseLogLevel converts a string log level to the corresponding LogLevel.
func parseLogLevel(level string) LogLevel {
	if logLevel, exists := levelMap[level]; exists {
		return logLevel
	}
	return InfoLevel // Default log level
}

// Add logger to the context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, logger)
}

// GetLogger retrieves the logger from the context
func GetLogger(c *gin.Context) *Logger {
	// Retrieve the logger from the Gin context
	logger, exists := c.Get("logger")
	if !exists {
		// If logger doesn't exist, create a new one
		defaultLogger := NewLogger("local", "default")
		log.Println("Logger not found, returning default logger")
		return defaultLogger
	}

	// Return the logger
	return logger.(*Logger)
}

// Log function writes log entries in JSON format and applies configurable log filtering
func (l *Logger) log(level LogLevel, levelStr, msg, trace string) {
	if level < l.logLevel {
		return
	}

	_, file, line, _ := runtime.Caller(3)
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     levelStr,
		Component: l.component,
		Message:   msg,
		Trace:     trace,
		Path:      l.requestPath,
		RequestID: l.requestID,
		File:      file,
		Line:      line,
	}

	if l.env == "prod" {
		l.logToKibana(entry)
	} else {
		l.logToConsole(entry)
	}
}

// Debugf logs a debug message without needing to pass gin.Context
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, "debug", fmt.Sprintf(format, args...), "")
}

// Infof logs an info message without needing to pass gin.Context
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, "info", fmt.Sprintf(format, args...), "")
}

// Warnf logs a warning message without needing to pass gin.Context
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, "warn", fmt.Sprintf(format, args...), "")
}

// Errorf logs an error message without needing to pass gin.Context
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, "error", fmt.Sprintf(format, args...), "")
}

// Helper method for logging to the console
func (l *Logger) logToConsole(entry LogEntry) {
	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}

// Simulated Kibana log persistence
func (l *Logger) logToKibana(entry LogEntry) {
	data, _ := json.Marshal(entry)
	// Replace this with your Kibana or log aggregation service integration
	fmt.Printf("KIBANA_LOG: %s\n", data)
}

// Middleware to add logger to the context in Gin
func LoggerMiddleware(env, component string) gin.HandlerFunc {
	logger := NewLogger(env, component)
	return func(c *gin.Context) {
		// Generate a unique request ID (can use a library like `github.com/google/uuid` for a UUID)
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Set("RequestID", requestID)

		// Add logger to the context
		logger.requestID = requestID
		logger.requestPath = c.Request.URL.Path
		c.Set("logger", logger)

		// Log the incoming request
		log := GetLogger(c)
		log.Infof("Request started", "path: %s", c.Request.URL.Path)

		// Continue processing the request
		c.Next()

		// Log the outgoing response after handling the request
		log.Infof("Request completed", "path: %s, status: %d", c.Request.URL.Path, c.Writer.Status())
	}
}
