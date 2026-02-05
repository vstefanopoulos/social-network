# This second version exists so that the outer make file can work, cause this file fixes the working directory issue

FROM golang:1.25-alpine AS base

RUN apk add --no-cache git build-base

WORKDIR /app

COPY backend/go.mod backend/go.sum ./

RUN go mod download
