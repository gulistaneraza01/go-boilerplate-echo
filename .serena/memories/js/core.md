# JS workspaces core

## Packages
- `packages/zod` (`@boilerplate/zod`): shared zod schemas; `index.ts` calls `extendZodWithOpenApi(z)` (anatine).
- `packages/openapi` (`@boilerplate/openapi`): ts-rest contracts in `src/contracts/` (root `apiContract` composes per-domain routers), `src/index.ts` builds OpenAPI via `@ts-rest/open-api` with security from route `metadata.openApiSecurity` (`getSecurityMetadata` in `utils.ts`). Export `./contracts` → `dist/contracts/index.js`.
  - `bun run gen` writes `openapi.json` here AND `apps/backend/static/openapi.json` (served at backend `/docs`). It also rewrites a `{type:"file"}` object schema to `string/binary`.
- `packages/emails`: React Email templates in `src/templates/`. Props default to Go template placeholders (`"{{.UserFirstName}}"`) so exported HTML is a Go `html/template`. `bun run export` → `apps/backend/templates/emails/`. `<Head>` must be inside `<Tailwind>` (react-email 6).

## Frontend (`apps/frontend`)
- Vite aliases `@boilerplate/openapi` and `@boilerplate/zod` to package `src/` (no rebuild needed in dev), plus `@` → `src`.
- Env: only via `src/config/env.ts` (zod/v4 schema over `import.meta.env`; `VITE_*` only). Never add `define: {"process.env": process.env}` to Vite config — leaks build env into bundle.
- API client: `src/api/index.ts` `useApiClient` = ts-rest `initClient(apiContract)` with axios, Clerk token (`getToken({template:"custom"})`), retries 401 up to 2×, base URL `${API_URL}/api`.
- Dev server port 3000 (matches backend default CORS origin).

## zod
- All workspaces on zod 3.25; see `mem:tech_stack` for why and how to use v4 API.
