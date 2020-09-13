# We need to export GOBIN to allow it to be set
# for processes spawned from the Makefile
export GOBIN ?= $(GOPATH)/bin
GO=go

store-mocks: ## Creates mock files.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir store -all -output store/storetest/mocks -note 'Regenerate this file using `make store-mocks`.'

test-run-coverage:
	go test -coverprofile=coverage_result ./...
	go tool cover -html=coverage_result
