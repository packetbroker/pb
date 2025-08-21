# Copyright © 2020 The Things Industries B.V.

SHELL = bash
GO = go

.PHONY: deps.tidy
deps.tidy:
	@$(GO) mod tidy

.PHONY: fmt
fmt:
	@$(GO) tool gofumpt -l -w .

.PHONY: quality
quality:
	$(GO) tool golangci-lint run --timeout 5m0s --issues-exit-code 0

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
