package handlers

import (
	"errors"

	"apiservice/internal/middleware"
	"apiservice/internal/store"

	"github.com/gofiber/fiber/v2"
)

// Handlers holds the dependencies shared across route handlers.
type Handlers struct {
	users *store.Users
}

// New constructs the handler set.
func New(users *store.Users) *Handlers {
	return &Handlers{users: users}
}

// Register wires routes onto the router. `protected` is a group that already has
// the Cognito auth middleware applied.
func (h *Handlers) Register(app *fiber.App, protected fiber.Router) {
	app.Get("/health", health)
	protected.Get("/me", h.me)
	protected.Get("/users/me", h.getUser)
}

// health is an unauthenticated liveness probe.
func health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// me echoes the raw verified token claims.
func (h *Handlers) me(c *fiber.Ctx) error {
	claims, ok := middleware.Claims(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "no claims")
	}
	return c.JSON(fiber.Map{
		"sub":      claims["sub"],
		"username": claims["username"],
		"scope":    claims["scope"],
	})
}

// getUser returns the stored profile for the authenticated subject.
func (h *Handlers) getUser(c *fiber.Ctx) error {
	sub := currentSub(c)
	if sub == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "no subject")
	}
	u, err := h.users.Get(c.UserContext(), sub)
	if errors.Is(err, store.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "user profile not found")
	}
	if err != nil {
		return err
	}
	return c.JSON(u)
}

func currentSub(c *fiber.Ctx) string {
	claims, ok := middleware.Claims(c)
	if !ok {
		return ""
	}
	sub, _ := claims["sub"].(string)
	return sub
}
