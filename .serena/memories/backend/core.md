# Backend core (`apps/backend`)

## Wiring (`cmd/go-boilerplate/main.go`)
config.LoadConfig → logger.NewLoggerService (New Relic) → database.Migrate (skipped if PRIMARY.ENV=local) → server.New (pgx pool, Redis client, Asynq JobService started) → repository.NewRepositories → service.NewServices → handler.NewHandlers → router.NewRouter → srv.SetupHTTPServer/Start. Shutdown on SIGINT/SIGTERM: HTTP → jobs → DB → Redis.
- Adding a feature = add field + constructor call in `Repositories`, `Services`, `Handlers`, then register route in `internal/router`.

## Config (`internal/config`)
- Env only, koanf env provider, prefix `BOILERPLATE_`, path delimiter `.` → vars look like `BOILERPLATE_DATABASE.HOST`. `godotenv/autoload` loads `.env` from CWD.
- `validate:"required"` on struct fields; missing → fatal at startup. Every new config field needs a `.env.sample` entry.
- Redis address is `host:port` (passed to `redis.Options.Addr` and `asynq.RedisClientOpt.Addr`), never a `redis://` URL.

## Request handling
- Typed wrappers in `internal/handler/base.go`: `Handle[Req,Res]`, `HandleNoContent`, `HandleFile`. Req implements `validation.Validatable` (`Validate() error`); wrapper does `validation.BindAndValidate`, logging, New Relic attrs.
- KNOWN BUG: wrappers capture a single `req` instance and bind every request into it → data race once used with a pointer Req; allocate a fresh Req per request when first using them.
- Errors: return `*errs.HTTPError`; `GlobalMiddlewares.GlobalErrorHandler` renders all errors; non-HTTP errors go through `sqlerr.HandleError` (maps pgx/Postgres codes like unique/FK violations, no rows → HTTP errors), so repositories may return raw pgx errors.
- Middleware order in `router.NewRouter`: Recover first, then rate limiter (20 rps, keyed by IP via `echo.ExtractIPFromXFFHeader` — trusts XFF only from private/loopback proxies), CORS, Secure, RequestID, New Relic, EnhanceTracing, ContextEnhancer, RequestLogger.
- Auth: `middlewares.Auth.RequireAuth` (Clerk header JWT) sets echo ctx keys `user_id`, `user_role`, `permissions`. Clerk key set in `service.NewAuthService`.
- Routes: system routes `/status`, `/docs`, `/static` in `router/system.go`. `/api/v1` group created but unused. Frontend client calls `${VITE_API_URL}/api<contract path>` → currently mismatched with `/status`.

## Data
- `server.DB.Pool` (pgxpool); pool limits from `DATABASE.MAX_OPEN_CONNS`, `CONN_MAX_LIFETIME`, `CONN_MAX_IDLE_TIME` (`MAX_IDLE_CONNS` unused — no pgxpool equivalent).
- Migrations: tern SQL in `internal/database/migrations/`, `go:embed`ed, version table `schema_version`, auto-applied at startup outside local.

## Background jobs (`internal/lib/job`)
- Asynq; queues critical/default/low. New task = type const + payload struct + `NewXTask` (see `email_tasks.go`), handler registered in `JobService.Start` mux, enqueue via `server.Job.Client`.
- Emails: `internal/lib/email` renders `templates/emails/<name>.html` with `html/template`, sends via Resend. Templates generated from `packages/emails` (see `mem:js/core`).

## Tests
- No `_test.go` yet. `internal/testing`: `SetupTest(t)` → testcontainers Postgres + migrations + test server; tx helpers; assertions.
