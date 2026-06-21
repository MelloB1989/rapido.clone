package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"apiservice/internal/config"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// userContextKey is the Fiber locals key under which the verified token claims
// are stored for downstream handlers.
const userContextKey = "user"

// Cognito returns a Fiber middleware that validates the `Authorization: Bearer <jwt>`
// header against the configured Cognito User Pool.
//
// On success the parsed jwt.MapClaims are stored in c.Locals("user").
func Cognito(cfg config.Config) (fiber.Handler, error) {
	if cfg.CognitoUserPoolID == "" {
		return nil, fmt.Errorf("cognito middleware: COGNITO_USER_POOL_ID is required")
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", cfg.Region, cfg.CognitoUserPoolID)
	jwksURL := issuer + "/.well-known/jwks.json"

	// keyfunc fetches and caches the JWKS, refreshing keys on rotation.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("cognito middleware: load JWKS from %s: %w", jwksURL, err)
	}

	return func(c *fiber.Ctx) error {
		raw, err := bearerToken(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		token, err := jwt.Parse(
			raw,
			jwks.Keyfunc,
			jwt.WithIssuer(issuer),
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}

		// Cognito access tokens carry token_use=access; reject id tokens used as bearers.
		if use, _ := claims["token_use"].(string); use != "access" {
			return fiber.NewError(fiber.StatusUnauthorized, "expected an access token")
		}

		// Access tokens use "client_id"; only validate when configured.
		if cfg.CognitoClientID != "" {
			if cid, _ := claims["client_id"].(string); cid != cfg.CognitoClientID {
				return fiber.NewError(fiber.StatusUnauthorized, "unexpected client")
			}
		}

		c.Locals(userContextKey, claims)
		return c.Next()
	}, nil
}

// Claims returns the verified token claims attached by the Cognito middleware.
func Claims(c *fiber.Ctx) (jwt.MapClaims, bool) {
	claims, ok := c.Locals(userContextKey).(jwt.MapClaims)
	return claims, ok
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", fmt.Errorf("malformed authorization header")
	}
	return parts[1], nil
}
