FROM golang:1.17-buster

RUN mkdir -p /valet
WORKDIR /valet
COPY ./ /valet

RUN make build
