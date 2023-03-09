module github.com/ossf/scorecard/v4

go 1.19

require (
	github.com/rhysd/actionlint v1.6.15
	gotest.tools v2.2.0+incompatible
)

require (
	cloud.google.com/go/bigquery v1.44.0
	cloud.google.com/go/monitoring v1.8.0 // indirect
	cloud.google.com/go/pubsub v1.27.0
	cloud.google.com/go/trace v1.4.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.12
	github.com/bombsimon/logrusr/v2 v2.0.1
	github.com/bradleyfalzon/ghinstallation/v2 v2.1.0
	github.com/go-git/go-git/v5 v5.5.2
	github.com/go-logr/logr v1.2.3
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry v0.12.1
	github.com/google/go-github/v38 v38.1.0
	github.com/grafeas/kritis v0.2.3-0.20210120183821-faeba81c520c
	github.com/h2non/filetype v1.1.3
	github.com/jszwec/csvutil v1.8.0
	github.com/moby/buildkit v0.11.4
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/gomega v1.24.2
	github.com/shurcooL/githubv4 v0.0.0-20201206200315-234843c633fa
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	go.opencensus.io v0.24.0
	gocloud.dev v0.26.0
	golang.org/x/text v0.7.0
	golang.org/x/tools v0.5.1-0.20230117180257-8aba49bb5ea2
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	mvdan.cc/sh/v3 v3.6.0
)

require (
	github.com/Masterminds/semver/v3 v3.2.0
	github.com/caarlos0/env/v6 v6.10.0
	github.com/gobwas/glob v0.2.3
	github.com/google/osv-scanner v1.2.1-0.20230302232134-592acbc2539b
	github.com/mcuadros/go-jsonschema-generator v0.0.0-20200330054847-ba7a369d4303
	github.com/onsi/ginkgo/v2 v2.7.0
	sigs.k8s.io/release-utils v0.6.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	cloud.google.com/go/containeranalysis v0.6.0 // indirect
	cloud.google.com/go/kms v1.6.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.22 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.17 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/CycloneDX/cyclonedx-go v0.7.0 // indirect
	github.com/cloudflare/circl v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jedib0t/go-pretty/v6 v6.4.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/package-url/packageurl-go v0.1.1-0.20220428063043-89078438f170 // indirect
	github.com/pjbgf/sha1cd v0.2.3 // indirect
	github.com/skeema/knownhosts v1.1.0 // indirect
	github.com/spdx/gordf v0.0.0-20221230105357-b735bd5aac89 // indirect
	github.com/spdx/tools-golang v0.4.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/vuln v0.0.0-20230118164824-4ec8867cc0e6 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.20.0 // indirect
	k8s.io/apimachinery v0.20.0 // indirect
	k8s.io/client-go v0.20.0 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.2 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	cloud.google.com/go v0.105.0 // indirect
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/iam v0.7.0 // indirect
	cloud.google.com/go/storage v1.27.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20221026131551-cf6655e29de4 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/aws/aws-sdk-go v1.43.31 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.13.0 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/docker/cli v23.0.0-rc.1+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v23.0.0-rc.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-github/v45 v45.2.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/prometheus v2.5.0+incompatible // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/xanzy/go-gitlab v0.78.0
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/oauth2 v0.3.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.103.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.50.1 // indirect
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
	// This replace is for https://osv.dev/vulnerability/GHSA-r48q-9g5r-8q2h
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.16.0+incompatible
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
