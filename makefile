build:
	@go build -o bin/savingsquadsbackend ./cmd/api-server

run: build
	./bin/savingsquadsbackend