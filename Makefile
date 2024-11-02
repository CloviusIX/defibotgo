BINARY_NAME=main
DOCKER_IMAGE_NAME=defibotharvest
DOCKER_IMAGE_TAG=latest

all: setup

build:
	go build -o $(BINARY_NAME)

setup:
	go mod tidy

run: build
	./$(BINARY_NAME)

rundev:
	go run main.go

test:
	go test ./tests/... -v

clean:
	go clean
	rm -f ./$(BINARY_NAME)

lint:
	go fmt ./...
	go vet ./...

docker/build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

docker/run:
	docker run -d --env-file .env --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

.PHONY: $(shell grep -E '^([a-zA-Z_-]|\/)+:' $(MAKEFILE_LIST) | awk -F':' '{print $$2}' | sed 's/:.*//')
