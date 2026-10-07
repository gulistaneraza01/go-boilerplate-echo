# Conventions

## Commits
- Enforced by husky `commit-msg` → commitlint `@commitlint/config-conventional` (config in root `package.json`).
- `type(scope?): subject`; types feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert; lowercase subject start; no trailing period; header/body/footer lines ≤100.

## Go
- Constructor DI: every layer is a struct of components built by `NewX(s *server.Server, ...)`; no DI framework, no interfaces for single implementations.
- `*server.Server` is the dependency container; don't add globals (exception: `job` package's `emailClient`).
- Request-scoped logging via `middleware.GetLogger(c)`; structured zerolog fields (`Str("request_id", ...)`), not formatted strings.
- Errors to clients are `*errs.HTTPError` built with `errs.New*Error` constructors; never write error JSON directly from handlers.
- Named constants for durations (e.g. `DefaultContextTimeout`, `DatabasePingTimeout`) expressed as ints × `time.Second`.

## TypeScript
- ESM packages (`"type": "module"`), NodeNext resolution in packages → relative imports need `.js` extension; `@/*` alias resolved by `tsc-alias` at build.
- Prettier with `@trivago/prettier-plugin-sort-imports` in frontend.
- Shared request/response shapes are defined once as zod schemas in `packages/zod`, consumed by contracts in `packages/openapi`.
