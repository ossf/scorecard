SHELL := /bin/bash
GIT_HASH := $(shell git rev-parse HEAD)
GIT_VERSION ?= $(shell git describe --tags --always --dirty)
SOURCE_DATE_EPOCH=$(shell git log --date=iso8601-strict -1 --pretty=%ct)
IMAGE_NAME = scorecard
OUTPUT = output
PLATFORM="linux/amd64,linux/arm64,linux/386,linux/arm"
LDFLAGS=$(shell ./scripts/version-ldflags)



############################### make help #####################################
.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; \
			printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ \
			{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
			/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

###############################################################################

##@ Tools
################################ make install #################################
TOOLS_DIR := tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)
GOBIN := $(shell go env GOBIN)

# Golang binaries.

GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/golangci/golangci-lint/cmd/golangci-lint

KO := $(TOOLS_BIN_DIR)/ko
$(KO): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/google/ko

STUNNING_TRIBBLE := $(TOOLS_BIN_DIR)/stunning-tribble
$(STUNNING_TRIBBLE): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/naveensrinivasan/stunning-tribble

MOCKGEN := $(TOOLS_BIN_DIR)/mockgen
$(MOCKGEN): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/golang/mock/mockgen

GINKGO := $(TOOLS_BIN_DIR)/ginkgo
$(GINKGO): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/onsi/ginkgo/v2/ginkgo

GORELEASER := $(TOOLS_BIN_DIR)/goreleaser
$(GORELEASER): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/goreleaser/goreleaser

PROTOC_GEN_GO := $(TOOLS_BIN_DIR)/protoc-gen-go
$(PROTOC_GEN_GO): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go

# Non-Golang binaries.
# TODO: Figure out how to install these binaries automatically.

