// Copyright 2026 OpenSSF Scorecard Authors
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

package fileparser

import (
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/internal/ghworkflow"
)

const (
	ecosystemPyPI     = "pypi"
	ecosystemNPM      = "npm"
	ecosystemRubyGems = "rubygems"
	ecosystemCrates   = "crates"
	ecosystemPub      = "pub"
)

var (
	reNPMPublish   = regexp.MustCompile(`npm.*publish`)
	reGemPush      = regexp.MustCompile(`gem.*push`)
	reCargoPublish = regexp.MustCompile(`cargo.*publish`)
	rePubPublish   = regexp.MustCompile(`(?:dart|flutter).*pub.*publish`)
)

// ClassifyPublishAuth reports whether a packaging workflow uses trusted
// publishing (GitHub Actions OIDC) or a long-lived credential.
//
// The empty auth means the workflow does not publish to an ecosystem whose
// trusted-publishing signal Scorecard understands. Callers should not treat
// that as a credential finding.
//
// A job-level permissions block replaces the workflow permissions, matching
// GitHub Actions. id-token is not granted unless the effective block sets it
// to write, or the block is write-all.
func ClassifyPublishAuth(workflow *ghworkflow.Workflow) (checker.PublishAuth, string) {
	if workflow == nil {
		return "", ""
	}

	var credentialEcosystem, oidcEcosystem string
	for _, job := range workflow.Jobs {
		if job == nil {
			continue
		}
		for _, ecosystem := range ecosystemsForJob(job) {
			if jobUsesTrustedPublishing(workflow, job, ecosystem) {
				oidcEcosystem = ecosystem
				continue
			}
			credentialEcosystem = ecosystem
		}
	}
	if credentialEcosystem != "" {
		return checker.PublishAuthCredential, credentialEcosystem
	}
	if oidcEcosystem != "" {
		return checker.PublishAuthOIDC, oidcEcosystem
	}
	return "", ""
}

func ecosystemsForJob(job *ghworkflow.Job) []string {
	var ecosystems []string
	seen := map[string]bool{}
	add := func(ecosystem string) {
		if ecosystem == "" || seen[ecosystem] {
			return
		}
		seen[ecosystem] = true
		ecosystems = append(ecosystems, ecosystem)
	}
	for _, step := range job.Steps {
		add(ecosystemForStep(job, step))
	}
	return ecosystems
}

func ecosystemForStep(job *ghworkflow.Job, step *ghworkflow.Step) string {
	switch {
	case usesAction(step, "pypa/gh-action-pypi-publish"):
		return ecosystemPyPI
	case usesAction(step, "rubygems/release-gem") || runMatches(step, reGemPush):
		return ecosystemRubyGems
	case runMatches(step, reCargoPublish):
		return ecosystemCrates
	case runMatches(step, rePubPublish):
		return ecosystemPub
	case runMatches(step, reNPMPublish) && jobPublishesToNPM(job):
		return ecosystemNPM
	default:
		return ""
	}
}

func jobPublishesToNPM(job *ghworkflow.Job) bool {
	for _, step := range job.Steps {
		if !usesAction(step, "actions/setup-node") {
			continue
		}
		with := getWith(step)
		input, ok := with["registry-url"]
		if ok && input != nil && input.Value != nil && input.Value.Value == "https://registry.npmjs.org" {
			return true
		}
	}
	return false
}

func jobUsesTrustedPublishing(workflow *ghworkflow.Workflow, job *ghworkflow.Job, ecosystem string) bool {
	if !idTokenWrite(workflow, job) {
		return false
	}
	switch ecosystem {
	case ecosystemPyPI:
		return !pypiPasswordSet(job)
	case ecosystemNPM:
		return !envSet(workflow, job, "NODE_AUTH_TOKEN", "NPM_TOKEN")
	case ecosystemRubyGems:
		return !envSet(workflow, job, "GEM_HOST_API_KEY")
	case ecosystemCrates:
		return !envSet(workflow, job, "CARGO_REGISTRY_TOKEN")
	case ecosystemPub:
		return true
	default:
		return false
	}
}

func idTokenWrite(workflow *ghworkflow.Workflow, job *ghworkflow.Job) bool {
	perms := workflow.Permissions
	if job != nil && job.Permissions != nil {
		perms = job.Permissions
	}
	if perms == nil {
		return false
	}
	if perms.All != nil && strings.EqualFold(strings.TrimSpace(perms.All.Value), "write-all") {
		return true
	}
	for key, scope := range perms.Scopes {
		name := key
		if scope != nil && scope.Name != nil && scope.Name.Value != "" {
			name = scope.Name.Value
		}
		if !strings.EqualFold(name, "id-token") || scope == nil || scope.Value == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(scope.Value.Value), "write") {
			return true
		}
	}
	return false
}

func pypiPasswordSet(job *ghworkflow.Job) bool {
	for _, step := range job.Steps {
		if !usesAction(step, "pypa/gh-action-pypi-publish") {
			continue
		}
		with := getWith(step)
		input, ok := with["password"]
		if ok && input != nil && input.Value != nil && strings.TrimSpace(input.Value.Value) != "" {
			return true
		}
	}
	return false
}

func envSet(workflow *ghworkflow.Workflow, job *ghworkflow.Job, names ...string) bool {
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[strings.ToUpper(name)] = struct{}{}
	}
	if envHas(workflow.Env, want) || envHas(job.Env, want) {
		return true
	}
	for _, step := range job.Steps {
		if step != nil && envHas(step.Env, want) {
			return true
		}
	}
	return false
}

func envHas(env *ghworkflow.Env, want map[string]struct{}) bool {
	if env == nil {
		return false
	}
	for key, variable := range env.Vars {
		if _, ok := want[strings.ToUpper(key)]; !ok {
			continue
		}
		if variable != nil && variable.Value != nil && strings.TrimSpace(variable.Value.Value) != "" {
			return true
		}
	}
	return false
}

func usesAction(step *ghworkflow.Step, name string) bool {
	uses := GetUses(step)
	if uses == nil {
		return false
	}
	return strings.HasPrefix(uses.Value, name+"@") || uses.Value == name
}

func runMatches(step *ghworkflow.Step, pattern *regexp.Regexp) bool {
	run := getRun(step)
	if run == nil {
		return false
	}
	return pattern.MatchString(run.Value)
}
