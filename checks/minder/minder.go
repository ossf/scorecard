package minder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	minder "github.com/mindersec/minder/pkg/engine/v1/interfaces"
	"github.com/mindersec/minder/pkg/engine/v1/rtengine"
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

type Evaluator = minder.Evaluator

// MakeEvaluator creates a new minder evaluator
func MakeEvaluator(ctx context.Context, rule *minderv1.RuleType, req *checker.CheckRequest) (*rtengine.RuleTypeEngine, error) {
	provider := GitHubProvider{
		repo: req.RepoClient,
	}
	rtengine, err := rtengine.NewRuleTypeEngine(ctx, rule, &provider)
	return rtengine, err
}

type NoopResultSink struct {
}

// SetIngestResult implements interfaces.ResultSink.
func (e *NoopResultSink) SetIngestResult(result *minder.Ingested) {
}

var _ minder.ResultSink = (*NoopResultSink)(nil)

// This error type is not currently exported, but contains the evaluation
// result details used by the Minder engine.
type ErrWithDetails interface {
	Error() string
	Details() string
}

func CheckRule(rule *minderv1.RuleType) checker.CheckFn {
	return func(req *checker.CheckRequest) checker.CheckResult {

		eval, err := MakeEvaluator(req.Ctx, rule, req)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return checker.CreateRuntimeErrorResult(rule.Name, e)
		}
		// TODO: fill in more, adapt for non-GitHub providers
		parts := strings.SplitN(req.Repo.Path(), "/", 2)
		entity := &minderv1.Repository{
			Owner:    parts[0],
			Name:     parts[1],
			CloneUrl: req.Repo.URI(),
		}
		// TODO: fill in entity.Properties

		// As rule definitions are hard-coded, we don't need to set any parameters.
		// In Minder, parameters can be used to customize rules and ingestion from
		// a common library of ruletypes.
		ruleVals := map[string]any{}
		ingestVals := map[string]any{}
		dataIngest := &NoopResultSink{}
		// Rule evaluation signals "denied" using an error interface, but may still
		// return evaluation results. Failure to evaluate a rule returns an error
		// without results.
		res, resErr := eval.Eval(req.Ctx, entity, ruleVals, ingestVals, dataIngest)

		if res == nil { // Failure to evaluate (e.g. parsing error, etc.)
			e := sce.WithMessage(sce.ErrScorecardInternal, resErr.Error())
			return checker.CreateRuntimeErrorResult(rule.Name, e)
		}

		reason := rule.Description
		if resErr != nil {
			reason = resErr.Error()
			if detailErr, ok := resErr.(ErrWithDetails); ok {
				reason = detailErr.Details()
			}
		}
		score, err := getScore(res.Output, resErr)
		if err != nil {
			return checker.CreateRuntimeErrorResult(rule.Name, err)
		}

		ret := checker.CreateResultWithScore(rule.Name, reason, score)
		// TODO: copy findings (JSON object) into ret.Findings [structured map]
		return ret
	}
}

func getScore(output any, err error) (int, error) {
	outputDict, ok := output.(map[string]any)
	if ok {
		if scoreNum, ok := outputDict["score"].(json.Number); ok {
			score64, err := scoreNum.Int64()
			if err != nil {
				return 0, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid score in output: %v", err))
			}
			score := int(score64)
			if score < checker.MinResultScore || score > checker.MaxResultScore {
				return 0, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid score in output: %d", score))
			}
			return score, nil
		}
	}
	if err != nil {
		return checker.MinResultScore, nil
	}
	return checker.MaxResultScore, nil
}
