// Copyright 2022 Allstar Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package format

import (
	"fmt"
	"os"

	"github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
	spol "github.com/ossf/scorecard/v4/policy"
	"github.com/ossf/scorecard/v4/refactor/options"
)

func FormatResults(
	opts *options.Options,
	results pkg.ScorecardResult,
	docs checks.Doc,
	policy *spol.ScorecardPolicy,
) error {
	var err error

	switch opts.Format {
	case options.FormatDefault:
		err = results.AsString(opts.ShowDetails, sclog.Level(opts.LogLevel), docs, os.Stdout)
	case options.FormatSarif:
		// TODO: support config files and update checker.MaxResultScore.
		err = results.AsSARIF(opts.ShowDetails, sclog.Level(opts.LogLevel), os.Stdout, docs, policy)
	case options.FormatJSON:
		err = results.AsJSON2(opts.ShowDetails, sclog.Level(opts.LogLevel), docs, os.Stdout)
	case options.FormatRaw:
		err = results.AsRawJSON(os.Stdout)
	default:
		err = sce.WithMessage(
			sce.ErrScorecardInternal,
			fmt.Sprintf(
				"invalid format flag: %v. Expected [default, json]",
				opts.Format,
			),
		)
	}

	if err != nil {
		return fmt.Errorf("Failed to output results: %v", err)
	}

	return nil
}
