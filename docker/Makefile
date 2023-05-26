SHELL := /bin/bash
#include .env

.PHONY: help

help:
	@grep -E '^[\%a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

build: ## Build LetsWifi portal
	docker compose build letswifi

up: ## Run LetsWifi portal
	docker compose up letswifi

stop: ## Stop a running development environment
	docker compose stop

down:  clean ## Stop production image
	docker compose down

clean: ## Clean containers and volatile data (keep db volume data)
	docker compose rm -svf

clean-images: ## Clean images
	docker compose down --remove-orphan --rmi=local
