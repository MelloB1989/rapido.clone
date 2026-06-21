# Build the API for AWS Lambda (custom runtime: provided.al2023, arm64).
# The CDK stack in ../../infra packages ./build as the Lambda asset.

BUILD_DIR := build
BINARY    := $(BUILD_DIR)/bootstrap

.PHONY: build run tidy clean

# Cross-compile a static linux/arm64 binary named `bootstrap` (required name
# for the provided.al2* custom runtime).
build:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(BINARY) ./cmd/api

# Run the server locally on $PORT (default 8080).
run:
	go run ./cmd/api

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)
