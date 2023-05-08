module github.com/ossf/scorecard/v4

go 1.19

require (
	github.com/rhysd/actionlint v1.6.15
	gotest.tools v2.2.0+incompatible
)

require (
	cloud.google.com/go/bigquery v1.51.1
	cloud.google.com/go/monitoring v1.13.0 // indirect
	cloud.google.com/go/pubsub v1.30.1
	cloud.google.com/go/trace v1.9.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.14
	github.com/bombsimon/logrusr/v2 v2.0.1
	github.com/bradleyfalzon/ghinstallation/v2 v2.4.0
	github.com/go-git/go-git/v5 v5.6.1
	github.com/go-logr/logr v1.2.4
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry v0.15.1
	github.com/google/go-github/v38 v38.1.0
	github.com/grafeas/kritis v0.2.3-0.20210120183821-faeba81c520c
	github.com/h2non/filetype v1.1.3
	github.com/jszwec/csvutil v1.8.0
	github.com/moby/buildkit v0.11.6
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/gomega v1.27.6
	github.com/shurcooL/githubv4 v0.0.0-20201206200315-234843c633fa
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.opencensus.io v0.24.0
	gocloud.dev v0.29.0
	golang.org/x/text v0.9.0
	golang.org/x/tools v0.8.0
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	mvdan.cc/sh/v3 v3.6.0
)

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/caarlos0/env/v6 v6.10.0
	github.com/gobwas/glob v0.2.3
	github.com/google/osv-scanner v1.3.2
	github.com/mcuadros/go-jsonschema-generator v0.0.0-20200330054847-ba7a369d4303
	github.com/onsi/ginkgo/v2 v2.9.4
	github.com/otiai10/copy v1.11.0
	sigs.k8s.io/release-utils v0.6.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/containeranalysis v0.9.0 // indirect
	cloud.google.com/go/kms v1.10.1 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/CycloneDX/cyclonedx-go v0.7.1 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/arrow/go/v12 v12.0.0 // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/cloudflare/circl v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-github/v52 v52.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20230111200839-76d1ae5aea2b // indirect
	github.com/google/s2a-go v0.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/jedib0t/go-pretty/v6 v6.4.6 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/package-url/packageurl-go v0.1.1-0.20220428063043-89078438f170 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/prometheus/prometheus v0.42.0 // indirect
	github.com/skeema/knownhosts v1.1.0 // indirect
	github.com/spdx/gordf v0.0.0-20221230105357-b735bd5aac89 // indirect
	github.com/spdx/tools-golang v0.4.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/term v0.7.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/vuln v0.0.0-20230303230808-d3042fecc4e3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.26.1 // indirect
	k8s.io/apimachinery v0.26.1 // indirect
	k8s.io/client-go v0.26.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221207184640-f3cff1453715 // indirect
	k8s.io/utils v0.0.0-20221128185143-99ec85e7a448 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/compute v1.19.1 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	cloud.google.com/go/storage v1.29.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/acomagu/bufpipe v1.0.4 // indirect
	github.com/aws/aws-sdk-go v1.44.200 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/docker/cli v23.0.5+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v23.0.5+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.8.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vbatts/tar-split v0.11.3 // indirect
	github.com/xanzy/go-gitlab v0.83.0
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20230425010034-47ecfdc1ba53
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/oauth2 v0.7.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.118.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.54.0 // indirect
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
