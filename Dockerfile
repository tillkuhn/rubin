# syntax = docker/dockerfile:1.2

# single stage version of Dockerfile, expects binaries to be built outside docker build
FROM gcr.io/distroless/static:nonroot@sha256:8dd8d3ca2cf283383304fd45a5c9c74d5f2cd9da8d3b077d720e264880077c65 AS prd
ARG PLATFORM="linux/amd64"
COPY /app/${PLATFORM}/rubin /rubin
COPY /app/${PLATFORM}/polly /polly
# RUN chmod ugo+x /rubin /polly can't run in distroless :-)
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
