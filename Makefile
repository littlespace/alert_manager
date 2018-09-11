all:
	$(MAKE) deps
	$(MAKE) alert_manager

deps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

alert_manager:
	go build ./cmd/alert_manager

debug:
	dep ensure
	go build -race ./cmd/alert_manager

test:
	dep ensure
	go test -v -race -short -failfast ./...

linux:
	dep ensure
	GOOS=linux GOARCH=amd64 go build -o alert_manager_linux ./cmd/alert_manager
