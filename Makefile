-include .env
export

.SILENT:
.PHONY: up down clean dev build sqlc migrate migrate-down migrate-status psql swag admin

up:
	docker compose up

down:
	docker compose down && clear

clean:
	docker compose down -v && clear

dev:
	air

build:
	GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

sqlc:
	cd db && sqlc generate

migrate:
	cd db && goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	cd db && goose -dir migrations postgres "$(DB_URL)" down

migrate-status:
	cd db && goose -dir migrations postgres "$(DB_URL)" status

psql:
	docker exec -it database psql -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)"

swag:
	go tool swag init -g cmd/server/main.go -o docs --parseInternal

admin:
	go run ./cmd/create-admin
