package main

import (
	"context"
	"log"
	"os"

	"apiservice/internal/config"
	"apiservice/internal/handlers"
	"apiservice/internal/middleware"
	"apiservice/internal/store"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	app := newApp(cfg)

	if onLambda() {
		serveLambda(app)
		return
	}

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

// newApp builds the Fiber app: shared middleware, public routes, and a
// Cognito-protected route group.
func newApp(cfg config.Config) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(recover.New())
	app.Use(logger.New())

	auth, err := middleware.Cognito(cfg)
	if err != nil {
		log.Fatalf("failed to init cognito middleware: %v", err)
	}

	db, err := store.NewClient(context.Background(), cfg.Region)
	if err != nil {
		log.Fatalf("failed to init dynamodb client: %v", err)
	}
	h := handlers.New(store.NewUsers(db, cfg.TableName))

	protected := app.Group("/api", auth)
	h.Register(app, protected)

	return app
}

// serveLambda adapts the Fiber app to API Gateway HTTP API (payload v2.0) events.
func serveLambda(app *fiber.App) {
	adapter := fiberadapter.New(app)
	lambda.Start(func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return adapter.ProxyWithContextV2(ctx, req)
	})
}

// onLambda reports whether the process is running inside the Lambda runtime.
func onLambda() bool {
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
}
