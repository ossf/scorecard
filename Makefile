SHELL := /bin/bash
GOPATH := $(go env GOPATH)
GINKGO := ginkgo
GOLANGGCI_LINT := golangci-lint
PROTOC_GEN_GO := protoc-gen-go
PROTOC := $(shell which protoc)
IMAGE_NAME = scorecard
OUTPUT = output
IGNORED_CI_TEST="E2E TEST:blob|E2E TEST:executable"

# generate VERSION_LDFLAGS
GIT_VERSION ?= $(shell git describe --tags --always --dirty)
GIT_HASH ?= $(shell git rev-parse HEAD)
DATE_FMT = +'%Y-%m-%dT%H:%M:%SZ'
SOURCE_DATE_EPOCH ?= $(shell git log -1 --pretty=%ct)
ifdef SOURCE_DATE_EPOCH
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u "$(DATE_FMT)")
else
    BUILD_DATE ?= $(shell date "$(DATE_FMT)")
endif
GIT_TREESTATE = clean
DIFF = $(shell git diff --quiet >/dev/null 2>&1; if [ $$? -eq 1 ]; then echo "1"; fi)
ifeq ($(DIFF), 1)
    GIT_TREESTATE = dirty
endif

# version should be injected in the cmd package
PKG=$(shell go list -m | head -n1)/cmd

VERSION_LDFLAGS=-X $(PKG).gitVersion=$(GIT_VERSION) -X $(PKG).gitCommit=$(GIT_HASH) -X $(PKG).gitTreeState=$(GIT_TREESTATE) -X $(PKG).buildDate=$(BUILD_DATE)

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
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

$(PROTOC):
	ifeq (,$(PROTOC))
		$(error download and install protobuf compiler package - https://developers.google.com/protocol-buffers/docs/downloads)
	endif
###############################################################################

################################## make all ###################################
all:  ## Runs build, test and verify
all-targets = update-dependencies build check-linter unit-test add-projects validate-projects tree-status
.PHONY: all $(all-targets)
all: $(all-targets)

update-dependencies: ## Update go dependencies for all modules
	# Update root go modules
	go mod tidy && go mod verify

$(GOLANGGCI_LINT): install
check-linter: ## Install and run golang linter
check-linter: $(GOLANGGCI_LINT)
	# Run golangci-lint linter
	golangci-lint run -c .golangci.yml

add-projects: ## Adds new projects to ./cron/data/projects.csv
add-projects: ./cron/data/projects.csv | build-add-script
	# Add new projects to ./cron/data/projects.csv
	./cron/data/add/add ./cron/data/projects.csv ./cron/data/projects.new.csv
	mv ./cron/data/projects.new.csv ./cron/data/projects.csv

validate-projects: ## Validates ./cron/data/projects.csv
validate-projects: ./cron/data/projects.csv | build-validate-script
	# Validate ./cron/data/projects.csv
	./cron/data/validate/validate ./cron/data/projects.csv

tree-status: ## Verify tree is clean and all changes are committed
	# Verify the tree is clean and all changes are commited
	./scripts/tree-status


###############################################################################

################################## make build #################################
build-targets = build-proto generate-docs build-scorecard build-pubsub build-bq-transfer \
	build-add-script build-validate-script build-update-script dockerbuild
.PHONY: build $(build-targets)
build: ## Build all binaries and images in the reepo.
build: $(build-targets)

build-proto: ## Compiles and generates all required protobufs
build-proto: cron/data/request.pb.go
cron/data/request.pb.go: cron/data/request.proto |  $(PROTOC)
	protoc --go_out=../../../ cron/data/request.proto

generate-docs: ## Generates docs
generate-docs: checks/checks.md
checks/checks.md: checks/checks.yaml checks/main/*.go
	# Generating checks.md
	cd ./checks/main && go run main.go

build-scorecard: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static" $(VERSION_LDFLAGS)'

build-pubsub: ## Runs go build on the PubSub cron job
	# Run go build and the PubSub cron job
	cd cron/controller && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "static"' -o controller
	cd cron/worker && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "static"' -o worker

build-bq-transfer: ## Runs go build on the BQ transfer cron job
build-bq-transfer: ./cron/bq/*.go
	# Run go build on the Copier cron job
	cd cron/bq && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "static"' -o data-transfer

build-add-script: ## Runs go build on the add script
build-add-script: cron/data/add/add
cron/data/add/add: cron/data/add/*.go cron/data/*.go cron/data/projects.csv
	# Run go build on the add script
	cd cron/data/add && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o add

build-validate-script: ## Runs go build on the validate script
build-validate-script: cron/data/validate/validate
cron/data/validate/validate: cron/data/validate/*.go cron/data/*.go cron/data/projects.csv
	# Run go build on the validate script
	cd cron/data/validate && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o validate

build-update-script: ## Runs go build on the update script
build-update-script: cron/data/update/projects-update
cron/data/update/projects-update:  cron/data/update/*.go cron/data/*.go
	# Run go build on the update script
	cd cron/data/update && CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'  -o projects-update

dockerbuild: ## Runs docker build
	# Build all Docker images in the Repo
	$(call ndef, GITHUB_AUTH_TOKEN)
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile --tag $(IMAGE_NAME)
	DOCKER_BUILDKIT=1 docker build . --file cron/controller/Dockerfile --tag $(IMAGE_NAME)-batch-controller
	DOCKER_BUILDKIT=1 docker build . --file cron/worker/Dockerfile --tag $(IMAGE_NAME)-batch-worker
	DOCKER_BUILDKIT=1 docker build . --file cron/bq/Dockerfile --tag $(IMAGE_NAME)-bq-transfer
###############################################################################

################################# make test ###################################
test-targets = unit-test e2e ci-e2e 
.PHONY: test $(test-targets)
test: $(test-targets)

unit-test: ## Runs unit test without e2e
	# Run unit tests, ignoring e2e tests
	go test -covermode atomic  `go list ./... | grep -v e2e`

e2e: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
e2e: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	$(GINKGO) --skip="E2E TEST:executable" -p -v -cover  ./...

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
