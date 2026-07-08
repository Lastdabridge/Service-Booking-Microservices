# Service Booking Microservices

Микросервисная backend-система онлайн-записи на услуги.

Проект разрабатывается на Go и демонстрирует работу с микросервисной архитектурой, REST API, PostgreSQL, Kafka, JWT-аутентификацией, авторизацией по ролям, API Gateway, Docker и командной разработкой.

---

# 1. Описание проекта

Система позволяет пользователям записываться на услуги.

Пример предметной области:

* барбершоп;
* клиника;
* автосервис;
* салон красоты;
* учебный центр;
* спортивный клуб.

В системе есть пользователи с разными ролями:

* `client` — клиент;
* `admin` — администратор;
* `specialist` — специалист.

Клиент может зарегистрироваться, войти в систему, посмотреть услуги и специалистов, создать запись и получить уведомление.

Администратор может создавать услуги, специалистов, расписание и смотреть audit log.

Специалист может смотреть свои записи и менять их статус.

---

# 2. Основной бизнес-сценарий

Главный сценарий проекта:

```text
1. Пользователь регистрируется.
2. Пользователь логинится и получает JWT.
3. Админ создаёт услугу.
4. Админ создаёт специалиста.
5. Админ задаёт расписание специалиста.
6. Клиент выбирает услугу и специалиста.
7. Клиент создаёт запись на свободное время.
8. Booking Service создаёт запись.
9. Booking Service публикует Kafka-событие booking.created.
10. Notification + Audit Service читает событие.
11. Notification + Audit Service создаёт уведомление клиенту.
12. Notification + Audit Service сохраняет событие в audit log.
13. Клиент видит уведомление.
14. Админ видит событие в audit log.
```

---

# 3. Архитектура проекта

Проект состоит из 4 основных сервисов:

```text
1. Gateway + Auth Service - Саид-Рахьман
2. Catalog Service - Шейх
3. Booking Service - Ислам
4. Notification + Audit Service - Ибрагим
```

Общая схема:

```text
Client / Postman / Frontend
          |
          v
   Gateway + Auth Service
          |
          | HTTP / REST
          v
------------------------------------------------
|                  |                           |
v                  v                           v
Catalog Service    Booking Service             Notification + Audit Service
|                  |                           |
v                  v                           v
catalog_db         booking_db                  notification_db / audit_db

Все сервисы также взаимодействуют через Kafka.
```

---

# 4. Технологии

В проекте используются:

```text
Go
Gin
PostgreSQL
GORM
Kafka
JWT
Docker
Docker Compose
Redis, если успеете
```

---

# 5. Структура проекта

Рекомендуемая структура:

```text
service-booking-microservices/
  docker-compose.yml
  README.md

  services/
    gateway-auth-service/
      cmd/
      internal/
      Dockerfile
      README.md
      .env.example

    catalog-service/
      cmd/
      internal/
      Dockerfile
      README.md
      .env.example

    booking-service/
      cmd/
      internal/
      Dockerfile
      README.md
      .env.example

    notification-audit-service/
      cmd/
      internal/
      Dockerfile
      README.md
      .env.example
```

Внутри каждого сервиса желательно использовать такую структуру:

```text
cmd/
  app/
    main.go

internal/
  config/
  transport/
  service/
  repository/
  model/
  dto/
  kafka/
```

---

# 6. Сервисы

---

# 6.1. Gateway + Auth Service

## Назначение

Gateway + Auth Service отвечает за:

```text
1. Регистрацию пользователей.
2. Логин пользователей.
3. Хранение пользователей.
4. Хэширование паролей.
5. Выдачу JWT.
6. Проверку JWT.
7. Проксирование запросов в другие сервисы.
8. Передачу X-User-ID и X-User-Role во внутренние сервисы.
```

---

## Основные эндпоинты

```http
POST /auth/register
POST /auth/login
GET /auth/me
PATCH /users/:id/role
```

---

## Регистрация

```http
POST /auth/register
```

Пример тела запроса:

```json
{
  "name": "Adam",
  "email": "adam@example.com",
  "password": "12345678"
}
```

После регистрации пользователь получает роль:

```text
client
```

---

## Логин

```http
POST /auth/login
```

Пример тела запроса:

```json
{
  "email": "adam@example.com",
  "password": "12345678"
}
```

Пример ответа:

