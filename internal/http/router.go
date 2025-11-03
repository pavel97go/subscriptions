package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Setup(app *fiber.App, h *Handler) {
	app.Use(logger.New())

	api := app.Group("/subscriptions")

	api.Get("/", h.List)
	api.Get("/summary", h.Summary)

	api.Post("/", h.Create)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}
