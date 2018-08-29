all:
	dep ensure
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
