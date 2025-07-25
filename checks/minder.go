package checks

import (
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/minder"
)

func init() {
	IngestToType := map[string][]checker.RequestType{
		"rest": nil,
		"git":  {checker.CommitBased, checker.FileBased},
	}

	for _, rule := range minder.AllRules {
		rulefunc := minder.CheckRule(rule)
		registerCheck(rule.Name, rulefunc, IngestToType[rule.Def.Ingest.Type])
	}
}