```json
{
  "access_token": "jwt_token_here"
}
```

---

## Передача данных пользователя во внутренние сервисы

Gateway после проверки JWT должен передавать во внутренние сервисы headers:

```http
X-User-ID: 15
X-User-Role: client
```

Дополнительно желательно передавать:

```http
X-Request-ID: request-uuid
```

---

## Авторизация внутри Auth Service

Auth Service сам проверяет права на свои эндпоинты.

```text
POST /auth/register      — public
POST /auth/login         — public
GET /auth/me             — authenticated
PATCH /users/:id/role    — admin
```

Важно:

```text
Gateway проверяет JWT.
Auth Service проверяет, имеет ли пользователь право выполнить действие внутри Auth Service.
```

---

## База данных Auth Service

Таблица `users`:

```text
id
name
email
password_hash
role
created_at
updated_at
```

Дополнительно можно сделать таблицу `security_events`:

```text
id
user_id
event_type
payload
created_at
```

---

## Kafka producer Auth Service

Auth Service публикует события в topic:

```text
users.events
```

События:

```text
user.registered
user.logged_in
user.role_changed
```

Пример `user.registered`:

```json
{
  "event": "user.registered",
  "user_id": 1,
  "email": "adam@example.com",
  "role": "client",
  "created_at": "2026-06-09T12:00:00Z"
}
```

---

## Kafka consumer Auth Service

Auth Service слушает topic:

```text
audit.events
```

Событие:

```text
suspicious.activity.detected
```

Минимальная логика:

```text
1. Получить событие.
2. Залогировать событие.
3. Сохранить событие в security_events, если таблица реализована.
```

---

# 6.2. Catalog Service

## Назначение

Catalog Service отвечает за справочную информацию:

```text
1. Услуги.
2. Специалистов.
3. Связь специалистов с услугами.
4. Расписание специалистов.
```

Booking Service использует данные Catalog Service, чтобы проверить, можно ли создать запись.

---

## Основные эндпоинты

### Услуги

```http
GET /services
POST /services
PATCH /services/:id
DELETE /services/:id
```

### Специалисты

```http
GET /specialists
POST /specialists
PATCH /specialists/:id
DELETE /specialists/:id
```

### Связь специалиста и услуг

```http
POST /specialists/:id/services
GET /specialists/:id/services
```

### Расписание

```http
POST /specialists/:id/schedule
GET /specialists/:id/schedule
```

### Internal API для Booking Service

```http
GET /internal/services/:id
GET /internal/specialists/:id
GET /internal/specialists/:id/services
GET /internal/specialists/:id/schedule
```

---

## Авторизация внутри Catalog Service

Catalog Service сам проверяет роль пользователя по headers:

```http
X-User-ID
X-User-Role
```

Правила доступа:

```text
GET /services                    — authenticated
GET /specialists                 — authenticated
GET /specialists/:id/services    — authenticated
GET /specialists/:id/schedule    — authenticated

POST /services                   — admin
PATCH /services/:id              — admin
DELETE /services/:id             — admin

POST /specialists                — admin
PATCH /specialists/:id           — admin
DELETE /specialists/:id          — admin

POST /specialists/:id/services   — admin
POST /specialists/:id/schedule   — admin
```

Если прав недостаточно:

```http
403 Forbidden
```

Пример ответа:

```json
{
  "error": "access denied"
}
```

---

## База данных Catalog Service

Таблица `services`:

```text
id
title
description
duration_minutes
price
is_active
created_at
updated_at
```

Таблица `specialists`:

```text
id
name
description
is_active
created_at
updated_at
```

Таблица `specialist_services`:

```text
id
specialist_id
service_id
created_at
```

Таблица `specialist_schedules`:

```text
id
specialist_id
weekday
start_time
end_time
created_at
updated_at
```

Таблица `catalog_events`, если реализуете хранение входящих событий:

```text
id
event_type
payload
created_at
```

---

## Kafka producer Catalog Service

Catalog Service публикует события в topic:

```text
catalog.events
```

События:

```text
service.created
service.updated
service.deleted
specialist.created
specialist.updated
specialist.deleted
specialist.service_attached
specialist.schedule_updated
```

Пример `service.created`:

```json
{
  "event": "service.created",
  "service_id": 3,
  "title": "Стрижка",
  "duration_minutes": 60,
  "price": 1500,
  "created_at": "2026-06-09T12:00:00Z"
}
```