PROTOC := $(shell which protoc)
$(PROTOC):
	ifeq (,$(PROTOC))
		$(error download and install protobuf compiler package - https://developers.google.com/protocol-buffers/docs/downloads)
	endif

# Installs required binaries into $(TOOLS_BIN_DIR) wherever possible.
# Keeping a local copy instead of a global install allows for:
# i) Controlling the binary version Scorecard depends on leading to consistent
# behavior across users.
# ii) Avoids installing a whole bunch of otherwise unnecessary tools in the user's workspace.
.PHONY: install
install: ## Installs required binaries.
install: $(GOLANGCI_LINT) \
	$(KO) \
	$(STUNNING_TRIBBLE) \
	$(PROTOC_GEN_GO) $(PROTOC) \
	$(MOCKGEN) \
	$(GINKGO) \
	$(GORELEASER)

###############################################################################

##@ Build
################################## make all ###################################
all:  ## Runs build, test and verify
all-targets = build unit-test check-linter validate-docs add-projects validate-projects
.PHONY: all all-targets-update-dependencies $(all-targets) update-dependencies tree-status
all-targets-update-dependencies: $(all-targets) | update-dependencies
all: update-dependencies all-targets-update-dependencies tree-status

update-dependencies: ## Update go dependencies for all modules
	# Update root go modules
	go mod tidy && go mod verify
	cd tools; go mod tidy && go mod verify; cd ../

check-linter: ## Install and run golang linter
check-linter: | $(GOLANGCI_LINT)
	# Run golangci-lint linter
	$(GOLANGCI_LINT) run -c .golangci.yml

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

build-targets = generate-mocks generate-docs build-scorecard build-cron build-proto build-attestor
.PHONY: build $(build-targets)
build: ## Build all binaries and images in the repo.
build: $(build-targets)

build-proto: ## Compiles and generates all required protobufs
build-proto: cron/data/request.pb.go cron/data/metadata.pb.go
cron/data/request.pb.go: cron/data/request.proto | $(PROTOC) $(PROTOC_GEN_GO)
	$(PROTOC) --plugin=$(PROTOC_GEN_GO) --go_out=. --go_opt=paths=source_relative cron/data/request.proto
cron/data/metadata.pb.go: cron/data/metadata.proto | $(PROTOC) $(PROTOC_GEN_GO)
	$(PROTOC) --plugin=$(PROTOC_GEN_GO) --go_out=. --go_opt=paths=source_relative cron/data/metadata.proto

generate-mocks: ## Compiles and generates all mocks using mockgen.
generate-mocks: clients/mockclients/repo_client.go \
	clients/mockclients/repo.go \
	clients/mockclients/cii_client.go \
	checks/mockclients/vulnerabilities.go \
	cmd/packagemanager_mockclient.go
clients/mockclients/repo_client.go: clients/repo_client.go | $(MOCKGEN)
	# Generating MockRepoClient
	$(MOCKGEN) -source=clients/repo_client.go -destination=clients/mockclients/repo_client.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
clients/mockclients/repo.go: clients/repo.go | $(MOCKGEN)
	# Generating MockRepo
	$(MOCKGEN) -source=clients/repo.go -destination=clients/mockclients/repo.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
clients/mockclients/cii_client.go: clients/cii_client.go | $(MOCKGEN)
	# Generating MockCIIClient
	$(MOCKGEN) -source=clients/cii_client.go -destination=clients/mockclients/cii_client.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
checks/mockclients/vulnerabilities.go: clients/vulnerabilities.go | $(MOCKGEN)
	# Generating MockCIIClient
	$(MOCKGEN) -source=clients/vulnerabilities.go -destination=clients/mockclients/vulnerabilities.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
cmd/packagemanager_mockclient.go: cmd/packagemanager_client.go | $(MOCKGEN)
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

SCORECARD_DEPS = $(shell find . -iname "*.go" | grep -v tools/)
build-scorecard: ## Build Scorecard CLI
build-scorecard: scorecard
scorecard: $(SCORECARD_DEPS)
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'
scorecard-docker: ## Build Scorecard CLI Docker image
scorecard-docker: scorecard.docker
scorecard.docker: Dockerfile $(SCORECARD_DEPS)
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile \
			--tag $(IMAGE_NAME) && \
			touch scorecard.docker

build-releaser: ## Build goreleaser for the Scorecard CLI
build-releaser: scorecard.releaser
scorecard.releaser: .goreleaser.yml $(SCORECARD_DEPS) | $(GORELEASER)
	# Run go releaser on the Scorecard repo
	$(GORELEASER) check && \
		VERSION_LDFLAGS="$(LDFLAGS)" $(GORELEASER) release \
		--snapshot --rm-dist --skip-publish --skip-sign && \
		touch scorecard.releaser

CRON_CONTROLLER_DEPS = $(shell find cron/internal/ -iname "*.go")
build-controller: ## Build cron controller
build-controller: cron/internal/controller/controller
cron/internal/controller/controller: $(CRON_CONTROLLER_DEPS)
	# Run go build on the cron PubSub controller
	cd cron/internal/controller && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o controller
cron-controller-docker: ## Build cron controller Docker image
cron-controller-docker: cron/internal/controller/controller.docker
cron/internal/controller/controller.docker: cron/internal/controller/Dockerfile $(CRON_CONTROLLER_DEPS)
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/controller/Dockerfile \
			--tag $(IMAGE_NAME)-batch-controller \
			&& touch cron/internal/controller/controller.docker

build-worker: ## Runs go build on the cron PubSub worker
	# Run go build on the cron PubSub worker
	cd cron/internal/worker && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o worker

CRON_CII_DEPS = $(shell find cron/internal/ clients/ -iname "*.go")
build-cii-worker: ## Build cron CII worker
build-cii-worker: cron/internal/cii/cii-worker
cron/internal/cii/cii-worker: $(CRON_CII_DEPS)
	# Run go build on the CII worker
	cd cron/internal/cii && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o cii-worker
cron-cii-worker-docker: # Build cron CII worker Docker image
cron-cii-worker-docker: cron/internal/cii/cii-worker.docker
cron/internal/cii/cii-worker.docker: cron/internal/cii/Dockerfile $(CRON_CII_DEPS)
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/cii/Dockerfile \
			--tag $(IMAGE_NAME)-cii-worker && \
			touch cron/internal/cii/cii-worker.docker

CRON_SHUFFLER_DEPS = $(shell find cron/data/ cron/internal/shuffle/ -iname "*.go")
build-shuffler: ## Build cron shuffle script
build-shuffler: cron/internal/shuffle/shuffle
cron/internal/shuffle/shuffle: $(CRON_SHUFFLER_DEPS)
	# Run go build on the cron shuffle script
	cd cron/internal/shuffle && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o shuffle

CRON_TRANSFER_DEPS = $(shell find cron/data/ cron/config/ cron/internal/bq/ -iname "*.go")
build-bq-transfer: ## Build cron BQ transfer worker
build-bq-transfer: cron/internal/bq/data-transfer
cron/internal/bq/data-transfer: $(CRON_TRANSFER_DEPS)
	# Run go build on the Copier cron job
	cd cron/internal/bq && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o data-transfer
cron-bq-transfer-docker: ## Build cron BQ transfer worker Docker image
cron-bq-transfer-docker: cron/internal/bq/data-transfer.docker
cron/internal/bq/data-transfer.docker: cron/internal/bq/Dockerfile $(CRON_TRANSFER_DEPS)
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/bq/Dockerfile \
			--tag $(IMAGE_NAME)-bq-transfer && \
			touch cron/internal/bq/data-transfer.docker

build-attestor: ## Runs go build on scorecard attestor
	# Run go build on scorecard attestor
	cd attestor/; CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)' -o scorecard-attestor


