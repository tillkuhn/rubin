# syntax = docker/dockerfile:1.2

# single stage version of Dockerfile, expects binaries to be built outside docker build
FROM gcr.io/distroless/static:nonroot@sha256:6732c3975d97fac664a5ed15a81a5915e023a7b5a7b58195e733c60b8dc7e684 AS prd
ARG PLATFORM="linux/amd64"
COPY /app/${PLATFORM}/rubin /rubin
COPY /app/${PLATFORM}/polly /polly
# RUN chmod ugo+x /rubin /polly can't run in distroless :-)
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
