FROM golang:1.25-alpine AS base

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
