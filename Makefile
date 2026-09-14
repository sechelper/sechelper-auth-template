.PHONY: help dev dev-down db-migrate test test-go test-web test-admin build openapi-lint docker-build architecture-check example-orders-test

COMPOSE_DEV := docker compose --project-name sechelper-auth-template-dev -f deploy/compose.dev.yaml

help:
	@printf '%s\n' \
		'make dev          Start the local PostgreSQL, Redis, API, and web stack' \
		'make dev-down     Stop the local stack and preserve named volumes' \
		'make db-migrate   Apply forward-only migrations to the configured database' \
		'make test         Run Go and frontend tests' \
		'make build        Build API, web, and admin production artifacts' \
		'make openapi-lint Validate the canonical OpenAPI contract' \
		'make architecture-check Enforce framework/business boundaries' \
		'make example-orders-test Run the opt-in orders example tests' \
		'make docker-build Build the production API and web images'

dev:
	$(COMPOSE_DEV) up --build

dev-down:
	$(COMPOSE_DEV) down

db-migrate:
	cd api && go run ./cmd/migrate --config ../config.yaml

test-go:
	tmp_dir=$$(mktemp -d /tmp/auth-template-test.XXXXXX); go_tmp=$$(mktemp -d /tmp/auth-template-go-tmp.XXXXXX); trap 'rm -rf "$$tmp_dir" "$$go_tmp"' EXIT; cd api && GOCACHE="$$tmp_dir/go-cache" GOTMPDIR="$$go_tmp" go test ./...

test-web:
	cd web && npm test

test-admin:
	cd web/admin && npm test

test: test-go test-web test-admin

architecture-check:
	bash scripts/check-boundaries.sh

example-orders-test:
	tmp_dir=$$(mktemp -d /tmp/auth-template-example.XXXXXX); go_tmp=$$(mktemp -d /tmp/auth-template-example-go.XXXXXX); trap 'rm -rf "$$tmp_dir" "$$go_tmp"' EXIT; cd api && GOCACHE="$$tmp_dir" GOTMPDIR="$$go_tmp" go test -tags=example ./internal/modules/orders/...

build:
	tmp_dir=$$(mktemp -d /tmp/auth-template-build.XXXXXX); go_tmp=$$(mktemp -d /tmp/auth-template-go-tmp.XXXXXX); trap 'rm -rf "$$tmp_dir" "$$go_tmp"' EXIT; cd api && GOCACHE="$$tmp_dir/go-cache" GOTMPDIR="$$go_tmp" go build -o "$$tmp_dir/auth-template" ./cmd/server && GOCACHE="$$tmp_dir/go-cache" GOTMPDIR="$$go_tmp" go build -o "$$tmp_dir/auth-template-migrate" ./cmd/migrate
	cd web && npm run build
	cd web/admin && npm run build

openapi-lint:
	npm_cache=$$(mktemp -d /tmp/auth-template-npm.XXXXXX); trap 'rm -rf "$$npm_cache"' EXIT; NPM_CONFIG_CACHE="$$npm_cache" npx --yes @redocly/cli@1.34.0 lint docs/contracts/openapi.yaml

docker-build:
	docker build -f deploy/Dockerfile.api -t sechelper-auth-template-api .
	docker build -f deploy/Dockerfile.web -t sechelper-auth-template-web .
	docker build -f deploy/Dockerfile.mock-idp -t sechelper-auth-template-mock-idp .
