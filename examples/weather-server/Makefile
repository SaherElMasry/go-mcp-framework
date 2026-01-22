# Weather MCP Server Makefile

.PHONY: help run build test clean docker-build docker-run

# Variables
BINARY_NAME=weather-server
DOCKER_IMAGE=weather-mcp-server
VERSION?=2.0.0

help: ## Show this help message
	@echo "Weather MCP Server - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run the server (requires WEATHER_API_KEY env var)
	@if [ -z "$$WEATHER_API_KEY" ]; then \
		echo "❌ WEATHER_API_KEY not set!"; \
		echo "   Get your key: https://www.weatherapi.com/signup.aspx"; \
		echo "   Then: export WEATHER_API_KEY=your-key-here"; \
		exit 1; \
	fi
	@echo "🚀 Starting Weather MCP Server..."
	go run cmd/server/main.go

build: ## Build the server binary
	@echo "🔨 Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) cmd/server/main.go
	@echo "✅ Built: ./$(BINARY_NAME)"

test: ## Run tests
	@echo "🧪 Running tests..."
	go test ./... -v

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf logs/
	@echo "✅ Cleaned"

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "✅ Built: $(DOCKER_IMAGE):$(VERSION)"

docker-run: ## Run Docker container
	@if [ -z "$$WEATHER_API_KEY" ]; then \
		echo "❌ WEATHER_API_KEY not set!"; \
		exit 1; \
	fi
	@echo "🐳 Running Docker container..."
	docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 8080:8080 \
		-p 9091:9091 \
		-e WEATHER_V1_API_KEY=$$WEATHER_V1_API_KEY \
		$(DOCKER_IMAGE):latest
	@echo "✅ Container started: $(DOCKER_IMAGE)"
	@echo "   HTTP: http://localhost:8080"
	@echo "   Metrics: http://localhost:9091/metrics"

docker-stop: ## Stop Docker container
	@echo "🛑 Stopping container..."
	docker stop $(DOCKER_IMAGE)
	docker rm $(DOCKER_IMAGE)
	@echo "✅ Container stopped"

logs: ## View server logs (if running in background)
	@tail -f logs/weather-server.log 2>/dev/null || echo "No log file found"

install: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

.DEFAULT_GOAL := help
