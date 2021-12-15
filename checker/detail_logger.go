// Copyright 2020 Security Scorecard Authors
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

package checker

// DetailLogger logs a CheckDetail struct.
type DetailLogger interface {
	Info(desc string, args ...interface{})
	Warn(desc string, args ...interface{})
	Debug(desc string, args ...interface{})

	// Functions to use for moving to SARIF format.
	// UPGRADEv3: to rename.
	Info3(msg *LogMessage)
	Warn3(msg *LogMessage)
	Debug3(msg *LogMessage)
}
