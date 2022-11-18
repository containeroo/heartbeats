FROM golang:1.19-alpine as builder

RUN apk add --no-cache git

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /heartbeats


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /heartbeats ./
RUN mkdir /config
ENTRYPOINT ["./heartbeats"]
