module github.com/ossf/scorecard/v4

go 1.17

// TODO(go.mod): Is there a reason these deps are kept separately from the
//               other `require`s?
require (
	github.com/rhysd/actionlint v1.6.15
	gotest.tools v2.2.0+incompatible
)

require (
	cloud.google.com/go/bigquery v1.34.1
	cloud.google.com/go/monitoring v1.4.0 // indirect
	cloud.google.com/go/pubsub v1.23.1
	cloud.google.com/go/trace v1.2.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.12
	github.com/bombsimon/logrusr/v2 v2.0.1
	github.com/bradleyfalzon/ghinstallation/v2 v2.0.4
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-logr/logr v1.2.3
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.8
	github.com/google/go-containerregistry v0.9.0
	github.com/google/go-github/v38 v38.1.0
	github.com/h2non/filetype v1.1.3
	github.com/jszwec/csvutil v1.7.0
	github.com/moby/buildkit v0.10.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/gomega v1.19.0
	github.com/shurcooL/githubv4 v0.0.0-20201206200315-234843c633fa
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	go.opencensus.io v0.23.0
	gocloud.dev v0.25.0
	golang.org/x/text v0.3.7
	golang.org/x/tools v0.1.10
	google.golang.org/genproto v0.0.0-20220622131801-db39fadba55f
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0
	mvdan.cc/sh/v3 v3.5.1
)

require (
	github.com/caarlos0/env/v6 v6.9.3
	github.com/mcuadros/go-jsonschema-generator v0.0.0-20200330054847-ba7a369d4303
	github.com/onsi/ginkgo/v2 v2.1.4
	sigs.k8s.io/release-utils v0.6.0
)

require (
	cloud.google.com/go v0.102.1 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/storage v1.22.1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/aws/aws-sdk-go v1.43.31 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.11.4 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/docker/cli v20.10.16+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.16+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-github/v41 v41.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/googleapis/go-type-adapters v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/klauspost/compress v1.15.4 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20220114050600-8b9d41f48198 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/prometheus v2.5.0+incompatible // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29 // indirect
	golang.org/x/net v0.0.0-20220617184016-355a448f1bc9 // indirect
	golang.org/x/oauth2 v0.0.0-20220622183110-fd043fe589d2 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220615213510-4f61da869c0c // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/api v0.85.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

replace (
	// https://deps.dev/advisory/OSV/GO-2021-0057?from=%2Fgo%2Fgithub.com%252Fbuger%252Fjsonparser%2Fv1.0.0
	github.com/buger/jsonparser => github.com/buger/jsonparser v1.1.1
	// https://deps.dev/advisory/OSV/GO-2020-0017?from=%2Fgo%2Fk8s.io%252Fclient-go%2Fv0.0.0-20200207030105-473926661c44
	github.com/dgrijalva/jwt-go v0.0.0-20170104182250-a601269ab70c => github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/golang-jwt/jwt v3.2.1+incompatible
	// This replace is for GHSA-qq97-vm5h-rrhg
	github.com/docker/distribution => github.com/docker/distribution v2.8.0+incompatible
	// https://go.googlesource.com/vulndb/+/refs/heads/master/reports/GO-2020-0020.yaml
	github.com/gorilla/handlers => github.com/gorilla/handlers v1.3.0
	// https://github.com/miekg/dns/issues/1037
	github.com/miekg/dns => github.com/miekg/dns v1.1.25-0.20191211073109-8ebf2e419df7
	//https://github.com/opencontainers/distribution-spec/security/advisories/GHSA-mc8v-mgrf-8f4m
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	// This replace is for https://nvd.nist.gov/vuln/detail/CVE-2021-3538
	github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181016170032-d91630c85102
	// This replace is for https://github.com/advisories/GHSA-25xm-hr59-7c27
	github.com/ulikunitz/xz => github.com/ulikunitz/xz v0.5.8

)
