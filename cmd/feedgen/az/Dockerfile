FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY ../../pkg ./pkg
COPY ../../cmd/feedgen/az ./cmd/feedgen/az

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o feedgen ./cmd/feedgen/az

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app

COPY --from=builder /src/feedgen .

ENTRYPOINT ["/app/feedgen"]
