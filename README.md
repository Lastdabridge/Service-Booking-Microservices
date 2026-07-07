# Service Booking Microservices

Система микросервисов для управления бронированием услуг на Go с событийно-ориентированной архитектурой (Event-Driven) на базе Apache Kafka. Проект состоит из API Gateway и четырёх независимых сервисов (Auth, Booking, Catalog, Notifications & Audit), взаимодействующих между собой асинхронно через Kafka.




## Стек технологий

- **Язык:** Go 1.25 / 1.26
- **HTTP-фреймворк:** [Gin](https://github.com/gin-gonic/gin)
- **База данных:** PostgreSQL 17
- **ORM:** GORM
- **Драйвер PostgreSQL:** [pgx/v5](https://github.com/jackc/pgx)
- **Брокер сообщений:** Apache Kafka 3.7.0 (KRaft, без ZooKeeper), клиент — [segmentio/kafka-go](https://github.com/segmentio/kafka-go)
- **UI для Kafka:** [Kafka UI](https://github.com/kafbat/kafka-ui)
- **Аутентификация:** JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt))
- **Валидация:** go-playground/validator
- **Конфигурация:** godotenv (.env)
- **Контейнеризация:** Docker, Docker Compose
- **Архитектура:** микросервисы, API Gateway, Event-Driven (Kafka)

## Архитектура
```
┌──────────────────────────────────────────────────────────────────┐
│                          API Gateway                             │
│                       (Port 8080)                                │
│  - JWT Validation & Token Processing                             │
│  - Reverse Proxy для микросервисов                               │
│  - Injection of X-User-ID и X-User-Role headers                  │
└────┬────────────┬──────────────┬──────────────┬──────────────────┘
     │            │              │              │
     ▼            ▼              ▼              ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐
│   Auth   │ │ Booking  │ │ Catalog  │ │ Notifications &  │
│ Service  │ │ Service  │ │ Service  │ │    Audit Service │
│ :8081    │ │ :8082    │ │ :8083    │ │      :8084       │
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────────┬─────────┘
     │            │            │                │
     └────────────┴────────────┴────────────────┘
              │
    ┌─────────▼──────────────┐
    │  Apache Kafka 3.7.0    │
    │   (Message Broker)     │
    │    Port 9092           │
    │                        │
    │ Topics:                │
    │ • users.events         │
    │ • bookings.events      │
    │ • catalog.events       │
    └────────┬───────────────┘
             │
    ┌────────┴──────────┐
    ▼                   ▼
┌─────────────┐   ┌──────────────────┐
│ PostgreSQL  │   │ PostgreSQL       │
│  :5433      │   │ :5434            │
│ (Auth DB)   │   │(Notifications DB)│
└─────────────┘   └──────────────────┘

Kafka UI: http://localhost:8090
```
---

## Установка и запуск

Требуется установленный [Docker](https://docs.docker.com/get-docker/) и [Docker Compose](https://docs.docker.com/compose/install/).

```bash
git clone https://github.com/Lastdabridge/Service-Booking-Microservices.git
cd Service-Booking-Microservices
docker-compose up --build
```

После запуска будут доступны:

| Сервис                        | Адрес                   |
|--------------------------------|--------------------------|
| API Gateway                    | http://localhost:8080   |
| Auth Service                   | http://localhost:8081   |
| Booking Service                | http://localhost:8082   |
| Catalog Service                | http://localhost:8083   |
| Notifications & Audit Service  | http://localhost:8084   |
| Kafka UI                       | http://localhost:8090   |

Не забудьте создать `.env`-файлы для каждого сервиса (по образцу переменных, используемых в `docker-compose.yml`), если они ещё не настроены.

---

## Участники разработки

- [Lastdabridge](https://github.com/Lastdabridge)
- [Veoler](https://github.com/Veoler)
- [Sheickh](https://github.com/Sheickh)
- [Nsa1d](https://github.com/Nsa1d)
