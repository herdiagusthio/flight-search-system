.PHONY: swagger swagger-serve test build run clean

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g internal/api/server.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated successfully!"

# Generate and serve Swagger UI
swagger-serve: swagger
	@echo "Starting server with Swagger UI..."
	@echo "Swagger UI available at: http://localhost:8080/swagger/index.html"
	go run cmd/api/main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Build the application
build:
	go build -o bin/api.exe ./cmd/api

# Run the application
run:
	go run cmd/api/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Install swagger CLI tool
install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

# Validate swagger documentation
swagger-validate:
	@echo "Validating Swagger documentation..."
	@if [ -f "docs/swagger.json" ]; then \
		echo "Swagger documentation is valid"; \
	else \
		echo "Error: Swagger documentation not found. Run 'make swagger' first"; \
		exit 1; \
	fi

# Help command
help:
	@echo "Available commands:"
	@echo "  make swagger         - Generate Swagger documentation"
	@echo "  make swagger-serve   - Generate docs and start server with Swagger UI"
	@echo "  make test            - Run all tests with coverage"
	@echo "  make build           - Build the application"
	@echo "  make run             - Run the application"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install-swagger - Install swag CLI tool"
	@echo "  make swagger-validate- Validate Swagger documentation"
	@echo "  make help            - Show this help message"
