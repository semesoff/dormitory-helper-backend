package serviceApp

import (
	"context"
	"dormitory-helper-service/internal/config"
	userServer "dormitory-helper-service/internal/grpc/user"
	userRepository "dormitory-helper-service/internal/repository/user"
	userService "dormitory-helper-service/internal/service/user"
	"fmt"
	"log"
	"net"

	userProto "dormitory-helper-service/generated/proto/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func Run() {
	ctx := context.Background()

	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Load()

	// Инициализация базы данных
	dbURL := fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DatabaseConfig.Driver,
		cfg.DatabaseConfig.User,
		cfg.DatabaseConfig.Password,
		cfg.DatabaseConfig.Host,
		cfg.DatabaseConfig.Port,
		cfg.DatabaseConfig.DBName,
	)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("failed to parse database config: %v", err)
	}

	// Настройка пула соединений
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}
	defer db.Close()

	// Проверка подключения к базе данных
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database")

	// Инициализация репозиториев
	userRepo := userRepository.NewRepository()

	// Инициализация сервисов
	userServ := userService.NewService(userRepo, db, cfg.ServerConfig.JWTSecretKey)

	// Инициализация gRPC сервера
	grpcServer := grpc.NewServer()

	// Регистрация сервисов
	userGrpcServer := userServer.NewServer(userServ)
	userProto.RegisterUserServiceServer(grpcServer, userGrpcServer)

	// Запуск gRPC сервера
	address := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", address, err)
	}

	log.Printf("gRPC server is running on %s", address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
