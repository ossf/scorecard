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

package azuredevopsrepo

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf16"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"

	"github.com/ossf/scorecard/v5/clients"
)

const (
	gitSecurityNamespaceID = "2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87"
	branchContribute       = 4
	branchForcePush        = 8
	branchBypassPush       = 128
)

func branchSecurityToken(repo *Repo, name string) string {
	var token strings.Builder
	fmt.Fprintf(&token, "repoV2/%s/%s/refs/heads/", repo.projectID, repo.id)
	// Security tokens ignore case, so encode case-sensitive Git refs as UTF-16LE.
	for _, unit := range utf16.Encode([]rune(name)) {
		if unit == '/' {
			token.WriteByte('/')
		} else {
			fmt.Fprintf(&token, "%02x%02x", unit&0xff, unit>>8)
		}
	}
	token.WriteByte('/')
	return token.String()
}

func (b *branchesHandler) applyBranchPermissions(branch *clients.BranchRef) error {
	namespace := uuid.MustParse(gitSecurityNamespaceID)
	token := branchSecurityToken(b.repourl, *branch.Name)
	var evaluations []security.PermissionEvaluation
	for _, permission := range []int{branchContribute, branchForcePush, branchBypassPush} {
		evaluations = append(evaluations, security.PermissionEvaluation{
			SecurityNamespaceId: &namespace, Token: &token, Permissions: &permission,
		})
	}
	alwaysAllowAdmins := false
	response, err := b.hasPermissionsBatch(b.ctx, security.HasPermissionsBatchArgs{
		EvalBatch: &security.PermissionEvaluationBatch{
			AlwaysAllowAdministrators: &alwaysAllowAdmins, Evaluations: &evaluations,
		},
	})
	if err != nil {
		if status, ok := azureDevOpsErrorStatusCode(err); ok &&
			(status == http.StatusUnauthorized || status == http.StatusForbidden) {
			log.Printf("Azure DevOps branch %q: permission evidence unavailable (HTTP %d); controls remain unknown",
				*branch.Name, status)
			return nil
		}
		return fmt.Errorf("evaluating permissions on branch %s: %w", *branch.Name, err)
	}
	if response == nil || response.Evaluations == nil || len(*response.Evaluations) != len(evaluations) {
		return fmt.Errorf("%w: missing permission evaluations for branch %s", errBranchProtectionData, *branch.Name)
	}
	values := make(map[int]bool)
	for _, evaluation := range *response.Evaluations {
		if evaluation.SecurityNamespaceId == nil || *evaluation.SecurityNamespaceId != namespace ||
			evaluation.Token == nil || *evaluation.Token != token ||
			evaluation.Permissions == nil || evaluation.Value == nil {
			return fmt.Errorf("%w: mismatched permission evaluation for branch %s", errBranchProtectionData, *branch.Name)
		}
		permission := *evaluation.Permissions
		if _, exists := values[permission]; exists ||
			(permission != branchContribute && permission != branchForcePush && permission != branchBypassPush) {
			return fmt.Errorf("%w: unexpected permission evaluation for branch %s", errBranchProtectionData, *branch.Name)
		}
		values[permission] = *evaluation.Value
	}
	if branch.Protected == nil {
		return nil
	}
	protected := *branch.Protected
	if values[branchContribute] && values[branchForcePush] && (!protected || values[branchBypassPush]) {
		allowed := true
		branch.BranchProtectionRule.AllowForcePushes = &allowed
		if !protected {
			deletions := true
			branch.BranchProtectionRule.AllowDeletions = &deletions
		}
		log.Printf("Azure DevOps branch %q: caller can force updates; other users' permissions remain unknown",
			*branch.Name)
	}
	if protected && values[branchContribute] && values[branchBypassPush] {
		log.Printf("Azure DevOps branch %q: caller can bypass policies when pushing; administrator coverage is unknown",
			*branch.Name)
	}
	return nil
}
