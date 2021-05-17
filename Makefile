SHELL := /bin/bash
GOBIN ?= $(GOPATH)/bin
GINKGO := $(GOBIN)/ginkgo
GOLANGGCI_LINT := $(GOBIN)/golangci-lint
PROTOC_GEN_GO := $(GOBIN)/protoc-gen-go
PROTOC := $(shell which protoc)
IMAGE_NAME = scorecard
OUTPUT = output
FOCUS_DISK_TEST="E2E TEST:Disk Cache|E2E TEST:executable"
IGNORED_CI_TEST="E2E TEST:blob|E2E TEST:Disk Cache|E2E TEST:executable"

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
install: | $(GINKGO) $(GOLANGGCI_LINT) $(PROTOC_GEN_GO) $(PROTOC)

$(GINKGO):
	go get -u github.com/onsi/ginkgo/ginkgo@v1.16.2

$(GOLANGGCI_LINT):
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.0

$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go

$(PROTOC):
	ifeq (,$(PROTOC))
		$(error download and install protobuf compiler package - https://developers.google.com/protocol-buffers/docs/downloads)
	endif
###############################################################################

################################## make all ###################################
all:  ## Runs build, test and verify
all-targets = update-dependencies build check-linter unit-test validate-projects tree-status
.PHONY: all $(all-targets)
all: $(all-targets)

update-dependencies: ## Update go dependencies for all modules
	# Update root go modules
	go mod tidy && go mod verify
	# Update ./scripts/ go modules
	cd scripts && go mod tidy && go mod verify
	# Update ./scripts/update go modules
	cd ./scripts/update && go mod tidy && go mod verify

check-linter: ## Install and run golang linter
check-linter: | $(GOLANGGCI_LINT)
	# Run golangci-lint linter
	golangci-lint run -n

validate-projects: ## Validates ./cron/projects.txt
validate-projects: build-scripts
	# Validate ./cron/projects.txt
	./scripts/validate ./cron/projects.txt

tree-status: ## Verify tree is clean and all changes are committed
	# Verify the tree is clean and all changes are commited
	./scripts/tree-status
###############################################################################

############################### make build ################################
build-targets = build-proto generate-docs build-scorecard build-cron build-scripts build-update dockerbuild
.PHONY: build $(build-targets)
build: ## Build all binaries and images in the reepo.
build: $(build-targets)

build-proto: ## Compiles and generates all required protobufs
build-proto: cron/data/request.pb.go
cron/data/request.pb.go: cron/data/request.proto | $(PROTOC_GEN_GO) $(PROTOC)
	protoc --go_out=../../../ cron/data/request.proto

generate-docs: ## Generates docs
generate-docs: checks/checks.md
checks/checks.md: checks/checks.yaml checks/main/*.go
	# Generating checks.md
	cd ./checks/main && go run main.go

build-scorecard: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'

build-cron: ## Runs go build on the cron job
	# Run go build on the cronjob
	cd cron && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o scorecardcron

build-scripts: ## Runs go build on the scripts
build-scripts: scripts/validate
scripts/validate: scripts/*.go
	# Run go build on the scripts
	cd scripts && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o validate

build-update: ## Runs go build on scripts/update
build-update: scripts/update/projects-update
scripts/update/projects-update: scripts/update/*.go
	# Run go build on projects-update
	cd scripts/update && CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'  -o projects-update

dockerbuild: ## Runs docker build
	# Build all Docker images in the Repo
	$(call ndef, GITHUB_AUTH_TOKEN)
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile --tag $(IMAGE_NAME)
	DOCKER_BUILDKIT=1 docker build . --file cron/Dockerfile --tag $(IMAGE_NAME)cron
###############################################################################

################################# make test ###################################
test-targets = unit-test e2e ci-e2e test-disk-cache e2e-cron
.PHONY: test $(test-targets)
test: $(test-targets)

unit-test: ## Runs unit test without e2e
	# Run unit tests, ignoring e2e tests
	go test -covermode atomic  `go list ./... | grep -v e2e`

e2e: ## Runs e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
e2e: build-scorecard check-env | $(GINKGO)
	# Run e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	ginkgo --skip="E2E TEST:executable" -p -v -cover  ./...

ci-e2e: ## Runs CI e2e tests. Requires GITHUB_AUTH_TOKEN env var to be set to GitHub personal access token
ci-e2e: build-scorecard check-env e2e-cron | $(GINKGO)
	# Run CI e2e tests. GITHUB_AUTH_TOKEN with personal access token must be exported to run this
	$(call ndef, GITHUB_AUTH_TOKEN)
	@echo Ignoring these test for ci-e2e $(IGNORED_CI_TEST)
	ginkgo -p  -v -cover --skip=$(IGNORED_CI_TEST)  ./e2e/...

test-disk-cache: ## Runs disk cache tests
test-disk-cache: build-scorecard | $(GINKGO)
	# Runs disk cache tests
	$(call ndef, GITHUB_AUTH_TOKEN)
	# Start with clean cache
	rm -rf $(OUTPUT)
	rm -rf cache
	mkdir $(OUTPUT)
	mkdir cache
	@echo Focusing on these tests $(FOCUS_DISK_TEST)
	USE_DISK_CACHE=1 DISK_CACHE_PATH="./cache" \
				   ./scorecard \
				   --repo=https://github.com/ossf/scorecard \
				   --show-details --metadata=openssf  --format json > ./$(OUTPUT)/results.json
	USE_DISK_CACHE=1 DISK_CACHE_PATH="./cache" ginkgo -p  -v -cover --focus=$(FOCUS_DISK_TEST)  ./e2e/...
	# Rerun the same test with the disk cache filled to make sure the cache is working.
	USE_DISK_CACHE=1 DISK_CACHE_PATH="./cache" \
				   ./scorecard \
				   --repo=https://github.com/ossf/scorecard --show-details \
				   --metadata=openssf  --format json > ./$(OUTPUT)/results.json
	USE_DISK_CACHE=1 DISK_CACHE_PATH="./cache" ginkgo -p  -v -cover --focus=$(FOCUS_DISK_TEST)  ./e2e/...

e2e-cron: ## Runs a e2e test cron job and validates its functionality
	# Validate cron
	GCS_BUCKET=ossf-scorecards-dev go run ./cron/main.go ./e2e/cron-projects.txt

check-env:
ifndef GITHUB_AUTH_TOKEN
	$(error GITHUB_AUTH_TOKEN is undefined)
endif
###############################################################################
