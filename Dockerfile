# syntax=docker/dockerfile:1.8

ARG GO_VERSION=1.25
ARG NODE_VERSION=22

FROM golang:${GO_VERSION}-bookworm AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM node:${NODE_VERSION}-bookworm AS assets
WORKDIR /assets
COPY assets ./assets
COPY uno.config.ts package.json package-lock.json ./
RUN npm install
RUN npm run unocss:build

FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/caatsm ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=builder /app/bin/caatsm ./caatsm
COPY --from=assets /assets/public ./public
COPY config ./config
EXPOSE 3002
ENTRYPOINT ["./caatsm","-config","/app/config/config.toml"]

