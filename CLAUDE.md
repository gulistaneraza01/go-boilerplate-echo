# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout

Bun + Turborepo monorepo. The Go backend lives inside it but is **not** a Turbo workspace — it has its own toolchain (`go`, `task`).

- `apps/backend` — Go 1.26 / Echo v4 API (module `github.com/gulistaneraza01/go-boilerplate-echo`)
- `apps/frontend` — React 19 + Vite + Tailwind v4, Clerk auth, ts-rest client, TanStack Query
- `packages/zod` — shared zod schemas (`@boilerplate/zod`)
- `packages/openapi` — ts-rest API contracts + OpenAPI generator (`@boilerplate/openapi`)
- `packages/emails` — React Email templates that compile to Go HTML templates

## Commands

### Backend (run from `apps/backend` — the app reads `.env` and `templates/emails/` relative to CWD)

```bash
cp .env.example .env
task run                       # go run ./cmd/go-boilerplate (needs Postgres + Redis)
go build ./... && go vet ./...
golangci-lint run ./...        # config is golangci-lint v2 format; v1 binaries refuse it
go test ./...
go test ./internal/<pkg> -run TestName   # single test
task migrations:new name=add_users       # creates a tern migration
BOILERPLATE_DB_DSN=postgres://... task migrations:up   # prompts for confirmation
task tidy                      # go fmt + go mod tidy + verify
```

`task test` / `task migrations:down` mentioned in the READMEs do not exist in the Taskfile.

### JS (run from repo root)

```bash
bun install                    # also installs husky hooks via "prepare"
bun run dev | build | lint | typecheck    # turbo across workspaces
bunx turbo run build --filter='./packages/*'
cd packages/openapi && bun run gen        # regenerate openapi.json (see below)
cd packages/emails && bun run dev         # email preview on :3001
cd packages/emails && bun run export      # compile templates into apps/backend/templates/emails
```

Frontend `dev` waits on `packages/openapi/dist/index.js`, so packages must be built first. The frontend Vite config aliases `@boilerplate/*` directly to package `src/`, so the frontend sees package source edits without a rebuild.

## Backend architecture

**Wiring (`cmd/go-boilerplate/main.go`)**: `config.LoadConfig` → logger (New Relic) → `database.Migrate` (skipped when `PRIMARY.ENV=local`) → `server.New` (pgx pool, Redis, Asynq job server) → `repository.NewRepositories` → `service.NewServices` → `handler.NewHandlers` → `router.NewRouter`. Each layer is a plain struct of sub-components (`Repositories`, `Services`, `Handlers`, `Middlewares`) built by constructor; a new feature is added by adding a field + constructor call at each layer. `*server.Server` is the shared dependency container (Config, Logger, DB, Redis, Job).

**Config**: env vars only, via koanf with `.` as the path delimiter — names look like `BOILERPLATE_DATABASE.HOST`, `BOILERPLATE_SERVER.PORT` (not underscores). `godotenv/autoload` loads `.env` from CWD. Struct tags in `internal/config/config.go` are validated with `validate:"required"`; a missing required var is fatal at startup. Redis address is `host:port`, not a `redis://` URL.

**Handlers**: wrap typed handlers with `handler.Handle[Req, Res](h, fn, status, &Req{})` (also `HandleNoContent`, `HandleFile`) from `internal/handler/base.go`. `Req` must implement `validation.Validatable` (`Validate() error`); the wrapper binds, validates, logs, and adds New Relic attributes. Use `middleware.GetLogger(c)` for the request-scoped logger (carries request_id, user_id, trace context).

**Errors**: return `*errs.HTTPError` (constructors in `internal/errs/types.go`). `GlobalMiddlewares.GlobalErrorHandler` renders every error as the `HTTPError` JSON shape; non-HTTP errors are first passed through `sqlerr.HandleError`, which maps pgx/Postgres errors (unique violation, FK, not found, …) to HTTP errors — so repositories can return raw pgx errors.

**Routing**: `internal/router/router.go` sets global middleware (Recover, rate limit 20 req/s per IP, CORS, secure headers, request ID, New Relic, context logger, request log). System routes (`/status`, `/docs`, `/static`) are in `router/system.go`. A `/api/v1` group is created but currently unused. Protect routes with `middlewares.Auth.RequireAuth` (Clerk JWT; sets `user_id`, `user_role`, `permissions` on the echo context).

**Database**: pgx v5 pool (`server.DB.Pool`), no ORM. Migrations are tern SQL files in `internal/database/migrations/`, embedded into the binary and auto-applied at startup outside `local`. Each file needs tern's `---- create above / drop below ----` separator.

**Background jobs**: Asynq on Redis (`internal/lib/job`). Define a task type constant + payload constructor (see `email_tasks.go`), register its handler in `JobService.Start`'s mux, enqueue via `server.Job.Client`. Queues: `critical`, `default`, `low`.

**Tests**: no `_test.go` files exist yet. `internal/testing` provides `SetupTest(t)` (spins up Postgres via testcontainers — Docker required — and runs migrations), transaction helpers, and assertions.

## Cross-package flows

**API contract → docs**: endpoints are described as ts-rest contracts in `packages/openapi/src/contracts/` using schemas from `packages/zod`. `bun run gen` in `packages/openapi` writes both `packages/openapi/openapi.json` and `apps/backend/static/openapi.json`, which the backend serves at `/docs`. The Go handlers are not generated from the contract — keep them in sync by hand. The frontend client (`apps/frontend/src/api/index.ts`) prefixes contract paths with `${VITE_API_URL}/api`.

**Emails**: templates are React components in `packages/emails/src/templates/`. Props default to Go template placeholders (e.g. `userFirstName = "{{.UserFirstName}}"`); `bun run export` renders them to static HTML in `apps/backend/templates/emails/`, which `internal/lib/email` loads with `html/template`. After editing a template, re-export and commit the generated HTML.

**zod version**: everything is on zod 3.25 because `@anatine/zod-openapi` and `@ts-rest/*` require zod 3. Frontend code that wants the v4 API imports from `"zod/v4"`. Don't bump to zod 4 without replacing those libraries.

**Frontend env**: read through `apps/frontend/src/config/env.ts` (`import.meta.env`, `VITE_*` vars only). Never reintroduce `define: { "process.env": process.env }` in the Vite config — it inlines the build machine's whole environment into the bundle.

## Commits

Husky runs commitlint (`@commitlint/config-conventional`, configured in root `package.json`) on every commit: `type(scope?): subject`, types `feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert`, lowercase subject, no trailing period, header ≤ 100 chars.

`AGENTS.md` is a block managed by `turbo` (re-added automatically); for Turborepo config questions, read the docs bundled in the installed `turbo` package as it describes.
