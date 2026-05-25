package main

import (
	"context"
	"log"
	nethttp "net/http"
	"os"

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
	logger.Printf("server started on %s", addr)

	if err := nethttp.ListenAndServe(addr, apphttp.NewRouter(subscriptionHandler)); err != nil {
		logger.Fatal(err)
	}
}
