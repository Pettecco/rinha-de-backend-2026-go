.PHONY: help build up down logs

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build api image with full dataset
	docker compose build

up: ## Start the stack (proxy + 2 APIs)
	docker compose up -d

down: ## Stop and remove containers
	docker compose down

logs: ## Tail compose logs
	docker compose logs -f
