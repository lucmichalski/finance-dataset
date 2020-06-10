FROM golang:alpine

RUN apk update && \
    apk add --no-cache nano jq bash git && \
    go get -v github.com/gdamore/govisor/...

CMD ["/bin/bash"]
