package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(time.RFC3339)
	level := strings.ToUpper(entry.Level.String())
	// Define the maximum length of the log level (e.g., DEBUG, INFO, WARN, ERROR)
	maxLevelLength := 7
	// Pad the level with spaces to ensure alignment
	paddedLevel := fmt.Sprintf("%-*s", maxLevelLength, level)
	return []byte(fmt.Sprintf("%s %s %s\n", timestamp, paddedLevel, entry.Message)), nil
}

func NewLogger(verbose bool) *LogrusLogger {
	log := logrus.New()
	log.SetFormatter(&PlainFormatter{})

	log.SetOutput(os.Stdout)
	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	return &LogrusLogger{logger: log}
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LogrusLogger) SetOutput(output io.Writer) {
	l.logger.SetOutput(output)
}

func (l *LogrusLogger) SetLevel(level string) {
	lvl, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		l.logger.SetLevel(logrus.InfoLevel)
		return
	}
	l.logger.SetLevel(lvl)
}

func (l *LogrusLogger) Write(level Level, args ...interface{}) {
	switch level {
	case DebugLevel:
		l.Debug(args...)
	case InfoLevel:
		l.Info(args...)
	case WarnLevel:
		l.Warn(args...)
	case ErrorLevel:
		l.Error(args...)
	default:
		l.Info(args...)
	}
}

func (l *LogrusLogger) Writef(level Level, format string, args ...interface{}) {
	switch level {
	case DebugLevel:
		l.Debugf(format, args...)
	case InfoLevel:
		l.Infof(format, args...)
	case WarnLevel:
		l.Warnf(format, args...)
	case ErrorLevel:
		l.Errorf(format, args...)
	default:
		l.Infof(format, args...)
	}
}
