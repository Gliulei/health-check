ROOT:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BIN_OUT:=$(ROOT)/bin/health-check
PKG:=$(shell go list -m)

.PHONY: all build dts clean test build_with_coverage
all: build #test

build: health-check

health-check:
	go build -o $(BIN_OUT) main.go

clean:
	@rm -rf bin