---

## Kafka consumer Catalog Service

Catalog Service слушает topic:

```text
booking.events
```

События:

```text
booking.created
booking.cancelled
```

Минимальная логика:

```text
1. Получить событие.
2. Залогировать событие.
3. Сохранить событие в catalog_events.
```

Улучшенная логика:

```text
1. Обновлять статистику популярности услуг.
2. Считать количество записей по специалистам.
3. Хранить занятые слоты.
```

---

# 6.3. Booking Service

## Назначение

Booking Service отвечает за основную бизнес-логику записи на услугу:

```text
1. Создание записи.
2. Отмена записи.
3. Изменение статуса записи.
4. Проверка свободного времени.
5. Проверка пересечений записей.
6. Работа с транзакциями.
```

---

## Основные эндпоинты

```http
POST /appointments
GET /appointments/my
GET /appointments/all
GET /appointments/specialist/:id
PATCH /appointments/:id/status
DELETE /appointments/:id
```

---

## Создание записи

```http
POST /appointments
```

Пример тела запроса:

```json
{
  "service_id": 3,
  "specialist_id": 5,
  "start_time": "2026-06-15T14:00:00Z"
}
```

Booking Service должен сам рассчитать `end_time` по длительности услуги.

Если услуга длится 60 минут:

```text
start_time: 2026-06-15T14:00:00Z
end_time:   2026-06-15T15:00:00Z
```

---

## Обязательные проверки при создании записи

Booking Service должен проверить:

```text
1. Пользователь авторизован.
2. Роль пользователя — client.
3. Услуга существует.
4. Услуга активна.
5. Специалист существует.
6. Специалист активен.
7. Специалист оказывает выбранную услугу.
8. Специалист работает в выбранный день.
9. Выбранное время входит в рабочее расписание специалиста.
10. Выбранное время не находится в прошлом.
11. У специалиста нет другой активной записи на это время.
12. Запись создаётся в транзакции.
```

Активные статусы:

```text
created
confirmed
```

Отменённые записи не блокируют слот.

---

## Проверка пересечений записей

Нельзя создать запись, если она пересекается с другой активной записью специалиста.

Пример уже существующей записи:

```text
14:00 - 15:00
```

Нельзя создать:

```text
14:30 - 15:30
13:30 - 14:30
14:00 - 15:00
```

Можно создать:

```text
13:00 - 14:00
15:00 - 16:00
```

---

## Авторизация внутри Booking Service

Booking Service сам проверяет роль пользователя по headers:

```http
X-User-ID
X-User-Role
```

Правила доступа:

```text
POST /appointments                 — client
GET /appointments/my               — client
GET /appointments/all              — admin
GET /appointments/specialist/:id   — admin, specialist
PATCH /appointments/:id/status     — admin, specialist
DELETE /appointments/:id           — client, only own appointment
```

Дополнительное правило:

```text
client может отменить только свою запись
```

То есть при отмене нужно проверить:

```text
appointment.client_id == X-User-ID
```

Если клиент пытается отменить чужую запись:

```http
403 Forbidden
```

---

## База данных Booking Service

Таблица `appointments`:

```text
id
client_id
specialist_id
service_id
start_time
end_time
status
created_at
updated_at
```

Статусы:

```text
created
confirmed
cancelled
completed
```

Таблица `booking_catalog_events`, если реализуете хранение входящих событий:

```text
id
event_type
payload
created_at
```

Для сильной версии можно добавить read-model таблицы:

```text
catalog_services_cache
catalog_specialists_cache
catalog_specialist_services_cache
catalog_schedules_cache
```

---

## Kafka producer Booking Service

Booking Service публикует события в topic:

```text
booking.events
```

События:

```text
booking.created
booking.cancelled
booking.status_changed
booking.completed
```

Пример `booking.created`:

```json
{
  "event": "booking.created",
  "booking_id": 101,
  "client_id": 15,
  "specialist_id": 5,
  "service_id": 3,
  "start_time": "2026-06-15T14:00:00Z",
  "end_time": "2026-06-15T15:00:00Z",
  "created_at": "2026-06-09T12:00:00Z"
}
```

---

## Kafka consumer Booking Service

Booking Service слушает topic:

```text
catalog.events
```

События:

