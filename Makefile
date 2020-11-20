# We need to export GOBIN to allow it to be set
# for processes spawned from the Makefile
export GOBIN ?= $(GOPATH)/bin
GO=go

store-mocks: ## Creates mock files.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir store -all -output store/storetest/mocks -note 'Regenerate this file using `make store-mocks`.'

einterfaces-mocks: ## Creates mock files for einterfaces.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir einterfaces -all -output einterfaces/mocks -note 'Regenerate this file using `make einterfaces-mocks`.'

test-run-coverage:
	go test -coverprofile=coverage_result ./...
	go tool cover -html=coverage_result

start-cmd-server:
	docker-compose up -d
	go run .\cmd\hser\main.go

run-fmt:
	go fmt ./...

gcloud-deploy:
	gcloud app deploy