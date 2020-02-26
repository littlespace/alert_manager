FROM golang:1.12-alpine as builder

ENV GO111MODULE=on

RUN apk update && \
    apk upgrade && \
    apk add --no-cache make git alpine-sdk

WORKDIR /go/src/github.com/mayuresh82/alert_manager

# Copy local files into the container "WORKDIR", however this dir will 
# be obscured by the bind mount in the docker-compose if you use it
COPY . .

# Install CompileDaemon to enable hot reloading 
RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -log-prefix=false -build="go build -mod=vendor ./cmd/alert_manager" -command="./alert_manager -logtostderr -v=4 -config config.toml -alert-config alert_config.yaml"