FROM golang:1.12-alpine as builder

ENV GO111MODULE=on

RUN apk update && \
    apk upgrade && \
    apk add --no-cache make git alpine-sdk
RUN mkdir -p /go/src/github.com/mayuresh82/alert_manager

COPY . /go/src/github.com/mayuresh82/alert_manager

WORKDIR /go/src/github.com/mayuresh82/alert_manager

RUN go mod download
RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/mayuresh82/alert_manager .

EXPOSE 8181/tcp 8282/tcp

ENTRYPOINT ["./alert_manager"]
