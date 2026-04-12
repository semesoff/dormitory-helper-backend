package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"dormitory-helper-backend/internal/config"
	"dormitory-helper-backend/internal/service"
	"dormitory-helper-backend/internal/storage"
	httptransport "dormitory-helper-backend/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	db, err := storage.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := storage.ApplyMigrations(ctx, db, cfg.MigrationsPath); err != nil {
		log.Fatal(err)
	}

	repo := storage.NewRepository(db)
	svc := service.New(repo, service.NewTokenManager(cfg.JWTSecret, cfg.TokenTTL))
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httptransport.NewServer(svc, cfg.CORSEnabled).Handler(),
	}

	go func() {
		log.Printf("api listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
