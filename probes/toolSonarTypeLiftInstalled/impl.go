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

package toolSonarTypeLiftInstalled

import (
	"embed"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/utils"
)

//go:embed *.yml
var fs embed.FS

var probe = "toolSonarTypeLiftInstalled"

type sonartypeLyft struct {}

func (t sonartypeLyft) Name() string{
	return  "Sonatype Lift"
}
func (t sonartypeLyft) Matches(tool checker.Tool) bool{
	return t.Name() == tool.Name
}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	tools := raw.DependencyUpdateToolResults.Tools
	var matcher sonartypeLyft
	return utils.ToolsRun(tools, fs, probe,
		finding.OutcomePositive, finding.OutcomeNegative, matcher)
}
