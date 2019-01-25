.PHONY: all alert_manager test

DOCKER_IMAGE := mayuresh82/alert_manager
DOCKER_IMAGE_TEST := mayuresh82/alert_manager_test

VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
DOCKER_TAG := $(COMMIT)
LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
    LDFLAGS += -X main.version=$(VERSION)
	DOCKER_TAG = $(VERSION)
endif

all:
	$(MAKE) alert_manager

docker:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

alert_manager:
	go build -ldflags "$(LDFLAGS)" ./cmd/alert_manager

debug:
	go build -race ./cmd/alert_manager

test:
	go test -v -race -short -failfast ./...

linux:
	GOOS=linux GOARCH=amd64 go build -o alert_manager_linux ./cmd/alert_manager

ci-test:
	docker build -t $(DOCKER_IMAGE_TEST):$(COMMIT) -f Dockerfile.test .
	docker run -t \
		-w /source \
		$(DOCKER_IMAGE_TEST):$(COMMIT) go test -v -race -short -failfast ./...