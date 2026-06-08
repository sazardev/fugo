SHELL := /bin/bash
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.0.0")
BINARY := fugo
# Pin the Dart protoc plugin to the version that matches the pubspec protobuf
# runtime (protobuf ^3.1.0). A mismatched global plugin breaks `flutter build`.
PROTOC_DART_VERSION := 21.1.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell git log -1 --format=%cd --date=format:'%Y-%m-%d_%H:%M:%S' 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)"

# --- OS-specific settings (Windows vs Unix) ---
ifeq ($(OS),Windows_NT)
	FLUTTER_BUILD_ARGS := windows
	FLUTTER_BINARY     := flutter_client/build/windows/x64/runner/Release/fugo_flutter_client.exe
	SPIKE_BIN          := bin/fugo-spike.exe
	DART_PROTOC_PLUGIN := $(LOCALAPPDATA)/Pub/Cache/bin/protoc-gen-dart.bat
else
	FLUTTER_BUILD_ARGS := linux --debug
	FLUTTER_BINARY     := flutter_client/build/linux/x64/debug/bundle/fugo_flutter_client
	SPIKE_BIN          := bin/fugo-spike
	DART_PROTOC_PLUGIN := $(HOME)/.pub-cache/bin/protoc-gen-dart
endif

.PHONY: help test bench build clean lint vet version changelog release install install-tools push pr pr-merge pr-list pr-update proto proto-tools flutter-build spike run run-spike cli cli-test install-cli

help:
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

version: ## Show current version
	@echo "$(VERSION)"

test: ## Run tests with race detector
	go test ./... -count=1 -race -shuffle=on -v

bench: ## Run engine benchmarks
	go test ./engine/ -bench=. -benchmem -run='^$$'

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

install-tools: ## Install dev tools (gofumpt, staticcheck, golangci-lint) built with local Go
	go install mvdan.cc/gofumpt@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "=== Tools installed to $$(go env GOPATH)/bin (built with $$(go version)) ==="

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

proto-tools: ## Pin the Dart protoc plugin to match the pubspec protobuf runtime
	dart pub global activate protoc_plugin $(PROTOC_DART_VERSION)

proto: proto-tools ## Generate protobuf code (Go + Dart)
	@echo "=== Generating Go protobuf code ==="
	protoc --proto_path=. --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		transport/proto/fugo/v1/fugo.proto
	@echo "=== Copying proto to flutter_client ==="
	@mkdir -p flutter_client/proto/fugo/v1
	cp transport/proto/fugo/v1/fugo.proto flutter_client/proto/fugo/v1/
	@echo "=== Generating Dart protobuf code ==="
	@mkdir -p flutter_client/lib/generated
	protoc --proto_path=flutter_client/proto \
		--dart_out=grpc:flutter_client/lib/generated \
		--plugin=protoc-gen-dart=$(DART_PROTOC_PLUGIN) \
		fugo/v1/fugo.proto
	@echo "=== Proto generation complete ==="

flutter-build: ## Build Flutter client for the current OS
	@echo "=== Building Flutter client ($(FLUTTER_BUILD_ARGS)) ==="
	cd flutter_client && flutter build $(FLUTTER_BUILD_ARGS)
	@echo "=== Flutter build complete ==="

spike: ## Build spike binary
	@mkdir -p bin
	go build -o $(SPIKE_BIN) ./cmd/fugo-spike/
	@echo "=== Spike binary built: $(SPIKE_BIN) ==="

run: run-spike ## Run the demo app (alias for run-spike)

run-spike: spike ## Run spike (starts Go server + Flutter client)
	@if [ ! -f "$(FLUTTER_BINARY)" ]; then \
		echo "Flutter client not built. Run: make flutter-build"; \
		exit 1; \
	fi
	./$(SPIKE_BIN)

cli: ## Build fugo CLI binary
	go build -o bin/fugo.exe ./cmd/fugo/
	go build -o bin/fugo-spike.exe ./cmd/fugo-spike/

install-cli: ## Install fugo globally (go install → GOPATH/bin)
	go install ./cmd/fugo/
	@echo "=== fugo installed! Run: fugo --help ==="

cli-test: cli ## Build CLI + create test project (init -> build)
	cmd /c "if exist testapp rmdir /s /q testapp"
	.\bin\fugo.exe init testapp
	cd testapp && go build -o bin\app.exe .
