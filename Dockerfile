FROM golang:1.20.2-alpine3.17

ADD . /build
WORKDIR /build

RUN go build -o rpid

EXPOSE 8095

CMD ["./rpid", "--config", "config/config.yml"]