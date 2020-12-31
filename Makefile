all: fmt tidy lint test 
build: 
	go build 

fmt:
	go fmt ./...

# ignoring e2e tests
test: 
	go test -covermode atomic -coverprofile=profile.out `go list ./... | grep -v e2e`

tidy:
	go mod tidy

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) v1.29.0 ;\
	}

lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run 

.PHONY: e2e
# export GITHUB_AUTH_TOKEN with  personal access token to run the e2e
e2e:
	ginkgo test -v -p ./e2e/...

