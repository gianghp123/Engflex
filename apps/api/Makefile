MIGRATIONS_DIR ?= internal/database/migrations

# Default DATABASE_URL from apps/api/.env so goose works without exporting it
# by hand. `?=` — NOT `include .env` — is the load-bearing part: a makefile
# assignment overrides the environment, so `include` would let the .env value
# clobber a real DATABASE_URL (the e2e suite migrates a dedicated database and
# was silently migrating the dev one). Missing file is fine (CI passes env).
ifneq (,$(wildcard .env))
	ENV_VARS := $(subst $(newline), ,$(file < .env))
	PARSED_DB_URL := $(patsubst DATABASE_URL=%,%,$(filter DATABASE_URL=%,$(ENV_VARS)))
	DATABASE_URL ?= $(strip $(firstword $(PARSED_DB_URL)))
endif

.PHONY: migration-create migration-up migration-down migration-status migration-reset mocks contracts swagger seed-audio pronounce-image pronounce-warm run

# usage: make migration-create name=your_migration_name
migration-create:
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql

migration-up:
	@goose -dir $(MIGRATIONS_DIR) -allow-missing postgres "$(DATABASE_URL)" up

migration-down:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migration-status:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migration-reset:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" reset

# regenerate repository mocks (see .mockery.yaml)
mocks:
	@mockery

# regenerate TypeScript contracts in packages/contracts (see tygo.yaml)
contracts:
	@tygo generate
	@node ../../packages/contracts/scripts/gen-enum-values.mjs

# regenerate swagger docs in docs/ (served at /swagger/*any)
swagger:
	@swag init -g cmd/server/main.go

# Generate + upload lesson audio (needs db migrated and minio up).
seed-audio:
	@uv run --with edge-tts --with minio --with "psycopg[binary]" --with python-dotenv scripts/seed_audio.py

# Seed stories from source JSON with edge-tts audio and DeepSeek translation.
seed-stories:
	@uv run --with edge-tts --with "psycopg[binary]" --with python-dotenv --with requests scripts/seed_stories.py

# Crawl a story from StoryWeaver by URL/slug and optionally ingest directly: make crawl-story url=... [ingest=1]
crawl-story:
	@uv run --with requests --with python-dotenv --with edge-tts --with "psycopg[binary]" scripts/crawl_storyweaver.py --url "$(url)" $(if $(ingest),--ingest,)

# Build the OpenPronounce model-host image from the pinned upstream tag
# (bundles the two Wav2Vec2 checkpoints; ~5 GB / a few minutes, network needed).
# `docker compose up -d openpronounce` builds it automatically on first run;
# this target prebuilds it without starting.
pronounce-image:
	docker compose build openpronounce

# Load the container's models now (~20 s once) instead of on the first attempt.
pronounce-warm:
	@python3 scripts/warm_pronounce.py

# run the server
run:
	@go run cmd/server/main.go
