FROM golang:1.25.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o gateway ./cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/gateway .
COPY .env .

EXPOSE 8080

CMD ["./gateway"]