# use alpine as base image
# build the image using go in a separate image
# copy the go binary to the final image

FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY config_daemon/go.mod /app/go.mod
COPY config_daemon/go.sum /app/go.sum

RUN go mod download

COPY config_daemon/. .

RUN go build -o /app/main .

FROM alpine:3.14

COPY --from=builder /app/main /app/main

RUN mkdir /configs

# set the working directory in the container
WORKDIR /configs

EXPOSE 51280/udp

CMD ["/app/main"]
