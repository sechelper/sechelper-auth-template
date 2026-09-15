.PHONY: help dev dev-down run build db-migrate test test-go test-web test-admin release-check release-id release-sync toolchain-check openapi-lint architecture-check example-orders-test

ENV ?= development
ENV_NORMALIZED := $(shell printf '%s' '$(ENV)' | tr '[:upper:]' '[:lower:]')
ACTION ?= serve
COMPONENT ?= all
CHECKED_OUT_REVISION := $(shell git rev-parse HEAD 2>/dev/null || true)
WORKTREE_CLEAN := $(shell git diff --quiet && git diff --cached --quiet && test -z "$$(git ls-files --others --exclude-standard)" && printf true || printf false)
RELEASE_VERSION := $(shell sed -n 's/^releaseVersion:[[:space:]]*//p' release.yaml)
RELEASE_BUILD_ID := $(shell sed -n 's/^releaseBuildId:[[:space:]]*//p' release.yaml)
NPM_VERSION := $(shell sed -n 's/^npm:[[:space:]]*//p' toolchain.versions)
ifeq ($(filter command line,$(origin BUILD_ID)),)
override BUILD_ID := $(if $(filter development,$(ENV_NORMALIZED)),dev-local,$(RELEASE_BUILD_ID))
endif
SOURCE_REVISION := $(if $(filter true,$(WORKTREE_CLEAN)),$(CHECKED_OUT_REVISION),working-tree)
ARTIFACT_MANIFEST ?= $(shell printf '%s' "$${TMPDIR:-/tmp}")/sechelper-auth-template-$(ENV_NORMALIZED)-artifacts.json
COMPOSE_DEV := docker compose --project-name sechelper-auth-template-dev --env-file deploy/environments/development.env -f deploy/compose.dev.yaml
COMPOSE_TEST := docker compose --project-name sechelper-auth-template-test --env-file .env --env-file deploy/environments/test.env -f deploy/compose.prod.yaml
COMPOSE_PRODUCTION := docker compose --project-name sechelper-auth-template --env-file .env --env-file deploy/environments/production.env -f deploy/compose.prod.yaml

help:
	@printf '%s\n' \
		'make build ENV=<development|test|production> [COMPONENT=all|api|web] Build through one entry point' \
		'make run ENV=<development|test|production> Start the application stack through one entry point' \
		'make release-check Validate canonical release metadata' \
		'make release-id    Allocate and store a new UTC release Build ID' \
		'make release-sync  Synchronize frontend package versions from release.yaml' \
		'make toolchain-check Verify declared compiler and package-manager versions' \
		'make dev-down     Stop the local stack and preserve named volumes' \
		'make db-migrate   Apply forward-only migrations to the configured database' \
		'make test         Run Go and frontend tests' \
		'make openapi-lint Validate the canonical OpenAPI contract' \
		'make architecture-check Enforce framework/business boundaries' \
		'make example-orders-test Run the opt-in orders example tests'

dev:
	$(MAKE) --no-print-directory run ENV=development

run:
	@case "$(ACTION)" in serve|migrate) ;; *) echo 'ACTION must be serve or migrate' >&2; exit 2 ;; esac
	@case "$(ENV_NORMALIZED)" in \
		development) $(MAKE) --no-print-directory build ENV=$(ENV_NORMALIZED) && if [ "$(ACTION)" = migrate ]; then RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) INCLUDE_EXAMPLE_MIGRATIONS=1 $(COMPOSE_DEV) run --rm migrate; elif [ "$(ACTION)" = serve ]; then RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) INCLUDE_EXAMPLE_MIGRATIONS=1 $(COMPOSE_DEV) up -d --no-build; else echo 'ACTION must be serve or migrate' >&2; exit 2; fi ;; \
		test) $(MAKE) --no-print-directory build ENV=$(ENV_NORMALIZED) && if [ "$(ACTION)" = migrate ]; then RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) INCLUDE_EXAMPLE_MIGRATIONS=1 $(COMPOSE_TEST) run --rm migrate; else RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) INCLUDE_EXAMPLE_MIGRATIONS=1 $(COMPOSE_TEST) up -d --no-build; fi ;; \
		production) $(MAKE) --no-print-directory build ENV=$(ENV_NORMALIZED) && if [ "$(ACTION)" = migrate ]; then RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) $(COMPOSE_PRODUCTION) run --rm migrate; else RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) $(COMPOSE_PRODUCTION) up -d --no-build; fi ;; \
		*) echo 'ENV must be development, test, or production' >&2; exit 2 ;; \
	esac

dev-down:
	RELEASE_VERSION=$(RELEASE_VERSION) BUILD_ID=$(BUILD_ID) $(COMPOSE_DEV) down

release-check:
	node deploy/scripts/check-release.mjs

release-id:
	node deploy/scripts/update-release-id.mjs release.yaml

release-sync:
	node deploy/scripts/sync-package-versions.mjs

toolchain-check:
	node deploy/scripts/check-toolchain.mjs

db-migrate:
	$(MAKE) --no-print-directory run ENV=$(ENV_NORMALIZED) ACTION=migrate

