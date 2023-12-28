IMAGE_NAME := savingsquadsbackend
CONTAINER_NAME := savingsquadsbackend-container

build:
	@go build -o bin/$(IMAGE_NAME) ./cmd/api

run: build
	./bin/$(IMAGE_NAME)

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run: docker-build
	docker run -p 3000:4000 --name $(CONTAINER_NAME) $(IMAGE_NAME)

docker-stop:
	docker stop $(CONTAINER_NAME)

docker-clean: docker-stop
	docker rm $(CONTAINER_NAME)

.PHONY: build run docker-build docker-run docker-stop docker-clean
