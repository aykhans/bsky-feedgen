FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY ../../pkg ./pkg
COPY ../../cmd/consumer ./cmd/consumer

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o consumer ./cmd/consumer

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app

COPY --from=builder /src/consumer .

ENTRYPOINT ["/app/consumer"]
