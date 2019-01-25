FROM golang:alpine as builder
RUN apk update && \
    apk upgrade && \
    apk add --no-cache make git alpine-sdk

RUN mkdir -p /source
COPY . /source
WORKDIR /source
RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /source/alert_manager .

EXPOSE 8181/tcp 8282/tcp

ENTRYPOINT ["./alert_manager"]
