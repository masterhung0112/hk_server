.PHONY: build package run stop run-client run-server run-haserver stop-client stop-server restart restart-server restart-client restart-haserver start-docker clean-dist clean nuke check-style check-client-style check-server-style check-unit-tests test dist prepare-enteprise run-client-tests setup-run-client-tests cleanup-run-client-tests test-client build-linux build-osx build-windows internal-test-web-client vet run-server-for-web-client-tests diff-config prepackaged-plugins prepackaged-binaries test-server test-server-ee test-server-quick test-server-race start-docker-check migrations-bindata new-migration migration-prereqs

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

ifeq ($(OS),Windows_NT)
	PLATFORM := Windows
else
	PLATFORM := $(shell uname)
endif

# Set an environment variable on Linux used to resolve `docker.host.internal` inconsistencies with
# docker. This can be reworked once https://github.com/docker/for-linux/issues/264 is resolved
# satisfactorily.
ifeq ($(PLATFORM),Linux)
	export IS_LINUX = -linux
else
	export IS_LINUX =
endif

IS_CI ?= false
# Build Flags
BUILD_NUMBER ?= $(BUILD_NUMBER:)
BUILD_DATE = $(shell date -u)
BUILD_HASH = $(shell git rev-parse HEAD)
# If we don't set the build number it defaults to dev
ifeq ($(BUILD_NUMBER),)
	BUILD_NUMBER := dev
endif
BUILD_ENTERPRISE_DIR ?= ../enterprise
BUILD_ENTERPRISE ?= true
BUILD_ENTERPRISE_READY = false
BUILD_TYPE_NAME = team
BUILD_HASH_ENTERPRISE = none
ifneq ($(wildcard $(BUILD_ENTERPRISE_DIR)/.),)
	ifeq ($(BUILD_ENTERPRISE),true)
		BUILD_ENTERPRISE_READY = true
		BUILD_TYPE_NAME = enterprise
		BUILD_HASH_ENTERPRISE = $(shell cd $(BUILD_ENTERPRISE_DIR) && git rev-parse HEAD)
	else
		BUILD_ENTERPRISE_READY = false
		BUILD_TYPE_NAME = team
	endif
else
	BUILD_ENTERPRISE_READY = false
	BUILD_TYPE_NAME = team
endif
BUILD_WEBAPP_DIR ?= ../hungknow-webapp
BUILD_CLIENT = false
BUILD_HASH_CLIENT = independant
ifneq ($(wildcard $(BUILD_WEBAPP_DIR)/.),)
	ifeq ($(BUILD_CLIENT),true)
		BUILD_CLIENT = true
		BUILD_HASH_CLIENT = $(shell cd $(BUILD_WEBAPP_DIR) && git rev-parse HEAD)
	else
		BUILD_CLIENT = false
	endif
else
	BUILD_CLIENT = false
endif

# We need current user's UID for `run-haserver` so docker compose does not run server
# as root and mess up file permissions for devs. When running like this HOME will be blank
# and docker will add '/', so we need to set the go-build cache location or we'll get
# permission errors on build as it tries to create a cache in filesystem root.
export CURRENT_UID = $(shell id -u):$(shell id -g)
ifeq ($(HOME),/)
	export XDG_CACHE_HOME = /tmp/go-cache/
endif
# Go Flags
GOFLAGS ?= $(GOFLAGS:)
# We need to export GOBIN to allow it to be set
# for processes spawned from the Makefile
export GOBIN ?= $(PWD)/bin
GO ?= /usr/local/go/bin/go

LDFLAGS += -X "github.com/masterhung0112/hk_server/v5/model.BuildNumber=$(BUILD_NUMBER)"
LDFLAGS += -X "github.com/masterhung0112/hk_server/v5/model.BuildDate=$(BUILD_DATE)"
LDFLAGS += -X "github.com/masterhung0112/hk_server/v5/model.BuildHash=$(BUILD_HASH)"
LDFLAGS += -X "github.com/masterhung0112/hk_server/v5/model.BuildHashEnterprise=$(BUILD_HASH_ENTERPRISE)"
LDFLAGS += -X "github.com/masterhung0112/hk_server/v5/model.BuildEnterpriseReady=$(BUILD_ENTERPRISE_READY)"

GO_MAJOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MINIMUM_SUPPORTED_GO_MAJOR_VERSION = 1
MINIMUM_SUPPORTED_GO_MINOR_VERSION = 15
GO_VERSION_VALIDATION_ERR_MSG = Golang version is not supported, please update to at least $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION).$(MINIMUM_SUPPORTED_GO_MINOR_VERSION)

