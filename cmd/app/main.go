package main

import (
	"context"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/config"
	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/db"
	apphttp "github.com/en7ka/Effective_Mobile_testovoe.git/internal/http"
	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/repository/postgres"
	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cfg := config.Load()

	database, err := db.NewPostgres(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal(err)
	}
	defer database.Close()

	repository := postgres.NewSubscriptionRepository(database)
	subscriptionService := service.NewSubscriptionService(repository)
	subscriptionHandler := apphttp.NewSubscriptionHandler(subscriptionService, logger)

	addr := ":" + cfg.HTTPPort
	server := &nethttp.Server{
		Addr:              addr,
		Handler:           apphttp.NewRouter(subscriptionHandler, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Printf("server started on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			logger.Fatalf("listen and serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Println("shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("server shutdown: %v", err)
	}

	logger.Println("server stopped")
}
