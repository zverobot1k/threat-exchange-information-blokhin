package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zverobot1k/threat-service-blokhin/internal/config"
	"github.com/zverobot1k/threat-service-blokhin/internal/handler"
	"github.com/zverobot1k/threat-service-blokhin/internal/repository"
	"github.com/zverobot1k/threat-service-blokhin/internal/router"
	"github.com/zverobot1k/threat-service-blokhin/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := config.OpenDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewGORMThreatRepository(db)
	threatService := service.NewThreatService(repo)
	threatHandler := handler.NewThreatHandler(threatService)

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           router.New(threatHandler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	case <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown failed: %v", err)
		}
	}
}
