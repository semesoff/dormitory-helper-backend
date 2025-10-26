# Итоговая документация: JWT Authentication + HTTP Gateway

## Что было реализовано

### 1. JWT Authentication (Безопасность) ✅

Все API эндпоинты теперь используют JWT токены вместо передачи `user_id`:

**До:**

```json
{
  "user_id": 123,
  "start_time": "2025-10-27T10:00:00Z",
  "end_time": "2025-10-27T12:00:00Z"
}
```

**После:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "start_time": "2025-10-27T10:00:00Z",
  "end_time": "2025-10-27T12:00:00Z"
}
```

#### Преимущества

- ✅ **Безопасность**: user_id нельзя подделать, он извлекается из подписанного токена
- ✅ **Аутентификация**: автоматическая проверка токена на каждом запросе
- ✅ **Валидация**: токен проверяется на валидность и срок действия

### 2. HTTP Gateway (REST API) ✅

Создан отдельный gateway сервис, который:

- Принимает HTTP/REST запросы на порту **8081**
- Преобразует их в gRPC вызовы к backend серверу
- Возвращает JSON ответы клиентам

**Архитектура:**

```
Client (Browser/App) 
    ↓ HTTP/REST (port 8081)
Gateway Service
    ↓ gRPC (internal)
Backend Service (port from .env)
    ↓
Database
```

#### Компоненты

1. **Proto файлы с HTTP аннотациями**
   - `proto/user/user.proto`
   - `proto/laundry/laundry_service.proto`
   - `proto/kitchen/kitchen_service.proto`

2. **Gateway Application**
   - `cmd/gateway/gateway.go` - точка входа
   - `internal/app/gateway/gateway.go` - HTTP сервер с middleware

3. **Middleware**
   - CORS - для работы с frontend
   - Logging - логирование всех запросов
   - Custom header matching - передача заголовков в gRPC

### 3. Обновленные файлы

#### Proto Definitions

- ✅ Заменен `user_id` на `token` во всех request сообщениях
- ✅ Добавлены HTTP аннотации (google.api.http)
- ✅ Определены REST endpoints для всех методов

#### gRPC Servers

- ✅ `internal/grpc/laundry/server.go` - валидация JWT токена
- ✅ `internal/grpc/kitchen/server.go` - валидация JWT токена
- ✅ `internal/utils/grpc/auth.go` - helper для валидации токенов

#### Configuration

- ✅ `Makefile` - команды для генерации proto с grpc-gateway
- ✅ `internal/app/service/app.go` - передача jwtSecret в серверы

### 4. Новая структура проекта

```
cmd/
├── app/          # Backend gRPC сервер
│   └── app.go
└── gateway/      # HTTP Gateway сервер
    └── gateway.go

internal/
├── app/
│   ├── gateway/  # Gateway application
│   │   └── gateway.go
│   └── service/  # Backend application
│       └── app.go
├── grpc/
│   ├── kitchen/
│   │   └── server.go    # JWT validation
│   ├── laundry/
│   │   └── server.go    # JWT validation
│   └── user/
│       └── server.go
└── utils/
    └── grpc/
        └── auth.go      # JWT validation helper

proto/
├── kitchen/
│   └── kitchen_service.proto  # HTTP annotations
├── laundry/
│   └── laundry_service.proto  # HTTP annotations
└── user/
    └── user.proto             # HTTP annotations

third_party/
└── googleapis/   # Google API proto files (для HTTP annotations)
```

## API Endpoints

### User Service

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/check` | Проверка токена / получение нового |

### Laundry Service

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/laundry/bookings` | Создать запись на стирку |
| GET | `/api/v1/laundry/bookings` | Получить все записи |
| GET | `/api/v1/laundry/bookings/my` | Получить мои записи |
| DELETE | `/api/v1/laundry/bookings/{id}` | Удалить запись |

### Kitchen Service

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/kitchen/bookings` | Создать запись на кухню |
| GET | `/api/v1/kitchen/bookings` | Получить все записи |
| GET | `/api/v1/kitchen/bookings/my` | Получить мои записи |
| DELETE | `/api/v1/kitchen/bookings/{id}` | Удалить запись |

## Пошаговая инструкция запуска

### Шаг 1: Установка зависимостей

```bash
# Установите protoc (если еще не установлен)
# Ubuntu/Debian:
sudo apt install -y protobuf-compiler

# Установите Go tools
make install-proto-tools

# Установите googleapis
make setup-googleapis

# Установите зависимости Go
go mod download
go mod tidy
```

### Шаг 2: Генерация proto файлов

```bash
make proto
```

Эта команда создаст:

- `generated/proto/user/*.pb.go` и `*.pb.gw.go`
- `generated/proto/laundry/*.pb.go` и `*.pb.gw.go`
- `generated/proto/kitchen/*.pb.go` и `*.pb.gw.go`

### Шаг 3: Раскомментировать код

После генерации proto раскомментируйте:

**В `internal/app/service/app.go`:**

