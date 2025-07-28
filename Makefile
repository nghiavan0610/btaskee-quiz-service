.PHONY: build dev prod di-generate lint tidy deps test test-silent test-cov
# Application
build:
	go build -o bin/btaskee-quiz-service cmd/app/main.go

dev:
	air -c .air.toml

prod:
	go run cmd/app/main.go

di-generate:
	wire ./app

sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc-generate:
	sqlc generate

# Migrate
migrate-add:
	@read -p "Enter migration name: (example: create-todo-table) " name; \
		go run ./cmd/migrate create $$name sql

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down
	
# Lint code
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Test
test:
	go test -v ./...

test-silent:
	go test ./...

test-cov:
	go test -cover ./...