build-attestor-docker: ## Build scorecard-attestor Docker image
build-attestor-docker:
	DOCKER_BUILDKIT=1 docker build . --file attestor/Dockerfile \
		--tag scorecard-attestor:latest \
		--tag scorecard-atttestor:$(GIT_HASH)

TOKEN_SERVER_DEPS = $(shell find clients/githubrepo/roundtripper/tokens/ -iname "*.go")
build-github-server: ## Build GitHub token server
build-github-server: clients/githubrepo/roundtripper/tokens/server/github-auth-server
clients/githubrepo/roundtripper/tokens/server/github-auth-server: $(TOKEN_SERVER_DEPS)
	# Run go build on the GitHub auth server
	cd clients/githubrepo/roundtripper/tokens/server && \
		CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o github-auth-server
cron-github-server-docker: ## Build GitHub token server Docker image
cron-github-server-docker: clients/githubrepo/roundtripper/tokens/server/github-auth-server.docker
clients/githubrepo/roundtripper/tokens/server/github-auth-server.docker: \
	clients/githubrepo/roundtripper/tokens/server/Dockerfile $(TOKEN_SERVER_DEPS)
	DOCKER_BUILDKIT=1 docker build . \
			--file clients/githubrepo/roundtripper/tokens/server/Dockerfile \
			--tag ${IMAGE_NAME}-github-server && \
			touch clients/githubrepo/roundtripper/tokens/server/github-auth-server.docker

CRON_WEBHOOK_DEPS = $(shell find cron/internal/webhook/ cron/data/ -iname "*.go")
build-webhook: ## Build cron webhook server
build-webhook: cron/internal/webhook/webhook
cron/internal/webhook/webhook: $(CRON_WEBHOOK_DEPS)
	# Run go build on the cron webhook
	cd cron/internal/webhook && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o webhook
cron-webhook-docker: ## Build cron webhook server Docker image
cron-webhook-docker: cron/internal/webhook/webhook.docker
cron/internal/webhook/webhook.docker: cron/internal/webhook/Dockerfile $(CRON_WEBHOOK_DEPS)
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/webhook/Dockerfile \
			--tag ${IMAGE_NAME}-webhook && \
			touch cron/internal/webhook/webhook.docker


