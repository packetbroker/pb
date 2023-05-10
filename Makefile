# Copyright © 2020 The Things Industries B.V.

SHELL = bash
GO = go
GOBIN = $(PWD)/.bin
export GOBIN

.PHONY: deps.dev
deps.dev:
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % $(GO) install %

.PHONY: deps.tidy
deps.tidy:
	@$(GO) mod tidy

.PHONY: fmt
fmt:
	@$(GOBIN)/gofumpt -l -w .

.PHONY: quality
quality:
	@$(GOBIN)/golint -set_exit_status ./... && \
		$(GO) vet ./...

.PHONY: test
test:
	@$(GO) test ./...

.PHONY: test.race
test.race:
	@$(GO) test -race -covermode=atomic ./...

.PHONY: test.cover
test.cover:
	@$(GO) test -cover ./...

.PHONY: git.nodiff
git.nodiff:
	@if [[ ! -z "`git diff`" ]]; then \
		git diff; \
		exit 1; \
	fi

# vim: ft=make
