# Meta info
NAME := mworker
VERSION := 0.2.0
MAINTAINER := "Otto Giron <ottog2486@gmail.com"
SOURCE_URL := https://github.com/ottogiron/metricsworker
DATE := $(shell date -u +%Y%m%d.%H%M%S)
COMMIT_ID := $(shell git rev-parse --short HEAD)
GIT_REPO := $(shell git config --get remote.origin.url)
# Go tools flags
LD_FLAGS := -X github.com/ottogiron/metricsworker/md.buildVersion=$(VERSION)
LD_FLAGS += -X github.com/ottogiron/metricsworker/md.buildCommit=$(COMMIT_ID)
LD_FLAGS += -X github.com/ottogiron/metricsworker/md.buildDate=$(DATE)
EXTRA_BUILD_VARS := CGO_ENABLED=0 GOARCH=amd64
SOURCE_DIRS := $(shell go list ./... | grep -v /vendor/)


all: test package-linux package-darwin

build-release: container

lint:
	@go fmt $(SOURCE_DIRS)
	@go vet $(SOURCE_DIRS)

test: install_dependencies lint
	 @go test -v $(SOURCE_DIRS) -cover -bench . -race 

install_dependencies: 
	glide install

cover: 
	gocov test $(SOURCE_DIRS) | gocov-html > coverage.html && open coverage.html
	
binaries: binary-darwin binary-linux

binary-darwin:
	@-rm -rf build/dist/darwin
	@-mkdir -p build/dist/darwin
	GOOS=darwin $(EXTRA_BUILD_VARS) go build -ldflags "$(LD_FLAGS)" -o build/dist/darwin/$(NAME)

binary-linux:
	@-rm -rf build/dist/linux
	@-mkdir -p build/dist/linux
	GOOS=linux $(EXTRA_BUILD_VARS) go build -ldflags "$(LD_FLAGS)" -o build/dist/linux/$(NAME)


package-darwin: binary-darwin
	@tar -czf build/dist/mworker.darwin-amd64.tar.gz -C build/dist/darwin mworker


package-linux: binary-linux
	@tar -czf build/dist/mworker.linux-amd64.tar.gz -C build/dist/linux mworker