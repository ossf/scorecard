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
