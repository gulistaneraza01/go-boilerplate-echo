# Task completion checks

## Backend change (CWD `apps/backend`)
1. `go build ./... && go vet ./...`
2. `golangci-lint run ./...` (v2 binary)
3. `go test ./...` (DB tests need Docker)
4. If deps changed: `go mod tidy` + govulncheck.
5. If migration added: confirm tern `---- create above / drop below ----` separator present.

## JS change
1. Packages: `bunx turbo run build --filter='./packages/*'`
2. Frontend (CWD `apps/frontend`): `bunx tsc -b && bunx eslint src && VITE_CLERK_PUBLISHABLE_KEY=pk_test_x bunx vite build`
3. Emails: `bunx tsc --noEmit -p packages/emails`; if template changed, `bun run export` in `packages/emails` and commit regenerated HTML.
4. If contract/schema changed: `bun run gen` in `packages/openapi` and commit `apps/backend/static/openapi.json`; update Go handler to match.
5. If deps changed: `bun audit`.

## Commit
- Message must pass commitlint (see `mem:conventions`).
