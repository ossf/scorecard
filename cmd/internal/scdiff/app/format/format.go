package format

import (
	"os"

	"github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

//nolint:wrapcheck
func JSON(r *pkg.ScorecardResult) error {
	const details = true
	docs, err := checks.Read()
	if err != nil {
		return err
	}
	// TODO standardize the input, and output it to a file
	return r.AsJSON2(details, log.DefaultLevel, docs, os.Stdout)
}
