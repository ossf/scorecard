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

package checks

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

// ---- RoundTripper helper to mock deps.dev ----

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

//nolint:gocognit,paralleltest // test function with many test cases
func TestMTTUDependencies_TableDriven(t *testing.T) {
	type tc struct {
		files       map[string]string
		ddResponses map[string]string
		name        string
		wantScore   int
		wantErr     bool
	}

	// Helpers for relative timestamps
	tsDaysAgo := func(d int) string {
		return time.Now().UTC().AddDate(0, 0, -d).Format(time.RFC3339)
	}
	const (
		longAgo   = 365 // ≥180d → High
		midAgo    = 30  // [14d,180d) → Low
		recentAgo = 6   // <14d → VeryLow
		d60       = 60
		d90       = 90
		d300      = 300
		d455      = 455
		d445      = 445
		d435      = 435
		d500      = 500
		d200      = 200
	)

	// ---- Per-testcase deps.dev responses built with relative dates ----

	respNPMA := fmt.Sprintf(`{
  "packageKey": {"system":"npm","name":"a"},
  "purl":"pkg:npm/a",
  "versions":[
    {"versionKey":{"system":"npm","name":"a","version":"0.9.0"},"purl":"pkg:npm/a@0.9.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"a","version":"0.9.1"},"purl":"pkg:npm/a@0.9.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"a","version":"0.9.2"},"purl":"pkg:npm/a@0.9.2","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"a","version":"1.0.0"},"purl":"pkg:npm/a@1.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	respNPMB := fmt.Sprintf(`{
  "packageKey": {"system":"npm","name":"b"},
  "purl":"pkg:npm/b",
  "versions":[
    {"versionKey":{"system":"npm","name":"b","version":"1.8.0"},"purl":"pkg:npm/b@1.8.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"1.9.0"},"purl":"pkg:npm/b@1.9.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"1.9.1"},"purl":"pkg:npm/b@1.9.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"2.0.0"},"purl":"pkg:npm/b@2.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	// go.mod testcase constants: system "go", name "github.com/mod/a".
	// Project uses v1.2.3, but newer versions exist and the oldest newer (v1.2.4) is long ago,
	// so meantime ≥ 180 days → High probe true → score 0 (deterministic).
	respGoModA := fmt.Sprintf(`{
  "packageKey": {"system":"go","name":"github.com/mod/a"},
  "purl":"pkg:go/github.com/mod/a",
  "versions":[
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.3"},"purl":"pkg:go/github.com/mod/a@v1.2.3","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.4"},"purl":"pkg:go/github.com/mod/a@v1.2.4","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.3.0"},"purl":"pkg:go/github.com/mod/a@v1.3.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v2.0.0"},"purl":"pkg:go/github.com/mod/a@v2.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d500), tsDaysAgo(longAgo), tsDaysAgo(d300), tsDaysAgo(d200))

	// go.mod: LOW window (mean in [14d, 180d)).
	// Project uses v1.2.3. Oldest newer is v1.2.4 published ~30 days ago.
	respGoModALow := fmt.Sprintf(`{
  "packageKey": {"system":"go","name":"github.com/mod/a"},
  "purl":"pkg:go/github.com/mod/a",
  "versions":[
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.2"},"purl":"pkg:go/github.com/mod/a@v1.2.2","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.3"},"purl":"pkg:go/github.com/mod/a@v1.2.3","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.4"},"purl":"pkg:go/github.com/mod/a@v1.2.4","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.3.0"},"purl":"pkg:go/github.com/mod/a@v1.3.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d90), tsDaysAgo(d60), tsDaysAgo(midAgo), tsDaysAgo(recentAgo))

	// go.mod: VERY LOW window (mean < 14d).
	// Project uses v1.2.3 and that is the default/latest version → contributes 0 → score 10.
	respGoModAVeryLowLatest := fmt.Sprintf(`{
  "packageKey": {"system":"go","name":"github.com/mod/a"},
  "purl":"pkg:go/github.com/mod/a",
  "versions":[
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.0"},"purl":"pkg:go/github.com/mod/a@v1.2.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.1"},"purl":"pkg:go/github.com/mod/a@v1.2.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.2"},"purl":"pkg:go/github.com/mod/a@v1.2.2","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"go","name":"github.com/mod/a","version":"v1.2.3"},"purl":"pkg:go/github.com/mod/a@v1.2.3","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d90), tsDaysAgo(d60), tsDaysAgo(midAgo), tsDaysAgo(recentAgo))

	respNPMAPackageJSON := fmt.Sprintf(`{
  "packageKey": {"system":"npm","name":"a"},
  "purl":"pkg:npm/a",
  "versions":[
    {"versionKey":{"system":"npm","name":"a","version":"0.9.0"},"purl":"pkg:npm/a@0.9.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"a","version":"0.9.1"},"purl":"pkg:npm/a@0.9.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"a","version":"0.9.2"},"purl":"pkg:npm/a@0.9.2","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"a","version":"1.0.0"},"purl":"pkg:npm/a@1.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	respNPMBPackageJSON := fmt.Sprintf(`{
  "packageKey": {"system":"npm","name":"b"},
  "purl":"pkg:npm/b",
  "versions":[
    {"versionKey":{"system":"npm","name":"b","version":"1.8.0"},"purl":"pkg:npm/b@1.8.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"1.9.0"},"purl":"pkg:npm/b@1.9.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"1.9.1"},"purl":"pkg:npm/b@1.9.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"2.0.0"},"purl":"pkg:npm/b@2.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	respNPMBPackageJSONLow := fmt.Sprintf(`{
  "packageKey": {"system":"npm","name":"b"},
  "purl":"pkg:npm/b",
  "versions":[
    {"versionKey":{"system":"npm","name":"b","version":"1.9.9"},"purl":"pkg:npm/b@1.9.9","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"2.0.0"},"purl":"pkg:npm/b@2.0.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"2.0.1"},"purl":"pkg:npm/b@2.0.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"npm","name":"b","version":"2.1.0"},"purl":"pkg:npm/b@2.1.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d90), tsDaysAgo(d60), tsDaysAgo(midAgo), tsDaysAgo(recentAgo))

	respPyPIALatest := fmt.Sprintf(`{
  "packageKey": {"system":"pypi","name":"aa"},
  "purl":"pkg:pypi/a",
  "versions":[
    {"versionKey":{"system":"pypi","name":"aa","version":"0.9.0"},"purl":"pkg:pypi/a@0.9.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"aa","version":"0.9.1"},"purl":"pkg:pypi/a@0.9.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"aa","version":"0.9.2"},"purl":"pkg:pypi/a@0.9.2","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"aa","version":"1.0.0"},"purl":"pkg:pypi/a@1.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	respPyPIBLatest := fmt.Sprintf(`{
  "packageKey": {"system":"pypi","name":"b"},
  "purl":"pkg:pypi/b",
  "versions":[
    {"versionKey":{"system":"pypi","name":"bb","version":"1.9.9"},"purl":"pkg:pypi/b@1.9.9","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"bb","version":"1.9.10"},"purl":"pkg:pypi/b@1.9.10","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"bb","version":"1.9.11"},"purl":"pkg:pypi/b@1.9.11","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d90), tsDaysAgo(d60), tsDaysAgo(d60))

	respPyPICLatest := fmt.Sprintf(`{
  "packageKey": {"system":"pypi","name":"cc"},
  "purl":"pkg:pypi/c",
  "versions":[
    {"versionKey":{"system":"pypi","name":"cc","version":"2.9.9"},"purl":"pkg:pypi/c@2.9.9","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"cc","version":"2.9.10"},"purl":"pkg:pypi/c@2.9.10","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"cc","version":"2.9.11"},"purl":"pkg:pypi/c@2.9.11","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"pypi","name":"cc","version":"3.0.0"},"purl":"pkg:pypi/c@3.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	respMavenALatest := fmt.Sprintf(`{
  "packageKey": {"system":"maven","name":"com.example:a"},
  "purl":"pkg:maven/com.example/a",
  "versions":[
    {"versionKey":{"system":"maven","name":"com.example:a","version":"0.9.0"},"purl":"pkg:maven/com.example/a@0.9.0","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:a","version":"0.9.1"},"purl":"pkg:maven/com.example/a@0.9.1","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:a","version":"0.9.2"},"purl":"pkg:maven/com.example/a@0.9.2","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:a","version":"1.0.0"},"purl":"pkg:maven/com.example/a@1.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	// For the "one stale → Low" case, make the oldest-newer ~60d ago.
	respMavenBLatest := fmt.Sprintf(`{
  "packageKey": {"system":"maven","name":"com.example:b"},
  "purl":"pkg:maven/com.example/b",
  "versions":[
    {"versionKey":{"system":"maven","name":"com.example:b","version":"1.9.9"},"purl":"pkg:maven/com.example/b@1.9.9","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:b","version":"1.9.10"},"purl":"pkg:maven/com.example/b@1.9.10","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:b","version":"1.9.11"},"purl":"pkg:maven/com.example/b@1.9.11","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d90), tsDaysAgo(d60+10), tsDaysAgo(d60))

	respMavenCLatest := fmt.Sprintf(`{
  "packageKey": {"system":"maven","name":"com.example:c"},
  "purl":"pkg:maven/com.example/c",
  "versions":[
    {"versionKey":{"system":"maven","name":"com.example:c","version":"2.9.9"},"purl":"pkg:maven/com.example/c@2.9.9","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:c","version":"2.9.10"},"purl":"pkg:maven/com.example/c@2.9.10","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:c","version":"2.9.11"},"purl":"pkg:maven/com.example/c@2.9.11","publishedAt":"%s","isDefault":false,"isDeprecated":false},
    {"versionKey":{"system":"maven","name":"com.example:c","version":"3.0.0"},"purl":"pkg:maven/com.example/c@3.0.0","publishedAt":"%s","isDefault":true,"isDeprecated":false}
  ]
}`, tsDaysAgo(d455), tsDaysAgo(d445), tsDaysAgo(d435), tsDaysAgo(longAgo))

	tests := []tc{
		{
			name: "npm-lock all-latest → VeryLow true → score 10",
			files: map[string]string{
				// Aligns with expected packages: a@1.0.0 and b@2.0.0.
				"package-lock.json": `{
		  "name": "example",
		  "version": "1.0.0",
		  "lockfileVersion": 1,
		  "dependencies": {
		    "a": { "version": "1.0.0" },
		    "b": { "version": "2.0.0" }
		  }
		}`,
			},
			ddResponses: map[string]string{
				"NPM/a": respNPMA, // latest is 1.0.0 (default)
				"NPM/b": respNPMB, // latest is 2.0.0 (default)
			},
			wantScore: 10, // mean = 0 → VeryLow probe true → score 10
		},
		{
			name: "javascript package.json all-latest → VeryLow true → score 10",
			files: map[string]string{
				"package.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "dependencies": {
				    "a": "1.0.0",
				    "b": "2.0.0"
				  }
				}`,
				"package-lock.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "lockfileVersion": 2,
				  "requires": true,
				  "packages": {
				    "": {
				      "name": "example",
				      "version": "1.0.0",
				      "dependencies": {
				        "a": "1.0.0",
				        "b": "2.0.0"
				      }
				    },
				    "node_modules/a": {
				      "version": "1.0.0"
				    },
				    "node_modules/b": {
				      "version": "2.0.0"
				    }
				  },
				  "dependencies": {
				    "a": {
				      "version": "1.0.0"
				    },
				    "b": {
				      "version": "2.0.0"
				    }
				  }
				}`,
			},
			ddResponses: map[string]string{
				"NPM/a": respNPMAPackageJSON,
				"NPM/b": respNPMBPackageJSON,
			},
			wantScore: 10, // mean = 0
		},
		{
			name: "javascript package.json very stale → High true → score 0",
			files: map[string]string{
				"package.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "dependencies": {
				    "a": "0.9.2",
				    "b": "1.9.1"
				  }
				}`,
				"package-lock.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "lockfileVersion": 2,
				  "requires": true,
				  "packages": {
				    "": {
				      "name": "example",
				      "version": "1.0.0",
				      "dependencies": {
				        "a": "0.9.2",
				        "b": "1.9.1"
				      }
				    },
				    "node_modules/a": {
				      "version": "0.9.2"
				    },
				    "node_modules/b": {
				      "version": "1.9.1"
				    }
				  },
				  "dependencies": {
				    "a": {
				      "version": "0.9.2"
				    },
				    "b": {
				      "version": "1.9.1"
				    }
				  }
				}`,
			},
			ddResponses: map[string]string{
				"NPM/a": respNPMAPackageJSON, // oldest newer for a is 1.0.0 (long ago)
				"NPM/b": respNPMBPackageJSON, // oldest newer for b is 2.0.0 (long ago)
			},
			wantScore: 0, // mean ≥ 180d
		},
		{
			name: "javascript package.json moderately stale → Low true → score 5",
			files: map[string]string{
				"package.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "dependencies": {
				    "a": "1.0.0",
				    "b": "2.0.0"
				  }
				}`,
				"package-lock.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "lockfileVersion": 2,
				  "requires": true,
				  "packages": {
				    "": {
				      "name": "example",
				      "version": "1.0.0",
				      "dependencies": {
				        "a": "1.0.0",
				        "b": "2.0.0"
				      }
				    },
				    "node_modules/a": {
				      "version": "1.0.0"
				    },
				    "node_modules/b": {
				      "version": "2.0.0"
				    }
				  },
				  "dependencies": {
				    "a": {
				      "version": "1.0.0"
				    },
				    "b": {
				      "version": "2.0.0"
				    }
				  }
				}`,
			},
			ddResponses: map[string]string{
				"NPM/a": respNPMAPackageJSON,    // a is latest → contributes 0
				"NPM/b": respNPMBPackageJSONLow, // b's oldest newer (2.0.1) published ~30d ago
			},
			wantScore: 5, // mean ~ (0 + ~30d)/2 ∈ [14d, 180d)
		},
		{
			name: "go.mod very stale → High true → score 0",
			files: map[string]string{
				"go.mod": "module example.com/mod\n\ngo 1.21\nrequire github.com/mod/a v1.2.3\n",
			},
			ddResponses: map[string]string{
				"GO/github.com/mod/a": respGoModA, // oldest newer published long ago → High
			},
			wantScore: 0,
		},
		{
			name: "go.mod moderately stale → Low true → score 5",
			files: map[string]string{
				"go.mod": "module example.com/mod\n\ngo 1.21\nrequire github.com/mod/a v1.2.3\n",
			},
			ddResponses: map[string]string{
				"GO/github.com/mod/a": respGoModALow, // oldest newer v1.2.4 @ ~30 days ago
			},
			wantScore: 5,
		},
		{
			name: "go.mod all-latest → VeryLow true → score 10",
			files: map[string]string{
				"go.mod": "module example.com/mod\n\ngo 1.21\nrequire github.com/mod/a v1.2.3\n",
			},
			ddResponses: map[string]string{
				"GO/github.com/mod/a": respGoModAVeryLowLatest, // v1.2.3 is default/latest → mean = 0
			},
			wantScore: 10,
		},
		{
			name: "javascript package.json all-latest → VeryLow true → score 10",
			files: map[string]string{
				"package.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "dependencies": {
				    "a": "1.0.0",
				    "b": "2.0.0"
				  }
				}`,
				"package-lock.json": `{
				  "name": "example",
				  "version": "1.0.0",
				  "lockfileVersion": 2,
				  "requires": true,
				  "packages": {
				    "": {
				      "name": "example",
				      "version": "1.0.0",
				      "dependencies": {
				        "a": "1.0.0",
				        "b": "2.0.0"
				      }
				    },
				    "node_modules/a": {
				      "version": "1.0.0"
				    },
				    "node_modules/b": {
				      "version": "2.0.0"
				    }
				  },
				  "dependencies": {
				    "a": {
				      "version": "1.0.0"
				    },
				    "b": {
				      "version": "2.0.0"
				    }
				  }
				}`,
			},
			ddResponses: map[string]string{
				"NPM/a": respNPMAPackageJSON, // latest is 1.0.0 (default)
				"NPM/b": respNPMBPackageJSON, // latest is 2.0.0 (default)
			},
			wantScore: 10, // mean = 0 → VeryLow probe true → score 10
		},
		{
			name: "requirements.txt all-latest → VeryLow true → score 10",
			files: map[string]string{
				"requirements.txt": "aa==1.0.0\nbb==1.9.11\ncc==3.0.0\n",
			},
			ddResponses: map[string]string{
				"PYPI/aa": respPyPIALatest,
				"PYPI/bb": respPyPIBLatest,
				"PYPI/cc": respPyPICLatest,
			},
			wantScore: 10, // mean = 0
		},
		{
			name: "requirements.txt one stale → Low true → score 5",
			files: map[string]string{
				"requirements.txt": "aa==1.0.0\nbb==1.9.10\ncc==3.0.0\n",
			},
			ddResponses: map[string]string{
				"PYPI/aa": respPyPIALatest, // aa latest 1.0.0
				"PYPI/bb": respPyPIBLatest, // bb oldest newer 1.9.11 @ ~30d
				"PYPI/cc": respPyPICLatest, // cc latest 3.0.0
			},
			wantScore: 5,
		},
		{
			name: "requirements.txt very stale → High true → score 0",
			files: map[string]string{
				"requirements.txt": "aa==0.9.2\nbb==1.9.11\ncc==2.9.11\n",
			},
			ddResponses: map[string]string{
				"PYPI/aa": respPyPIALatest, // oldest newer 1.0.0 @ ≥365d
				"PYPI/bb": respPyPIBLatest, // latest, contributes 0
				"PYPI/cc": respPyPICLatest, // oldest newer 3.0.0 @ ≥365d
			},
			wantScore: 0,
		},
		// ---- Add these three pom.xml testcases ----

		{
			name: "pom.xml all-latest → VeryLow true → score 10",
			files: map[string]string{
				"pom.xml": `
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>example</groupId>
  <artifactId>example</artifactId>
  <version>1.0.0</version>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>a</artifactId>
      <version>1.0.0</version>
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>b</artifactId>
      <version>1.9.11</version>
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>c</artifactId>
      <version>3.0.0</version>
    </dependency>
  </dependencies>
</project>`,
			},
			ddResponses: map[string]string{
				"MAVEN/com.example:a": respMavenALatest,
				"MAVEN/com.example:b": respMavenBLatest,
				"MAVEN/com.example:c": respMavenCLatest,
			},
			wantScore: 10, // mean = 0
		},
		{
			name: "pom.xml one stale → Low true → score 5",
			files: map[string]string{
				"pom.xml": `
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>example</groupId>
  <artifactId>example</artifactId>
  <version>1.0.0</version>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>a</artifactId>
      <version>1.0.0</version>
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>b</artifactId>
      <version>1.9.10</version> <!-- one behind; oldest-newer ~60d -->
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>c</artifactId>
      <version>3.0.0</version>
    </dependency>
  </dependencies>
</project>`,
			},
			ddResponses: map[string]string{
				"MAVEN/com.example:a": respMavenALatest, // latest → 0
				"MAVEN/com.example:b": respMavenBLatest, // oldest-newer ~60d → Low mean
				"MAVEN/com.example:c": respMavenCLatest, // latest → 0
			},
			wantScore: 5,
		},
		{
			name: "pom.xml very stale → High true → score 0",
			files: map[string]string{
				"pom.xml": `
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>example</groupId>
  <artifactId>example</artifactId>
  <version>1.0.0</version>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>a</artifactId>
      <version>0.9.2</version> <!-- oldest-newer 1.0.0 long ago -->
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>b</artifactId>
      <version>1.9.11</version> <!-- latest, contributes 0 -->
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>c</artifactId>
      <version>2.9.11</version> <!-- oldest-newer 3.0.0 long ago -->
    </dependency>
  </dependencies>
</project>`,
			},
			ddResponses: map[string]string{
				"MAVEN/com.example:a": respMavenALatest, // oldest-newer ≥180d
				"MAVEN/com.example:b": respMavenBLatest, // latest
				"MAVEN/com.example:c": respMavenCLatest, // oldest-newer ≥180d
			},
			wantScore: 0,
		},
	}

	//nolint:paralleltest // test setup uses shared mocks
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Intercept deps.dev calls using a RoundTripper that serves testcase-constant JSON bodies.
			orig := http.DefaultTransport
			http.DefaultTransport = rtFunc(func(req *http.Request) (*http.Response, error) {
				// Expect path like /v3alpha/systems/{system}/packages/{name}
				parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
				var system, name string
				for i := 0; i < len(parts)-1; i++ {
					if parts[i] == "systems" && i+1 < len(parts) {
						system = parts[i+1]
					}
					if parts[i] == "packages" && i+1 < len(parts) {
						name = strings.Join(parts[i+1:], "/")
					}
				}
				key := system + "/" + name
				body, ok := tt.ddResponses[key]
				if !ok {
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(strings.NewReader("no mock for " + key)),
						Header:     make(http.Header),
						Request:    req,
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Request:    req,
				}, nil
			})
			t.Cleanup(func() { http.DefaultTransport = orig })

			// Create temp directory with test files for scalibr to scan.
			tmpDir := t.TempDir()
			for fileName, fileContents := range tt.files {
				filePath := tmpDir + "/" + fileName
				if err := os.WriteFile(filePath, []byte(fileContents), 0o600); err != nil {
					t.Fatalf("failed to write temp file %s: %v", fileName, err)
				}
			}

			// If there's a package.json but no package-lock.json, generate a minimal one
			// since scalibr's NPM extractor requires a lockfile
			//nolint:nestif // test setup requires nested structure
			if _, hasPackageJSON := tt.files["package.json"]; hasPackageJSON {
				if _, hasLock := tt.files["package-lock.json"]; !hasLock {
					// Parse package.json to extract dependencies
					lockContent := `{"name":"example","version":"1.0.0","lockfileVersion":2,"requires":true,"packages":{"":{"name":"example","version":"1.0.0","dependencies":{`
					first := true
					for _, line := range strings.Split(tt.files["package.json"], "\n") {
						if strings.Contains(line, `"dependencies"`) {
							continue
						}
						// Simple parser for "name": "version" pairs
						if strings.Contains(line, `":`) && !strings.Contains(line, `"name"`) && !strings.Contains(line, `"version"`) {
							parts := strings.Split(line, `":`)
							if len(parts) == 2 {
								name := strings.Trim(strings.TrimSpace(parts[0]), `"`)
								version := strings.Trim(strings.TrimSpace(parts[1]), `,}"`)
								if name != "" && version != "" {
									if !first {
										lockContent += ","
									}
									lockContent += `"` + name + `":"` + version + `"`
									first = false
								}
							}
						}
					}
					lockContent += `}}},"dependencies":{`
					first = true
					for _, line := range strings.Split(tt.files["package.json"], "\n") {
						if strings.Contains(line, `"dependencies"`) {
							continue
						}
						if strings.Contains(line, `":`) && !strings.Contains(line, `"name"`) && !strings.Contains(line, `"version"`) {
							parts := strings.Split(line, `":`)
							if len(parts) == 2 {
								name := strings.Trim(strings.TrimSpace(parts[0]), `"`)
								version := strings.Trim(strings.TrimSpace(parts[1]), `,}"`)
								if name != "" && version != "" {
									if !first {
										lockContent += ","
									}
									lockContent += `"` + name + `":{"version":"` + version + `"}`
									first = false
								}
							}
						}
					}
					lockContent += `}}`
					lockPath := tmpDir + "/package-lock.json"
					if err := os.WriteFile(lockPath, []byte(lockContent), 0o600); err != nil {
						t.Fatalf("failed to write generated package-lock.json: %v", err)
					}
				}
			}

			// gomock repo that returns the temp directory path.
			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			main := "main"
			mockRepo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
			mockRepo.EXPECT().GetDefaultBranchName().Return(main, nil).AnyTimes()
			mockRepo.EXPECT().LocalPath().Return(tmpDir, nil).AnyTimes()

			dl := scut.TestDetailLogger{}
			req := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := MTTUDependencies(req)

			if tt.wantErr {
				if res.Error == nil {
					t.Fatalf("expected error, got nil (score=%d)", res.Score)
				}
				return
			}
			if res.Error != nil {
				t.Fatalf("unexpected error: %v", res.Error)
			}
			if res.Score != tt.wantScore {
				t.Fatalf("unexpected score: want %d, got %d", tt.wantScore, res.Score)
			}
			if res.Name != "MTTUDependencies" {
				t.Fatalf("unexpected check name: %s", res.Name)
			}
		})
	}
}
