# Build the frontend assets
FROM node:24-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Prepare writable dirs for distroless runtime
FROM alpine:3.23 AS prep
RUN mkdir -p /outfs/work /outfs/tmp \
  # Change group ownership of /work and /tmp to GID 0 (root group),
  # because OpenShift assigns containers a random UID but always includes them in group 0.
  && chgrp -R 0 /outfs/work /outfs/tmp \
  # Give group 0 read/write/execute (X only applies to dirs or already-executable files).
  # This makes the dirs writable by arbitrary UIDs in group 0.
  && chmod -R g+rwX /outfs/work /outfs/tmp \
  # Set the setgid bit on the dirs so that any new files/dirs created inside
  # will inherit group 0 instead of the creator's primary group.
  && chmod g+s /outfs/work /outfs/tmp

# Build the manager binary
FROM golang:1.25 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG LDFLAGS="-s -w"
ENV CGO_ENABLED=0

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY internal/ internal
COPY --from=frontend /web/dist web/dist

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -ldflags="$LDFLAGS" -a -o heartbeats main.go

RUN mkdir -p /outfs/work /outfs/tmp \
  # Change group ownership of /work and /tmp to GID 0 (root group),
  # because OpenShift assigns containers a random UID but always includes them in group 0.
  && chgrp -R 0 /outfs/work /outfs/tmp \
  # Give group 0 read/write/execute (X only applies to dirs or already-executable files).
  # This makes the dirs writable by arbitrary UIDs in group 0.
  && chmod -R g+rwX /outfs/work /outfs/tmp \
  # Set the setgid bit on the dirs so that any new files/dirs created inside
  # will inherit group 0 instead of the creator's primary group.
  && chmod g+s /outfs/work /outfs/tmp


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=prep /outfs/work /work
COPY --from=prep /outfs/tmp  /tmp
ENV HOME=/tmp
WORKDIR /work
COPY --from=builder /workspace/heartbeats .
USER 65532:65532
ENTRYPOINT ["/heartbeats"]


