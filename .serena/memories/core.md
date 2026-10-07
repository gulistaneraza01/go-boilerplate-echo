# go-boilerplate-echo — core

Bun + Turborepo monorepo; Go backend lives inside but is NOT a turbo/bun workspace (own toolchain: `go`, `task`).

## Source map
- `apps/backend` — Go/Echo REST API. Layering, wiring, config, errors, jobs, DB: `mem:backend/core`
- `apps/frontend` — React 19 + Vite SPA; `packages/{zod,openapi,emails}` shared TS packages. Package roles, contract→docs flow, email template pipeline, zod pin: `mem:js/core`
- Stack/version pins and why they can't be bumped: `mem:tech_stack`
- Commands (run, build, migrate, gen, export): `mem:suggested_commands`
- Code/commit conventions: `mem:conventions`
- Verification before calling a task done: `mem:task_completion`

## Project-wide invariants
- Backend must run with CWD=`apps/backend` (reads `.env` and `templates/emails/` relative to CWD).
- Go handlers are NOT generated from the ts-rest contracts; contract (`packages/openapi`) and Go routes are kept in sync by hand.
- Generated artifacts are committed: `apps/backend/static/openapi.json` (from `packages/openapi` gen) and `apps/backend/templates/emails/*.html` (from `packages/emails` export). Edit the sources, regenerate, commit output.
- `CLAUDE.md` at repo root mirrors this guidance; README files contain stale facts (`task test`, `migrations:down`, underscore env names) — trust code/Taskfile over READMEs.
- `AGENTS.md` is auto-managed by turbo; don't hand-edit.
