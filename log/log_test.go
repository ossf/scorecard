// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package log

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logLevel Level
	}{
		{
			name:     "debug",
			logLevel: DebugLevel,
		},
		{
			name:     "info",
			logLevel: InfoLevel,
		},
		{
			name:     "warn",
			logLevel: WarnLevel,
		},
		{
			name:     "error",
			logLevel: ErrorLevel,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := NewLogger(tt.logLevel)
			if logger == nil {
				t.Errorf("NewLogger() returned nil")
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		levelStr      string
		expectedLevel Level
	}{
		{
			name:          "panic level",
			levelStr:      "panic",
			expectedLevel: PanicLevel,
		},
		{
			name:          "fatal level",
			levelStr:      "fatal",
			expectedLevel: FatalLevel,
		},
		{
			name:          "error level",
			levelStr:      "error",
			expectedLevel: ErrorLevel,
		},
		{
			name:          "warn level",
			levelStr:      "warn",
			expectedLevel: WarnLevel,
		},
		{
			name:          "info level",
			levelStr:      "info",
			expectedLevel: InfoLevel,
		},
		{
			name:          "debug level",
			levelStr:      "debug",
			expectedLevel: DebugLevel,
		},
		{
			name:          "trace level",
			levelStr:      "trace",
			expectedLevel: TraceLevel,
		},
		{
			name:          "default level",
			levelStr:      "invalid",
			expectedLevel: DefaultLevel,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			level := ParseLevel(tt.levelStr)
			if level != tt.expectedLevel {
				t.Errorf("ParseLevel(%s) = %v, expected %v", tt.levelStr, level, tt.expectedLevel)
			}
		})
	}
}
