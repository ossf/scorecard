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
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint

KO := $(TOOLS_BIN_DIR)/ko
$(KO): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/google/ko

MOCKGEN := $(TOOLS_BIN_DIR)/mockgen
$(MOCKGEN): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install go.uber.org/mock/mockgen

GINKGO := $(TOOLS_BIN_DIR)/ginkgo
$(GINKGO): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/onsi/ginkgo/v2/ginkgo

GORELEASER := $(TOOLS_BIN_DIR)/goreleaser
$(GORELEASER): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); GOBIN=$(TOOLS_BIN_DIR) go install github.com/goreleaser/goreleaser/v2

# Installs required binaries into $(TOOLS_BIN_DIR) wherever possible.
# Keeping a local copy instead of a global install allows for:
# i) Controlling the binary version Scorecard depends on leading to consistent
# behavior across users.
# ii) Avoids installing a whole bunch of otherwise unnecessary tools in the user's workspace.
.PHONY: install
install: ## Installs required binaries.
install: $(GOLANGCI_LINT) \
	$(KO) \
	$(MOCKGEN) \
	$(GINKGO) \
	$(GORELEASER)

###############################################################################

##@ Build
################################## make all ###################################
all:  ## Runs build, test and verify
all-targets = build unit-test check-linter validate-docs
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

fix-linter: ## Install and run golang linter, with fixes
fix-linter: | $(GOLANGCI_LINT)
	# Run golangci-lint linter
	$(GOLANGCI_LINT) run -c .golangci.yml --fix

tree-status: | all-targets-update-dependencies ## Verify tree is clean and all changes are committed
	# Verify the tree is clean and all changes are committed
	./scripts/tree-status

###############################################################################

################################## make build #################################
build-targets = generate-mocks generate-docs build-scorecard build-attestor
.PHONY: build $(build-targets)
build: ## Build all binaries and images in the repo.
build: $(build-targets)

generate-mocks: ## Compiles and generates all mocks using mockgen.
generate-mocks: clients/mockclients/repo_client.go \
	clients/mockclients/repo.go \
	clients/mockclients/cii_client.go \
	checks/mockclients/vulnerabilities.go \
	cmd/internal/packagemanager/packagemanager_mockclient.go \
	cmd/internal/nuget/nuget_mockclient.go
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
cmd/internal/packagemanager/packagemanager_mockclient.go: cmd/internal/packagemanager/client.go | $(MOCKGEN)
	# Generating MockPackageManagerClient
	$(MOCKGEN) -source=cmd/internal/packagemanager/client.go -destination=cmd/internal/packagemanager/packagemanager_mockclient.go -package=packagemanager -copyright_file=clients/mockclients/license.txt
cmd/internal/nuget/nuget_mockclient.go: cmd/internal/nuget/client.go | $(MOCKGEN)
	# Generating MockNugetClient
	$(MOCKGEN) -source=cmd/internal/nuget/client.go -destination=cmd/internal/nuget/nuget_mockclient.go -package=nuget -copyright_file=clients/mockclients/license.txt

