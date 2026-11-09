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

// Package v20260828 is OSPS Baseline release v2026.08.28 and Scorecard's
// mappings for it.
//
// The requirement IDs and applicability come from the release tag
// (commit a26a7963f6fd098c1e82829322ec9e1196a4f294 of
// github.com/ossf/security-baseline). They must not change after this
// package ships. A fix to how a requirement is evaluated goes in package
// ospsbaseline and is identified by the Scorecard version.
package v20260828

import (
	"strings"

	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/internal/conformance/ospsbaseline"
)

const (
	// Version is the release version.
	Version = "v2026.08.28"

	maturity1 = "maturity-1"
	maturity2 = "maturity-2"
	maturity3 = "maturity-3"
)

var (
	m1   = []string{maturity1}
	m3   = []string{maturity3}
	m23  = []string{maturity2, maturity3}
	m123 = []string{maturity1, maturity2, maturity3}
)

// Release returns the release and its mappings.
func Release() *conformance.Release {
	return &conformance.Release{
		Catalog:      ospsbaseline.CatalogID,
		Title:        ospsbaseline.CatalogTitle,
		Version:      Version,
		URL:          "https://github.com/ossf/security-baseline/releases/tag/" + Version,
		Profiles:     []string{maturity1, maturity2, maturity3},
		Requirements: requirements(),
		Mappings: []conformance.Mapping{
			ospsbaseline.AC0301Mapping(),
			ospsbaseline.AC0302Mapping(),
		},
	}
}

// req builds a requirement. An OSPS requirement ID is its control's ID
// followed by "." and a number.
func req(id string, applicability []string) conformance.Requirement {
	control, _, _ := strings.Cut(id, ".")
	return conformance.Requirement{ID: id, Control: control, Applicability: applicability}
}

func requirements() []conformance.Requirement {
	return []conformance.Requirement{
		req("OSPS-AC-01.01", m123),
		req("OSPS-AC-02.01", m123),
		req("OSPS-AC-03.01", m123),
		req("OSPS-AC-03.02", m123),
		req("OSPS-AC-04.01", m23),
		req("OSPS-AC-04.02", m3),
		req("OSPS-BR-01.01", m123),
		req("OSPS-BR-01.02", m123),
		req("OSPS-BR-01.03", m123),
		req("OSPS-BR-01.04", m3),
		req("OSPS-BR-02.01", m23),
		req("OSPS-BR-02.02", m3),
		req("OSPS-BR-03.01", m123),
		req("OSPS-BR-03.02", m123),
		req("OSPS-BR-04.01", m23),
		req("OSPS-BR-05.01", m23),
		req("OSPS-BR-06.01", m23),
		req("OSPS-BR-07.01", m1),
		req("OSPS-BR-07.02", m3),
		req("OSPS-DO-01.01", m123),
		req("OSPS-DO-02.01", m123),
		req("OSPS-DO-03.01", m3),
		req("OSPS-DO-03.02", m3),
		req("OSPS-DO-04.01", m3),
		req("OSPS-DO-05.01", m3),
		req("OSPS-DO-06.01", m23),
		req("OSPS-DO-07.01", m23),
		req("OSPS-GV-01.01", m23),
		req("OSPS-GV-01.02", m23),
		req("OSPS-GV-02.01", m123),
		req("OSPS-GV-03.01", m123),
		req("OSPS-GV-03.02", m23),
		req("OSPS-GV-04.01", m3),
		req("OSPS-LE-01.01", m23),
		req("OSPS-LE-02.01", m123),
		req("OSPS-LE-02.02", m123),
		req("OSPS-LE-03.01", m123),
		req("OSPS-LE-03.02", m123),
		req("OSPS-QA-01.01", m123),
		req("OSPS-QA-01.02", m123),
		req("OSPS-QA-02.01", m123),
		req("OSPS-QA-02.02", m3),
		req("OSPS-QA-03.01", m23),
		req("OSPS-QA-04.01", m123),
		req("OSPS-QA-04.02", m3),
		req("OSPS-QA-05.01", m123),
		req("OSPS-QA-05.02", m123),
		req("OSPS-QA-06.01", m23),
		req("OSPS-QA-06.02", m3),
		req("OSPS-QA-06.03", m3),
		req("OSPS-QA-07.01", m3),
		req("OSPS-SA-01.01", m23),
		req("OSPS-SA-02.01", m23),
		req("OSPS-SA-03.01", m23),
		req("OSPS-SA-03.02", m3),
		req("OSPS-VM-01.01", m23),
		req("OSPS-VM-02.01", m1),
		req("OSPS-VM-03.01", m23),
		req("OSPS-VM-04.01", m23),
		req("OSPS-VM-04.02", m3),
		req("OSPS-VM-05.01", m3),
		req("OSPS-VM-05.02", m3),
		req("OSPS-VM-05.03", m3),
		req("OSPS-VM-06.01", m3),
		req("OSPS-VM-06.02", m3),
	}
}
