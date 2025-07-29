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
	"embed"
	"strings"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

var AllRules = []*minderv1.RuleType{}

//go:embed *.rego
var ruleFiles embed.FS

func init() {
	files, err := ruleFiles.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, ruleFile := range files {
		data, err := ruleFiles.ReadFile(ruleFile.Name())
		if err != nil {
			panic(err)
		}

		metadata := string(data)
		if !strings.HasPrefix(metadata, "# METADATA") {
			panic("Rule file " + ruleFile.Name() + " has no metadata")
		}

		// Extract all lines starting with # after # METADATA
		lines := strings.Split(metadata, "\n")
		var metadataLines []string
		foundMetadata := false

		for _, line := range lines {
			if line == "# METADATA" {
				foundMetadata = true
				continue
			}
			if foundMetadata && !strings.HasPrefix(line, "# ") {
				break
			}
			metadataLine := strings.TrimPrefix(line, "# ")
			if metadataLine != "" {
				metadataLines = append(metadataLines, metadataLine)
			}
		}

		metadataStruct := struct {
			Title       string `yaml:"title"`
			Ingest      string `yaml:"ingest"`
			Eval        string `yaml:"eval"`
			Description string `yaml:"description"`
		}{}
		yamlContent := strings.Join(metadataLines, "\n")
		if err := yaml.Unmarshal([]byte(yamlContent), &metadataStruct); err != nil {
			panic("Error parsing YAML metadata in " + ruleFile.Name() + ": " + err.Error())
		}

		projectName := "unused"
		rule := &minderv1.RuleType{
			Name: metadataStruct.Title,
			// Required by the engine, but not used in this context.
			Context: &minderv1.Context{
				Project: &projectName,
			},
			Description: metadataStruct.Description,
			Def: &minderv1.RuleType_Definition{
				InEntity: minderv1.Entity_ENTITY_REPOSITORIES.ToString(),
				// TODO: support REST ingest
				Ingest: &minderv1.RuleType_Definition_Ingest{
					Type: metadataStruct.Ingest,
					Git:  &minderv1.GitType{},
				},
				RuleSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{},
				},
				Eval: &minderv1.RuleType_Definition_Eval{
					Type: "rego",
					Rego: &minderv1.RuleType_Definition_Eval_Rego{
						Type: metadataStruct.Eval,
						Def:  string(data),
					},
				},
			},
		}

		AllRules = append(AllRules, rule)
	}
}
