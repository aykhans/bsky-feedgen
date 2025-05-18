FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY ../../pkg ./pkg
COPY ../../cmd/manager ./cmd/manager

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o manager ./cmd/manager/main.go

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app

COPY --from=builder /src/manager .

ENTRYPOINT ["/app/manager"]
CMD ["help"]
