# Этап 1: сборка бинарника
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 — статическая линковка, бинарник не зависит от libc хост-системы.
# Без этого на Alpine (musl) иногда возникают проблемы совместимости.
RUN CGO_ENABLED=0 go build -o notification-audit-service ./cmd/app

# Этап 2: минимальный образ для запуска
FROM alpine:3.19

WORKDIR /app

# ca-certificates — нужны если сервис когда-либо будет делать HTTPS-запросы наружу.
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/notification-audit-service .

EXPOSE 8084

CMD ["./notification-audit-service"]