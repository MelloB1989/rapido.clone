package shared

import "github.com/gofiber/fiber/v2"

type Handlers interface {
	// Register wires routes onto the router. `protected` is a group that already has
	// the Cognito auth middleware applied.
	Register(app *fiber.App, protected fiber.Router)
}
