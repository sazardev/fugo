SHELL := /bin/bash
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.0.0")
BINARY := fugo
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)"

.PHONY: help test build clean lint version changelog release install push

help:
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

version: ## Show current version
	@echo "$(VERSION)"

test: ## Run tests with race detector
	go test ./... -count=1 -race -shuffle=on -v

build: ## Build binary
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY) .

lint: ## Run golangci-lint
	golangci-lint run --timeout 10m ./...

vet: ## Run go vet
	go vet ./...

clean: ## Remove build artifacts
	rm -rf bin/

changelog: ## Show unreleased changes
	@awk '/^## \[Unreleased\]/{flag=1; next} /^## \[/{flag=0} flag' CHANGELOG.md | sed '/^$$/d' || true

release: ## Create release: make release TYPE=patch|minor|major MSG="description"
	@$(eval TYPE := $(TYPE))
	@$(eval MSG := $(MSG))
	@if [ -z "$(MSG)" ]; then \
		echo "ERROR: MSG is required. Usage: make release TYPE=patch MSG='description'"; \
		exit 1; \
	fi
	@if [ "$(TYPE)" != "patch" ] && [ "$(TYPE)" != "minor" ] && [ "$(TYPE)" != "major" ]; then \
		echo "ERROR: TYPE must be patch, minor, or major. Usage: make release TYPE=patch MSG='description'"; \
		exit 1; \
	fi
	@current=$$(cat VERSION); \
	newver=$$(awk -F. -v type="$(TYPE)" ' \
		{ if (type == "patch") print $$1"."$$2"."$$3+1; \
		  else if (type == "minor") print $$1"."$$2+1".0"; \
		  else print $$1+1".0.0"; }' VERSION); \
	echo "=== Release v$$newver ==="; \
	echo "$$newver" > VERSION; \
	date=$$(date -u '+%Y-%m-%d'); \
	entry="## [$$newver] - $$date"; \
	sed -i "s/^## \[Unreleased\]$$/## [Unreleased]\n\n$$entry\n\n### Added\n- $(MSG)/" CHANGELOG.md; \
	git add VERSION CHANGELOG.md; \
	git commit -m "release: v$$newver"; \
	git tag "v$$newver"; \
	echo ""; \
	echo "=== Release v$$newver created ==="; \
	echo "Run: git push --follow-tags origin main"

install: ## Install lefthook hooks
	go tool lefthook install

push: ## Push commits and tags
	@echo "Pushing commits and tags..."
	git push --follow-tags origin main
