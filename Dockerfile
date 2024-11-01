# syntax = docker/dockerfile:1.2

# single stage version of Dockerfile, expects binaries to be built outside docker build
FROM gcr.io/distroless/static:nonroot@sha256:3a03fc0826340c7deb82d4755ca391bef5adcedb8892e58412e1a6008199fa91 AS prd
ARG PLATFORM="linux/amd64"
COPY /app/${PLATFORM}/rubin /rubin
COPY /app/${PLATFORM}/polly /polly
# RUN chmod ugo+x /rubin /polly can't run in distroless :-)
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
