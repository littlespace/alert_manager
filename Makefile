.PHONY: all alert_manager test

VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
    LDFLAGS += -X main.version=$(VERSION)
endif

all:
	$(MAKE) alert_manager

alert_manager:
	go build -ldflags "$(LDFLAGS)" ./cmd/alert_manager

debug:
	go build -race ./cmd/alert_manager

test:
	go test -v -race -short -failfast ./...

linux:
	GOOS=linux GOARCH=amd64 go build -o alert_manager_linux ./cmd/alert_manager
