# syntax = docker/dockerfile:1.2

# get modules, if they don't change the cache can be used for faster builds
FROM golang:1.19@sha256:7ffa70183b7596e6bc1b78c132dbba9a6e05a26cd30eaa9832fecad64b83f029 AS base
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
# disable, let docker buildx action handle the platform
# ENV GOARCH=amd64
WORKDIR /src
# avoid go.* (sonar security issue)
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# build th application
FROM base AS build
# temp mount all files instead of loading into image with COPY
# temp mount module cache
# temp mount go build cache

# Build arguments for this image (used as -X args in ldflags)
# goreleaser defaults: '-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser'
# go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" \
ARG APP_VERSION=""
ARG APP_COMMIT=""
ARG APP_DATE=""
ARG APP_BUILT_BY=""

RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s \
      -X 'main.version=${APP_VERSION}' \
      -X 'main.commit=${APP_COMMIT}' \
      -X 'main.date=${APP_DATE}' \
      -X 'main.builtBy=${APP_BUILT_BY}' \
      -extldflags '-static'" \
    -o /app/ ./...

# Import the binary from build stage
FROM gcr.io/distroless/static:nonroot@sha256:55c636171053dbc8ae07a280023bd787d2921f10e569f3e319f1539076dbba11 AS prd
COPY --from=build /app/rubin /rubin
COPY --from=build /app/polly /polly
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
