# Booking Functionality Implementation

## Что было добавлено

Реализован полный функционал для записи пользователей на стирку (laundry) и кухню (kitchen).

### Созданные файлы

1. **Proto definitions:**
   - `proto/laundry/laundry_service.proto` - определения для сервиса стирки
   - `proto/kitchen/kitchen_service.proto` - определения для сервиса кухни

2. **Repository слой:**
   - `internal/repository/laundry/repository.go` - методы для работы с БД для стирки и кухни

3. **Service слой:**
   - `internal/service/laundry/service.go` - бизнес-логика для бронирования

4. **gRPC серверы:**
   - `internal/grpc/laundry/server.go` - gRPC сервер для стирки
   - `internal/grpc/kitchen/server.go` - gRPC сервер для кухни

5. **Обновленные файлы:**
   - `Makefile` - добавлены команды для генерации proto файлов
   - `internal/app/service/app.go` - интеграция новых сервисов

## Функционал

### Сервис стирки (LaundryService)

- `CreateLaundryBooking` - создание записи на стирку (максимум 2 часа)
- `GetLaundryBookings` - получение всех записей (с фильтрацией по времени)
- `GetUserLaundryBookings` - получение записей конкретного пользователя
- `DeleteLaundryBooking` - удаление записи

### Сервис кухни (KitchenService)

- `CreateKitchenBooking` - создание записи на кухню (максимум 3 часа)
- `GetKitchenBookings` - получение всех записей (с фильтрацией по времени)
- `GetUserKitchenBookings` - получение записей конкретного пользователя
- `DeleteKitchenBooking` - удаление записи

## Валидация

Реализована следующая валидация:

- Проверка на пересечение временных слотов (нельзя забронировать занятое время)
- Ограничение длительности: 2 часа для стирки, 3 часа для кухни (на уровне БД и сервиса)
- Проверка что end_time > start_time
- Проверка прав пользователя при удалении (только владелец может удалить свою запись)

## Следующие шаги для завершения

### 1. Сгенерировать proto файлы

```bash
make proto
```

Или вручную:

```bash
# Для laundry
rm -rf ./generated/proto/laundry
mkdir -p ./generated/proto/laundry
cd ./proto && protoc -I . --go_out=paths=source_relative:../generated/proto --go-grpc_out=paths=source_relative:../generated/proto laundry/*.proto

# Для kitchen
rm -rf ./generated/proto/kitchen
mkdir -p ./generated/proto/kitchen
cd ./proto && protoc -I . --go_out=paths=source_relative:../generated/proto --go-grpc_out=paths=source_relative:../generated/proto kitchen/*.proto
```

### 2. Обновить app.go после генерации proto

После генерации proto файлов, в `internal/app/service/app.go`:

1. Раскомментируйте импорты:

```go
laundryProto "dormitory-helper-service/generated/proto/laundry"
kitchenProto "dormitory-helper-service/generated/proto/kitchen"
```

2. Раскомментируйте регистрацию сервисов:

```go
laundryProto.RegisterLaundryServiceServer(grpcServer, laundryGrpcServer)
kitchenProto.RegisterKitchenServiceServer(grpcServer, kitchenGrpcServer)
```

3. Удалите временные строки с `_ =`

### 3. Обновить серверы после генерации proto

В файлах `internal/grpc/laundry/server.go` и `internal/grpc/kitchen/server.go`:

1. Замените placeholder типы на сгенерированные из proto
2. Импортируйте сгенерированные пакеты:

```go
laundryProto "dormitory-helper-service/generated/proto/laundry"
```

3. Используйте `laundryProto.UnimplementedLaundryServiceServer` вместо локального placeholder

### 4. Запуск и тестирование

После выполнения шагов выше:

```bash
go mod tidy
go run cmd/app/app.go
```

## Примеры использования

### Создание записи на стирку

```go
req := &laundry.CreateLaundryBookingRequest{
    UserId: 1,
    StartTime: timestamppb.New(time.Now().Add(1 * time.Hour)),
    EndTime: timestamppb.New(time.Now().Add(2 * time.Hour)),
}
```

### Получение всех записей

```go
req := &laundry.GetLaundryBookingsRequest{
    StartTime: timestamppb.New(time.Now()),
    EndTime: timestamppb.New(time.Now().Add(24 * time.Hour)),
}
```

### Удаление записи

```go
req := &laundry.DeleteLaundryBookingRequest{
    BookingId: 1,
    UserId: 1,
}
```

## Структура базы данных

Используются существующие таблицы из миграции `20251020134950_init.sql`:

- `laundry_bookings` - записи на стирку
- `kitchen_bookings` - записи на кухню

Обе таблицы имеют:

- `id` - уникальный идентификатор
- `user_id` - ссылка на пользователя
- `start_time` - время начала
- `end_time` - время окончания
- Constraints на валидацию времени
