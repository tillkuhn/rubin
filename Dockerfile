# syntax = docker/dockerfile:1.2

# single stage version of Dockerfile, expects binaries to be built outside docker build
FROM gcr.io/distroless/static:nonroot@sha256:55c636171053dbc8ae07a280023bd787d2921f10e569f3e319f1539076dbba11 AS prd
ARG PLATFORM="linux/amd64"
COPY --from=build /app/${PLATFORM}/rubin /rubin
COPY --from=build /app/${PLATFORM}/polly /polly
# this is the numeric version of user nonroot:nonroot to check runAsNonRoot in kubernetes
USER 65532:65532
ENTRYPOINT ["/rubin"]
