# Краткая инструкция по запуску

## Шаг 1: Установите необходимые инструменты

Убедитесь, что у вас установлены:

- protoc (Protocol Buffers compiler)
- protoc-gen-go
- protoc-gen-go-grpc

Для установки плагинов:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Шаг 2: Сгенерируйте proto файлы

```bash
make proto
```

Или запустите вручную команды из Makefile.

## Шаг 3: Обновите internal/app/service/app.go

После генерации proto файлов:

1. Раскомментируйте импорты laundryProto и kitchenProto
2. Раскомментируйте строки с RegisterLaundryServiceServer и RegisterKitchenServiceServer
3. Удалите временные строки с `_ = laundryGrpcServer` и `_ = kitchenGrpcServer`

## Шаг 4: Обновите gRPC серверы

В файлах `internal/grpc/laundry/server.go` и `internal/grpc/kitchen/server.go`:

1. Замените placeholder структуры на импорты из сгенерированных proto пакетов
2. Обновите `UnimplementedLaundryServiceServer` и `UnimplementedKitchenServiceServer`

## Шаг 5: Запустите приложение

```bash
go mod tidy
go run cmd/app/app.go
```

## API методы

### Laundry Service (стирка)

- `CreateLaundryBooking(user_id, start_time, end_time)` - создать запись (макс 2 часа)
- `GetLaundryBookings(start_time?, end_time?)` - получить все записи
- `GetUserLaundryBookings(user_id)` - получить записи пользователя
- `DeleteLaundryBooking(booking_id, user_id)` - удалить запись

### Kitchen Service (кухня)

- `CreateKitchenBooking(user_id, start_time, end_time)` - создать запись (макс 3 часа)
- `GetKitchenBookings(start_time?, end_time?)` - получить все записи
- `GetUserKitchenBookings(user_id)` - получить записи пользователя
- `DeleteKitchenBooking(booking_id, user_id)` - удалить запись

## Особенности реализации

✅ Проверка на пересечение временных слотов  
✅ Ограничение длительности (2ч для стирки, 3ч для кухни)  
✅ Валидация времени начала и окончания  
✅ Проверка прав при удалении  
✅ Транзакции для целостности данных  
✅ Использование pgx для работы с PostgreSQL
