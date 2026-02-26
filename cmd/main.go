package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BountyM/effectiveMobileTestTask/internal/config"
	"github.com/BountyM/effectiveMobileTestTask/internal/handler"
	"github.com/BountyM/effectiveMobileTestTask/internal/logger"
	"github.com/BountyM/effectiveMobileTestTask/internal/repository"
	server "github.com/BountyM/effectiveMobileTestTask/internal/server"
	"github.com/BountyM/effectiveMobileTestTask/internal/service"
	_ "github.com/lib/pq"

	_ "github.com/BountyM/effectiveMobileTestTask/docs"
)

// @title Subscription API
// @version 1.0
// @description API для управления подписками
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
		return
	}
	log := logger.New(cfg.Logger)
	log.Info("Logger initialized")
	db, err := repository.NewPostgresDB(cfg.DB)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		return
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Error("Error occurred on DB connection close", "error", closeErr)
		} else {
			log.Info("Database connection closed successfully")
		}
	}()

	repo := repository.New(db)
	services := service.New(repo)
	handlers := handler.New(services, log) // переименовано для избежания конфликта с пакетом

	srv := &server.Server{}

	// Канал для ошибок от HTTP сервера
	serverErr := make(chan error, 1)

	go func() {
		defer close(serverErr)
		// Формируем правильный адрес с двоеточием
		addr := ":" + cfg.Port
		if err := srv.Run(addr, handlers.InitRoutes()); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	log.Info("Subscription API started")

	// Ожидание сигнала завершения или ошибки сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
		log.Info("Subscription API shutting down")
	case err := <-serverErr:
		log.Error("Server failed with error", "error", err)
		return
	}

	// Graceful shutdown сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Error occurred on server shutdown", "error", err)
	} else {
		log.Info("Server stopped gracefully")
	}
}
