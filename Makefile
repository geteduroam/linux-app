SHELL := /usr/bin/env bash
.PHONY: help

help:  ## Print this help message
	@grep -E '^[\%a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

.build-notifcheck:
	go build -o geteduroam-notifcheck ./cmd/geteduroam-notifcheck

.build-cli:
	go build -o geteduroam-cli ./cmd/geteduroam-cli

.build-gui:
	go build -o geteduroam-gui ./cmd/geteduroam-gui

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
	./geteduroam-notifcheck

run-cli: .build-cli ## Run CLI version
	./geteduroam-cli

run-gui: .build-gui  ## Run GUI version
	./geteduroam-gui

clean: ## Clean the project
	go clean
	rm -rf geteduroam-notifcheck
	rm -rf geteduroam-cli
	rm -rf geteduroam-gui

test:  ## Run tests
	go test ./...
