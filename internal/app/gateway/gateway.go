package gatewayApp

import (
	"context"
	"dormitory-helper-service/internal/config"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// Uncomment these after running `make proto`:
	// userProto "dormitory-helper-service/generated/proto/user"
	// laundryProto "dormitory-helper-service/generated/proto/laundry"
	// kitchenProto "dormitory-helper-service/generated/proto/kitchen"
)

func Run() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Load()

	// Создаем HTTP gateway
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
	)

	// gRPC адрес backend сервера
	grpcAddress := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)

	// Опции подключения к gRPC
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Регистрация User Service
	// Uncomment after running `make proto`:
	// err := userProto.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	// if err != nil {
	// 	log.Fatalf("Failed to register user service handler: %v", err)
	// }

	// Регистрация Laundry Service
	// Uncomment after running `make proto`:
	// err = laundryProto.RegisterLaundryServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	// if err != nil {
	// 	log.Fatalf("Failed to register laundry service handler: %v", err)
	// }

	// Регистрация Kitchen Service
	// Uncomment after running `make proto`:
	// err = kitchenProto.RegisterKitchenServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	// if err != nil {
	// 	log.Fatalf("Failed to register kitchen service handler: %v", err)
	// }

	// Placeholder для подавления ошибки компиляции
	_ = opts

	// HTTP сервер с middleware
	gatewayAddress := ":8081" // Gateway будет слушать на порту 8081
	handler := corsMiddleware(loggingMiddleware(mux))

	httpServer := &http.Server{
		Addr:         gatewayAddress,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting HTTP gateway on %s", gatewayAddress)
	log.Printf("Proxying to gRPC server at %s", grpcAddress)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTP gateway: %v", err)
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