test-go: toolchain-check
	tmp_dir=$$(mktemp -d /tmp/auth-template-test.XXXXXX); go_tmp=$$(mktemp -d /tmp/auth-template-go-tmp.XXXXXX); trap 'rm -rf "$$tmp_dir" "$$go_tmp"' EXIT; cd api && GOCACHE="$$tmp_dir/go-cache" GOTMPDIR="$$go_tmp" go test ./...

test-web: toolchain-check
	cd web && npm test

test-admin: toolchain-check
	cd web/admin && npm test

test: test-go test-web test-admin example-orders-test

architecture-check:
	bash deploy/scripts/check-boundaries.sh

example-orders-test: toolchain-check
	tmp_dir=$$(mktemp -d /tmp/auth-template-example.XXXXXX); go_tmp=$$(mktemp -d /tmp/auth-template-example-go.XXXXXX); trap 'rm -rf "$$tmp_dir" "$$go_tmp"' EXIT; cd api && GOCACHE="$$tmp_dir" GOTMPDIR="$$go_tmp" go test -tags=example ./internal/business/orders/...

build: release-check toolchain-check
	@case "$(ENV_NORMALIZED)" in development|test|production) ;; *) echo 'ENV must be development, test, or production' >&2; exit 2 ;; esac
	@case "$(COMPONENT)" in all|api|web) ;; *) echo 'COMPONENT must be all, api, or web' >&2; exit 2 ;; esac
	@if [ "$(ENV_NORMALIZED)" != development ] && [ "$(BUILD_ID)" != "$(RELEASE_BUILD_ID)" ]; then echo 'test/production BUILD_ID must match release.yaml releaseBuildId' >&2; exit 2; fi
	@if [ "$(ENV_NORMALIZED)" != development ] && { [ "$(WORKTREE_CLEAN)" != true ] || ! printf '%s' "$(SOURCE_REVISION)" | grep -Eq '^[0-9a-f]{40}$$' || [ "$(SOURCE_REVISION)" != "$(CHECKED_OUT_REVISION)" ]; }; then echo 'test/production builds require a clean committed source revision' >&2; exit 2; fi
	@if [ "$(COMPONENT)" = all ] || [ "$(COMPONENT)" = api ]; then docker build --build-arg CANONICAL_BUILD=1 --build-arg BUILD_ENV=$(ENV_NORMALIZED) --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" --build-arg BUILD_ID="$(BUILD_ID)" --build-arg SOURCE_REVISION="$(SOURCE_REVISION)" -f deploy/Dockerfile.api -t sechelper-auth-template-api:$(RELEASE_VERSION)-$(BUILD_ID)-$(ENV_NORMALIZED) -t sechelper-auth-template-api:$(ENV_NORMALIZED) .; fi
	@if [ "$(COMPONENT)" = all ] || [ "$(COMPONENT)" = web ]; then docker build --build-arg CANONICAL_BUILD=1 --build-arg BUILD_ENV=$(ENV_NORMALIZED) --build-arg NPM_VERSION="$(NPM_VERSION)" --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" --build-arg BUILD_ID="$(BUILD_ID)" --build-arg SOURCE_REVISION="$(SOURCE_REVISION)" -f deploy/Dockerfile.web -t sechelper-auth-template-web:$(RELEASE_VERSION)-$(BUILD_ID)-$(ENV_NORMALIZED) -t sechelper-auth-template-web:$(ENV_NORMALIZED) .; fi
	@if [ "$(COMPONENT)" = all ] && [ "$(ENV_NORMALIZED)" = development ]; then docker build --build-arg CANONICAL_BUILD=1 --build-arg BUILD_ENV=development -f deploy/Dockerfile.mock-idp -t sechelper-auth-template-mock-idp:development .; fi
	@if [ "$(COMPONENT)" = all ] && [ "$(ENV_NORMALIZED)" = production ]; then \
		web_id=$$(docker create sechelper-auth-template-web:production); \
		api_id=$$(docker create sechelper-auth-template-api:production); \
		tmp_dir=$$(mktemp -d /tmp/auth-template-prod-check.XXXXXX); \
		trap 'docker rm -f "$$web_id" "$$api_id" >/dev/null 2>&1 || true; rm -rf "$$tmp_dir"' EXIT; \
		docker cp "$$web_id:/usr/share/nginx/html/admin" "$$tmp_dir/admin"; \
		docker cp "$$api_id:/app/migrations" "$$tmp_dir/migrations"; \
		if grep -R -E -q '/admin/(component-reference|orders)' "$$tmp_dir/admin"; then echo 'test/example route leaked into production image' >&2; exit 1; fi; \
		if find "$$tmp_dir/migrations/business" -name MODULE_KIND -type f -exec grep -l -x example {} + | grep -q .; then echo 'example migrations leaked into production image' >&2; exit 1; fi; \
	fi
	@node deploy/scripts/write-artifact-manifest.mjs "$(ARTIFACT_MANIFEST)" "$(ENV_NORMALIZED)" "$(COMPONENT)" "$(RELEASE_VERSION)" "$(BUILD_ID)" "$(SOURCE_REVISION)"

openapi-lint: toolchain-check
	npm_cache=$$(mktemp -d /tmp/auth-template-npm.XXXXXX); trap 'rm -rf "$$npm_cache"' EXIT; NPM_CONFIG_CACHE="$$npm_cache" npx --yes @redocly/cli@1.34.0 lint docs/contracts/openapi.yaml docs/contracts/business/orders/openapi.yaml
