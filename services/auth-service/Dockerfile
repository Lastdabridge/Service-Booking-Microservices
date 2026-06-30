# Этап сборки
FROM golang:1.25.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-service ./cmd/app

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/auth-service .
COPY .env .

EXPOSE 8081

CMD ["./auth-service"]