docker run --rm -it \
  -v "$(pwd)":/go/src/github.com/ghjnut/pingwave \
  -w /go/src/github.com/ghjnut/pingwave \
  golang:1.11-alpine3.7 sh
