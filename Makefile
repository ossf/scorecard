SHELL := /bin/bash
GOBIN ?= $(GOPATH)/bin
GINKGO ?= $(GOBIN)/ginkgo
all: fmt tidy lint test
build: 
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'

fmt:
	go fmt ./...

# ignoring e2e tests
test: 
	go test -covermode atomic  `go list ./... | grep -v e2e`

tidy:
	go mod tidy

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci-lint:
	rm -f $(GOLANGCI_LINT) || :
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) v1.36.0 ;\

lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run  -n

check-env:
ifndef GITHUB_AUTH_TOKEN
	$(error GITHUB_AUTH_TOKEN is undefined)
endif

.PHONY: e2e
# export GITHUB_AUTH_TOKEN with personal access token to run the e2e
e2e: build check-env ginkgo
	$(GINKGO) --skip="E2E TEST:executable" -p -v -cover  ./...

ginkgo:
	GO111MODULE=off go get -u github.com/onsi/ginkgo/ginkgo
.PHONY: ci-e2e
# export GITHUB_AUTH_TOKEN with personal access token to run the e2e
ci-e2e: build check-env
	$(call ndef, GITHUB_AUTH_TOKEN)
	mkdir -p bin
	./scorecard --repo=https://github.com/ossf/scorecard --format json > ./bin/results.json
	ginkgo -p  -v -cover  ./...

