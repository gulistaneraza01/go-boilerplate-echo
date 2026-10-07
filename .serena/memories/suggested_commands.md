# Suggested commands

## Backend (CWD `apps/backend`)
- Setup env: `cp .env.sample .env` (file is `.env.sample`).
- Run: `task run` (needs Postgres + Redis reachable per `.env`).
- Build/vet: `go build ./... && go vet ./...`
- Lint: `golangci-lint run ./...` (needs golangci-lint v2).
- Tests: `go test ./...`; single: `go test ./internal/<pkg> -run TestName`. DB tests need Docker (testcontainers).
- Vulns: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- New migration: `task migrations:new name=<name>`; apply manually: `BOILERPLATE_DB_DSN=postgres://... task migrations:up` (interactive confirm prompt).
- Deps: `task tidy` (fmt + mod tidy + verify); upgrade: `go get -u ./... && go mod tidy`.
- Nonexistent despite README: `task test`, `task migrations:down`.

## JS (CWD repo root unless noted)
- `bun install` (also installs husky hooks via `prepare`).
- `bun run dev | build | lint | typecheck` (turbo across workspaces).
- Build shared packages only: `bunx turbo run build --filter='./packages/*'` (frontend `dev` waits for `packages/openapi/dist/index.js`).
- Frontend checks (CWD `apps/frontend`): `bunx tsc -b`, `bunx eslint src`, `bunx vite build` (needs `VITE_CLERK_PUBLISHABLE_KEY`; `bun run build` additionally needs `.env.local` via env-cmd).
- Regenerate OpenAPI (CWD `packages/openapi`): `bun run gen`.
- Emails (CWD `packages/emails`): `bun run dev` (preview :3001), `bun run export`.
- Audit: `bun audit`, `bun outdated --filter='*'`.

## Darwin notes
- BSD `sed`: in-place edit needs `sed -i ''`.
