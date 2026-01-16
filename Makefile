BINARY_NAME=metachan
BUILD_PATH=bin/$(BINARY_NAME)
MAIN_PATH=$(BINARY_NAME)/main.go
ENV_PATH=.env
DOCS_PATH=docs

.PHONY: setup clean build run dev docs docs-serve all

define ensure_setup
	@if [ ! -f $(ENV_PATH) ]; then \
		echo "Running setup first..."; \
		$(MAKE) -s setup; \
	fi
endef

setup:
	@echo "Setting up environment..."
	@go mod download
	@go mod tidy
	@if [ ! -f $(ENV_PATH) ]; then cp .env.example $(ENV_PATH); fi
	@echo "Environment setup complete."

clean:
	@echo "Cleaning up..."
	@rm -rf bin
	@echo "Cleanup complete."

build:
	$(call ensure_setup)
	@echo "Building..."
	@go build -o $(BUILD_PATH) $(MAIN_PATH) || true
	@echo "Build complete."

run:
	$(call ensure_setup)
	@if [ ! -f $(BUILD_PATH) ]; then echo "Binary not found. Building binary..."; $(MAKE) -s build; fi
	@echo "Running..."
	@$(BUILD_PATH) || true

dev:
	$(call ensure_setup)
	@echo "Running in development mode..."
	@go run $(MAIN_PATH) || true

docs:
	@echo "Generating API documentation..."
	@command -v redocly >/dev/null 2>&1 || { echo "Redocly CLI not found. Install with: npm install -g @redocly/cli"; exit 1; }
	@mkdir -p public
	@redocly build-docs $(DOCS_PATH)/openapi.yaml --output public/index.html --title "Metachan API Documentation"
	@echo "Documentation generated at public/index.html"
	@echo "Open public/index.html in your browser to view"

docs-serve: docs
	@echo "Documentation server starting at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@cd public && python3 -m http.server 8080

all: setup clean build run

.SILENT: