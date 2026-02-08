.PHONY: build run stop

build:
	docker build -t drl-api-server .

build-network:
	docker network create drl-network

run:
	docker run --detach --name drl-redis --network drl-network --rm redis:latest
	docker run --detach --name drl-api-server --network drl-network --publish 8080:8080 --rm drl-api-server

stop:
	docker stop drl-api-server
	docker stop drl-redis
