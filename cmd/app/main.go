package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/pavel97go/subscriptions/internal/config"
	httpapi "github.com/pavel97go/subscriptions/internal/http"
	"github.com/pavel97go/subscriptions/internal/repo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg := config.Load()
	if os.Getenv("MIGRATIONS_DIR") == "" {
		_ = os.Setenv("MIGRATIONS_DIR", "./migrations")
	}
	r, err := repo.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to init repo: %v", err)
	}
	defer r.Close()
	app := fiber.New(fiber.Config{
		AppName:               "Subscriptions Service",
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		DisableStartupMessage: false,
	})
	h := httpapi.NewHandler(r)
	httpapi.Setup(app, h)
	go func() {
		addr := ":" + cfg.AppPort
		log.Printf("listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			if err.Error() != "server closed" {
				log.Fatalf("fiber.Listen: %v", err)
			}
		}
	}()
	<-ctx.Done()
	log.Println("shutdown signal received")
	if err := app.Shutdown(); err != nil {
		log.Printf("fiber shutdown error: %v", err)
	}
	r.Close()
	log.Println("server stopped gracefully")
}
