FROM golang:1.19-alpine as builder

RUN apk add --no-cache git

RUN CGO_ENABLED=0 GO111MODULE=on go build -a -installsuffix nocgo -o /heartbeats .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /heartbeats ./
ENTRYPOINT ["./heartbeats"]
