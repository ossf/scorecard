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

package releasesHaveVerifiedSignatures

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "No packages to verify",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "Package with verified GPG signature",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:test-artifact",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:         checker.SignatureTypeGPG,
									IsVerified:   true,
									ArtifactURL:  "https://repo1.maven.org/maven2/org/example/test-artifact/1.0.0/test-artifact-1.0.0.jar",
									SignatureURL: "https://repo1.maven.org/maven2/org/example/test-artifact/1.0.0/test-artifact-1.0.0.jar.asc",
									KeyID:        "ABCD1234",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "Package with verified Sigstore signature",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "pypi",
							Name:    "example-package",
							Version: "2.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeSigstore,
									IsVerified:  true,
									ArtifactURL: "https://files.pythonhosted.org/packages/example-package-2.0.0.whl",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "Package with failed signature verification",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:bad-artifact",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:         checker.SignatureTypeGPG,
									IsVerified:   false,
									ArtifactURL:  "https://repo1.maven.org/maven2/org/example/bad-artifact/1.0.0/bad-artifact-1.0.0.jar",
									SignatureURL: "https://repo1.maven.org/maven2/org/example/bad-artifact/1.0.0/bad-artifact-1.0.0.jar.asc",
									ErrorMsg:     "signature verification failed: key not found",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "Multiple packages with mixed verification results",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:good-artifact",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  true,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/good-artifact/1.0.0/good-artifact-1.0.0.jar",
									KeyID:       "ABCD1234",
								},
							},
						},
						{
							System:  "pypi",
							Name:    "bad-package",
							Version: "2.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeSigstore,
									IsVerified:  false,
									ArtifactURL: "https://files.pythonhosted.org/packages/bad-package-2.0.0.whl",
									ErrorMsg:    "bundle validation failed",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeFalse,
			},
		},
		{
			name: "Package with multiple verified signatures",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:multi-sig",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  true,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/multi-sig/1.0.0/multi-sig-1.0.0.jar",
									KeyID:       "KEY1",
								},
								{
									Type:        checker.SignatureTypeSigstore,
									IsVerified:  true,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/multi-sig/1.0.0/multi-sig-1.0.0.jar",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "Package with multiple artifacts and mixed results",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:mixed",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  true,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/mixed/1.0.0/mixed-1.0.0.jar",
									KeyID:       "VERIFIED_KEY",
								},
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  false,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/mixed/1.0.0/mixed-1.0.0-sources.jar",
									ErrorMsg:    "key expired",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeFalse,
			},
		},
		{
			name: "Package with no signatures",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:     "maven",
							Name:       "org.example:unsigned",
							Version:    "1.0.0",
							Signatures: []checker.PackageSignature{},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeNotApplicable,
			},
		},
		{
			name: "Multiple packages, some without signatures",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:signed",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  true,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/signed/1.0.0/signed-1.0.0.jar",
								},
							},
						},
						{
							System:     "maven",
							Name:       "org.example:unsigned",
							Version:    "2.0.0",
							Signatures: []checker.PackageSignature{},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "PyPI package with PEP 740 attestation",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "pypi",
							Name:    "sigstore-python",
							Version: "3.5.1",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeSigstore,
									IsVerified:  true,
									ArtifactURL: "https://files.pythonhosted.org/packages/sigstore_python-3.5.1-py3-none-any.whl",
								},
								{
									Type:        checker.SignatureTypeSigstore,
									IsVerified:  true,
									ArtifactURL: "https://files.pythonhosted.org/packages/sigstore-python-3.5.1.tar.gz",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
			},
		},
		{
			name: "All packages have failed verification",
			raw: &checker.RawResults{
				SignedReleasesResults: checker.SignedReleasesData{
					Packages: []checker.ProjectPackage{
						{
							System:  "maven",
							Name:    "org.example:bad1",
							Version: "1.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  false,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/bad1/1.0.0/bad1-1.0.0.jar",
									ErrorMsg:    "key not trusted",
								},
							},
						},
						{
							System:  "maven",
							Name:    "org.example:bad2",
							Version: "2.0.0",
							Signatures: []checker.PackageSignature{
								{
									Type:        checker.SignatureTypeGPG,
									IsVerified:  false,
									ArtifactURL: "https://repo1.maven.org/maven2/org/example/bad2/2.0.0/bad2-2.0.0.jar",
									ErrorMsg:    "signature mismatch",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
				finding.OutcomeFalse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("expected %d findings, got %d findings", len(tt.outcomes), len(findings))
			}
			for i := range findings {
				if i >= len(tt.outcomes) {
					break
				}
				if diff := cmp.Diff(tt.outcomes[i], findings[i].Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestRun_NilRawResults(t *testing.T) {
	t.Parallel()

	_, _, err := Run(nil)
	if err == nil {
		t.Error("expected error for nil raw results")
	}
}

func TestRun_FindingValues(t *testing.T) {
	t.Parallel()

	raw := &checker.RawResults{
		SignedReleasesResults: checker.SignedReleasesData{
			Packages: []checker.ProjectPackage{
				{
					System:  "maven",
					Name:    "org.example:test",
					Version: "1.0.0",
					Signatures: []checker.PackageSignature{
						{
							Type:        checker.SignatureTypeGPG,
							IsVerified:  true,
							ArtifactURL: "https://repo1.maven.org/maven2/org/example/test/1.0.0/test-1.0.0.jar",
							KeyID:       "ABCD1234",
						},
					},
				},
			},
		},
	}

	findings, probeName, err := Run(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if probeName != Probe {
		t.Errorf("expected probe name %s, got %s", Probe, probeName)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]

	// Check that all expected values are present
	expectedValues := map[string]string{
		"packageSystem":  "maven",
		"packageName":    "org.example:test",
		"packageVersion": "1.0.0",
		"signatureType":  "gpg",
		"artifactURL":    "https://repo1.maven.org/maven2/org/example/test/1.0.0/test-1.0.0.jar",
		"keyID":          "ABCD1234",
	}

	for key, expectedValue := range expectedValues {
		actualValue, ok := f.Values[key]
		if !ok {
			t.Errorf("expected key %s not found in values", key)
		}
		if actualValue != expectedValue {
			t.Errorf("expected %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestRun_FailedVerificationFindingValues(t *testing.T) {
	t.Parallel()

	raw := &checker.RawResults{
		SignedReleasesResults: checker.SignedReleasesData{
			Packages: []checker.ProjectPackage{
				{
					System:  "pypi",
					Name:    "failed-package",
					Version: "2.0.0",
					Signatures: []checker.PackageSignature{
						{
							Type:        checker.SignatureTypeSigstore,
							IsVerified:  false,
							ArtifactURL: "https://files.pythonhosted.org/packages/failed-package-2.0.0.whl",
							ErrorMsg:    "bundle validation failed: invalid certificate",
						},
					},
				},
			},
		},
	}

	findings, _, err := Run(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]

	if f.Outcome != finding.OutcomeFalse {
		t.Errorf("expected outcome False, got %s", f.Outcome)
	}

	// Check that error message is included
	expectedValues := map[string]string{
		"packageSystem":  "pypi",
		"packageName":    "failed-package",
		"packageVersion": "2.0.0",
		"signatureType":  "sigstore",
		"errorMsg":       "bundle validation failed: invalid certificate",
	}

	for key, expectedValue := range expectedValues {
		actualValue, ok := f.Values[key]
		if !ok {
			t.Errorf("expected key %s not found in values", key)
		}
		if actualValue != expectedValue {
			t.Errorf("expected %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}
