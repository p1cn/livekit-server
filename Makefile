ENV ?= prod
SERVICE_NAME ?= duet-livekit-server

BUILD_DIR := $(CURDIR)/build

bootstrap:
	sh bootstrap.sh

clean: 
	rm -rf $(BUILD_DIR)

build: clean bootstrap
	mage build

pack: build
	mkdir -p $(BUILD_DIR)/{bin,conf}
	cp -r $(CURDIR)/bin/* $(BUILD_DIR)/bin/
	cp -r $(CURDIR)/conf/$(ENV)/* $(BUILD_DIR)/conf/
	cp -r $(CURDIR)/deploy/$(ENV)/$(SERVICE_NAME).service $(BUILD_DIR)/

.PHONY: bootstrap, build