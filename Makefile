TARGETS := $(shell ls scripts)
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

vendor:
	GO111MODULE=on go mod vendor

fmt:
	gofmt -s -w $(GOFMT_FILES)

$(TARGETS): .dapper
	./.dapper $@

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS) vendor fmt
