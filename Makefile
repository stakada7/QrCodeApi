# http://mshim.jp/web-develop/programming/golang/golang-makefile/

VERSION=0.1.0
SRC=main/main.go

.PHONY: $(/bin/bash egrep -o ^[a-zA-Z_-]+: $(MAKEFILE_LIST) | sed 's/://')

all: setup deps lint test build-cross  ## setup & deps & lint & test & build

setup:  ## setup dev tool
	go get github.com/golang/dep/cmd/dep
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	go get github.com/pilu/fresh

build-cross: $(SRC)  ## build binary linux and darwin
	GOOS=linux GOARCH=amd64 go build -a -tags netgo -installsuffix netgo -o  bin/linux/amd64/qrcodeapi-${VERSION}/qrcodeapi ${SRC}
	GOOS=darwin GOARCH=amd64 go build -a -tags netgo -installsuffix netgo -o  bin/darwin/amd64/qrcodeapi-${VERSION}/qrcodeapi ${SRC}

fresh:  ## go run hot rebuild
	fresh

run:  ## go run
	go run main/main.go

test:  ## unit test
	go test -v -cover ./...

lint:  ## code check
	go vet ./...
	golint -set_exit_status main/main.go

deps:  ## dependency install
	dep ensure -vendor-only -v

up:  ## dependency update
	dep ensure update -v

clean:  ## clean bin, vendor, tmp dir
	rm -rf bin/*
	rm -rf vendor/*
	rm -rf tmp

help:  ## show help
	@echo Usage: make [target]
	@echo ${\n}
	@echo Targets:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
