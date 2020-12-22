build: 
	go build 

fmt:
	go fmt ./...

test: 
	go test -covermode atomic -coverprofile=profile.out ./...

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

