BINARY_NAME=main
DOCKER_IMAGE_NAME=defibotharvest
DOCKER_IMAGE_TAG=latest

all: setup

build:
	go build -o $(BINARY_NAME)

setup:
	go mod tidy

run: build
	APP_ENV=development ./$(BINARY_NAME) -chain=$(CHAIN) -protocol=$(PROTOCOL) -pool=$(POOL)

rundev:
	APP_ENV=development go run main.go -chain=$(CHAIN) -protocol=$(PROTOCOL) -pool=$(POOL)

test:
	APP_ENV=test go test ./tests/... -v

clean:
	go clean
	rm -f ./$(BINARY_NAME)

lint:
	go fmt ./...
	go vet ./...

docker/build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

# update -chain flag as you want
docker/run:
	docker run -d --env-file .env --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) -chain=$(CHAIN) -protocol=$(PROTOCOL) -pool=$(POOL)

.PHONY: $(shell grep -E '^([a-zA-Z_-]|\/)+:' $(MAKEFILE_LIST) | awk -F':' '{print $$2}' | sed 's/:.*//')
