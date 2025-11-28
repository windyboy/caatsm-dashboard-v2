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

FROM node:${NODE_VERSION}-bookworm AS frontend
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/caatsm ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=builder /app/bin/caatsm ./caatsm
COPY --from=assets /assets/public ./public
# Copy SvelteKit build output
# Deno adapter outputs to .svelte-kit/deno, copy the entire .svelte-kit directory
COPY --from=frontend /frontend/.svelte-kit ./frontend/.svelte-kit
COPY config ./config
EXPOSE 3002
ENTRYPOINT ["./caatsm","-config","/app/config/config.toml"]

