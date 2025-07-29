// Copyright 2025 OpenSSF Scorecard Authors
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

package minder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	minder "github.com/mindersec/minder/pkg/engine/v1/interfaces"
	"github.com/mindersec/minder/pkg/engine/v1/rtengine"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
)

type Evaluator = minder.Evaluator

// MakeEvaluator creates a new minder evaluator.
func MakeEvaluator(
	ctx context.Context,
	rule *minderv1.RuleType,
	req *checker.CheckRequest,
) (*rtengine.RuleTypeEngine, error) {
	provider := GitHubProvider{
		repo: req.RepoClient,
	}
	rtengine, err := rtengine.NewRuleTypeEngine(ctx, rule, &provider)
	if err != nil {
		err = fmt.Errorf("error creating Minder engine: %w", err)
	}
	return rtengine, err
}

// noopResultSink is a no-op implementation of the ResultSink interface.
type noopResultSink struct{}

// SetIngestResult implements interfaces.ResultSink.
func (e *noopResultSink) SetIngestResult(result *minder.Ingested) {
}

var _ minder.ResultSink = (*noopResultSink)(nil)

// ErrWithDetails is not currently exported by the interface, but contains the
// evaluation result details used by the Minder engine.
type ErrWithDetails interface {
	Error() string
	Details() string
}

func CheckRule(rule *minderv1.RuleType) checker.CheckFn {
	return func(req *checker.CheckRequest) checker.CheckResult {
		eval, err := MakeEvaluator(req.Ctx, rule, req)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return checker.CreateRuntimeErrorResult(rule.GetName(), e)
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
		dataIngest := &noopResultSink{}
		// Rule evaluation signals "denied" using an error interface, but may still
		// return evaluation results. Failure to evaluate a rule returns an error
		// without results.
		res, resErr := eval.Eval(req.Ctx, entity, ruleVals, ingestVals, dataIngest)

		if res == nil { // Failure to evaluate (e.g. parsing error, etc.)
			e := sce.WithMessage(sce.ErrScorecardInternal, resErr.Error())
			return checker.CreateRuntimeErrorResult(rule.GetName(), e)
		}

		reason := rule.GetDescription()
		if resErr != nil {
			reason = resErr.Error()
			var detailErr ErrWithDetails
			if errors.As(resErr, &detailErr) {
				reason = detailErr.Details()
			}
		}
		score, err := getScore(res.Output, resErr)
		if err != nil {
			return checker.CreateRuntimeErrorResult(rule.GetName(), err)
		}

		ret := checker.CreateResultWithScore(rule.GetName(), reason, score)
		// TODO: copy findings (JSON object) into ret.Findings [structured map]
		return ret
	}
}

func getScore(output any, err error) (int, error) {
	// extract an int number field from a JSON object, false if not found
	numberFromMap := func(data any, key string) (int, bool) {
		asMap, ok := data.(map[string]any)
		if !ok {
			return 0, false
		}
		num, ok := asMap[key].(json.Number)
		if !ok {
			return 0, false
		}
		score64, err := num.Int64()
		if err != nil {
			return 0, false
		}
		return int(score64), true
	}

	score, ok := numberFromMap(output, "score")
	if ok {
		if score < checker.MinResultScore || score > checker.MaxResultScore {
			return 0, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid score in output: %d", score))
		}
		return score, nil
	}
	if err == nil {
		return checker.MaxResultScore, nil
	}
	return checker.MinResultScore, nil
}
