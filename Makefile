NAME     := rosculus
VERSION  := v0.0.1
REVISION := $(shell git rev-parse --short HEAD)

SRCS    := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""

DIST_DIRS := find * -type d -exec

.DEFAULT_GOAL := bin/$(NAME)

bin/$(NAME): $(SRCS)
	go build $(LDFLAGS) -o bin/$(NAME)

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf vendor/*

.PHONY: deps
deps: dep
	dep ensure

.PHONY: dep
dep:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: test
test:
	go test -cover -v `go list ./... | grep -v /vendor/`

.PHONY: update-deps
update-deps: dep
	dep ensure -update
