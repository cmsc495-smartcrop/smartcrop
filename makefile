# Include variables from the .envrc file
include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


.PHONY: serve/dev
## serve/dev: run the web server and mqtt broker locally, both hot-reloaded via air
serve/dev:
	npm run css:watch & \
	go tool air -c .air.broker.toml & \
	go tool air

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/prod: build a production binary with embedded assets
.PHONY: build/prod
build/prod:
	go build -tags production -o bin/smartcrop ./cmd/web

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

DB_DSN ?= postgres://smartcrop:smartcrop@localhost:5432/smartcrop?sslmode=disable

## db/start: start the local Postgres container via Podman
.PHONY: db/start
db/start:
	podman kube play infra/postgres.yaml

## db/stop: stop and remove the local Postgres container
.PHONY: db/stop
db/stop:
	podman kube down infra/postgres.yaml

## db/migrate: run all pending goose migrations
.PHONY: db/migrate
db/migrate:
	goose -dir db/migrations postgres "$(DB_DSN)" up

## db/seed: load seed data into the local database
.PHONY: db/seed
db/seed:
	go run ./cmd/seed

# ==================================================================================== #
# MQTT
# ==================================================================================== #

## broker/run: run the mqtt broker locally
.PHONY: broker/run
broker/run:
	go run ./cmd/broker

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

DB_DSN ?= postgres://smartcrop:smartcrop@localhost:5432/smartcrop?sslmode=disable

## db/start: start the local Postgres container via Podman
.PHONY: db/start
db/start:
	podman kube play infra/postgres.yaml

## db/stop: stop and remove the local Postgres container
.PHONY: db/stop
db/stop:
	podman kube down infra/postgres.yaml

## db/migrate: run all pending goose migrations
.PHONY: db/migrate
db/migrate:
	goose -dir db/migrations postgres "$(DB_DSN)" up

## db/seed: load seed data into the local database
.PHONY: db/seed
db/seed:
	go run ./cmd/seed