PROBE_DEFINITION_FILES = $(shell find ./probes/ -name "def.yml")
generate-docs: ## Generates docs
generate-docs: validate-docs docs/checks.md docs/checks/internal/checks.yaml docs/checks/internal/*.go docs/checks/internal/generate/*.go \
		docs/probes.md $(PROBE_DEFINITION_FILES) docs/probes/internal/generate/*.go
	# Generating checks.md
	go run ./docs/checks/internal/generate/main.go docs/checks.md
	# Generating probes.md
	go run ./docs/probes/internal/generate/main.go probes/ > docs/probes.md

validate-docs: docs/checks/internal/generate/main.go
	# Validating checks.yaml
	go run ./docs/checks/internal/validate/main.go

setup-probe:
	go run ./probes/internal/scripts/setup.go $(probeName)

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
		--snapshot --clean --skip=publish,sign && \
		touch scorecard.releaser

build-attestor: ## Runs go build on scorecard attestor
	# Run go build on scorecard attestor
	cd attestor/; CGO_ENABLED=0 go build -trimpath -a -tags netgo -ldflags '$(LDFLAGS)' -o scorecard-attestor


build-attestor-docker: ## Build scorecard-attestor Docker image
build-attestor-docker:
	DOCKER_BUILDKIT=1 docker build . --file attestor/Dockerfile \
		--tag scorecard-attestor:latest \
		--tag scorecard-attestor:$(GIT_HASH)

docker-targets = scorecard-docker
.PHONY: dockerbuild $(docker-targets)
dockerbuild: $(docker-targets)

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

check-env-gitlab:
ifndef GITLAB_AUTH_TOKEN
	$(error GITLAB_AUTH_TOKEN is undefined)
endif

check-env-azure-devops:
ifndef AZURE_DEVOPS_AUTH_TOKEN
	$(error AZURE_DEVOPS_AUTH_TOKEN is undefined)
endif

e2e-pat: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
e2e-pat: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	TOKEN_TYPE="PAT" $(GINKGO) --race -p -v --flake-attempts=3 -coverprofile=e2e-coverage.out -coverpkg=./... -r ./...

e2e-gh-token: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to default GITHUB_TOKEN
e2e-gh-token: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN set to secrets.GITHUB_TOKEN must be used to run this.
	GITLAB_AUTH_TOKEN="" TOKEN_TYPE="GITHUB_TOKEN" $(GINKGO) --race -p -v --flake-attempts=3 -coverprofile=e2e-coverage.out --keep-separate-coverprofiles ./...

e2e-gitlab-token: ## Runs e2e tests that require a GITLAB_TOKEN
e2e-gitlab-token: build-scorecard check-env-gitlab | $(GINKGO)
	TEST_GITLAB_EXTERNAL=1 TOKEN_TYPE="GITLAB_PAT" $(GINKGO) --race -p -vv --flake-attempts=3 -coverprofile=e2e-coverage.out --keep-separate-coverprofiles --focus '.*GitLab' ./...

e2e-gitlab: ## Runs e2e tests for GitLab only. TOKEN_TYPE is not used (since these are public APIs), but must be set to something
e2e-gitlab: build-scorecard | $(GINKGO)
	TEST_GITLAB_EXTERNAL=1 TOKEN_TYPE="PAT" $(GINKGO) --race -p -vv --flake-attempts=3 -coverprofile=e2e-coverage.out --keep-separate-coverprofiles --focus ".*GitLab" ./...

e2e-azure-devops-token: ## Runs e2e tests that require a AZURE_DEVOPS_AUTH_TOKEN
e2e-azure-devops-token: build-scorecard check-env-azure-devops | $(GINKGO)
	SCORECARD_EXPERIMENTAL=1 TEST_AZURE_DEVOPS_EXTERNAL=1 TOKEN_TYPE="AZURE_DEVOPS_PAT" $(GINKGO) --race -p -vv --flake-attempts=3 -coverprofile=e2e-coverage.out --keep-separate-coverprofiles --focus "Azure DevOps" ./...

e2e-attestor: ## Runs e2e tests for scorecard-attestor
	cd attestor/e2e; go test -covermode=atomic -coverprofile=e2e-coverage.out; cd ../..

###############################################################################

##@ TODO(#744)
################################## make ko-images #############################
ko-targets = scorecard-ko
.PHONY: ko-images $(ko-targets)
ko-images: $(ko-targets)

KOCACHE_PATH=/tmp/ko

$(KOCACHE_PATH):
	mkdir -p $(KOCACHE_PATH)

scorecard-ko: | $(KO) $(KOCACHE_PATH)
	KO_DATA_DATE_EPOCH=$(SOURCE_DATE_EPOCH) \
			   KO_DOCKER_REPO=ghcr.io/ossf/scorecard \
			   LDFLAGS="$(LDFLAGS)" \
			   KO_CACHE=$(KOCACHE_PATH) \
			   $(KO) build --bare \
			   --sbom=none \
			   --platform=$(PLATFORM) \
			   --tags latest,$(GIT_VERSION),$(GIT_HASH) \
			   github.com/ossf/scorecard/v5

###############################################################################
