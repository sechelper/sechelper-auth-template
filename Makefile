.PHONY: help dev dev-down run build db-migrate test test-go test-web test-admin release-check release-id release-sync toolchain-check openapi-lint architecture-check example-orders-test

ENV ?= development
ENV_NORMALIZED := $(shell printf '%s' '$(ENV)' | tr '[:upper:]' '[:lower:]')
ACTION ?= serve
COMPONENT ?= all
ifeq ($(origin CHECKED_OUT_REVISION),undefined)
CHECKED_OUT_REVISION := $(shell git rev-parse HEAD 2>/dev/null || true)
endif
ifeq ($(origin WORKTREE_CLEAN),undefined)
WORKTREE_CLEAN := $(shell if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then git diff --quiet && git diff --cached --quiet && test -z "$$(git ls-files --others --exclude-standard)" && printf true || printf false; else printf false; fi)
endif
RELEASE_VERSION := $(shell sed -n 's/^releaseVersion:[[:space:]]*//p' release.yaml)
RELEASE_BUILD_ID := $(shell sed -n 's/^releaseBuildId:[[:space:]]*//p' release.yaml)
NPM_VERSION := $(shell sed -n 's/^npm:[[:space:]]*//p' toolchain.versions)
ifeq ($(filter command line,$(origin BUILD_ID)),)
override BUILD_ID := $(if $(filter development,$(ENV_NORMALIZED)),dev-local,$(RELEASE_BUILD_ID))
endif
SOURCE_REVISION := $(if $(filter true,$(WORKTREE_CLEAN)),$(CHECKED_OUT_REVISION),working-tree)
COMPOSE_DEV := docker compose --project-name sechelper-auth-template-dev --env-file deploy/environments/development.env -f deploy/compose.dev.yaml
COMPOSE_TEST := docker compose --project-name sechelper-auth-template-test --env-file .env --env-file deploy/environments/test.env -f deploy/compose.prod.yaml
COMPOSE_PRODUCTION := docker compose --project-name sechelper-auth-template --env-file .env --env-file deploy/environments/production.env -f deploy/compose.prod.yaml
ifeq ($(ENV_NORMALIZED),test)
ifeq ($(origin BUILD_OUTPUT_DIR),undefined)
BUILD_OUTPUT_DIR := $(shell mktemp -d "$${TMPDIR:-/tmp}/sechelper-auth-template-test-build.XXXXXX")
endif
ARTIFACT_MANIFEST ?= $(BUILD_OUTPUT_DIR)/artifact-manifest.json
else
ARTIFACT_MANIFEST ?= $(shell printf '%s' "$${TMPDIR:-/tmp}")/sechelper-auth-template-$(ENV_NORMALIZED)-artifacts.json
endif

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
	@if [ "$(ENV_NORMALIZED)" = test ]; then \
	  build_dir="$(BUILD_OUTPUT_DIR)"; mkdir -p "$$build_dir"; \
	  if [ "$(COMPONENT)" = all ] || [ "$(COMPONENT)" = api ]; then \
	    mkdir -p "$$build_dir/api" "$$build_dir/go-cache" "$$build_dir/go-tmp"; \
	    (cd api && GOCACHE="$$build_dir/go-cache" GOTMPDIR="$$build_dir/go-tmp" CGO_ENABLED=0 go build -tags=example -trimpath -ldflags="-s -w -X main.releaseVersion=$(RELEASE_VERSION) -X main.buildID=$(BUILD_ID) -X main.buildRevision=$(SOURCE_REVISION) -X main.buildEnvironment=test" -o "$$build_dir/api/auth-template" ./cmd/server); \
	    (cd api && GOCACHE="$$build_dir/go-cache" GOTMPDIR="$$build_dir/go-tmp" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.buildEnvironment=test" -o "$$build_dir/api/auth-template-migrate" ./cmd/migrate); \
	    cp -a api/migrations "$$build_dir/api/"; \
	  fi; \
	  if [ "$(COMPONENT)" = all ] || [ "$(COMPONENT)" = web ]; then \
	    mkdir -p "$$build_dir/web" "$$build_dir/admin"; \
	    (cd web && NPM_CONFIG_CACHE="$$build_dir/npm-cache" npm ci --no-audit --no-fund && NPM_CONFIG_CACHE="$$build_dir/npm-cache" npm run build -- --mode test --outDir "$$build_dir/web" --emptyOutDir); \
	    printf '{"component":"web","applicationVersion":"%s","buildId":"%s","sourceRevision":"%s","environment":"test"}\n' "$(RELEASE_VERSION)" "$(BUILD_ID)" "$(SOURCE_REVISION)" > "$$build_dir/web/build-info.json"; \
	    (cd web/admin && NPM_CONFIG_CACHE="$$build_dir/npm-cache" npm ci --no-audit --no-fund && NPM_CONFIG_CACHE="$$build_dir/npm-cache" npm run build -- --mode test --outDir "$$build_dir/admin" --emptyOutDir); \
	    printf '{"component":"admin","applicationVersion":"%s","buildId":"%s","sourceRevision":"%s","environment":"test"}\n' "$(RELEASE_VERSION)" "$(BUILD_ID)" "$(SOURCE_REVISION)" > "$$build_dir/admin/build-info.json"; \
	    grep -R -F -q '/admin/component-reference' "$$build_dir/admin" || { echo 'test build is missing component reference route' >&2; exit 1; }; \
	    grep -R -F -q '/admin/orders' "$$build_dir/admin" || { echo 'test build is missing example orders route' >&2; exit 1; }; \
	  fi; \
	  node deploy/scripts/write-artifact-manifest.mjs "$(ARTIFACT_MANIFEST)" test "$(COMPONENT)" "$(RELEASE_VERSION)" "$(BUILD_ID)" "$(SOURCE_REVISION)" "$$build_dir"; \
	  printf 'test build artifacts: %s\n' "$$build_dir"; \
	else \
	  if [ "$(COMPONENT)" = all ] || [ "$(COMPONENT)" = api ]; then docker build --build-arg CANONICAL_BUILD=1 --build-arg BUILD_ENV=$(ENV_NORMALIZED) --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" --build-arg BUILD_ID="$(BUILD_ID)" --build-arg SOURCE_REVISION="$(SOURCE_REVISION)" -f deploy/Dockerfile.api -t sechelper-auth-template-api:$(RELEASE_VERSION)-$(BUILD_ID)-$(ENV_NORMALIZED) -t sechelper-auth-template-api:$(ENV_NORMALIZED) .; fi; \
	  if [ "$(COMPONENT)" = all ] || [ "$(COMPONENT)" = web ]; then docker build --build-arg CANONICAL_BUILD=1 --build-arg BUILD_ENV=$(ENV_NORMALIZED) --build-arg NPM_VERSION="$(NPM_VERSION)" --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" --build-arg BUILD_ID="$(BUILD_ID)" --build-arg SOURCE_REVISION="$(SOURCE_REVISION)" -f deploy/Dockerfile.web -t sechelper-auth-template-web:$(RELEASE_VERSION)-$(BUILD_ID)-$(ENV_NORMALIZED) -t sechelper-auth-template-web:$(ENV_NORMALIZED) .; fi; \
	  if [ "$(COMPONENT)" = all ] && [ "$(ENV_NORMALIZED)" = development ]; then docker build --build-arg CANONICAL_BUILD=1 --build-arg BUILD_ENV=development -f deploy/Dockerfile.mock-idp -t sechelper-auth-template-mock-idp:development .; fi; \
	  if [ "$(ENV_NORMALIZED)" = production ] && [ "$(COMPONENT)" = all ]; then \
	    web_id=$$(docker create sechelper-auth-template-web:production); \
	    api_id=$$(docker create sechelper-auth-template-api:production); \
	    tmp_dir=$$(mktemp -d /tmp/auth-template-prod-check.XXXXXX); \
	    trap 'docker rm -f "$$web_id" "$$api_id" >/dev/null 2>&1 || true; rm -rf "$$tmp_dir"' EXIT; \
	    docker cp "$$web_id:/usr/share/nginx/html/admin" "$$tmp_dir/admin"; \
	    docker cp "$$api_id:/app/migrations" "$$tmp_dir/migrations"; \
	    if grep -R -E -q '/admin/(component-reference|orders)' "$$tmp_dir/admin"; then echo 'test/example route leaked into production image' >&2; exit 1; fi; \
	    if find "$$tmp_dir/migrations/business" -name MODULE_KIND -type f -exec grep -l -x example {} + | grep -q .; then echo 'test/example migrations leaked into production image' >&2; exit 1; fi; \
	  fi; \
	  node deploy/scripts/write-artifact-manifest.mjs "$(ARTIFACT_MANIFEST)" "$(ENV_NORMALIZED)" "$(COMPONENT)" "$(RELEASE_VERSION)" "$(BUILD_ID)" "$(SOURCE_REVISION)"; \
	fi

openapi-lint: toolchain-check
	npm_cache=$$(mktemp -d /tmp/auth-template-npm.XXXXXX); trap 'rm -rf "$$npm_cache"' EXIT; NPM_CONFIG_CACHE="$$npm_cache" npx --yes @redocly/cli@1.34.0 lint docs/contracts/openapi.yaml docs/contracts/business/orders/openapi.yaml
