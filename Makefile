.PHONY: up dev swag migrate run tidy

up:
	 docker compose up --build

dev: swag migrate run

swag:
	 go install github.com/swaggo/swag/cmd/swag@latest
	 swag init -g cmd/sca/main.go -o docs

migrate:
	 go run cmd/sca/main.go --migrate-only

run:
	 go run cmd/sca/main.go

tidy:
	 go mod tidy
