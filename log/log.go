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
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger exposes logging capabilities.
// The struct name was chosen to closely mimic other logging facilities within
// to project to make them easier to search/replace.
// Initial implementation was designed to encapsulate calls to `zap`, but
// future iterations will seek to directly expose logging methods.
type Logger struct {
	Zap *zap.Logger
}

// NewLogger creates an instance of *zap.Logger.
// Copied from clients/githubrepo/client.go.
func NewLogger(logLevel Level) (*Logger, error) {
	zapCfg := zap.NewProductionConfig()
	zapLevel := parseLogLevelZap(string(logLevel))
	zapCfg.Level.SetLevel(zapLevel)

	zapLogger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("configuring zap logger: %w", err)
	}

	logger := &Logger{
		Zap: zapLogger,
	}

	return logger, nil
}

// Level is a string representation of log level, which can easily be passed as
// a parameter, in lieu of defined types in upstream logging packages.
type Level string

// Log levels
// TODO(log): Revisit if all levels are required. The current list mimics zap
//            log levels.
const (
	DefaultLevel       = InfoLevel
	DebugLevel   Level = "debug"
	InfoLevel    Level = "info"
	WarnLevel    Level = "warn"
	ErrorLevel   Level = "error"
	DPanicLevel  Level = "dpanic"
	PanicLevel   Level = "panic"
	FatalLevel   Level = "fatal"
)

func (l Level) String() string {
	return string(l)
}

// parseLogLevelZap parses a log level string and returning a zapcore.Level,
// which defaults to `zapcore.InfoLevel` when the provided string is not
// recognized.
// It is an inversion of go.uber.org/zap/zapcore.Level.String().
// TODO(log): Should we include a strict option here, which fails if the
//            provided log level is not recognized or is it fine to default to
//            InfoLevel?
func parseLogLevelZap(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
