GO_FLAGS   ?=
NAME       := aws-creds-injector
OUTPUT_BIN ?= execs/awssec-cp
PACKAGE    := github.com/Mr-Mark2112/$(NAME)
GIT_REV    ?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
VERSION    ?= v0.1.0
IMG_NAME   := Mr-Mark2112/aws-creds-injector/awssec-cp
IMAGE      := ${IMG_NAME}:${VERSION}

default: help

# test:   ## Run all tests
# 	@go clean --testcache && go test ./...

# cover:  ## Run test coverage suite
# 	@go test ./... --coverprofile=cov.out
# 	@go tool cover --html=cov.out

build:  ## Builds the CLI
	@go build ${GO_FLAGS} \
	-ldflags "-w -s -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT_REV} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o ${OUTPUT_BIN} main.go

