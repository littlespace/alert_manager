FROM golang:alpine as builder
RUN apk update && \
    apk upgrade && \
    apk add --no-cache make git alpine-sdk

RUN mkdir -p /go/src/github.com/mayuresh82
COPY . /go/src/github.com/mayuresh82
WORKDIR /go/src/github.com/mayuresh82
RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/mayuresh82/alert_manager .

EXPOSE 8181/tcp 8282/tcp

ENTRYPOINT ["./alert_manager"]
