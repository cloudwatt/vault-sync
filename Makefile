NAME=vault-sync
AUTHOR=Orange
REGISTRY=docker.io
IMPORT_PATH=github.com/cloudwatt/vault-sync
HARDWARE=$(shell uname -m)
VERSION=$(shell git describe --tags --always)
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

.PHONY: test authors changelog build docker static release lint cover vet

default: build

build:
	@echo "--> Compiling the project"
	mkdir -p bin
	go build -ldflags "-X main.Version=${VERSION}" -o bin/${NAME} .

static:
	@echo "--> Compiling the static binary"
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags "-w -X main.Version=${VERSION}" -o bin/${NAME} .

docker-build:
	@echo "--> Compiling the project"
	docker run --rm -v ${ROOT_DIR}:/go/src/${IMPORT_PATH} \
		-w /go/src/${IMPORT_PATH}/ -e GOOS=linux golang:${GOVERSION} make static

authors:
	@echo "--> Updating the AUTHORS"
	git log --format='%aN <%aE>' | sort -u > AUTHORS

test: deps
	@echo "--> Running the tests"
	go test -v $(glide nv)

changelog: release
	git log $(shell git tag | tail -n1)..HEAD --no-merges --format=%B > changelog
