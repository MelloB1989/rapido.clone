package users

import (
	ie "apiservice/internal/errors"
	"apiservice/internal/middleware"
	"apiservice/internal/shared"
	"apiservice/internal/store/users"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	users *users.Users
}

func CreateUserHandlers(deps *shared.Deps) shared.Handlers {
	us := users.NewUsers(deps)
	return &UserHandler{
		users: us,
	}
}

func (h *UserHandler) Register(app *fiber.App, protected fiber.Router) {
	protected.Get("/me", h.me)
	protected.Get("/users/me", h.getUser)
}

// me echoes the raw verified token claims.
func (h *UserHandler) me(c *fiber.Ctx) error {
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
func (h *UserHandler) getUser(c *fiber.Ctx) error {
	sub := currentSub(c)
	if sub == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "no subject")
	}
	u, err := h.users.Get(c.UserContext(), sub)
	if errors.Is(err, ie.ErrNotFound) {
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
