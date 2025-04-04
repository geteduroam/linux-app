SHELL := /usr/bin/env bash
.PHONY: help

help:  ## Print this help message
	@grep -E '^[\%a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
VARIANT := geteduroam
ifneq ($(VARIANT),geteduroam)
	GOBUILDFLAGS := "-tags=$(VARIANT)"
endif

.build-notifcheck:
	go build $(GOBUILDFLAGS) -o $(VARIANT)-notifcheck ./cmd/geteduroam-notifcheck

.build-cli:
	go build $(GOBUILDFLAGS) -o $(VARIANT)-cli ./cmd/geteduroam-cli

.build-gui:
	go build $(GOBUILDFLAGS) -o $(VARIANT)-gui ./cmd/geteduroam-gui

lint:  ## Lint the codebase using Golangci-lint
	golangci-lint run -E stylecheck,revive,gocritic --timeout 5m

fmt:  ## Format the codebase using Gofumpt
	gofumpt -w .

build-notifcheck: .build-notifcheck ## Build notification checker
	@echo "Done building, run 'make run-notifcheck' to run the notification checker"

build-cli: .build-cli ## Build CLI version
	@echo "Done building, run 'make run-cli' to run the CLI"

build-gui: .build-gui ## Build GUI version
	@echo "Done building, run 'make run-gui' to run the GUI"

run-notifcheck: .build-notifcheck ## Run notification checker
	./$(VARIANT)-notifcheck

run-cli: .build-cli ## Run CLI version
	./$(VARIANT)-cli

run-gui: .build-gui  ## Run GUI version
	./$(VARIANT)-gui

clean: ## Clean the project
	go clean
	rm -rf $(VARIANT)-notifcheck
	rm -rf $(VARIANT)-cli
	rm -rf $(VARIANT)-gui

test:  ## Run tests
	go test ./...
