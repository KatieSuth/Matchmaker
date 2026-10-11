# Matchmaker

# General Principles
- Do not connect remotely (web searches, etc) without asking first.
- Be concise. Do not explain standard patterns unless asked.
- Use early returns to minimize nesting.
- **Prod infra:** never `terraform apply`, `make infra-apply`, `make gcp-deploy`, or mutating `gcloud` until the operator has reviewed the plan/command and OK'd that run. See [.claude/rules/infra.md](.claude/rules/infra.md).
- No placeholders like "// TODO" or "// implement logic here" unless asked to create a placeholder, and any placeholders that have been replaced with real content should have such placeholder comments removed.
- If a change is significant, explain the architectural "Why" briefly.
- All added/edited functions must include/update a comment at the top documenting what it does. Inside functions, add inline comments only for non-obvious logic — do not comment throughout or restate self-evident code.

# Tech Stack
- Backend: Go (Golang) using the Gin framework.
- Database: PostgreSQL with `sqlc` for type-safe queries and Goose for migrations.
- Frontend: Next.js (App Router), TypeScript, Tailwind CSS.
- Containerization: Docker and Docker Compose.
- Edge / reverse proxy: **Caddy** for local Compose; **Cloudflare Worker** for GCP production (see `infra/`).

# Environment Variables
- Any new environment variable MUST be added to the root [`.env.example`](.env.example) and documented in the "Environment Variables" section of `README.md`. Do not create separate `backend/.env.example` or `frontend/.env.example`.

## Backend
- All variables are processed and validated in `backend/cmd/server/main.go`.
- Inject into handlers via `handler.New` — never call `os.Getenv` inside handlers.
- If a required variable has no default, call `fatalExit` (which uses `slog.Error`) before exiting — never silently proceed with an empty value.
- Whenever `handler.New` is updated, you MUST also update `newTestHandler` in `backend/internal/handler/test_helpers_test.go` with a suitable test value (e.g. a mock key or dummy string).

## Frontend
- Prefix any variable that must be accessible in the browser with `NEXT_PUBLIC_`.

