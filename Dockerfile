FROM golang:1.17-buster as build

RUN mkdir -p /valet
WORKDIR /valet
COPY ./ /valet

# Install default-mysql-client for DB creation and migrations
RUN apt-get update && apt-get install -y default-mysql-client

RUN make build

FROM alpine as prod

RUN mkdir -p /valet
WORKDIR /valet

COPY --from=build /valet/valetd /valet/

