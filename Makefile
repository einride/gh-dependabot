SHELL := /usr/bin/env bash

.PHONY: all
all: \
	commitlint \
	go-lint \
	go-review \
	go-test \
	go-mod-tidy \
	git-verify-nodiff

include .tools/commitlint/rules.mk
include .tools/git-verify-nodiff/rules.mk
include .tools/golangci-lint/rules.mk
include .tools/goreview/rules.mk
include .tools/semantic-release/rules.mk

.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@go mod tidy -v

.PHONY: go-test
go-test:
	$(info [$@] running Go tests...)
	@mkdir -p .build/coverage
	@go test -short -race -coverprofile=.build/coverage/$@.txt -covermode=atomic ./...