```go
// Строки 18-20: импорты
laundryProto "dormitory-helper-service/generated/proto/laundry"
kitchenProto "dormitory-helper-service/generated/proto/kitchen"

// Строки 90-94: регистрация
laundryProto.RegisterLaundryServiceServer(grpcServer, laundryGrpcServer)
kitchenProto.RegisterKitchenServiceServer(grpcServer, kitchenGrpcServer)

// Удалите строки с `_ = laundryGrpcServer` и `_ = kitchenGrpcServer`
```

**В `internal/app/gateway/gateway.go`:**

```go
// Строки 15-17: импорты
userProto "dormitory-helper-service/generated/proto/user"
laundryProto "dormitory-helper-service/generated/proto/laundry"
kitchenProto "dormitory-helper-service/generated/proto/kitchen"

// Строки 44-62: регистрация handlers
err := userProto.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
if err != nil {
    log.Fatalf("Failed to register user service handler: %v", err)
}

err = laundryProto.RegisterLaundryServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
if err != nil {
    log.Fatalf("Failed to register laundry service handler: %v", err)
}

err = kitchenProto.RegisterKitchenServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
if err != nil {
    log.Fatalf("Failed to register kitchen service handler: %v", err)
}

// Удалите строку `_ = opts`
```

**В `internal/grpc/laundry/server.go` и `internal/grpc/kitchen/server.go`:**

Замените placeholder типы на импорты из generated proto:

```go
// Вместо локальных placeholder типов используйте:
import laundryProto "dormitory-helper-service/generated/proto/laundry"

// И замените типы в сигнатурах методов
```

### Шаг 4: Запуск сервисов

**Опция A: Запуск в разных терминалах**

Терминал 1:

```bash
make run-backend
# Backend запустится на порту из .env (например, 50051)
```

Терминал 2:

```bash
make run-gateway
# Gateway запустится на порту 8081
```

**Опция B: Запуск всех сервисов одной командой**

```bash
make run-all
```

### Шаг 5: Тестирование

```bash
# Дайте права на выполнение
chmod +x test_api.sh

# Запустите тесты
./test_api.sh
```

Или вручную:

```bash
# Получить токен
curl -X POST http://localhost:8081/api/v1/auth/check \
  -H "Content-Type: application/json" \
  -d '{"token": ""}'

# Создать запись
TOKEN="your_token_here"
curl -X POST http://localhost:8081/api/v1/laundry/bookings \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$TOKEN\",
    \"start_time\": \"2025-10-27T10:00:00Z\",
    \"end_time\": \"2025-10-27T12:00:00Z\"
  }"
```

## Безопасность

### JWT Token Flow

1. **Получение токена**

   ```
   POST /api/v1/auth/check с пустым token
   → Создается новый пользователь
   → Возвращается JWT token
   ```

2. **Использование токена**

   ```
   POST /api/v1/laundry/bookings с token
   → Gateway передает token в gRPC сервер
   → Server валидирует token через ValidateTokenAndGetUserID()
   → Извлекается user_id из токена
   → Выполняется операция от имени этого пользователя
   ```

3. **Защита данных**
   - User не может подделать user_id
   - Только владелец может удалить свою запись
   - Токен подписан секретным ключом (JWT_SECRET_KEY)

### CORS Configuration

По умолчанию разрешены все origins (`*`) для удобства разработки.

**Для production измените в `internal/app/gateway/gateway.go`:**

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Замените * на конкретный домен
        w.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
        // ...
```

## Полезные команды

```bash
# Установка инструментов
make install-proto-tools
make setup-googleapis

# Генерация proto
make proto

# Запуск сервисов
make run-backend      # Backend (gRPC)
make run-gateway      # Gateway (HTTP)
make run-all          # Оба сервиса

# Тестирование
./test_api.sh
```

## Документы

- `GATEWAY_SETUP.md` - подробная инструкция по настройке gateway
- `BOOKING_IMPLEMENTATION.md` - документация по функционалу бронирования
- `test_api.sh` - скрипт для автоматического тестирования API

## Что дальше?

### Рекомендации для production

1. **TLS/HTTPS**
   - Добавить TLS в gateway
   - Использовать защищенное соединение между gateway и backend

2. **Rate Limiting**
   - Добавить middleware для ограничения запросов
   - Защита от DDoS

3. **Monitoring**
   - Prometheus metrics
   - Логирование в ELK stack

4. **Authentication Enhancement**
   - Refresh tokens
   - Token revocation
   - OAuth2 integration

5. **API Documentation**
   - Генерация OpenAPI/Swagger документации
   - API versioning

6. **Container Deployment**
   - Dockerfile для backend и gateway
   - docker-compose.yml для локальной разработки
   - Kubernetes манифесты для production

## Поддержка

Если возникли проблемы, проверьте:

1. Все ли зависимости установлены
2. Сгенерированы ли proto файлы
3. Раскомментированы ли импорты после генерации
4. Запущены ли оба сервиса (backend и gateway)
5. Правильно ли настроен .env файл

Для подробностей смотрите `GATEWAY_SETUP.md`.
