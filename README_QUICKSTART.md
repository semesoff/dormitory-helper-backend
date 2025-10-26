# Dormitory Helper Backend - Quick Start

REST API сервис для управления записями на стирку и кухню в общежитии.

## 🚀 Быстрый запуск

### 1. Установка

```bash
# Установите protoc
sudo apt install -y protobuf-compiler  # Ubuntu/Debian
brew install protobuf                   # macOS

# Установите Go tools и googleapis
make install-proto-tools
make setup-googleapis

# Установите зависимости
go mod tidy
```

### 2. Генерация proto файлов

```bash
make proto
```

### 3. Раскомментируйте код

После генерации proto раскомментируйте импорты и регистрации в:

- `internal/app/service/app.go`
- `internal/app/gateway/gateway.go`

Подробности в `JWT_AND_GATEWAY_SUMMARY.md`.

### 4. Запуск

```bash
# Вариант 1: Оба сервиса одной командой
make run-all

# Вариант 2: В разных терминалах
make run-backend  # Terminal 1 - gRPC backend
make run-gateway  # Terminal 2 - HTTP gateway
```

### 5. Тестирование

```bash
chmod +x test_api.sh
./test_api.sh
```

## 📚 API Endpoints

**Base URL:** `http://localhost:8081`

### Authentication

- `POST /api/v1/auth/check` - Get JWT token

### Laundry

- `POST /api/v1/laundry/bookings` - Create booking
- `GET /api/v1/laundry/bookings` - Get all bookings
- `GET /api/v1/laundry/bookings/my` - Get my bookings
- `DELETE /api/v1/laundry/bookings/{id}` - Delete booking

### Kitchen

- `POST /api/v1/kitchen/bookings` - Create booking
- `GET /api/v1/kitchen/bookings` - Get all bookings
- `GET /api/v1/kitchen/bookings/my` - Get my bookings
- `DELETE /api/v1/kitchen/bookings/{id}` - Delete booking

## 📖 Документация

- `JWT_AND_GATEWAY_SUMMARY.md` - Полная документация по JWT и Gateway
- `GATEWAY_SETUP.md` - Детальная настройка Gateway
- `BOOKING_IMPLEMENTATION.md` - Функционал бронирования

## 🔧 Технологии

- **Backend:** Go + gRPC
- **Gateway:** grpc-gateway (gRPC → HTTP/REST)
- **Auth:** JWT tokens
- **Database:** PostgreSQL
- **Middleware:** CORS, Logging

## 🏗️ Архитектура

```
Client → HTTP Gateway (8081) → gRPC Backend → PostgreSQL
```

## 📝 Пример использования

```bash
# 1. Получить токен
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/check \
  -H "Content-Type: application/json" \
  -d '{"token": ""}' | jq -r '.token')

# 2. Создать запись на стирку
curl -X POST http://localhost:8081/api/v1/laundry/bookings \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$TOKEN\",
    \"start_time\": \"2025-10-27T10:00:00Z\",
    \"end_time\": \"2025-10-27T12:00:00Z\"
  }"

# 3. Получить мои записи
curl -X GET "http://localhost:8081/api/v1/laundry/bookings/my?token=$TOKEN"
```

## 🔐 Безопасность

Все эндпоинты используют JWT токены:

- User ID извлекается из токена (нельзя подделать)
- Только владелец может удалить свою запись
- Автоматическая валидация токена на каждом запросе

## ⚙️ Конфигурация

Создайте `.env` файл:

```env
SERVER_HOST=localhost
SERVER_PORT=50051
JWT_SECRET_KEY=your-secret-key-here

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=dormitory_db
DATABASE_USER=postgres
DATABASE_PASSWORD=password
DRIVER=postgres
```

## 🐳 Docker (опционально)

```bash
# Запуск базы данных
cd dev/database
docker-compose up -d

# Миграции
goose postgres "postgresql://user:password@localhost:5432/dormitory_db" up
```

## 🛠️ Разработка

```bash
# Форматирование кода
go fmt ./...

# Проверка
go vet ./...

# Тесты
go test ./...

# Пересборка proto
make proto
```

## 📦 Структура проекта

```
├── cmd/
│   ├── app/        # gRPC backend
│   └── gateway/    # HTTP gateway
├── internal/
│   ├── app/
│   ├── grpc/       # gRPC servers
│   ├── repository/ # Database layer
│   ├── service/    # Business logic
│   └── utils/      # Helpers
├── proto/          # Protobuf definitions
├── generated/      # Generated code
└── migrations/     # Database migrations
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

MIT License

---

**Подробная документация:** См. `JWT_AND_GATEWAY_SUMMARY.md`
