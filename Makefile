.PHONY: all alert_manager test

VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
    LDFLAGS += -X main.version=$(VERSION)
endif

all:
	$(MAKE) deps
	$(MAKE) alert_manager

deps:
	go get -u golang.org/x/lint/golint
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

alert_manager:
	go build -ldflags "$(LDFLAGS)" ./cmd/alert_manager

debug:
	dep ensure
	go build -race ./cmd/alert_manager

test:
	dep ensure
	go test -v -race -short -failfast ./...

linux:
	dep ensure
	GOOS=linux GOARCH=amd64 go build -o alert_manager_linux ./cmd/alert_manager
