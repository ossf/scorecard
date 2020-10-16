package checks

import (
	"fmt"
	"strings"

	"github.com/dlorenc/scorecard/checker"
)

//go:generate ../gen_github.sh

var ossFuzzRepos map[string]struct{}

func init() {
	ossFuzzRepos = map[string]struct{}{}
	for _, r := range strings.Split(fuzzRepos, "\n") {
		if r == "" {
			continue
		}
		r = strings.TrimSuffix(r, ".git")
		ossFuzzRepos[r] = struct{}{}
	}

	registerCheck("Fuzzing", Fuzzing)
}

func Fuzzing(c checker.Checker) checker.CheckResult {
	url := fmt.Sprintf("github.com/%s/%s", c.Owner, c.Repo)
	if _, ok := ossFuzzRepos[url]; ok {
		return checker.CheckResult{
			Pass:       true,
			Confidence: 10,
		}
	}
	return checker.CheckResult{
		Pass:       false,
		Confidence: 3,
	}
}
