// Copyright 2022 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"log"
	"os"
	"strings"

	"github.com/bombsimon/logrusr/v2"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// Logger exposes logging capabilities using
// https://pkg.go.dev/github.com/go-logr/logr.
type Logger struct {
	*logr.Logger
}

// NewLogger creates an instance of *Logger.
// TODO(log): Consider adopting production config from zap.
func NewLogger(logLevel Level) *Logger {
	logrusLog := logrus.New()

	// Set log level from logrus
	logrusLevel := parseLogrusLevel(logLevel)
	logrusLog.SetLevel(logrusLevel)

	return NewLogrusLogger(logrusLog)
}

// NewCronLogger creates an instance of *Logger.
func NewCronLogger(logLevel Level) *Logger {
	logrusLog := logrus.New()

	// Configure logger
	// Don't log to stderr by default (stackdriver treats stderr as error severity)
	logrusLog.SetOutput(os.Stdout)
	// for stackdriver, see: https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
	logrusLog.SetFormatter(&logrus.JSONFormatter{FieldMap: logrus.FieldMap{
		logrus.FieldKeyLevel: "severity",
		logrus.FieldKeyMsg:   "message",
	}})

	// Set log level from logrus
	logrusLevel := parseLogrusLevel(logLevel)
	logrusLog.SetLevel(logrusLevel)

	return NewLogrusLogger(logrusLog)
}

// NewLogrusLogger creates an instance of *Logger backed by the supplied
// logrusLog instance.
func NewLogrusLogger(logrusLog *logrus.Logger) *Logger {
	logrLogger := logrusr.New(logrusLog)
	logger := &Logger{
		&logrLogger,
	}
	return logger
}

// ParseLevel takes a string level and returns the sclog Level constant.
// If the level is not recognized, it defaults to `sclog.InfoLevel` to swallow
// potential configuration errors/typos when specifying log levels.
// https://pkg.go.dev/github.com/sirupsen/logrus#ParseLevel
func ParseLevel(lvl string) Level {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	case "error":
		return ErrorLevel
	case "warn":
		return WarnLevel
	case "info":
		return InfoLevel
	case "debug":
		return DebugLevel
	case "trace":
		return TraceLevel
	}

	return DefaultLevel
}

// Level is a string representation of log level, which can easily be passed as
// a parameter, in lieu of defined types in upstream logging packages.
type Level string

// Log levels.
const (
	DefaultLevel       = InfoLevel
	TraceLevel   Level = "trace"
	DebugLevel   Level = "debug"
	InfoLevel    Level = "info"
	WarnLevel    Level = "warn"
	ErrorLevel   Level = "error"
	PanicLevel   Level = "panic"
	FatalLevel   Level = "fatal"
)

func (l Level) String() string {
	return string(l)
}

func parseLogrusLevel(lvl Level) logrus.Level {
	logrusLevel, err := logrus.ParseLevel(lvl.String())
	if err != nil {
		log.Printf(
			"defaulting to INFO log level, as %s is not a valid log level: %+v",
			lvl,
			err,
		)

		logrusLevel = logrus.InfoLevel
	}

	return logrusLevel
}
