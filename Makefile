SHELL := /bin/bash
GINKGO := ginkgo
GIT_HASH := $(shell git rev-parse HEAD)
GIT_VERSION ?= $(shell git describe --tags --always --dirty)
SOURCE_DATE_EPOCH=$(shell git log --date=iso8601-strict -1 --pretty=%ct)
GOLANGGCI_LINT := golangci-lint
PROTOC_GEN_GO := protoc-gen-go
MOCKGEN := mockgen
PROTOC := $(shell which protoc)
GORELEASER := goreleaser
IMAGE_NAME = scorecard
OUTPUT = output
IGNORED_CI_TEST="E2E TEST:blob|E2E TEST:executable"
PLATFORM="linux/amd64,linux/arm64,linux/386,linux/arm"
LDFLAGS=$(shell ./scripts/version-ldflags)

############################### make help #####################################
.PHONY: help
help:  ## Display this help
	@awk \
		-v "col=${COLOR}" -v "nocol=${NOCOLOR}" \
		' \
			BEGIN { \
				FS = ":.*##" ; \
				printf "Available targets:\n"; \
			} \
			/^[a-zA-Z0-9_-]+:.*?##/ { \
				printf "  %s%-25s%s %s\n", col, $$1, nocol, $$2 \
			} \
			/^##@/ { \
				printf "\n%s%s%s\n", col, substr($$0, 5), nocol \
			} \
		' $(MAKEFILE_LIST)
###############################################################################

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

################################## make all ###################################
all:  ## Runs build, test and verify
all-targets = build check-linter check-osv unit-test validate-docs add-projects validate-projects
.PHONY: all all-targets-update-dependencies $(all-targets) update-dependencies tree-status
all-targets-update-dependencies: $(all-targets) | update-dependencies
all: update-dependencies all-targets-update-dependencies tree-status

update-dependencies: ## Update go dependencies for all modules
	# Update root go modules
	go mod tidy && go mod verify
	cd tools
	go mod tidy && go mod verify

$(GOLANGGCI_LINT): install
check-linter: ## Install and run golang linter
check-linter: $(GOLANGGCI_LINT)
	# Run golangci-lint linter
	golangci-lint run -c .golangci.yml

check-osv: ## Checks osv.dev for any vulnerabilities
check-osv: $(install)
	# Run stunning-tribble for checking the dependencies have any OSV
	go list -m -f '{{if not (or  .Main)}}{{.Path}}@{{.Version}}_{{.Replace}}{{end}}' all \
			| stunning-tribble
	# Checking the tools which also has go.mod
	cd tools 
	go list -m -f '{{if not (or  .Main)}}{{.Path}}@{{.Version}}_{{.Replace}}{{end}}' all \
			| stunning-tribble 

add-projects: ## Adds new projects to ./cron/data/projects.csv
add-projects: ./cron/data/projects.csv | build-add-script
	# Add new projects to ./cron/data/projects.csv
	./cron/data/add/add ./cron/data/projects.csv ./cron/data/projects.new.csv
	mv ./cron/data/projects.new.csv ./cron/data/projects.csv

validate-projects: ## Validates ./cron/data/projects.csv
validate-projects: ./cron/data/projects.csv | build-validate-script
	# Validate ./cron/data/projects.csv
	./cron/data/validate/validate ./cron/data/projects.csv

tree-status: | all-targets-update-dependencies ## Verify tree is clean and all changes are committed
	# Verify the tree is clean and all changes are commited
	./scripts/tree-status

###############################################################################

################################## make build #################################
## Build all cron-related targets
build-cron: build-controller build-worker build-cii-worker \
	build-shuffler build-bq-transfer build-github-server \
	build-webhook build-add-script build-validate-script build-update-script

build-targets = generate-mocks generate-docs build-proto build-scorecard build-releaser build-cron ko-build-everything dockerbuild
.PHONY: build $(build-targets)
build: ## Build all binaries and images in the repo.
build: $(build-targets)

build-proto: ## Compiles and generates all required protobufs
build-proto: cron/data/request.pb.go cron/data/metadata.pb.go
cron/data/request.pb.go: cron/data/request.proto |  $(PROTOC)
	protoc --go_out=../../../ cron/data/request.proto
cron/data/metadata.pb.go: cron/data/metadata.proto |  $(PROTOC)
	protoc --go_out=../../../ cron/data/metadata.proto

generate-mocks: ## Compiles and generates all mocks using mockgen.
generate-mocks: clients/mockclients/repo_client.go clients/mockclients/repo.go clients/mockclients/cii_client.go
clients/mockclients/repo_client.go: clients/repo_client.go
	# Generating MockRepoClient
	$(MOCKGEN) -source=clients/repo_client.go -destination=clients/mockclients/repo_client.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
clients/mockclients/repo.go: clients/repo.go
	# Generating MockRepo
	$(MOCKGEN) -source=clients/repo.go -destination=clients/mockclients/repo.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