```text
service.created
service.updated
service.deleted
specialist.created
specialist.updated
specialist.deleted
specialist.service_attached
specialist.schedule_updated
```

Минимальная логика:

```text
1. Получить событие.
2. Залогировать событие.
3. Сохранить событие в booking_catalog_events.
```

Улучшенная логика:

```text
1. Хранить локальную копию услуг.
2. Хранить локальную копию специалистов.
3. Хранить локальную копию расписания.
4. Использовать локальную копию при создании записи.
```

---

# 6.4. Notification + Audit Service

## Назначение

Notification + Audit Service отвечает за:

```text
1. Создание уведомлений пользователям.
2. Хранение уведомлений.
3. Отметку уведомлений как прочитанных.
4. Хранение audit log.
5. Сохранение важных событий системы.
6. Обнаружение подозрительной активности, если успеете.
```

---

## Основные эндпоинты

### Уведомления

```http
GET /notifications/my
PATCH /notifications/:id/read
```

### Audit

```http
GET /audit/events
GET /audit/events/:id
```

---

## Авторизация внутри Notification + Audit Service

Notification + Audit Service сам проверяет роль пользователя по headers:

```http
X-User-ID
X-User-Role
```

Правила доступа:

```text
GET /notifications/my          — authenticated
PATCH /notifications/:id/read  — authenticated, only own notification

GET /audit/events              — admin
GET /audit/events/:id          — admin
```

Дополнительное правило:

```text
Пользователь может читать и изменять только свои уведомления.
```

То есть нужно проверить:

```text
notification.user_id == X-User-ID
```

Audit log доступен только администратору.

---

## База данных Notification + Audit Service

Таблица `notifications`:

```text
id
user_id
type
title
message
is_read
created_at
updated_at
```

Таблица `audit_logs`:

```text
id
event_type
actor_id
entity_type
entity_id
source_service
payload
created_at
```

---

## Kafka consumer Notification + Audit Service

Сервис слушает topics:

```text
users.events
catalog.events
booking.events
notifications.events
gateway.events
```

События:

```text
user.registered
user.logged_in

service.created
service.updated
service.deleted
specialist.created
specialist.updated
specialist.schedule_updated

booking.created
booking.cancelled
booking.status_changed
booking.completed

notification.created
notification.read
notification.failed

access.denied
```

Логика:

```text
1. Получить событие.
2. Сохранить событие в audit_logs.
3. Если событие требует уведомления — создать notification.
```

Примеры уведомлений:

```text
user.registered      -> "Добро пожаловать!"
booking.created      -> "Ваша запись успешно создана."
booking.cancelled    -> "Ваша запись была отменена."
booking.completed    -> "Ваша запись завершена."
```

---

## Kafka producer Notification + Audit Service

Сервис публикует события в topics:

```text
notifications.events
audit.events
```

События в `notifications.events`:

```text
notification.created
notification.read
notification.failed
```

События в `audit.events`:

```text
audit.logged
suspicious.activity.detected
```

Пример `notification.created`:

```json
{
  "event": "notification.created",
  "notification_id": 88,
  "user_id": 15,
  "type": "booking_created",
  "source_event": "booking.created",
  "created_at": "2026-06-09T12:00:01Z"
}
```

Пример `audit.logged`:

```json
{
  "event": "audit.logged",
  "audit_id": 501,
  "source_event": "booking.created",
  "source_service": "booking-service",
  "created_at": "2026-06-09T12:00:02Z"
}
```

---

# 7. Kafka topics

В проекте используются topics:

```text
users.events
catalog.events
booking.events
notifications.events
audit.events
gateway.events
```

---

## Кто куда пишет

```text
Gateway + Auth Service           -> users.events
Gateway                          -> gateway.events
Catalog Service                  -> catalog.events
Booking Service                  -> booking.events
Notification + Audit Service     -> notifications.events
Notification + Audit Service     -> audit.events
```

---

## Кто что читает

```text
Gateway + Auth Service           <- audit.events
Catalog Service                  <- booking.events
Booking Service                  <- catalog.events
Notification + Audit Service     <- users.events
Notification + Audit Service     <- catalog.events
Notification + Audit Service     <- booking.events
Notification + Audit Service     <- notifications.events
Notification + Audit Service     <- gateway.events
```

---

# 7. Авторизация: общий принцип

В системе используется двухуровневая схема:

