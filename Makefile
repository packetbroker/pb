# SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.

SHELL = bash
GO = go
GIT = git

.PHONY: deps.tidy
deps.tidy:
	@$(GO) mod tidy

.PHONY: fmt
fmt:
	@$(GO) tool gofumpt -w -extra -l .

.PHONY: quality
quality:
	$(GO) tool golangci-lint run --timeout 5m0s --allow-parallel-runners --max-issues-per-linter 0 --max-same-issues 0 $(GO_LINT_FLAGS) ./...

BASE_REF ?= master
.PHONY: quality.new
quality.new:
	@$(MAKE) quality GO_LINT_FLAGS="$(strip $(GO_LINT_FLAGS) --new-from-rev=origin/$(BASE_REF))"

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
	@$(GIT) diff --exit-code

# vim: ft=make
