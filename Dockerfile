FROM golang:1.17-buster as build

RUN mkdir -p /valet
WORKDIR /valet
COPY ./ /valet

RUN make build-linux
