include .env
export

.SILENT:
.PHONY: up down clean dev sqlc migrate migrate-down migrate-status

up:
	COMPOSE_BAKE=true docker compose up --build

down:
	docker compose down && clear

clean:
	docker compose down -v && clear

dev:
	air

sqlc:
	cd db && sqlc generate

migrate:
	cd db && goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	cd db && goose -dir migrations postgres "$(DB_URL)" down

migrate-status:
	cd db && goose -dir migrations postgres "$(DB_URL)" status