# GOOS/GOARCH of the build host, used to determine whether we're cross-compiling or not
BUILDER_GOOS_GOARCH="$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)"

PLATFORM_FILES="./cmd/hkserver/main.go"

# Output paths
DIST_ROOT=dist
DIST_PATH=$(DIST_ROOT)/hkserver

# Tests
TESTS=.

# Packages lists
TE_PACKAGES=$(shell $(GO) list ./... | grep -v ./data)

TEMPLATES_DIR=templates

# Plugins Packages
PLUGIN_PACKAGES ?= mattermost-plugin-antivirus-v0.1.2
PLUGIN_PACKAGES += mattermost-plugin-autolink-v1.2.2
PLUGIN_PACKAGES += mattermost-plugin-aws-SNS-v1.2.0
PLUGIN_PACKAGES += mattermost-plugin-channel-export-v0.2.2
PLUGIN_PACKAGES += mattermost-plugin-custom-attributes-v1.3.0
PLUGIN_PACKAGES += mattermost-plugin-github-v2.0.1
PLUGIN_PACKAGES += mattermost-plugin-gitlab-v1.3.0
PLUGIN_PACKAGES += mattermost-plugin-incident-collaboration-v1.12.0
PLUGIN_PACKAGES += mattermost-plugin-jenkins-v1.1.0
PLUGIN_PACKAGES += mattermost-plugin-jira-v2.4.0
PLUGIN_PACKAGES += mattermost-plugin-nps-v1.1.0
PLUGIN_PACKAGES += mattermost-plugin-welcomebot-v1.2.0
PLUGIN_PACKAGES += mattermost-plugin-zoom-v1.5.0
PLUGIN_PACKAGES += focalboard-v0.6.7-plugin

# Prepares the enterprise build if exists. The IGNORE stuff is a hack to get the Makefile to execute the commands outside a target
ifeq ($(BUILD_ENTERPRISE_READY),true)
	IGNORE:=$(shell echo Enterprise build selected, preparing)
	IGNORE:=$(shell rm -f imports/imports.go)
	IGNORE:=$(shell cp $(BUILD_ENTERPRISE_DIR)/imports/imports.go imports/)
	IGNORE:=$(shell rm -f enterprise)
	IGNORE:=$(shell ln -s $(BUILD_ENTERPRISE_DIR) enterprise)
else
	IGNORE:=$(shell rm -f imports/imports.go)
endif

EE_PACKAGES=$(shell $(GO) list ./enterprise/...)

ifeq ($(BUILD_ENTERPRISE_READY),true)
ALL_PACKAGES=$(TE_PACKAGES) $(EE_PACKAGES)
else
ALL_PACKAGES=$(TE_PACKAGES)
endif

all: run ## Alias for 'run'.

