package checks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/dlorenc/scorecard/checker"
)

func init() {
	registerCheck("CII-Best-Practices", CiiBestPractices)
}

type response struct {
	BadgeLevel string `json:"badge_level"`
}

func CiiBestPractices(c checker.Checker) checker.CheckResult {
	repoUrl := fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo)
	url := fmt.Sprintf("https://bestpractices.coreinfrastructure.org/projects.json?url=%s", repoUrl)
	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return checker.RetryResult(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return checker.RetryResult(err)
	}

	parsedResponse := []response{}
	if err := json.Unmarshal(b, &parsedResponse); err != nil {
		return checker.RetryResult(err)
	}

	if len(parsedResponse) < 1 {
		return checker.CheckResult{
			Pass:       false,
			Confidence: 10,
		}
	}

	result := parsedResponse[0]
	if result.BadgeLevel != "" {
		return checker.CheckResult{
			Pass:       true,
			Confidence: 10,
			Details:    result.BadgeLevel,
		}
	}

	return checker.CheckResult{
		Pass:       false,
		Confidence: 10,
	}
}
