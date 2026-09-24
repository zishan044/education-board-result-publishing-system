-include .env
export

DB_URL  := postgres://zishan044:$(POSTGRES_PASSWORD)@localhost:5432/results?sslmode=disable
GOOSE   := goose -dir ./migrations postgres "$(DB_URL)"
COMPOSE := docker compose -f deploy/docker-compose.yml --env-file .env

.PHONY: db-up db-down migrate-up migrate-down migrate-redo migrate-status migrate-create load run

db-up:
	$(COMPOSE) up -d postgres

db-down:
	$(COMPOSE) down

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-redo:
	$(GOOSE) redo

migrate-status:
	$(GOOSE) status

migrate-create:
	goose -dir ./migrations -s create $(name) sql

load:
	go run ./cmd/loader -db "$(DB_URL)" -count $(or $(count),100000)

run:
	DATABASE_URL="$(DB_URL)" go run ./cmd/api