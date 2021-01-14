build-linux:
	@echo Build Linux amd64
ifeq ($(BUILDER_GOOS_GOARCH),"linux_amd64")
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(GOBIN) $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./...
else
	mkdir -p $(GOBIN)/linux_amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(GOBIN)/linux_amd64 $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./...
endif

build-osx:
	@echo Build OSX amd64, goarch $(BUILDER_GOOS_GOARCH)
ifeq ($(BUILDER_GOOS_GOARCH),"darwin_amd64")
	env GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOBIN) $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./...
else
	mkdir -p $(GOBIN)/darwin_amd64
	env GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOBIN)/darwin_amd64 $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./...
endif

build-windows:
	@echo Build Windows amd64
ifeq ($(BUILDER_GOOS_GOARCH),"windows_amd64")
	env GOOS=windows GOARCH=amd64 $(GO) build -o $(GOBIN) $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./...
else
	mkdir -p $(GOBIN)/windows_amd64
	env GOOS=windows GOARCH=amd64 $(GO) build -o $(GOBIN)/windows_amd64 $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./...
endif

build: build-linux # build-windows build-osx


build-cmd-linux:
	@echo Build Linux amd64
ifeq ($(BUILDER_GOOS_GOARCH),"linux_amd64")
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(GOBIN) $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./cmd/...
else
	mkdir -p $(GOBIN)/linux_amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(GOBIN)/linux_amd64 $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./cmd/...
endif

build-cmd-osx:
	@echo Build OSX amd64
ifeq ($(BUILDER_GOOS_GOARCH),"darwin_amd64")
	env GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOBIN) $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./cmd/...
else
	mkdir -p $(GOBIN)/darwin_amd64
	env GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOBIN)/darwin_amd64 $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./cmd/...
endif

build-cmd-windows:
	@echo Build Windows amd64
ifeq ($(BUILDER_GOOS_GOARCH),"windows_amd64")
	env GOOS=windows GOARCH=amd64 $(GO) build -o $(GOBIN) $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./cmd/...
else
	mkdir -p $(GOBIN)/windows_amd64
	env GOOS=windows GOARCH=amd64 $(GO) build -o $(GOBIN)/windows_amd64 $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' ./cmd/...
endif

build-cmd: build-cmd-linux # build-cmd-windows build-cmd-osx

package:
	@ echo Packaging hkserver
	@# Remove any old files
	rm -Rf $(DIST_ROOT)

	@# Create needed directories
	mkdir -p $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/logs
	mkdir -p $(DIST_PATH)/prepackaged_plugins

	@# Resource directories
	mkdir -p $(DIST_PATH)/config
	OUTPUT_CONFIG=$(PWD)/$(DIST_PATH)/config/config.json go generate ./config
	cp -RL fonts $(DIST_PATH)
	cp -RL templates $(DIST_PATH)
	cp -RL i18n $(DIST_PATH)

	@# Disable developer settings
	sed -i'' -e 's|"ConsoleLevel": "DEBUG"|"ConsoleLevel": "INFO"|g' $(DIST_PATH)/config/config.json
	sed -i'' -e 's|"SiteURL": "http://localhost:8065"|"SiteURL": ""|g' $(DIST_PATH)/config/config.json

	@# Reset email sending to original configuration
	sed -i'' -e 's|"SendEmailNotifications": true,|"SendEmailNotifications": false,|g' $(DIST_PATH)/config/config.json
	sed -i'' -e 's|"FeedbackEmail": "test@example.com",|"FeedbackEmail": "",|g' $(DIST_PATH)/config/config.json
	sed -i'' -e 's|"ReplyToAddress": "test@example.com",|"ReplyToAddress": "",|g' $(DIST_PATH)/config/config.json
	sed -i'' -e 's|"SMTPServer": "localhost",|"SMTPServer": "",|g' $(DIST_PATH)/config/config.json
	sed -i'' -e 's|"SMTPPort": "2500",|"SMTPPort": "",|g' $(DIST_PATH)/config/config.json

	@# ----- PLATFORM SPECIFIC -----

	@# Make osx package
	@# Copy binary
ifeq ($(BUILDER_GOOS_GOARCH),"darwin_amd64")
	#cp $(GOBIN)/hkserver $(DIST_PATH)/bin # from native bin dir, not cross-compiled
else
	#cp $(GOBIN)/darwin_amd64/hkserver $(DIST_PATH)/bin # from cross-compiled bin dir
endif
	@# Package
	#tar -C dist -czf $(DIST_PATH)-$(BUILD_TYPE_NAME)-osx-amd64.tar.gz hkserver
	@# Cleanup
	# rm -f $(DIST_PATH)/bin/hkserver
	# rm -f $(DIST_PATH)/bin/mmctl
	# rm -f $(DIST_PATH)/prepackaged_plugins/*

	@# Make linux package
	@# Copy binary
ifeq ($(BUILDER_GOOS_GOARCH),"linux_amd64")
	cp $(GOBIN)/hkserver $(DIST_PATH)/bin # from native bin dir, not cross-compiled
else
	cp $(GOBIN)/linux_amd64/hkserver $(DIST_PATH)/bin # from cross-compiled bin dir
endif
	@# Package
	tar -C dist -czf $(DIST_PATH)-$(BUILD_TYPE_NAME)-linux-amd64.tar.gz hkserver
	@# Don't clean up native package so dev machines will have an unzipped package available
	@#rm -f $(DIST_PATH)/bin/hkserver

	rm -rf tmpprepackaged