-include config.override.mk
include config.mk
include build/*.mk

LDFLAGS += -X "github.com/masterhung0112/hk_server/v5/model.MockCWS=$(MM_ENABLE_CWS_MOCK)"
RUN_IN_BACKGROUND ?=
ifeq ($(RUN_SERVER_IN_BACKGROUND),true)
	RUN_IN_BACKGROUND := &
endif

start-docker-check:
ifeq (,$(findstring minio,$(ENABLED_DOCKER_SERVICES)))
  TEMP_DOCKER_SERVICES:=$(TEMP_DOCKER_SERVICES) minio
endif
ifeq ($(BUILD_ENTERPRISE_READY),true)
  ifeq (,$(findstring openldap,$(ENABLED_DOCKER_SERVICES)))
    TEMP_DOCKER_SERVICES:=$(TEMP_DOCKER_SERVICES) openldap
  endif
  ifeq (,$(findstring elasticsearch,$(ENABLED_DOCKER_SERVICES)))
    TEMP_DOCKER_SERVICES:=$(TEMP_DOCKER_SERVICES) elasticsearch
  endif
endif
ENABLED_DOCKER_SERVICES:=$(ENABLED_DOCKER_SERVICES) $(TEMP_DOCKER_SERVICES)

start-docker: ## Starts the docker containers for local development.
ifneq ($(IS_CI),false)
	@echo CI Build: skipping docker start
else ifeq ($(MM_NO_DOCKER),true)
	@echo No Docker Enabled: skipping docker start
else
	@echo Starting docker containers

	$(GO) run ./build/docker-compose-generator/main.go $(ENABLED_DOCKER_SERVICES) | docker-compose -f docker-compose.makefile.yml -f /dev/stdin run --rm start_dependencies
ifneq (,$(findstring openldap,$(ENABLED_DOCKER_SERVICES)))
	cat tests/${LDAP_DATA}-data.ldif | docker-compose -f docker-compose.makefile.yml exec -T openldap bash -c 'ldapadd -x -D "cn=admin,dc=mm,dc=test,dc=com" -w mostest || true';
endif
ifneq (,$(findstring mysql-read-replica,$(ENABLED_DOCKER_SERVICES)))
	./scripts/replica-mysql-config.sh
endif
endif

run-haserver:
ifeq ($(BUILD_ENTERPRISE_READY),true)
	@echo Starting mattermost in an HA topology '(3 node cluster)'

	docker-compose -f docker-compose.yaml up --remove-orphans haproxy
endif
stop-haserver:
	@echo Stopping docker containers for HA topology
	docker-compose stop

stop-docker: ## Stops the docker containers for local development.
ifeq ($(MM_NO_DOCKER),true)
	@echo No Docker Enabled: skipping docker stop
else
	@echo Stopping docker containers

	docker-compose stop
endif
clean-docker: ## Deletes the docker containers for local development.
ifeq ($(MM_NO_DOCKER),true)
	@echo No Docker Enabled: skipping docker clean
else
	@echo Removing docker containers

	docker-compose down -v
	docker-compose rm -v
endif

plugin-checker:
	$(GO) run $(GOFLAGS) ./plugin/checker

prepackaged-plugins: ## Populate the prepackaged-plugins directory
	@echo Downloading prepackaged plugins
	mkdir -p prepackaged_plugins
	@cd prepackaged_plugins && for plugin_package in $(PLUGIN_PACKAGES) ; do \
		curl -f -O -L https://plugins-store.test.mattermost.com/release/$$plugin_package.tar.gz; \
		curl -f -O -L https://plugins-store.test.mattermost.com/release/$$plugin_package.tar.gz.sig; \
	done

prepackaged-binaries: ## Populate the prepackaged-binaries to the bin directory
ifeq ($(shell test -f bin/mmctl && printf "yes"),yes)
	@echo "MMCTL already exists in bin/mmctl not downloading a new version."
else
	@scripts/download_mmctl_release.sh
endif

golangci-lint: ## Run golangci-lint on codebase
# https://stackoverflow.com/a/677212/1027058 (check if a command exists or not)
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...
ifeq ($(BUILD_ENTERPRISE_READY),true)
ifneq ($(MM_NO_ENTERPRISE_LINT),true)
	golangci-lint run ./enterprise/...
endif
endif

app-layers: ## Extract interface from App struct
	$(GO) get -modfile=go.tools.mod github.com/reflog/struct2interface
	$(GOBIN)/struct2interface -f "app" -o "app/app_iface.go" -p "app" -s "App" -i "AppIface" -t ./app/layer_generators/app_iface.go.tmpl
	$(GO) run ./app/layer_generators -in ./app/app_iface.go -out ./app/opentracing/opentracing_layer.go -template ./app/layer_generators/opentracing_layer.go.tmpl

i18n-extract: ## Extract strings for translation from the source code
	$(GO) get -modfile=go.tools.mod github.com/mattermost/mattermost-utilities/mmgotool
	$(GOBIN)/mmgotool i18n extract --portal-dir=""

i18n-check: ## Exit on empty translation strings and translation source strings
	$(GO) get -modfile=go.tools.mod github.com/mattermost/mattermost-utilities/mmgotool
	$(GOBIN)/mmgotool i18n clean-empty --portal-dir="" --check
	$(GOBIN)/mmgotool i18n check-empty-src --portal-dir=""

store-mocks: ## Creates mock files.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir store -all -output store/storetest/mocks -note 'Regenerate this file using `make store-mocks`.'
telemetry-mocks: ## Creates mock files.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir services/telemetry -all -output services/telemetry/mocks -note 'Regenerate this file using `make telemetry-mocks`.'

store-layers: ## Generate layers for the store
	$(GO) generate $(GOFLAGS) ./store
migration-prereqs: ## Builds prerequisite packages for migrations
	$(GO) get -modfile=go.tools.mod github.com/golang-migrate/migrate/v4/cmd/migrate

new-migration: migration-prereqs ## Creates a new migration
	@echo "Generating new migration for mysql"
	$(GOBIN)/migrate create -ext sql -dir db/migrations/mysql -seq $(name)

	@echo "Generating new migration for postgres"
	$(GOBIN)/migrate create -ext sql -dir db/migrations/postgres -seq $(name)

	@echo "When you are done writing your migration, run 'make migrations'"

migrations-bindata: ## Generates bindata migrations
	$(GO) get -modfile=go.tools.mod github.com/go-bindata/go-bindata/...

	@echo Generating bindata for migrations
	$(GO) generate $(GOFLAGS) ./db/migrations/

filestore-mocks: ## Creates mock files.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir shared/filestore -all -output shared/filestore/mocks -note 'Regenerate this file using `make filestore-mocks`.'

ldap-mocks: ## Creates mock files for ldap.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir enterprise/ldap -all -output enterprise/ldap/mocks -note 'Regenerate this file using `make ldap-mocks`.'

plugin-mocks: ## Creates mock files for plugins.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir plugin -name API -output plugin/plugintest -outpkg plugintest -case underscore -note 'Regenerate this file using `make plugin-mocks`.'
	$(GOBIN)/mockery -dir plugin -name Hooks -output plugin/plugintest -outpkg plugintest -case underscore -note 'Regenerate this file using `make plugin-mocks`.'
	$(GOBIN)/mockery -dir plugin -name Helpers -output plugin/plugintest -outpkg plugintest -case underscore -note 'Regenerate this file using `make plugin-mocks`.'

einterfaces-mocks: ## Creates mock files for einterfaces.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir einterfaces -all -output einterfaces/mocks -note 'Regenerate this file using `make einterfaces-mocks`.'

searchengine-mocks: ## Creates mock files for searchengines.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir services/searchengine -all -output services/searchengine/mocks -note 'Regenerate this file using `make searchengine-mocks`.'

sharedchannel-mocks: ## Creates mock files for shared channels.
	$(GO) get -modfile=go.tools.mod github.com/vektra/mockery/...
	$(GOBIN)/mockery -dir=./services/sharedchannel -name=ServerIface -output=./services/sharedchannel -inpkg -outpkg=sharedchannel -testonly -note 'Regenerate this file using `make sharedchannel-mocks`.'
	$(GOBIN)/mockery -dir=./services/sharedchannel -name=AppIface -output=./services/sharedchannel -inpkg -outpkg=sharedchannel -testonly -note 'Regenerate this file using `make sharedchannel-mocks`.'
pluginapi: ## Generates api and hooks glue code for plugins
	$(GO) generate $(GOFLAGS) ./plugin

check-prereqs: ## Checks prerequisite software status.
	./scripts/prereq-check.sh

check-prereqs-enterprise: ## Checks prerequisite software status for enterprise.
ifeq ($(BUILD_ENTERPRISE_READY),true)
	./scripts/prereq-check-enterprise.sh
endif

check-style: golangci-lint plugin-checker vet ## Runs style/lint checks


do-cover-file: ## Creates the test coverage report file.
	@echo "mode: count" > cover.out

go-junit-report:
	$(GO) get -modfile=go.tools.mod github.com/jstemmer/go-junit-report

test-compile: ## Compile tests.
	@echo COMPILE TESTS

	for package in $(TE_PACKAGES) $(EE_PACKAGES); do \
		$(GO) test $(GOFLAGS) -c $$package; \
	done

test-db-migration: start-docker ## Gets diff of upgrade vs new instance schemas.
	./scripts/mysql-migration-test.sh
	./scripts/psql-migration-test.sh

gomodtidy:
	@cp go.mod go.mod.orig
	@cp go.sum go.sum.orig
	$(GO) mod tidy
	@if [ "$$(diff go.mod go.mod.orig)" != "" -o "$$(diff go.sum go.sum.orig)" != "" ]; then \
		echo "go.mod/go.sum was modified. \ndiff- $$(diff go.mod go.mod.orig) \n$$(diff go.sum go.sum.orig) \nRun \"go mod tidy\"."; \
		rm go.*.orig; \
		exit 1; \
	fi;
	@rm go.*.orig;

test-server-pre: check-prereqs-enterprise start-docker-check start-docker go-junit-report do-cover-file ## Runs tests.
ifeq ($(BUILD_ENTERPRISE_READY),true)
	@echo Running all tests
else
	@echo Running only TE tests
endif
	./scripts/test.sh "$(GO)" "$(GOFLAGS)" "$(ALL_PACKAGES)" "$(TESTS)" "$(TESTFLAGS)" "$(GOBIN)"
  ifneq ($(IS_CI),true)
    ifneq ($(MM_NO_DOCKER),true)
      ifneq ($(TEMP_DOCKER_SERVICES),)
	      @echo Stopping temporary docker services
	      docker-compose stop $(TEMP_DOCKER_SERVICES)
      endif
    endif
  endif

test-server-ci:
	./scripts/test.sh "$(GO)" "$(GOFLAGS)" "$(ALL_PACKAGES)" "$(TESTS)" "$(TESTFLAGS)" "$(GOBIN)"
