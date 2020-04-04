-include .env

VERSION := $(shell git describe --always --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

# Go related variables.
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard **/*.go)

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)" 

DOCKER_IMAGE := donilan/dot
DOCKER_IMAGE_TAG ?= latest

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## install: Install missing dependencies. Runs `go get` internally. e.g; make install get=github.com/foo/bar
install: go-get
## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

## exec: Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
exec:
	@$(run)

## clean: Clean build files. Runs `go clean` internally.
clean:
	@-rm $(GOBIN)/$(PROJECTNAME) 2> /dev/null
	@-$(MAKE) go-clean

go-compile: go-get go-build

go-build:
	@echo "  >  Building binary..."
	@go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-generate:
	@echo "  >  Generating dependency files..."
	@go generate $(generate)

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	@go get $(get)

go-install:
	@go install $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@go clean

docker-build:
	@docker build . -t ${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG}

docker-release: docker-build
	@docker push ${DOCKER_IMAGE}:${DOCKER_IMAGE_TAG}

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo