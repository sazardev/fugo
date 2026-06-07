SHELL := /bin/bash
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.0.0")
BINARY := fugo
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)"

.PHONY: help test build clean lint version changelog release install push pr pr-merge pr-list pr-update

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

pr: ## Create PR: make pr MSG="type: description"
	@if [ -z "$(MSG)" ]; then \
		echo "ERROR: MSG is required. Usage: make pr MSG='fix: login bug'"; \
		exit 1; \
	fi
	@branch=$$(echo "$(MSG)" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9/_-]/-/g; s/:/\//g; s/--*/-/g; s/^-//; s/-$$//'); \
	current=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$current" != "main" ]; then \
		echo "ERROR: You must be on 'main' to create a PR branch. Currently on '$$current'."; \
		exit 1; \
	fi; \
	echo "=== Creating branch: $$branch ==="; \
	git checkout -b "$$branch" && \
	git add . && \
	if git diff --cached --quiet; then \
		echo "No changes staged, creating empty commit."; \
		git commit --allow-empty -m "$(MSG)"; \
	else \
		git commit -m "$(MSG)"; \
	fi && \
	echo "=== Pushing branch ===" && \
	git push origin "$$branch" && \
	echo "=== Creating PR ===" && \
	printf '## Description\n\n%s\n\n## Checklist\n\n- [ ] `gofumpt -w .` run\n- [ ] `go vet ./...` passes\n- [ ] `golangci-lint run` passes\n- [ ] `go test ./... -race -shuffle=on` passes\n- [ ] `go mod tidy` run\n- [ ] CHANGELOG updated\n' "$(MSG)" > /tmp/pr-body-$$$$.md && \
	gh pr create \
		--base main \
		--head "$$branch" \
		--title "$(MSG)" \
		--body-file /tmp/pr-body-$$$$.md && \
	rm -f /tmp/pr-body-$$$$.md && \
	echo "=== Done: https://github.com/sazardev/fugo/pull/$$(gh pr list --head $$branch --json number -q '.[0].number') ==="

pr-merge: ## Merge PR: make pr-merge PR=<number>
	@if [ -z "$(PR)" ]; then \
		echo "ERROR: PR is required. Usage: make pr-merge PR=5"; \
		exit 1; \
	fi
	@echo "=== Merging PR #$(PR) ==="; \
	head_branch=$$(gh pr view $(PR) --json headRefName -q '.headRefName'); \
	echo "Branch: $$head_branch"; \
	gh pr merge $(PR) --squash --admin --subject "$$(gh pr view $(PR) --json title -q '.title')" && \
	echo "=== Cleaning up ===" && \
	git checkout main 2>/dev/null; \
	git pull origin main 2>/dev/null; \
	git branch -D "$$head_branch" 2>/dev/null || true; \
	git push origin --delete "$$head_branch" 2>/dev/null || true; \
	echo "=== PR #$(PR) merged and branch deleted ==="

pr-list: ## List open PRs
	@gh pr list --json number,title,headRefName,author,createdAt,mergeStateStatus \
		-q '.[] | "PR #\(.number): \(.title)\n  author: \(.author.login)  branch: \(.headRefName)  status: \(.mergeStateStatus)  created: \(.createdAt)\n"' \
		|| gh pr list

pr-update: ## Update current branch with main
	@current=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$current" = "main" ]; then \
		echo "Already on main."; \
		exit 0; \
	fi; \
	echo "=== Updating $$current with main ==="; \
	git fetch origin main && \
	git rebase origin/main && \
	echo "=== Branch updated ==="
