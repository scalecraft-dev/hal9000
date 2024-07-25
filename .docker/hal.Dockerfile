FROM golang:1.21.5-bullseye

COPY . /hal

WORKDIR /hal

RUN go mod tidy
