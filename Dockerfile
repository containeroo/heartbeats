FROM golang:1.24-alpine as builder

ARG LDFLAGS="-s -w"
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY go.mod go.sum ./
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .

RUN go build -ldflags="$LDFLAGS" -o /heartbeats ./main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /heartbeats ./
USER nonroot:nonroot

ENTRYPOINT ["./heartbeats"]
