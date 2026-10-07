# Tech stack

## Backend (`apps/backend`, module `github.com/gulistaneraza01/go-boilerplate-echo`)
- Go: go.mod `go 1.26.0` + `toolchain go1.26.6` (pinned toolchain fixes stdlib CVEs; go auto-downloads it).
- Echo v4, pgx v5 pool (no ORM), tern v2 migrations, Asynq (Redis jobs), go-redis v9, koanf v2 config, go-playground/validator, zerolog, New Relic go-agent v3 (+ nrecho, nrpgx5, nrredis), Clerk SDK v2 (auth), Resend (email), testcontainers-go (tests).
- Task runner: Taskfile v3 (`task`). Lint: golangci-lint **v2** config (`.golangci.yml`); v1 binaries refuse it.

## JS (root + `apps/frontend` + `packages/*`)
- Package manager: Bun (lockfile `bun.lock`), Turborepo for task orchestration; root has husky + commitlint.
- Frontend: React 19, Vite 7, Tailwind v4 (`@tailwindcss/vite`), Clerk React, ts-rest client, TanStack Query, react-hook-form, react-router 7, axios.
- TypeScript 5.9 (editors may run TS 6; configs avoid deprecated `baseUrl` — `paths` are tsconfig-relative `./src/*`).
- zod pinned to **v3.25** everywhere: `@anatine/zod-openapi` and `@ts-rest/core`/`@ts-rest/open-api` (stable) require zod 3. v4 API available via `import { z } from "zod/v4"`. Upgrading to zod 4 requires replacing those libs.
- Emails: react-email 6 + `@react-email/components` 1.x (exact pins). v6 requires `<Head>` inside `<Tailwind>`.
