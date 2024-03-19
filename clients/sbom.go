// Copyright 2024 OpenSSF Scorecard Authors
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

package clients

import (
	"regexp"
	"time"
)

// Sbom represents a customized struct for sbom used by clients.
type Sbom struct {
	ID            string
	Name          string
	Origin        string
	URL           string
	Created       time.Time
	Path          string
	Tool          string
	Schema        string
	SchemaVersion string
}

var ReSbomFile = regexp.MustCompile(
	`(?i).+\.(cdx.json|cdx.xml|spdx|spdx.json|spdx.xml|spdx.y[a?]ml|spdx.rdf|spdx.rdf.xm)`,
)

type SecurityInsightsSchema struct {
	Schema      string   `yaml:"$schema"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Required    []string `yaml:"required"`
	Properties  struct {
		Header struct {
			Properties struct {
				ProjectURL struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"project-url"`
				Changelog struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"changelog"`
				License struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"license"`
				ExpirationDate struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
				} `yaml:"expiration-date"`
				LastUpdated struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
				} `yaml:"last-updated"`
				LastReviewed struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
				} `yaml:"last-reviewed"`
				CommitHash struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"commit-hash"`
				ProjectRelease struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"project-release"`
				SchemaVersion struct {
					ID          string   `yaml:"$id"`
					Description string   `yaml:"description"`
					Type        string   `yaml:"type"`
					Enum        []string `yaml:"enum"`
				} `yaml:"schema-version"`
			} `yaml:"properties"`
			ID                   string   `yaml:"$id"`
			Type                 string   `yaml:"type"`
			Required             []string `yaml:"required"`
			AdditionalProperties bool     `yaml:"additionalProperties"`
		} `yaml:"header"`
		SecurityTesting struct {
			ID    string `yaml:"$id"`
			Type  string `yaml:"type"`
			Items struct {
				ID    string `yaml:"$id"`
				AnyOf []struct {
					ID         string   `yaml:"$id"`
					Type       string   `yaml:"type"`
					Required   []string `yaml:"required"`
					Properties struct {
						ToolURL struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Format      string `yaml:"format"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"tool-url"`
						Comment struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"comment"`
						ToolName struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"tool-name"`
						ToolVersion struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"tool-version"`
						ToolType struct {
							ID          string   `yaml:"$id"`
							Description string   `yaml:"description"`
							Type        string   `yaml:"type"`
							Enum        []string `yaml:"enum"`
						} `yaml:"tool-type"`
						Integration struct {
							Properties struct {
								AdHoc struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
								} `yaml:"ad-hoc"`
								BeforeRelease struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
								} `yaml:"before-release"`
								Ci struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
								} `yaml:"ci"`
							} `yaml:"properties"`
							ID                   string   `yaml:"$id"`
							Type                 string   `yaml:"type"`
							Required             []string `yaml:"required"`
							AdditionalProperties bool     `yaml:"additionalProperties"`
						} `yaml:"integration"`
						ToolRulesets struct {
							ID    string `yaml:"$id"`
							Type  string `yaml:"type"`
							Items struct {
								ID    string `yaml:"$id"`
								AnyOf []struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
								} `yaml:"anyOf"`
							} `yaml:"items"`
							AdditionalItems bool `yaml:"additionalItems"`
							UniqueItems     bool `yaml:"uniqueItems"`
						} `yaml:"tool-rulesets"`
					} `yaml:"properties"`
					AdditionalProperties bool `yaml:"additionalProperties"`
				} `yaml:"anyOf"`
			} `yaml:"items"`
			AdditionalItems bool `yaml:"additionalItems"`
		} `yaml:"security-testing"`
		Documentation struct {
			ID    string `yaml:"$id"`
			Type  string `yaml:"type"`
			Items struct {
				ID    string `yaml:"$id"`
				AnyOf []struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"anyOf"`
			} `yaml:"items"`
			AdditionalItems bool `yaml:"additionalItems"`
			UniqueItems     bool `yaml:"uniqueItems"`
		} `yaml:"documentation"`
		DistributionPoints struct {
			ID    string `yaml:"$id"`
			Type  string `yaml:"type"`
			Items struct {
				ID    string `yaml:"$id"`
				AnyOf []struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"anyOf"`
			} `yaml:"items"`
			AdditionalItems bool `yaml:"additionalItems"`
			UniqueItems     bool `yaml:"uniqueItems"`
		} `yaml:"distribution-points"`
		SecurityAssessments struct {
			ID    string `yaml:"$id"`
			Type  string `yaml:"type"`
			Items struct {
				ID    string `yaml:"$id"`
				AnyOf []struct {
					Properties struct {
						AuditorName struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"auditor-name"`
						AuditorReport struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Format      string `yaml:"format"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"auditor-report"`
						AuditorURL struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Format      string `yaml:"format"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"auditor-url"`
						ReportYear struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"report-year"`
						Comment struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"comment"`
					} `yaml:"properties"`
					ID                   string   `yaml:"$id"`
					Type                 string   `yaml:"type"`
					Required             []string `yaml:"required"`
					AdditionalProperties bool     `yaml:"additionalProperties"`
				} `yaml:"anyOf"`
			} `yaml:"items"`
			AdditionalItems bool `yaml:"additionalItems"`
			UniqueItems     bool `yaml:"uniqueItems"`
		} `yaml:"security-assessments"`
		SecurityContacts struct {
			ID    string `yaml:"$id"`
			Type  string `yaml:"type"`
			Items struct {
				ID    string `yaml:"$id"`
				AnyOf []struct {
					Properties struct {
						Value struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"value"`
						Primary struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"primary"`
						Type struct {
							ID          string   `yaml:"$id"`
							Description string   `yaml:"description"`
							Type        string   `yaml:"type"`
							Enum        []string `yaml:"enum"`
						} `yaml:"type"`
					} `yaml:"properties"`
					ID                   string   `yaml:"$id"`
					Type                 string   `yaml:"type"`
					Required             []string `yaml:"required"`
					AdditionalProperties bool     `yaml:"additionalProperties"`
				} `yaml:"anyOf"`
			} `yaml:"items"`
			AdditionalItems bool `yaml:"additionalItems"`
			UniqueItems     bool `yaml:"uniqueItems"`
		} `yaml:"security-contacts"`
		VulnerabilityReporting struct {
			ID   string `yaml:"$id"`
			Type string `yaml:"type"`
			Then struct {
				Required []string `yaml:"required"`
			} `yaml:"then"`
			Required   []string `yaml:"required"`
			Properties struct {
				BugBountyURL struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"bug-bounty-url"`
				SecurityPolicy struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"security-policy"`
				EmailContact struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
				} `yaml:"email-contact"`
				PgpKey struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"pgp-key"`
				Comment struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"comment"`
				AcceptsVulnerabilityReports struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"accepts-vulnerability-reports"`
				BugBountyAvailable struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"bug-bounty-available"`
				InScope struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Items       struct {
						ID    string `yaml:"$id"`
						AnyOf []struct {
							ID   string   `yaml:"$id"`
							Type string   `yaml:"type"`
							Enum []string `yaml:"enum"`
						} `yaml:"anyOf"`
					} `yaml:"items"`
					AdditionalItems bool `yaml:"additionalItems"`
					UniqueItems     bool `yaml:"uniqueItems"`
				} `yaml:"in-scope"`
				OutScope struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Items       struct {
						ID    string `yaml:"$id"`
						AnyOf []struct {
							ID   string   `yaml:"$id"`
							Type string   `yaml:"type"`
							Enum []string `yaml:"enum"`
						} `yaml:"anyOf"`
					} `yaml:"items"`
					AdditionalItems bool `yaml:"additionalItems"`
					UniqueItems     bool `yaml:"uniqueItems"`
				} `yaml:"out-scope"`
			} `yaml:"properties"`
			AdditionalProperties bool `yaml:"additionalProperties"`
			If                   struct {
				Properties struct {
					AcceptsVulnerabilityReports struct {
						Const bool `yaml:"const"`
					} `yaml:"accepts-vulnerability-reports"`
				} `yaml:"properties"`
			} `yaml:"if"`
		} `yaml:"vulnerability-reporting"`
		Dependencies struct {
			ID         string `yaml:"$id"`
			Type       string `yaml:"type"`
			Properties struct {
				DependenciesLifecycle struct {
					Properties struct {
						PolicyURL struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Format      string `yaml:"format"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"policy-url"`
						Comment struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"comment"`
					} `yaml:"properties"`
					ID                   string `yaml:"$id"`
					Type                 string `yaml:"type"`
					AdditionalProperties bool   `yaml:"additionalProperties"`
				} `yaml:"dependencies-lifecycle"`
				EnvDependenciesPolicy struct {
					Properties struct {
						PolicyURL struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Format      string `yaml:"format"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"policy-url"`
						Comment struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"comment"`
					} `yaml:"properties"`
					ID                   string `yaml:"$id"`
					Type                 string `yaml:"type"`
					AdditionalProperties bool   `yaml:"additionalProperties"`
				} `yaml:"env-dependencies-policy"`
				ThirdPartyPackages struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"third-party-packages"`
				Sbom struct {
					ID    string `yaml:"$id"`
					Type  string `yaml:"type"`
					Items struct {
						AnyOf []struct {
							Properties struct {
								SbomFile struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Format      string `yaml:"format"`
									Pattern     string `yaml:"pattern"`
								} `yaml:"sbom-file"`
								SbomFormat struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
								} `yaml:"sbom-format"`
								SbomURL struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Format      string `yaml:"format"`
									Pattern     string `yaml:"pattern"`
								} `yaml:"sbom-url"`
								SbomCreation struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Pattern     string `yaml:"pattern"`
								} `yaml:"sbom-creation"`
							} `yaml:"properties"`
							ID                   string `yaml:"$id"`
							AdditionalProperties bool   `yaml:"additionalProperties"`
						} `yaml:"anyOf"`
					} `yaml:"items"`
					AdditionalItems bool `yaml:"additionalItems"`
					UniqueItems     bool `yaml:"uniqueItems"`
				} `yaml:"sbom"`
				DependenciesLists struct {
					ID    string `yaml:"$id"`
					Items struct {
						ID    string `yaml:"$id"`
						AnyOf []struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Format      string `yaml:"format"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"anyOf"`
					} `yaml:"items"`
					AdditionalItems bool `yaml:"additionalItems"`
				} `yaml:"dependencies-lists"`
			} `yaml:"properties"`
			AdditionalProperties bool `yaml:"additionalProperties"`
		} `yaml:"dependencies"`
		ProjectLifecycle struct {
			ID string `yaml:"$id"`
			If struct {
				Properties struct {
					Status struct {
						Const string `yaml:"const"`
					} `yaml:"status"`
				} `yaml:"properties"`
			} `yaml:"if"`
			Type       string `yaml:"type"`
			Properties struct {
				Roadmap struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"roadmap"`
				ReleaseCycle struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"release-cycle"`
				ReleaseProcess struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"release-process"`
				BugFixesOnly struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"bug-fixes-only"`
				Status struct {
					ID          string   `yaml:"$id"`
					Description string   `yaml:"description"`
					Type        string   `yaml:"type"`
					Enum        []string `yaml:"enum"`
				} `yaml:"status"`
				CoreMaintainers struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Items       struct {
						ID    string `yaml:"$id"`
						AnyOf []struct {
							ID   string `yaml:"$id"`
							Type string `yaml:"type"`
						} `yaml:"anyOf"`
					} `yaml:"items"`
					AdditionalItems bool `yaml:"additionalItems"`
					UniqueItems     bool `yaml:"uniqueItems"`
				} `yaml:"core-maintainers"`
			} `yaml:"properties"`
			Then struct {
				Required []string `yaml:"required"`
			} `yaml:"then"`
			Required []string `yaml:"required"`
			Else     struct {
				Then struct {
					Required []string `yaml:"required"`
				} `yaml:"then"`
				If struct {
					Properties struct {
						BugFixesOnly struct {
							Const bool `yaml:"const"`
						} `yaml:"bug-fixes-only"`
					} `yaml:"properties"`
				} `yaml:"if"`
			} `yaml:"else"`
			AdditionalProperties bool `yaml:"additionalProperties"`
		} `yaml:"project-lifecycle"`
		ContributionPolicy struct {
			ID         string   `yaml:"$id"`
			Type       string   `yaml:"type"`
			Required   []string `yaml:"required"`
			Properties struct {
				ContributingPolicy struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"contributing-policy"`
				CodeOfConduct struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Format      string `yaml:"format"`
					Pattern     string `yaml:"pattern"`
				} `yaml:"code-of-conduct"`
				AcceptsPullRequests struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"accepts-pull-requests"`
				AcceptsAutomatedPullRequests struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
				} `yaml:"accepts-automated-pull-requests"`
				AutomatedToolsList struct {
					ID          string `yaml:"$id"`
					Description string `yaml:"description"`
					Type        string `yaml:"type"`
					Items       struct {
						AnyOf []struct {
							ID         string   `yaml:"$id"`
							Type       string   `yaml:"type"`
							Required   []string `yaml:"required"`
							Properties struct {
								Comment struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Pattern     string `yaml:"pattern"`
								} `yaml:"comment"`
								AutomatedTool struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
								} `yaml:"automated-tool"`
								Action struct {
									ID          string   `yaml:"$id"`
									Description string   `yaml:"description"`
									Type        string   `yaml:"type"`
									Enum        []string `yaml:"enum"`
								} `yaml:"action"`
								Path struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Items       struct {
										ID    string `yaml:"$id"`
										AnyOf []struct {
											ID          string `yaml:"$id"`
											Description string `yaml:"description"`
											Type        string `yaml:"type"`
										} `yaml:"anyOf"`
									} `yaml:"items"`
									AdditionalItems bool `yaml:"additionalItems"`
									UniqueItems     bool `yaml:"uniqueItems"`
								} `yaml:"path"`
							} `yaml:"properties"`
							AdditionalProperties bool `yaml:"additionalProperties"`
						} `yaml:"anyOf"`
					} `yaml:"items"`
					AdditionalItems bool `yaml:"additionalItems"`
					UniqueItems     bool `yaml:"uniqueItems"`
				} `yaml:"automated-tools-list"`
			} `yaml:"properties"`
			AdditionalProperties bool `yaml:"additionalProperties"`
		} `yaml:"contribution-policy"`
		SecurityArtifacts struct {
			ID         string `yaml:"$id"`
			Type       string `yaml:"type"`
			Properties struct {
				OtherArtifacts struct {
					ID    string `yaml:"$id"`
					AnyOf []struct {
						ID   string `yaml:"$id"`
						Type string `yaml:"type"`
						Then struct {
							Required []string `yaml:"required"`
						} `yaml:"then"`
						Properties struct {
							Comment struct {
								ID          string `yaml:"$id"`
								Description string `yaml:"description"`
								Type        string `yaml:"type"`
								Pattern     string `yaml:"pattern"`
							} `yaml:"comment"`
							ArtifactName struct {
								ID          string `yaml:"$id"`
								Description string `yaml:"description"`
								Type        string `yaml:"type"`
							} `yaml:"artifact-name"`
							ArtifactCreated struct {
								ID          string `yaml:"$id"`
								Description string `yaml:"description"`
								Type        string `yaml:"type"`
							} `yaml:"artifact-created"`
							EvidenceURL struct {
								ID    string `yaml:"$id"`
								Type  string `yaml:"type"`
								Items struct {
									ID    string `yaml:"$id"`
									AnyOf []struct {
										ID          string `yaml:"$id"`
										Description string `yaml:"description"`
										Type        string `yaml:"type"`
										Format      string `yaml:"format"`
										Pattern     string `yaml:"pattern"`
									} `yaml:"anyOf"`
								} `yaml:"items"`
								AdditionalItems bool `yaml:"additionalItems"`
								UniqueItems     bool `yaml:"uniqueItems"`
							} `yaml:"evidence-url"`
						} `yaml:"properties"`
						AdditionalProperties bool `yaml:"additionalProperties"`
						If                   struct {
							Properties struct {
								ArtifactCreated struct {
									Const bool `yaml:"const"`
								} `yaml:"artifact-created"`
							} `yaml:"properties"`
						} `yaml:"if"`
					} `yaml:"anyOf"`
				} `yaml:"other-artifacts"`
				ThreatModel struct {
					ID   string `yaml:"$id"`
					Type string `yaml:"type"`
					Then struct {
						Required []string `yaml:"required"`
					} `yaml:"then"`
					Required   []string `yaml:"required"`
					Properties struct {
						Comment struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"comment"`
						ThreatModelCreated struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"threat-model-created"`
						EvidenceURL struct {
							ID    string `yaml:"$id"`
							Type  string `yaml:"type"`
							Items struct {
								ID    string `yaml:"$id"`
								AnyOf []struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Format      string `yaml:"format"`
									Pattern     string `yaml:"pattern"`
								} `yaml:"anyOf"`
							} `yaml:"items"`
							AdditionalItems bool `yaml:"additionalItems"`
							UniqueItems     bool `yaml:"uniqueItems"`
						} `yaml:"evidence-url"`
					} `yaml:"properties"`
					AdditionalProperties bool `yaml:"additionalProperties"`
					If                   struct {
						Properties struct {
							ThreatModelCreated struct {
								Const bool `yaml:"const"`
							} `yaml:"threat-model-created"`
						} `yaml:"properties"`
					} `yaml:"if"`
				} `yaml:"threat-model"`
				SelfAssessment struct {
					ID   string `yaml:"$id"`
					Type string `yaml:"type"`
					Then struct {
						Required []string `yaml:"required"`
					} `yaml:"then"`
					Required   []string `yaml:"required"`
					Properties struct {
						Comment struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
							Pattern     string `yaml:"pattern"`
						} `yaml:"comment"`
						SelfAssessmentCreated struct {
							ID          string `yaml:"$id"`
							Description string `yaml:"description"`
							Type        string `yaml:"type"`
						} `yaml:"self-assessment-created"`
						EvidenceURL struct {
							ID    string `yaml:"$id"`
							Type  string `yaml:"type"`
							Items struct {
								ID    string `yaml:"$id"`
								AnyOf []struct {
									ID          string `yaml:"$id"`
									Description string `yaml:"description"`
									Type        string `yaml:"type"`
									Format      string `yaml:"format"`
									Pattern     string `yaml:"pattern"`
								} `yaml:"anyOf"`
							} `yaml:"items"`
							AdditionalItems bool `yaml:"additionalItems"`
							UniqueItems     bool `yaml:"uniqueItems"`
						} `yaml:"evidence-url"`
					} `yaml:"properties"`
					AdditionalProperties bool `yaml:"additionalProperties"`
					If                   struct {
						Properties struct {
							SelfAssessmentCreated struct {
								Const bool `yaml:"const"`
							} `yaml:"self-assessment-created"`
						} `yaml:"properties"`
					} `yaml:"if"`
				} `yaml:"self-assessment"`
			} `yaml:"properties"`
			AdditionalProperties bool `yaml:"additionalProperties"`
		} `yaml:"security-artifacts"`
	} `yaml:"properties"`
	AdditionalProperties bool `yaml:"additionalProperties"`
}
