ENV ?= prod
SERVICE_NAME ?= duet-livekit-server

BUILD_DIR := $(CURDIR)/build

bootstrap:
	sh bootstrap.sh

clean: 
	rm -rf $(BUILD_DIR)

prebuild:
	mkdir -p $(BUILD_DIR)/{bin,conf}

build: clean prebuild 
	go build -o $(BUILD_DIR)/bin/livekit-server ./cmd/server/

pack: build
	cp -r $(CURDIR)/conf/$(SERVICE_NAME)/$(ENV)/* $(BUILD_DIR)/conf/
	cp -r $(CURDIR)/deploy/$(ENV)/$(SERVICE_NAME).service $(BUILD_DIR)/

.PHONY: bootstrap, build