FROM golang:1.19-alpine as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.sum ./
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /heartbeats

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /heartbeats ./
USER 65532:65532

ENTRYPOINT ["./heartbeats"]
