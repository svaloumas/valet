FROM golang:1.17-buster as build

RUN mkdir -p /valet
WORKDIR /valet
COPY ./ /valet

RUN make build

FROM alpine as prod

RUN mkdir -p /valet
WORKDIR /valet

COPY --from=build /valet/valetd /valet/

