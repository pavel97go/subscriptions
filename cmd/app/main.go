package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pavel97go/subscriptions/internal/config"
	httpapi "github.com/pavel97go/subscriptions/internal/http"
	"github.com/pavel97go/subscriptions/internal/logger"
	"github.com/pavel97go/subscriptions/internal/repo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger.Init()
	log := logger.Log
	log.Info("starting Subscriptions-service")
	cfg := config.Load()
	log.Infof("config loaded (port=%s, db=%s)", cfg.AppPort, cfg.DB.Host)
	if os.Getenv("MIGRATIONS_DIR") == "" {
		_ = os.Setenv("MIGRATIONS_DIR", "./migrations")
	}
	r, err := repo.New(ctx, cfg.DB.DSN)
	if err != nil {
		log.Fatalf("failed to init repo: %v", err)
	}
	defer r.Close()
	app := fiber.New(fiber.Config{
		AppName:      "Subscriptions Service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recovermw.New())

	h := httpapi.NewHandler(r)
	httpapi.Setup(app, h)

	go func() {
		addr := ":" + cfg.AppPort
		log.Infof("listening on %s", addr)
		if err := app.Listen(addr); err != nil && err.Error() != "server closed" {
			log.Fatalf("fiber.Listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Warn("shutdown signal received")

	if err := app.Shutdown(); err != nil {
		log.Errorf("fiber shutdown error: %v", err)
	}

	log.Info("server stopped gracefully")
}
