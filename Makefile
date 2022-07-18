GO_FLAGS   ?=
NAME       := aws-secret-injector
OUTPUT_BIN ?= execs/awssec-inj
PACKAGE    := github.com/Mr-Mark2112/$(NAME)
GIT_REV    ?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
VERSION    ?= v0.1.0
IMG_NAME   := Mr-Mark2112/aws-secret-injector/awssec-inj
IMAGE      := ${IMG_NAME}:${VERSION}

default: help

test:   ## Run tests
	@go clean --testcache && go test ./pkg/secrets/

build:  ## Builds the CLI
	@go build ${GO_FLAGS} \
	-ldflags "-w -s -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT_REV} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o ${OUTPUT_BIN} main.go

