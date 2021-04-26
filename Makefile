SHELL := /bin/bash
GOBIN ?= $(GOPATH)/bin
GINKGO ?= $(GOBIN)/ginkgo
IMAGE_NAME = scorecard
OUTPUT = output
FOCUS_DISK_TEST="E2E TEST:Disk Cache|E2E TEST:executable"
IGNORED_CI_TEST="E2E TEST:blob|E2E TEST:Disk Cache|E2E TEST:executable"
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

all:  ## Runs build, test and verify
.PHONY: all
all: build test check-projects build-cron build-scripts verify  projects-update

.PHONY: build
build: ## Runs go build and generates executable
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'

.PHONY: build-cron
build-cron: ## Runs go build on the cronjob
	cd cron && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o scorecardcron


.PHONY: build-scripts
build-scripts: ## Runs go build on the scripts
	cd scripts && CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o validate

.PHONY: test
test: ## Runs unit test
	# ignoring e2e tests
	go test -covermode atomic  `go list ./... | grep -v e2e`

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci-lint:
	rm -f $(GOLANGCI_LINT) || :
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) v1.36.0 ;\

lint: golangci-lint ## Runs golangci-lint linter
	$(GOLANGCI_LINT) run -n 

check-env:
ifndef GITHUB_AUTH_TOKEN
	$(error GITHUB_AUTH_TOKEN is undefined)
endif

e2e:  ## Runs e2e tests
.PHONY: e2e
# export GITHUB_AUTH_TOKEN with personal access token to run the e2e
e2e: build check-env ginkgo
	$(GINKGO) --skip="E2E TEST:executable" -p -v -cover  ./...

ginkgo:
	go get -u github.com/onsi/ginkgo/ginkgo

unexport USE_DISK_CACHE
unexport USE_BLOB_CACHE
ci-e2e:  ## Runs ci e2e tests
.PHONY: ci-e2e
# export GITHUB_AUTH_TOKEN with personal access token to run the e2e
ci-e2e: build check-env
	$(call ndef, GITHUB_AUTH_TOKEN)
	@echo Ignoring these test for ci-e2e $(IGNORED_CI_TEST)
	ginkgo -p  -v -cover --skip=$(IGNORED_CI_TEST)  ./e2e/...

.PHONY: test-disk-cache
test-disk-cache: build  ## Runs disk cache tests
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



# Verification targets
.PHONY: verify
verify: verify-go-mod verify-go-mod-cron verify-go-mod-scripts  lint ## Run all verification targets

.PHONY: verify-go-mod
verify-go-mod: ## Verify the go modules
		go mod tidy && \
		go mod verify
	./scripts/tree-status 

verify-go-mod-cron: ## Verify the go modules for cron
	cd cron && \
    go mod tidy && \
	go mod verify 
	./scripts/tree-status

verify-go-mod-scripts: ## Verify the go modules for scripts
	cd scripts && \
    go mod tidy && \
	go mod verify 
	./scripts/tree-status

.PHONY: dockerbuild
dockerbuild: ## Runs docker build
	$(call ndef, GITHUB_AUTH_TOKEN)
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile --tag $(IMAGE_NAME) 
	DOCKER_BUILDKIT=1 docker build . --file Dockerfile.gsutil --tag $(IMAGE_NAME)-gsutil

.PHONY: check-projects
check-projects: ## Validates ./cron/projects.txt
	cd ./scripts && go build  -o validate
	./scripts/validate ./cron/projects.txt

.PHONY: projects-update
projects-update: ## builds the scripts/update binary
	cd ./scripts/update && go mod tidy && go mod verify	
	cd ./scripts/update && CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'  -o projects-update .
	./scripts/tree-status
