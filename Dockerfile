FROM golang:1.4.2

ADD . /go/src/github.com/arvinkulagin/wschat

RUN go get github.com/gorilla/websocket

RUN go install github.com/arvinkulagin/wschat

WORKDIR /go/src/github.com/arvinkulagin/wschat

EXPOSE 8888

ENTRYPOINT wschat