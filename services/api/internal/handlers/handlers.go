package handlers

import (
	"apiservice/internal/handlers/users"
	"apiservice/internal/shared"

	"github.com/gofiber/fiber/v2"
)

// Handlers holds the dependencies shared across route handlers.
type RootHandler struct {
	deps *shared.Deps
}

func CreateRootHandler(deps *shared.Deps) shared.Handlers {
	return &RootHandler{
		deps: deps,
	}
}

func (h *RootHandler) Register(app *fiber.App, protected fiber.Router) {
	app.Get("/health", health)

	// Register user handlers
	uh := users.CreateUserHandlers(h.deps)
	uh.Register(app, protected)
}

// health is an unauthenticated liveness probe.
func health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
