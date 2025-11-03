package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Setup(app *fiber.App, h *Handler) {
	app.Use(logger.New())

	app.Post("/subscriptions", h.Create)
	app.Get("/subscriptions", h.List)
	app.Get("/subscriptions/:id", h.Get)
	app.Put("/subscriptions/:id", h.Update)
	app.Delete("/subscriptions/:id", h.Delete)

	app.Get("/subscriptions/summary", h.Summary)
}
