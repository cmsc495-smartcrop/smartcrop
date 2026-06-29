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
## serve/dev: run the web server locally (uses local ui/ directory for templates)
serve/dev:
	npm run css:watch & \
	go tool air

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/prod: build a production binary with embedded assets
.PHONY: build/prod
build/prod:
	go build -tags production -o bin/smartcrop ./cmd/web