.PHONY: build build-network lint tidy test run stop

build:
	docker build -t drl-api-server .

build-network:
	docker network create drl-network

lint:
	go vet ./...

tidy:
	go fmt ./...

test:
	go test -v -race -cover ./...

run:
	docker run --detach --name drl-redis --network drl-network --rm redis:latest
	docker run --detach --name drl-api-server --network drl-network --publish 8080:8080 --rm drl-api-server

stop:
	docker stop drl-api-server
	docker stop drl-redis
