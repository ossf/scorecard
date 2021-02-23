SHELL := /bin/bash
GOBIN ?= $(GOPATH)/bin
GINKGO ?= $(GOBIN)/ginkgo

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
all: build test verify 

.PHONY: build
build: ## Runs go build and generates executable
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'

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
	$(GOLANGCI_LINT) run  -n

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
	GO111MODULE=off go get -u github.com/onsi/ginkgo/ginkgo

ci-e2e:  ## Runs ci e2e tests
.PHONY: ci-e2e
# export GITHUB_AUTH_TOKEN with personal access token to run the e2e
ci-e2e: build check-env
	$(call ndef, GITHUB_AUTH_TOKEN)
	mkdir -p bin
	mkdir -p cache
	USE_DISK_CACHE=1 DISK_CACHE_PATH="./cache" ./scorecard --repo=https://github.com/ossf/scorecard --show-details --metadata=openssf  --format json > ./bin/results.json
	@sleep 30
	ginkgo -p  -v -cover --skip="E2E TEST:blob"  ./...


# Verification targets
.PHONY: verify
verify: verify-go-mod lint ## Run all verification targets

.PHONY: verify-go-mod
verify-go-mod: ## Verify the go modules
	export GO111MODULE=on && \
		go mod tidy && \
		go mod verify
	./hack/tree-status 
