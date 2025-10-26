package serviceApp

import (
	"context"
	"dormitory-helper-service/internal/config"
	userServer "dormitory-helper-service/internal/grpc/user"
	laundryRepository "dormitory-helper-service/internal/repository/laundry"
	userRepository "dormitory-helper-service/internal/repository/user"
	laundryService "dormitory-helper-service/internal/service/laundry"
	userService "dormitory-helper-service/internal/service/user"
	"fmt"
	"log"
	"net/http"
	"time"

	userProto "dormitory-helper-service/generated/proto/user"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
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
	_ = laundryRepository.NewRepository() // TODO: will be used when laundry/kitchen services are fixed

	// Инициализация сервисов
	userServ := userService.NewService(userRepo, db, cfg.ServerConfig.JWTSecretKey)
	_ = laundryService.NewService(laundryRepository.NewRepository(), db) // TODO: will be used when laundry/kitchen services are fixed

	// Создание HTTP gateway с grpc-gateway
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
	)

	// Инициализация gRPC серверов
	userGrpcServer := userServer.NewServer(userServ)
	// laundryGrpcServer := laundryServer.NewServer(laundryServ, cfg.ServerConfig.JWTSecretKey)
	// kitchenGrpcServer := kitchenServer.NewServer(laundryServ, cfg.ServerConfig.JWTSecretKey)

	// Регистрация сервисов напрямую в gateway (in-process)
	err = userProto.RegisterUserServiceHandlerServer(ctx, mux, userGrpcServer)
	if err != nil {
		log.Fatalf("Failed to register user service handler: %v", err)
	}

	// TODO: Uncomment after fixing laundry and kitchen servers to use generated proto types
	// err = laundryProto.RegisterLaundryServiceHandlerServer(ctx, mux, laundryGrpcServer)
	// if err != nil {
	// 	log.Fatalf("Failed to register laundry service handler: %v", err)
	// }

	// err = kitchenProto.RegisterKitchenServiceHandlerServer(ctx, mux, kitchenGrpcServer)
	// if err != nil {
	// 	log.Fatalf("Failed to register kitchen service handler: %v", err)
	// }

	// HTTP сервер с middleware
	httpAddress := ":8081"
	handler := corsMiddleware(loggingMiddleware(mux))

	httpServer := &http.Server{
		Addr:         httpAddress,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting HTTP server on %s", httpAddress)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// customHeaderMatcher определяет какие заголовки будут переданы в gRPC контекст
func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization", "X-Request-Id", "X-User-Agent":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// corsMiddleware добавляет CORS заголовки
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем CORS заголовки
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-Id")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Обработка preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware логирует все HTTP запросы
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Логируем запрос
		log.Printf("Started %s %s", r.Method, r.RequestURI)

		// Создаем wrapper для ResponseWriter чтобы перехватить status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Вызываем следующий handler
		next.ServeHTTP(wrapped, r)

		// Логируем ответ
		duration := time.Since(start)
		log.Printf("Completed %s %s with %d in %v", r.Method, r.RequestURI, wrapped.statusCode, duration)
	})
}

// responseWriter wrapper для перехвата status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
