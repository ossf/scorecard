SHELL := /bin/bash
GINKGO := ginkgo
GIT_HASH := $(shell git rev-parse HEAD)
GIT_VERSION ?= $(shell git describe --tags --always --dirty)
SOURCE_DATE_EPOCH=$(shell git log --date=iso8601-strict -1 --pretty=%ct)
GOLANGCI_LINT := golangci-lint
PROTOC_GEN_GO := protoc-gen-go
MOCKGEN := mockgen
PROTOC := $(shell which protoc)
GORELEASER := goreleaser
IMAGE_NAME = scorecard
OUTPUT = output
PLATFORM="linux/amd64,linux/arm64,linux/386,linux/arm"
LDFLAGS=$(shell ./scripts/version-ldflags)
KOCACHE_PATH=/tmp/ko

define create_kocache_path
  mkdir -p $(KOCACHE_PATH)
endef



############################### make help #####################################
.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; \
			printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ \
			{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
			/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

###############################################################################

##@ Development
################################ make install #################################
.PHONY: install
install: ## Installs all dependencies needed to compile Scorecard
install: | $(PROTOC)
	@echo Installing tools from tools/tools.go
	cd tools; cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

$(PROTOC):
	ifeq (,$(PROTOC))
		$(error download and install protobuf compiler package - https://developers.google.com/protocol-buffers/docs/downloads)
	endif
###############################################################################

##@ Build
################################## make all ###################################
all:  ## Runs build, test and verify
all-targets = build check-linter check-osv validate-docs add-projects validate-projects 
.PHONY: all all-targets-update-dependencies $(all-targets) update-dependencies tree-status
all-targets-update-dependencies: $(all-targets) | update-dependencies
all: update-dependencies all-targets-update-dependencies tree-status

update-dependencies: ## Update go dependencies for all modules
	# Update root go modules
	go mod tidy && go mod verify
	cd tools; go mod tidy && go mod verify; cd ../
	cd attestor; go mod tidy && go mod verify; cd ../

$(GOLANGCI_LINT): install
check-linter: ## Install and run golang linter
check-linter: $(GOLANGCI_LINT)
	# Run golangci-lint linter
	golangci-lint run -c .golangci.yml

check-osv: ## Checks osv.dev for any vulnerabilities
check-osv: $(install)
	# Run stunning-tribble for checking the dependencies have any OSV
	go list -m -f '{{if not (or  .Main)}}{{.Path}}@{{.Version}}_{{.Replace}}{{end}}' all \
			| stunning-tribble
	# Checking the tools which also has go.mod
	cd tools; go list -m -f '{{if not (or  .Main)}}{{.Path}}@{{.Version}}_{{.Replace}}{{end}}' all \
			| stunning-tribble ; cd ..
	# Checking the attestor module for vulns
	cd attestor; go list -m -f '{{if not (or  .Main)}}{{.Path}}@{{.Version}}_{{.Replace}}{{end}}' all \
			| stunning-tribble ; cd ..

add-projects: ## Adds new projects to ./cron/internal/data/projects.csv
add-projects: ./cron/internal/data/projects.csv | build-add-script
	# Add new projects to ./cron/internal/data/projects.csv
	./cron/internal/data/add/add ./cron/internal/data/projects.csv ./cron/internal/data/projects.new.csv
	mv ./cron/internal/data/projects.new.csv ./cron/internal/data/projects.csv

validate-projects: ## Validates ./cron/internal/data/projects.csv
validate-projects: ./cron/internal/data/projects.csv | build-validate-script
	# Validate ./cron/internal/data/projects.csv
	./cron/internal/data/validate/validate ./cron/internal/data/projects.csv

tree-status: | all-targets-update-dependencies ## Verify tree is clean and all changes are committed
	# Verify the tree is clean and all changes are commited
	./scripts/tree-status

###############################################################################

################################## make build #################################
## Build all cron-related targets
build-cron: build-controller build-worker build-cii-worker \
	build-shuffler build-bq-transfer build-github-server \
	build-webhook build-add-script build-validate-script build-update-script

build-targets = generate-mocks generate-docs build-proto build-scorecard build-cron 
.PHONY: build $(build-targets)
build: ## Build all binaries and images in the repo.
build: $(build-targets)

build-proto: ## Compiles and generates all required protobufs
build-proto: cron/internal/data/request.pb.go cron/internal/data/metadata.pb.go
cron/internal/data/request.pb.go: cron/internal/data/request.proto |  $(PROTOC) install
	protoc --go_out=../../../ cron/internal/data/request.proto
cron/internal/data/metadata.pb.go: cron/internal/data/metadata.proto |  $(PROTOC) install
	protoc --go_out=../../../ cron/internal/data/metadata.proto

generate-mocks: ## Compiles and generates all mocks using mockgen.
generate-mocks: clients/mockclients/repo_client.go clients/mockclients/repo.go clients/mockclients/cii_client.go checks/mockclients/vulnerabilities.go cmd/packagemanager_mockclient.go
clients/mockclients/repo_client.go: clients/repo_client.go
	# Generating MockRepoClient
	$(MOCKGEN) -source=clients/repo_client.go -destination=clients/mockclients/repo_client.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
clients/mockclients/repo.go: clients/repo.go
	# Generating MockRepo
	$(MOCKGEN) -source=clients/repo.go -destination=clients/mockclients/repo.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
clients/mockclients/cii_client.go: clients/cii_client.go
	# Generating MockCIIClient
	$(MOCKGEN) -source=clients/cii_client.go -destination=clients/mockclients/cii_client.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
checks/mockclients/vulnerabilities.go: clients/vulnerabilities.go
	# Generating MockCIIClient
	$(MOCKGEN) -source=clients/vulnerabilities.go -destination=clients/mockclients/vulnerabilities.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
cmd/packagemanager_mockclient.go: cmd/packagemanager_client.go
	# Generating MockPackageManagerClient
	$(MOCKGEN) -source=cmd/packagemanager_client.go -destination=cmd/packagemanager_mockclient.go -package=cmd -copyright_file=clients/mockclients/license.txt
generate-docs: ## Generates docs
generate-docs: validate-docs docs/checks.md
docs/checks.md: docs/checks/internal/checks.yaml docs/checks/internal/*.go docs/checks/internal/generate/*.go
	# Generating checks.md
	go run ./docs/checks/internal/generate/main.go docs/checks.md

validate-docs: docs/checks/internal/generate/main.go
	# Validating checks.yaml
	go run ./docs/checks/internal/validate/main.go

build-scorecard: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'

build-releaser: ## Runs goreleaser on the repo
	# Run go releaser on the Scorecard repo
	$(GORELEASER) check
	VERSION_LDFLAGS="$(LDFLAGS)" $(GORELEASER) release --snapshot --rm-dist --skip-publish --skip-sign

build-controller: ## Runs go build on the cron PubSub controller
	# Run go build on the cron PubSub controller
	cd cron/internal/controller && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o controller

build-worker: ## Runs go build on the cron PubSub worker
	# Run go build on the cron PubSub worker
	cd cron/internal/worker && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o worker

build-cii-worker: ## Runs go build on the CII worker
	# Run go build on the CII worker
	cd cron/internal/cii && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o cii-worker

build-shuffler: ## Runs go build on the cron shuffle script
	# Run go build on the cron shuffle script
	cd cron/internal/shuffle && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o shuffle

build-bq-transfer: ## Runs go build on the BQ transfer cron job
build-bq-transfer: ./cron/internal/bq/*.go
	# Run go build on the Copier cron job
	cd cron/internal/bq && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o data-transfer

build-github-server: ## Runs go build on the GitHub auth server
build-github-server: ./clients/githubrepo/roundtripper/tokens/*
	# Run go build on the GitHub auth server
	cd clients/githubrepo/roundtripper/tokens/server && \
		CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o github-auth-server

build-webhook: ## Runs go build on the cron webhook
	# Run go build on the cron webhook
	cd cron/internal/webhook && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o webhook

build-add-script: ## Runs go build on the add script
build-add-script: cron/internal/data/add/add
cron/internal/data/add/add: cron/internal/data/add/*.go cron/internal/data/*.go cron/internal/data/projects.csv
	# Run go build on the add script
	cd cron/internal/data/add && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o add

build-validate-script: ## Runs go build on the validate script
build-validate-script: cron/internal/data/validate/validate
cron/internal/data/validate/validate: cron/internal/data/validate/*.go cron/internal/data/*.go cron/internal/data/projects.csv
	# Run go build on the validate script
	cd cron/internal/data/validate && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o validate

build-update-script: ## Runs go build on the update script
build-update-script: cron/internal/data/update/projects-update
cron/internal/data/update/projects-update:  cron/internal/data/update/*.go cron/internal/data/*.go
	# Run go build on the update script
	cd cron/internal/data/update && CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'  -o projects-update

ko-targets = scorecard-ko cron-controller-ko cron-worker-ko cron-cii-worker-ko cron-bq-transfer-ko cron-webhook-ko cron-github-server-ko
.PHONY: ko-build-everything $(ko-targets)
ko-build-everything: $(ko-targets)

scorecard-ko:
	$(call create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) LDFLAGS="$(LDFLAGS)" \
	KO_CACHE=$(KOCACHE_PATH) ko build -B \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4
cron-controller-ko:
	$(call_create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-batch-controller LDFLAGS="$(LDFLAGS)" \
	KOCACHE=$(KOCACHE_PATH) ko build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4/cron/internal/controller
cron-worker-ko:
	$(call_create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-batch-worker LDFLAGS="$(LDFLAGS)" \
	KOCACHE=$(KOCACHE_PATH) ko build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4/cron/internal/worker
cron-cii-worker-ko:
	$(call_create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-cii-worker LDFLAGS="$(LDFLAGS)" \
	KOCACHE=$(KOCACHE_PATH) ko build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4/cron/internal/cii
cron-bq-transfer-ko:
	$(call_create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-bq-transfer LDFLAGS="$(LDFLAGS)" \
	KOCACHE=$(KOCACHE_PATH) ko build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4/cron/internal/bq
cron-webhook-ko:
	$(call_create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-cron-webhook LDFLAGS="$(LDFLAGS)" \
	KOCACHE=$(KOCACHE_PATH) ko build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4/cron/internal/webhook
cron-github-server-ko:
	$(call_create_kocache_path)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-github-server LDFLAGS="$(LDFLAGS)" \
	KOCACHE=$(KOCACHE_PATH) ko build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper/tokens/server

docker-targets = scorecard-docker cron-controller-docker cron-worker-docker cron-cii-worker-docker cron-bq-transfer-docker cron-webhook-docker cron-github-server-docker
.PHONY: dockerbuild $(docker-targets)
dockerbuild: $(docker-targets)

scorecard-docker:
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile --tag $(IMAGE_NAME)
cron-controller-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/controller/Dockerfile --tag $(IMAGE_NAME)-batch-controller
cron-worker-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/worker/Dockerfile --tag $(IMAGE_NAME)-batch-worker
cron-cii-worker-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/cii/Dockerfile --tag $(IMAGE_NAME)-cii-worker
cron-bq-transfer-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/bq/Dockerfile --tag $(IMAGE_NAME)-bq-transfer
cron-webhook-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/webhook/Dockerfile --tag ${IMAGE_NAME}-webhook
cron-github-server-docker:
	DOCKER_BUILDKIT=1 docker build . --file clients/githubrepo/roundtripper/tokens/server/Dockerfile --tag ${IMAGE_NAME}-github-server
###############################################################################

##@ Tests
################################# make test ###################################
test-targets = unit-test unit-test-attestor e2e-pat e2e-gh-token ci-e2e
.PHONY: test $(test-targets)
test: $(test-targets)

unit-test: ## Runs unit test without e2e
	# Run unit tests, ignoring e2e tests
	# run the go tests and gen the file coverage-all used to do the integration with codecov
	SKIP_GINKGO=1 go test -race -covermode=atomic  -coverprofile=unit-coverage.out `go list ./...`

unit-test-attestor: ## Runs unit tests on scorecard-attestor
	cd attestor; SKIP_GINKGO=1 go test -covermode=atomic -coverprofile=unit-coverage.out `go list ./...`; cd ..;

$(GINKGO): install

check-env:
ifndef GITHUB_AUTH_TOKEN
	$(error GITHUB_AUTH_TOKEN is undefined)
endif

e2e-pat: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
e2e-pat: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	TOKEN_TYPE="PAT" $(GINKGO) --race -p -v -cover -coverprofile=e2e-coverage.out --keep-separate-coverprofiles ./...

e2e-gh-token: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to default GITHUB_TOKEN
e2e-gh-token: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN set to secrets.GITHUB_TOKEN must be used to run this.
	TOKEN_TYPE="GITHUB_TOKEN" $(GINKGO) --race -p -v -cover -coverprofile=e2e-coverage.out --keep-separate-coverprofiles ./...
###############################################################################