```text
1. Gateway проверяет JWT.
2. Каждый микросервис сам проверяет роль пользователя для своих эндпоинтов.
```

Gateway отвечает за вопрос:

```text
Кто этот пользователь?
```

Микросервис отвечает за вопрос:

```text
Можно ли этому пользователю выполнить это действие?
```

---

## Что делает Gateway

Gateway должен:

```text
1. Принять запрос от клиента.
2. Проверить, публичный эндпоинт или защищённый.
3. Если эндпоинт защищённый — проверить JWT.
4. Извлечь user_id и role из JWT.
5. Передать запрос в нужный сервис.
6. Добавить X-User-ID и X-User-Role.
```

---

## Что делает микросервис

Каждый микросервис должен:

```text
1. Прочитать X-User-ID.
2. Прочитать X-User-Role.
3. Проверить, разрешено ли этой роли выполнить действие.
4. Вернуть 403 Forbidden, если прав нет.
```

---

## Пример

Клиент пытается создать услугу:

```http
POST /services
```

Gateway:

```text
1. Проверяет JWT.
2. Видит, что пользователь авторизован.
3. Передаёт запрос в Catalog Service.
4. Передаёт X-User-Role: client.
```

Catalog Service:

```text
1. Получает запрос.
2. Проверяет правило для POST /services.
3. Видит, что нужна роль admin.
4. Видит, что пришла роль client.
5. Возвращает 403 Forbidden.
```

---

# 8. Базы данных

Каждый сервис должен иметь свою базу данных.

Допустимые варианты:

```text
auth_db
catalog_db
booking_db
notification_db
```

Важное правило:

```text
Сервис не должен напрямую читать таблицы другого сервиса.
```

Неправильно:

```text
Booking Service напрямую читает таблицу services из catalog_db.
```

Правильно:

```text
Booking Service получает данные через HTTP-запрос в Catalog Service.
```

Или в более сильной версии:

```text
Booking Service слушает catalog.events и хранит свою локальную read-model копию.
```

---

# 11. Docker Compose

Проект должен запускаться одной командой:

```bash
docker compose up --build
```

В `docker-compose.yml` должны быть:

```text
gateway-auth-service
catalog-service
booking-service
notification-audit-service
postgres
kafka
redis, если используется
```

---

# 12. Redis

Redis не обязателен для минимальной версии, но желателен.

---

# 13. Транзакции

Транзакции обязательны в Booking Service при создании записи.

---

# 14. Ошибки API

Примеры ошибок:

```json
{
  "error": "invalid email or password"
}
```

```json
{
  "error": "access denied"
}
```

```json
{
  "error": "service not found"
}
```

```json
{
  "error": "specialist is not available at this time"
}
```

```json
{
  "error": "appointment time overlaps with existing appointment"
}
```

---

# 15. Логирование

Каждый сервис должен логировать:

```text
1. Старт сервиса.
2. Подключение к БД.
3. Ошибки подключения к БД.
4. Входящие HTTP-запросы.
5. Ошибки Kafka producer.
6. Ошибки Kafka consumer.
7. Важные бизнес-действия.
```

---

# 16. Документация API

Команда должна подготовить готоую коллекцию запросов в Postman:

Минимальный набор запросов:

```text
register
login
get me
create service
create specialist
attach service to specialist
create specialist schedule
create appointment
get my appointments
cancel appointment
get notifications
get audit events
```

---

# 17. Git workflow

Команда должна работать через Git.

Правила:

```text
1. Основная ветка — main.
2. Каждый участник работает в своей feature-ветке.
3. Нельзя пушить всё напрямую в main.
4. Изменения вливаются через Pull Request / Merge Request.
5. Коммиты должны быть осмысленными.
```

Примеры веток:

```text
feature/gateway-auth-service
feature/catalog-service
feature/booking-service
feature/notification-audit-service
feature/docker-compose
```

Примеры коммитов:

```text
feat: add user registration
feat: add kafka producer for booking events
fix: prevent appointment time overlap
docs: update README
```

---

# 18. Чекпоинты готовности

## Обязательный минимум

```text
1. 4 сервиса.
2. Gateway.
3. Регистрация и логин.
4. JWT.
5. Проверка ролей внутри каждого микросервиса.
6. PostgreSQL.
7. REST API.
8. Kafka producer и consumer у каждого сервиса.
9. Создание услуги.
10. Создание специалиста.
11. Создание расписания.
12. Создание записи.
13. Уведомление после создания записи.
14. Audit log важных событий.
15. Docker Compose.
16. Проверка пересечения записей.
17. README.
```