clients/mockclients/cii_client.go: clients/cii_client.go
	# Generating MockCIIClient
	$(MOCKGEN) -source=clients/cii_client.go -destination=clients/mockclients/cii_client.go -package=mockrepo -copyright_file=clients/mockclients/license.txt

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
	cd cron/controller && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o controller

build-worker: ## Runs go build on the cron PubSub worker
	# Run go build on the cron PubSub worker
	cd cron/worker && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o worker

build-cii-worker: ## Runs go build on the CII worker
	# Run go build on the CII worker
	cd cron/cii && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o cii-worker

build-shuffler: ## Runs go build on the cron shuffle script
	# Run go build on the cron shuffle script
	cd cron/shuffle && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o shuffle

build-bq-transfer: ## Runs go build on the BQ transfer cron job
build-bq-transfer: ./cron/bq/*.go
	# Run go build on the Copier cron job
	cd cron/bq && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o data-transfer

build-github-server: ## Runs go build on the GitHub auth server
build-github-server: ./clients/githubrepo/roundtripper/tokens/*
	# Run go build on the GitHub auth server
	cd clients/githubrepo/roundtripper/tokens/server && \
		CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o github-auth-server

build-webhook: ## Runs go build on the cron webhook
	# Run go build on the cron webhook
	cd cron/webhook && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o webhook

build-add-script: ## Runs go build on the add script
build-add-script: cron/data/add/add
cron/data/add/add: cron/data/add/*.go cron/data/*.go cron/data/projects.csv
	# Run go build on the add script
	cd cron/data/add && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o add

build-validate-script: ## Runs go build on the validate script
build-validate-script: cron/data/validate/validate
cron/data/validate/validate: cron/data/validate/*.go cron/data/*.go cron/data/projects.csv
	# Run go build on the validate script
	cd cron/data/validate && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o validate

build-update-script: ## Runs go build on the update script
build-update-script: cron/data/update/projects-update
cron/data/update/projects-update:  cron/data/update/*.go cron/data/*.go
	# Run go build on the update script
	cd cron/data/update && CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'  -o projects-update

ko-targets = scorecard-ko cron-controller-ko cron-worker-ko cron-cii-worker-ko cron-bq-transfer-ko cron-webhook-ko cron-github-server-ko
.PHONY: ko-build-everything $(ko-targets)
ko-build-everything: $(ko-targets)

scorecard-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/scorecard CGO_ENABLED=0 LDFLAGS="$(LDFLAGS)" \
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3
cron-controller-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-batch-controller CGO_ENABLED=0 LDFLAGS="$(LDFLAGS)" \
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3/cron/controller
cron-worker-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-batch-worker
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3/cron/worker
cron-cii-worker-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-cii-worker
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3/cron/cii
cron-bq-transfer-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-bq-transfer
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3/cron/bq
cron-webhook-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-cron-webhook
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3/cron/webhook
cron-github-server-ko:
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-github-server
	ko publish -B --bare --local \
			   --platform=$(PLATFORM)\
			   --push=false \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) github.com/ossf/scorecard/v3/clients/githubrepo/roundtripper/tokens/server

docker-targets = scorecard-docker cron-controller-docker cron-worker-docker cron-cii-worker-docker cron-bq-transfer-docker cron-webhook-docker cron-github-server-docker
.PHONY: dockerbuild $(docker-targets)
dockerbuild: $(docker-targets)

scorecard-docker:
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile --tag $(IMAGE_NAME)
cron-controller-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/controller/Dockerfile --tag $(IMAGE_NAME)-batch-controller
cron-worker-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/worker/Dockerfile --tag $(IMAGE_NAME)-batch-worker
cron-cii-worker-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/cii/Dockerfile --tag $(IMAGE_NAME)-cii-worker
cron-bq-transfer-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/bq/Dockerfile --tag $(IMAGE_NAME)-bq-transfer
cron-webhook-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/webhook/Dockerfile --tag ${IMAGE_NAME}-webhook
cron-github-server-docker:
	DOCKER_BUILDKIT=1 docker build . --file clients/githubrepo/roundtripper/tokens/server/Dockerfile --tag ${IMAGE_NAME}-github-server
###############################################################################

################################# make test ###################################
test-targets = unit-test e2e ci-e2e
.PHONY: test $(test-targets)
test: $(test-targets)

unit-test: ## Runs unit test without e2e
	# Run unit tests, ignoring e2e tests
	go test -race -covermode atomic  `go list ./... | grep -v e2e`

e2e: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
e2e: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	$(GINKGO) --race --skip="E2E TEST:executable" -p -v -cover  ./...

$(GINKGO): install

ci-e2e: ## Runs CI e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
ci-e2e: build-scorecard check-env | $(GINKGO)
	# Run CI e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	$(call ndef, GITHUB_AUTH_TOKEN)
	@echo Ignoring these test for ci-e2e $(IGNORED_CI_TEST)
	$(GINKGO) -p  -v -cover --skip=$(IGNORED_CI_TEST)  ./e2e/...


check-env:
ifndef GITHUB_AUTH_TOKEN
	$(error GITHUB_AUTH_TOKEN is undefined)
endif
###############################################################################
