FROM golang:1.12.6-alpine3.10

RUN rm -rf /var/cache/apk;mkdir /var/cache/apk;apk add --update --no-cache make git

WORKDIR /go/src/github.com/theskyinflames/sshexecutor
COPY . .
RUN GO111MODULE=on GOBIN=/usr/local/bin make build

FROM alpine:latest  
WORKDIR /root/
COPY --from=0 /usr/local/bin/cmd .
CMD ["./cmd"]  