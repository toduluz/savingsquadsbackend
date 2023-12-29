# This target builds the application
build:
	@go build -o bin/savingsquadsbackend ./cmd

# This target starts the production environment
docker-prod-up:
	docker-compose -f docker-compose.prod.yml up --build

# This target stops the production environment
docker-prod-down:
	docker-compose -f docker-compose.prod.yml down

# This target runs the integration tests and cleans up afterwards
docker-integration-test:
	docker-compose -f docker-compose.integration.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.integration.yml down --volumes --remove-orphans
	docker rmi savingsquadsbackend-app

# This target runs the handler unit tests and cleans up afterwards
docker-unit-test-handler:
	docker-compose -f docker-compose.unit-handler.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.unit-handler.yml down --volumes --remove-orphans
	docker rmi savingsquadsbackend-app

# This target runs the data unit tests and cleans up afterwards
docker-unit-test-data:
	docker-compose -f docker-compose.unit-data.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.unit-data.yml down --volumes --remove-orphans
	docker rmi savingsquadsbackend-app

# This target runs all tests: handler unit tests, data unit tests, and integration tests
docker-test-handler-data-integration: docker-unit-test-handler docker-unit-test-data docker-integration-test

# This target displays help information
help:
	@echo "Available targets:"
	@echo "  build                           - Builds the application"
	@echo "  docker-prod-up                  - Starts the production environment"
	@echo "  docker-prod-down                - Stops the production environment"
	@echo "  docker-integration-test         - Runs the integration tests and cleans up afterwards"
	@echo "  docker-unit-test-handler        - Runs the handler unit tests and cleans up afterwards"
	@echo "  docker-unit-test-data           - Runs the data unit tests and cleans up afterwards"
	@echo "  docker-test-handler-data-integration - Runs all tests: handler unit tests, data unit tests, and integration tests"

.PHONY: build run docker-prod-up docker-prod-down docker-integration-test docker-unit-test-handler docker-unit-test-data docker-test-handler-data-integration help