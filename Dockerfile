FROM golang:alpine as builder
RUN apk update && \
    apk upgrade && \
    apk add --no-cache git && \
    apk add make
RUN mkdir -p /opt/alert_manager
RUN mkdir -p /go/src/github.com/mayuresh82
RUN cd /go/src/github.com/mayuresh82 && \
    git clone --branch dev https://github.com/mayuresh82/alert_manager
WORKDIR /go/src/github.com/mayuresh82/alert_manager
RUN make
RUN cp alert_manager /opt/alert_manager/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /opt/alert_manager/alert_manager .

EXPOSE 8181/tcp 8282/tcp

ENTRYPOINT ["./alert_manager"]
