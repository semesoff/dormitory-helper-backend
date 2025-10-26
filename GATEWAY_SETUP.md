# Gateway Setup Guide

## Архитектура

```
Client (HTTP/REST) -> Gateway (port 8081) -> Backend (gRPC, port указан в .env)
```

Gateway преобразует HTTP запросы в gRPC и обратно, используя grpc-gateway.

## Предварительные требования

1. **Установите protoc** (Protocol Buffers compiler)

   ```bash
   # Ubuntu/Debian
   sudo apt install -y protobuf-compiler
   
   # macOS
   brew install protobuf
   
   # Windows
   # Скачайте с https://github.com/protocolbuffers/protobuf/releases
   ```

2. **Установите Go tools**

   ```bash
   make install-proto-tools
   ```

3. **Клонируйте googleapis**

   ```bash
   make setup-googleapis
   ```

4. **Установите зависимости**

   ```bash
   go mod download
   go mod tidy
   ```

## Генерация proto файлов

```bash
make proto
```

Эта команда сгенерирует:

- `*.pb.go` - Protocol Buffers структуры
- `*_grpc.pb.go` - gRPC серверы и клиенты
- `*.pb.gw.go` - HTTP handlers для grpc-gateway

## Настройка после генерации

После выполнения `make proto` раскомментируйте следующие строки:

### В `internal/app/service/app.go`

```go
laundryProto "dormitory-helper-service/generated/proto/laundry"
kitchenProto "dormitory-helper-service/generated/proto/kitchen"

// И регистрацию сервисов:
laundryProto.RegisterLaundryServiceServer(grpcServer, laundryGrpcServer)
kitchenProto.RegisterKitchenServiceServer(grpcServer, kitchenGrpcServer)
```

### В `internal/app/gateway/gateway.go`

```go
userProto "dormitory-helper-service/generated/proto/user"
laundryProto "dormitory-helper-service/generated/proto/laundry"
kitchenProto "dormitory-helper-service/generated/proto/kitchen"

// И регистрацию handlers:
err := userProto.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
err = laundryProto.RegisterLaundryServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
err = kitchenProto.RegisterKitchenServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
```

### В `internal/grpc/laundry/server.go` и `internal/grpc/kitchen/server.go`

Замените placeholder типы на сгенерированные из proto пакетов.

## Запуск сервисов

### Вариант 1: Запуск по отдельности

**Терминал 1 - Backend (gRPC):**

```bash
make run-backend
# или
go run cmd/app/app.go
```

**Терминал 2 - Gateway (HTTP):**

```bash
make run-gateway
# или
go run cmd/gateway/gateway.go
```

### Вариант 2: Запуск всех сервисов одновременно

```bash
make run-all
```

## API Endpoints

После запуска gateway доступен на `http://localhost:8081`

### User Service

- `POST /api/v1/auth/check` - Проверка аутентификации

### Laundry Service

- `POST /api/v1/laundry/bookings` - Создать запись на стирку
- `GET /api/v1/laundry/bookings` - Получить все записи
- `GET /api/v1/laundry/bookings/my` - Получить мои записи
- `DELETE /api/v1/laundry/bookings/{booking_id}` - Удалить запись

### Kitchen Service

- `POST /api/v1/kitchen/bookings` - Создать запись на кухню
- `GET /api/v1/kitchen/bookings` - Получить все записи
- `GET /api/v1/kitchen/bookings/my` - Получить мои записи
- `DELETE /api/v1/kitchen/bookings/{booking_id}` - Удалить запись

## Примеры запросов

### Проверка аутентификации (получение токена)

```bash
curl -X POST http://localhost:8081/api/v1/auth/check \
  -H "Content-Type: application/json" \
  -d '{"token": ""}'
```

Ответ:

```json
{
  "user_id": 1,
  "username": "user_1234567890",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Создание записи на стирку

```bash
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8081/api/v1/laundry/bookings \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$TOKEN\",
    \"start_time\": \"2025-10-27T10:00:00Z\",
    \"end_time\": \"2025-10-27T11:30:00Z\"
  }"
```

### Получение всех записей на стирку

```bash
curl -X GET "http://localhost:8081/api/v1/laundry/bookings"
```

### Получение моих записей на стирку

```bash
TOKEN="your_jwt_token_here"

curl -X GET "http://localhost:8081/api/v1/laundry/bookings/my?token=$TOKEN"
```

### Удаление записи

```bash
TOKEN="your_jwt_token_here"
BOOKING_ID=1

curl -X DELETE "http://localhost:8081/api/v1/laundry/bookings/$BOOKING_ID?token=$TOKEN"
```

### Создание записи на кухню

```bash
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8081/api/v1/kitchen/bookings \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$TOKEN\",
    \"start_time\": \"2025-10-27T14:00:00Z\",
    \"end_time\": \"2025-10-27T16:30:00Z\"
  }"
```

## Безопасность

### JWT Token

Все эндпоинты, требующие аутентификации, принимают JWT токен:

- В теле запроса (для POST/PUT)
- В query параметре (для GET/DELETE)

Token содержит:

- `user_id` - ID пользователя
- `username` - имя пользователя
- `exp` - время истечения

### CORS

Gateway настроен с поддержкой CORS:

- Разрешены все origins (`*`)
- Методы: GET, POST, PUT, DELETE, OPTIONS
- Headers: Accept, Authorization, Content-Type, X-Request-Id

Для production окружения рекомендуется ограничить `Access-Control-Allow-Origin`.

## Логирование

Gateway логирует все HTTP запросы:

```
2025/10/26 10:15:23 Started POST /api/v1/laundry/bookings
2025/10/26 10:15:23 Completed POST /api/v1/laundry/bookings with 200 in 45.2ms
```

## Troubleshooting

### Ошибка: "Import google/api/annotations.proto not found"

Выполните: `make setup-googleapis`

### Ошибка: "connection refused" в gateway

Убедитесь, что backend (gRPC сервер) запущен первым.

### Ошибка: "invalid token"

Сначала получите новый токен через `/api/v1/auth/check` с пустым token.

### Ошибка компиляции после генерации proto

1. Выполните `go mod tidy`
2. Раскомментируйте импорты в app.go и gateway.go
3. Замените placeholder типы в серверах

## Дополнительно

### OpenAPI/Swagger документация

При необходимости можно сгенерировать OpenAPI спецификацию:

```bash
# Добавьте в Makefile опцию --openapiv2_out
```

### TLS/HTTPS

Для production добавьте TLS в gateway:

```go
httpServer.ListenAndServeTLS("cert.pem", "key.pem")
```

### Rate Limiting

Рекомендуется добавить middleware для rate limiting в gateway.

### Authentication Middleware

Можно добавить middleware для автоматической валидации токена в gateway.
