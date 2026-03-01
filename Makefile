.PHONY: build lint tidy test run stop

build:
	docker build -t drl-api-server .

lint:
	go vet ./...

tidy:
	go fmt ./...

# The integration test is run separately from the unit tests because it flushes the Redis cache.
test:
	docker compose -f docker-compose.test.yml run --build --rm test
	docker compose -f docker-compose.test.yml run --rm integration-test
	docker compose -f docker-compose.test.yml down

run:
	docker compose up --build -d

stop:
	docker compose down
