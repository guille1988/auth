CONTAINER_NAME=api

.PHONY: up down migrate seed build-dev build-prod run-dev run-prod update test init

up:
	@if [ ! -f .env ]; then \
		echo ".env file not found, copying .env.example..."; \
		cp .env.example .env; \
	fi
	docker-compose up -d

down:
	docker-compose down

migrate:
	docker exec $(CONTAINER_NAME) go build -o bin/migrate cmd/migrate/main.go
	docker exec $(CONTAINER_NAME) ./bin/migrate

migrate-fresh:
	docker exec $(CONTAINER_NAME) go build -o bin/migrate cmd/migrate/main.go
	docker exec $(CONTAINER_NAME) ./bin/migrate --fresh

seed:
	docker exec $(CONTAINER_NAME) go build -o bin/seed cmd/seed/main.go
	docker exec $(CONTAINER_NAME) ./bin/seed

build-dev:
	docker exec $(CONTAINER_NAME) go build -o bin/migrate cmd/migrate/main.go
	docker exec $(CONTAINER_NAME) go build -o bin/seed cmd/seed/main.go

build-prod:
	docker exec $(CONTAINER_NAME) go build -o bin/migrate cmd/migrate/main.go
	docker exec $(CONTAINER_NAME) go build -o bin/api cmd/api/main.go

run-dev:
	docker exec $(CONTAINER_NAME) ./bin/migrate --fresh
	docker exec $(CONTAINER_NAME) ./bin/seed
	docker logs -f api

run-prod: build-prod
	docker exec $(CONTAINER_NAME) ./bin/migrate
	docker exec -it $(CONTAINER_NAME) ./bin/api

update:
	docker exec $(CONTAINER_NAME) go mod tidy

test:
	docker exec $(CONTAINER_NAME) go test -v ./tests/...

init: up update test