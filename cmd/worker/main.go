package main

import (
	"context"
	"log"
	"time"

	"dormitory-helper-backend/internal/config"
	"dormitory-helper-backend/internal/service"
	"dormitory-helper-backend/internal/storage"
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

	ticker := time.NewTicker(cfg.WorkerInterval)
	defer ticker.Stop()

	log.Printf("worker started with interval %s", cfg.WorkerInterval)
	for {
		cycleCtx, cancel := context.WithTimeout(context.Background(), cfg.WorkerInterval)
		if err := svc.RunWorkerCycle(cycleCtx); err != nil {
			log.Printf("worker cycle failed: %v", err)
		}
		cancel()
		<-ticker.C
	}
}
