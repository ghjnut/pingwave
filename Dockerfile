FROM golang:1.11-alpine3.7 as builder

RUN apk add git

WORKDIR /go/src/github.com/ghjnut/pingwave
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:3.8

COPY --from=builder /go/bin/pingwave /usr/local/bin/

VOLUME /etc/pingwave.hcl

ENTRYPOINT ["pingwave"]
