# syntax = docker/dockerfile:1.2

# single stage version of Dockerfile, expects binaries to be built outside docker build
FROM gcr.io/distroless/static:nonroot@sha256:6cd937e9155bdfd805d1b94e037f9d6a899603306030936a3b11680af0c2ed58 AS prd
ARG PLATFORM="linux/amd64"
COPY /app/${PLATFORM}/rubin /rubin
COPY /app/${PLATFORM}/polly /polly
# RUN chmod ugo+x /rubin /polly can't run in distroless :-)
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
