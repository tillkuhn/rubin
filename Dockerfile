# syntax = docker/dockerfile:1.2

# single stage version of Dockerfile, expects binaries to be built outside docker build
FROM gcr.io/distroless/static:nonroot@sha256:188ddfb9e497f861177352057cb21913d840ecae6c843d39e00d44fa64daa51c AS prd
ARG PLATFORM="linux/amd64"
COPY /app/${PLATFORM}/rubin /rubin
COPY /app/${PLATFORM}/polly /polly
# RUN chmod ugo+x /rubin /polly can't run in distroless :-)
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
