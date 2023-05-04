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

package toolRenovateInstalled

import (
	"embed"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/utils"
)

//go:embed *.yml
var fs embed.FS

const probe = "toolRenovateInstalled"

type renovate struct{}

func (t renovate) Name() string {
	return "RenovateBot"
}

func (t renovate) Matches(tool checker.Tool) bool {
	return t.Name() == tool.Name
}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	tools := raw.DependencyUpdateToolResults.Tools
	var matcher renovate
	// Check whether Renovate tool is installed on the repo,
	// and create the correponding findings.
	return utils.ToolsRun(tools, fs, probe,
		finding.OutcomePositive, finding.OutcomeNegative, matcher)
}
