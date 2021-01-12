
# Go parameters
GOCMD=go
GOGET=$(GOCMD) get
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOLINT=golint
BUILD_TARGET=lockfree

# Pkgs
ALL_PKGS := $(shell go list ./... )
PKGS := $(shell go list ./... | grep -v /test/ )

# Package Info
PACKAGE_VERSION := $(shell git describe --tags)
PACKAGE_COMMIT_ID := $(shell git rev-parse HEAD)
GIT_STATUS := $(shell git status --porcelain)
ifdef GIT_STATUS
	GIT_STATUS := "dirty"
else
	GIT_STATUS := "clean"
endif
GO_VERSION := $(shell go version)
BUILD_TIME=$(shell date +%F-%Z/%T)
VersionImportPath := github.com/dustinxie/lockfree
PackageFlags += -X '$(VersionImportPath).PackageVersion=$(PACKAGE_VERSION)'
PackageFlags += -X '$(VersionImportPath).PackageCommitID=$(PACKAGE_COMMIT_ID)'
PackageFlags += -X '$(VersionImportPath).GitStatus=$(GIT_STATUS)'
PackageFlags += -X '$(VersionImportPath).GoVersion=$(GO_VERSION)'
PackageFlags += -X '$(VersionImportPath).BuildTime=$(BUILD_TIME)'
PackageFlags += -s -w

V ?= 0
ifeq ($(V),0)
	ECHO_V = @
else
	VERBOSITY_FLAG = -v
	DEBUG_FLAG = -debug
endif

default: clean build test

.PHONY: build
build:
	$(GOBUILD) -ldflags "$(PackageFlags)" -v ./...

.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

.PHONY: lint
lint:
	go list ./... | xargs $(GOLINT)

.PHONY: test
test: fmt
	$(GOTEST) -short -race ./...

.PHONY: clean
clean:
	@echo "Cleaning..."
	$(ECHO_V)rm -rf ./bin/$(BUILD_TARGET)
