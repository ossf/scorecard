// Copyright 2022 Security Scorecard Authors
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

	logrLogger := logrusr.New(logrusLog)
	logger := &Logger{
		&logrLogger,
	}

	return logger
}

// ParseLevel takes a string level and returns the Logrus log level constant.
// If the level is not recognized, it defaults to `logrus.InfoLevel` to swallow
// potential configuration errors/typos when specifying log levels.
// https://pkg.go.dev/github.com/sirupsen/logrus#ParseLevel
func ParseLevel(lvl string) logrus.Level {
	logLevel := DefaultLevel

	switch strings.ToLower(lvl) {
	case "panic":
		logLevel = PanicLevel
	case "fatal":
		logLevel = FatalLevel
	case "error":
		logLevel = ErrorLevel
	case "warn":
		logLevel = WarnLevel
	case "info":
		logLevel = InfoLevel
	case "debug":
		logLevel = DebugLevel
	case "trace":
		logLevel = TraceLevel
	}

	return parseLogrusLevel(logLevel)
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