---

## Было бы хорошо

```text
2. Отмена записи.
3. Статусы записи.
4. Redis-кэш.
5. Rate limit в Gateway.
6. Postman collection.
7.  Graceful shutdown.
```

---

---

# 19. Демонстрационный сценарий для защиты

На защите нужно показать:

```text
1. Запуск проекта через docker compose up --build.
2. Регистрацию клиента.
3. Логин клиента и получение JWT.
4. Логин админа.
5. Создание услуги админом.
6. Создание специалиста админом.
7. Привязку услуги к специалисту.
8. Создание расписания специалиста.
9. Создание записи клиентом.
10. Публикацию booking.created в Kafka.
11. Создание уведомления клиенту.
12. Получение уведомлений клиентом.
13. Просмотр audit log админом.
14. Попытку запрещённого действия.
15. Получение 403 Forbidden.
```

---

# 20. Критерии оценки

## Архитектура

```text
1. Есть разделение на микросервисы.
2. Каждый сервис отвечает за свою область.
3. Сервисы не читают чужие таблицы напрямую.
4. Есть Gateway.
5. Есть Kafka-взаимодействие.
```

---

## Авторизация

```text
1. Gateway проверяет JWT.
2. Gateway передаёт X-User-ID и X-User-Role.
3. Каждый микросервис проверяет роли самостоятельно.
4. Client не может выполнять admin-действия.
5. Client не может отменить чужую запись.
6. Не-admin не может смотреть audit log.
```

---

## Backend-логика

```text
1. Регистрация работает.
2. Логин работает.
3. Создание услуги работает.
4. Создание специалиста работает.
5. Создание расписания работает.
6. Создание записи работает.
7. Нельзя записаться на занятое время.
8. Нельзя создать запись в прошлое.
```

---

## Kafka

```text
1. Каждый сервис имеет producer.
2. Каждый сервис имеет consumer.
3. События публикуются в правильные topics.
4. События читаются и обрабатываются.
5. Ошибки Kafka логируются.
```

---

## База данных

```text
1. Таблицы спроектированы корректно.
2. Есть связи между сущностями.
3. Используется GORM.
4. Есть миграции или auto migration.
5. Booking Service использует транзакцию.
```

---

## Инфраструктура

```text
1. Проект запускается через Docker Compose.
2. Все сервисы поднимаются.
3. PostgreSQL поднимается.
4. Kafka поднимается.
5. Redis поднимается, если используется.
```

---

## Командная работа

```text
1. Каждый участник внёс вклад.
2. Есть понятная история коммитов.
3. Есть ветки.
4. Есть README.
5. Каждый участник может объяснить свой сервис.
```

---

# 21. Важные ограничения

```text
1. Не делать монолит.
2. Не обращаться напрямую к таблицам другого сервиса.
3. Не смешивать ответственность сервисов.
4. Auth Service не должен заниматься уведомлениями.
5. Catalog Service не должен создавать записи.
6. Booking Service не должен управлять пользователями.
7. Notification Service не должен менять записи.
8. Gateway проверяет JWT, но не заменяет проверку ролей внутри сервисов.
9. Каждый микросервис обязан сам проверять права доступа к своим эндпоинтам.
```

---

# 22. Итоговая цель

В результате должна получиться микросервисная backend-система, где:

```text
1. Пользователь регистрируется.
2. Получает JWT.
3. Делает запросы через Gateway.
4. Gateway проверяет JWT.
5. Микросервисы сами проверяют роли.
6. Admin создаёт услуги, специалистов и расписание.
7. Client создаёт запись.
8. Booking Service проверяет бизнес-логику.
9. Booking Service публикует Kafka-событие.
10. Notification + Audit Service создаёт уведомление.
11. Notification + Audit Service сохраняет audit log.
12. Все сервисы работают отдельно.
13. Каждый участник команды понимает свой сервис и может объяснить его на защите.
```

Главная задача проекта — показать не просто CRUD, а понимание микросервисной архитектуры, событийного взаимодействия через Kafka, JWT-аутентификации, авторизации внутри сервисов и командной backend-разработки на Go.
