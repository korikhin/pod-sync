# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.20 AS build-stage
WORKDIR /build

ENV USER=vortex
ENV UID=10001
ARG TARGETOS
ARG TARGETARCH

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN go env -w GOCACHE=/go-cache \
    && go env -w GOMODCACHE=/gomod-cache

COPY ./go.* ./
RUN --mount=type=cache,target=/gomod-cache go mod download

COPY . .
RUN --mount=type=cache,target=/go-cache \
    --mount=type=cache,target=/gomod-cache \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags "-s -w -extldflags '-static'" -o ./service ./cmd/service

FROM alpine:3.20 AS release-stage
WORKDIR /service

COPY --from=build-stage /etc/passwd /etc/passwd
COPY --from=build-stage /etc/group /etc/group
COPY --from=build-stage /build/service ./service

# Metadata
LABEL org.opencontainers.image.description="Client Algorithms Synchronization Service"
LABEL org.opencontainers.image.source="https://github.com/korikhin/vortex-assignment"
LABEL org.opencontainers.image.licenses="MIT"

USER vortex:vortex

EXPOSE 8080
ENTRYPOINT ["./service"]
