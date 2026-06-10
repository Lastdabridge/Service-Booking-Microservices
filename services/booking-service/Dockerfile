# --- Этап 1: Сборка приложения ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Кешируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статически скомпилированный бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o booking-service cmd/main.go


# --- Этап 2: Финальный минимальный образ ---
FROM alpine:3.20

WORKDIR /app

# Устанавливаем таймзоны и сертификаты (нужны для HTTPS-запросов и работы со временем)
RUN apk --no-cache add ca-certificates tzdata

# Копируем только скомпилированный файл из первого этапа
COPY --from=builder /app/booking-service .

# Запуск приложения
CMD [ "./booking-service" ]