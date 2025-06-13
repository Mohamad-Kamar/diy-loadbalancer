package logger

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

func New(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logger.Printf("[FATAL] "+format, args...)
	os.Exit(1)
}

// WithContext adds contextual information to log messages
func (l *Logger) WithRequestID(requestID string) *ContextLogger {
	return &ContextLogger{
		logger:    l,
		requestID: requestID,
	}
}

type ContextLogger struct {
	logger    *Logger
	requestID string
}

func (cl *ContextLogger) Debug(format string, args ...interface{}) {
	cl.logger.Debug(fmt.Sprintf("[%s] %s", cl.requestID, format), args...)
}

func (cl *ContextLogger) Info(format string, args ...interface{}) {
	cl.logger.Info(fmt.Sprintf("[%s] %s", cl.requestID, format), args...)
}

func (cl *ContextLogger) Warn(format string, args ...interface{}) {
	cl.logger.Warn(fmt.Sprintf("[%s] %s", cl.requestID, format), args...)
}

func (cl *ContextLogger) Error(format string, args ...interface{}) {
	cl.logger.Error(fmt.Sprintf("[%s] %s", cl.requestID, format), args...)
}
