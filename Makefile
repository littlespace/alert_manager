all:
	dep ensure
	go build ./cmd/alert_manager

debug:
	dep ensure
	go build -race ./cmd/alert_manager

test:
	go test -race -short -failfast ./...
