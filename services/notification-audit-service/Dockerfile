# Этап 1: сборка бинарника
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o notification-audit-service ./cmd/app

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/notification-audit-service .

EXPOSE 8084

CMD ["./notification-audit-service"]