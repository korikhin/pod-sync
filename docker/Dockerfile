# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.20 AS build-stage
WORKDIR /build
COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN go env -w GOCACHE=/go-cache \
    && go env -w GOMODCACHE=/gomod-cache

RUN --mount=type=cache,target=/go-cache \
    --mount=type=cache,target=/gomod-cache \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags "-s -w -extldflags '-static'" -o ./service ./cmd/service

RUN apk add --no-cache upx \
    && upx --best --lzma ./service

FROM alpine:3.20 AS release-stage
WORKDIR /service
COPY --from=build-stage /build/service ./service

ENV UID=10001
ENV USER=vortex

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Metadata
LABEL org.opencontainers.image.description="Algorithm Synchronization Service"
LABEL org.opencontainers.image.source="https://github.com/korikhin/vortex-assignment"
LABEL org.opencontainers.image.licenses="MIT"

EXPOSE 8080
USER vortex:vortex
ENTRYPOINT ["./service"]
