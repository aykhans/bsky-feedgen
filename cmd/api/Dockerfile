FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY ../../pkg ./pkg
COPY ../../cmd/api ./cmd/api

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o api ./cmd/api

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app

COPY --from=builder /src/api .

ENTRYPOINT ["/app/api"]
