# Google API Proto Files

Этот каталог содержит необходимые proto файлы из googleapis для поддержки HTTP аннотаций (google/api/annotations.proto) в gRPC-gateway.

## Установка

Файлы должны быть загружены из репозитория googleapis:

```bash
# Клонировать googleapis (если еще не сделано)
git clone https://github.com/googleapis/googleapis.git third_party/googleapis

# Или добавить как submodule
git submodule add https://github.com/googleapis/googleapis.git third_party/googleapis
```

## Структура

```
third_party/
└── googleapis/
    └── google/
        └── api/
            ├── annotations.proto
            ├── http.proto
            └── ...
```

## Использование в protoc

При компиляции proto файлов нужно добавить путь к googleapis:

```bash
protoc -I . -I third_party/googleapis \
  --go_out=paths=source_relative:../generated/proto \
  --go-grpc_out=paths=source_relative:../generated/proto \
  --grpc-gateway_out=paths=source_relative:../generated/proto \
  user/*.proto
```
