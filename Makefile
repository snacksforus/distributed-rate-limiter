.PHONY: build run stop

build:
	docker build -t api-server .

run:
	docker run --detach --publish 8080:8080 --rm --name api-server api-server

stop:
	docker stop api-server