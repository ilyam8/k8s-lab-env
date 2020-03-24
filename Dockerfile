FROM golang:alpine as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o bouncer ./cmd/bouncer

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bouncer /

ENTRYPOINT ["/bouncer"]
