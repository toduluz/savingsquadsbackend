build:
	@go build -o bin/savingsquadsbackend

run: build
	./bin/savingsquadsbackend