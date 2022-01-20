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

type Logger struct {
	Zap *zap.Logger
}

// NewLogger creates an instance of *zap.Logger.
// Copied from clients/githubrepo/client.go
func NewLogger(logLevel zapcore.Level) (*Logger, error) {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level.SetLevel(logLevel)

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