build-add-script: ## Runs go build on the add script
build-add-script: cron/internal/data/add/add
cron/internal/data/add/add: cron/internal/data/add/*.go cron/data/*.go cron/internal/data/projects.csv
	# Run go build on the add script
	cd cron/internal/data/add && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o add

build-validate-script: ## Runs go build on the validate script
build-validate-script: cron/internal/data/validate/validate
cron/internal/data/validate/validate: cron/internal/data/validate/*.go cron/data/*.go cron/internal/data/projects.csv
	# Run go build on the validate script
	cd cron/internal/data/validate && CGO_ENABLED=0 go build -trimpath -a -ldflags '$(LDFLAGS)' -o validate

build-update-script: ## Runs go build on the update script
build-update-script: cron/internal/data/update/projects-update
cron/internal/data/update/projects-update:  cron/internal/data/update/*.go cron/data/*.go
	# Run go build on the update script
	cd cron/internal/data/update && CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'  -o projects-update

docker-targets = scorecard-docker cron-controller-docker cron-worker-docker cron-cii-worker-docker cron-bq-transfer-docker cron-webhook-docker cron-github-server-docker
.PHONY: dockerbuild $(docker-targets)
dockerbuild: $(docker-targets)

cron-worker-docker:
	DOCKER_BUILDKIT=1 docker build . --file cron/internal/worker/Dockerfile --tag $(IMAGE_NAME)-batch-worker
###############################################################################

##@ Tests
################################# make test ###################################
test-targets = unit-test e2e-pat e2e-gh-token ci-e2e
.PHONY: test $(test-targets)
test: $(test-targets)

unit-test: ## Runs unit test without e2e
	# Run unit tests, ignoring e2e tests
	# run the go tests and gen the file coverage-all used to do the integration with codecov
	SKIP_GINKGO=1 go test -race -covermode=atomic  -coverprofile=unit-coverage.out -coverpkg=./... `go list ./...`

unit-test-attestor: ## Runs unit tests on scorecard-attestor
	cd attestor; SKIP_GINKGO=1 go test -covermode=atomic -coverprofile=unit-coverage.out `go list ./...`; cd ..;

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

e2e-attestor: ## Runs e2e tests for scorecard-attestor
	cd attestor/e2e; go test -covermode=atomic -coverprofile=e2e-coverage.out; cd ../..

###############################################################################

##@ TODO(#744)
################################## make ko-images #############################
ko-targets = scorecard-ko cron-controller-ko cron-worker-ko cron-cii-worker-ko cron-bq-transfer-ko cron-webhook-ko cron-github-server-ko
.PHONY: ko-images $(ko-targets)
ko-images: $(ko-targets)

KOCACHE_PATH=/tmp/ko

$(KOCACHE_PATH):
	mkdir -p $(KOCACHE_PATH)

scorecard-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/${IMAGE_NAME}
			   LDFLAGS="$(LDFLAGS)" \
			   KO_CACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4

cron-controller-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-batch-controller \
			   LDFLAGS="$(LDFLAGS)" \
			   KOCACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4/cron/internal/controller

cron-worker-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-batch-worker \
			   LDFLAGS="$(LDFLAGS)" \
			   KOCACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4/cron/internal/worker

cron-cii-worker-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-cii-worker \
			   LDFLAGS="$(LDFLAGS)" \
			   KOCACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM)\
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4/cron/internal/cii

cron-bq-transfer-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-bq-transfer \
			   LDFLAGS="$(LDFLAGS)" \
			   KOCACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4/cron/internal/bq

cron-webhook-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-cron-webhook \
			   LDFLAGS="$(LDFLAGS)" \
			   KOCACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4/cron/internal/webhook

cron-github-server-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=${KO_PREFIX}/$(IMAGE_NAME)-github-server \
			   LDFLAGS="$(LDFLAGS)" \
			   KOCACHE=$(KOCACHE_PATH) \
			   $(KO) build -B \
			   --push=false \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper/tokens/server

###############################################################################
