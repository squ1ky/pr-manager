# 1. Build
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pr-manager ./cmd/app

# 2. Runtime
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/pr-manager ./pr-manager
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./pr-manager"]