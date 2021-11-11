module github.com/ossf/scorecard/v3

go 1.17

require (
	cloud.google.com/go/bigquery v1.24.0
	cloud.google.com/go/pubsub v1.17.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.8
	github.com/bradleyfalzon/ghinstallation/v2 v2.0.3
	github.com/go-git/go-git/v5 v5.4.2
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.6.0
	github.com/google/go-github/v38 v38.1.0
	github.com/h2non/filetype v1.1.1
	github.com/jszwec/csvutil v1.5.1
	github.com/moby/buildkit v0.8.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	github.com/shurcooL/githubv4 v0.0.0-20201206200315-234843c633fa
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	github.com/spf13/cobra v1.2.1
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.19.1
	gocloud.dev v0.24.0
	golang.org/x/tools v0.1.7
	google.golang.org/genproto v0.0.0-20210924002016-3dee208752a0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	mvdan.cc/sh/v3 v3.4.0
)

require (
	github.com/rhysd/actionlint v1.6.7
	gotest.tools v2.2.0+incompatible
)

replace (
	// https://deps.dev/advisory/OSV/GO-2021-0057?from=%2Fgo%2Fgithub.com%252Fbuger%252Fjsonparser%2Fv1.0.0
	github.com/buger/jsonparser => github.com/buger/jsonparser v1.1.1
	// https://deps.dev/advisory/OSV/GO-2020-0017?from=%2Fgo%2Fk8s.io%252Fclient-go%2Fv0.0.0-20200207030105-473926661c44
	github.com/dgrijalva/jwt-go v0.0.0-20170104182250-a601269ab70c => github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/golang-jwt/jwt v3.2.1+incompatible
	// https://go.googlesource.com/vulndb/+/refs/heads/master/reports/GO-2020-0020.yaml
	github.com/gorilla/handlers => github.com/gorilla/handlers v1.3.0
	// https://github.com/miekg/dns/issues/1037
	github.com/miekg/dns => github.com/miekg/dns v1.1.25-0.20191211073109-8ebf2e419df7
	// This replace is for https://nvd.nist.gov/vuln/detail/CVE-2021-3538
	github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181016170032-d91630c85102
	// This replace is for https://github.com/advisories/GHSA-25xm-hr59-7c27
	github.com/ulikunitz/xz => github.com/ulikunitz/xz v0.5.8
)
