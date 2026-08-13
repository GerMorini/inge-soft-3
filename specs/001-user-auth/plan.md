# Implementation Plan: Registro e inicio de sesión

**Branch**: `feature/login-register` | **Date**: 2026-08-13 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-user-auth/spec.md`

## Summary

Implementar registro e inicio de sesión en un único módulo `identity`. El backend Go expondrá tres
operaciones HTTP, aplicará reglas en service y persistirá una cuenta mediante pgx en PostgreSQL. Una
sola tabla guardará datos personales, domicilio y hash Argon2id; constraints e índices únicos
preservarán invariantes y concurrencia. Un login válido emitirá un JWT HS256 de 30 minutos; un
middleware validará firma, algoritmo y expiración antes de `GET /api/auth/me`, sin persistir tokens.
El frontend React tendrá dos formularios simples, estado local, `sessionStorage` y componentes
daisyUI sobre Tailwind CSS. El registro pedirá confirmación de contraseña y ofrecerá un control
independiente para mostrar u ocultar cada campo, sin ampliar el contrato del backend. No habrá
refresh tokens, revocación, recuperación, validación de email, MFA, roles, permisos ni limitación
de intentos.

## Technical Context

**Language/Version**: Go 1.26.5 (backend), TypeScript 7.0 + React 19.2 (frontend), Node.js 26.7
(tooling)

**Primary Dependencies**: Go standard library `net/http`, `encoding/json`, `log/slog` and
`crypto/rand`; `github.com/jackc/pgx/v5` 5.10; `golang.org/x/crypto` 0.54; React/React DOM 19.2;
`github.com/golang-jwt/jwt/v5` 5.3.1; Vite 8.1, Tailwind CSS 4.3, `@tailwindcss/vite` and daisyUI 5
for frontend build and styles; no router, form library, ORM or global state library

**Storage**: PostgreSQL 18.4, one relational `users` table, one SQL migration

**Testing**: Go `testing`, `httptest` and PostgreSQL integration tests; Vitest 4 with React Testing
Library and user-event; no mocking framework or test-container dependency

**Target Platform**: Linux container or local Linux process; evergreen browsers supported by Vite 8

**Project Type**: Modular monolith web application

**Performance Goals**: With a warm local database, registration and login complete within 1 second
at p95 for 20 sequential requests; authenticated identity lookup completes within 200 ms; Argon2id
work remains intentionally dominant during login and registration

**Constraints**: Environment-only configuration; password plaintext never persisted, returned or
logged; 16 KiB maximum JSON body; generic login failures; no server-side session persistence or
login throttling; one application process and PostgreSQL only; JWT secret loaded from `JWT_SECRET`;
fixed 30-minute token TTL; tokens never logged or persisted; only HS256 accepted; registration
submission blocked until password confirmation matches exactly

**Scale/Scope**: Academic single-instance deployment, one identity module, two screens, three HTTP
operations and hundreds to low thousands of accounts; no horizontal-scaling design work

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Pre-design gate**: PASS. No constitutional violations require justification.

- [x] Scope is deliberately small and contains no speculative capabilities.
- [x] Design stays within one React + Go + PostgreSQL modular monolith.
- [x] Backend code is grouped by functional module and follows
      `controller -> service -> repository` without forbidden direct dependencies.
- [x] HTTP request and response structures live in `identity/dto`; PostgreSQL query inputs and
      scanned rows live in `identity/dao`. Both packages solve existing coupling without adding
      layers or redundant domain models.
- [x] Data design is relational, minimal, constrained where reasonable, and avoids structured JSON.
- [x] Environment-specific configuration remains outside source; no secrets enter the repository.
- [x] JWT state remains client-side and ephemeral; PostgreSQL gains no session or token tables.
- [x] Meaningful backend and frontend behaviors have explicit test coverage contributing toward the
      project minimum of 8 useful backend tests and 4 useful frontend tests.
- [x] Frontend uses daisyUI over Tailwind CSS, prefers existing components, and needs no custom
      theme, wrapper, or additional visual library.
- [x] Every new dependency, abstraction, pattern, or infrastructure component solves a documented
      current requirement; otherwise it is omitted.
- [x] Build, test, and run workflows remain clear and locally reproducible.

**Post-design re-check**: PASS after constitution 1.2.0 amendment. Research and Phase 1 artifacts
retain the same boundaries. The single table, direct wiring, standard-library HTTP server and local
UI state remove unnecessary layers. Argon2id and pgx address current persistence requirements. The
focused JWT library avoids implementing signing and validation manually; one middleware and typed
request metadata serve the current protected operation. daisyUI and its minimal Tailwind/Vite
integration satisfy the mandatory visual stack without another component or styling library.

## Project Structure

### Documentation (this feature)

```text
specs/001-user-auth/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── openapi.yaml
└── tasks.md                 # generated later by /speckit-tasks
```

### Source Code (repository root)

```text
backend/
├── Dockerfile
├── cmd/api/main.go
├── internal/
│   ├── platform/
│   │   ├── config/config.go
│   │   ├── database/database.go
│   │   └── requestctx/identity.go
│   └── identity/
│       ├── dto/dto.go
│       ├── dao/dao.go
│       ├── controller/
│       │   ├── controller.go
│       │   ├── middleware.go
│       │   └── controller_test.go
│       ├── service/
│       │   ├── service.go
│       │   ├── validation.go
│       │   ├── password.go
│       │   ├── token.go
│       │   └── service_test.go
│       ├── repository/
│       │   ├── repository.go
│       │   └── repository_test.go
│       ├── errors/errors.go
│       └── tests/integration_test.go
├── migrations/
│   ├── 001_create_users.up.sql
│   └── 001_create_users.down.sql
└── go.mod

frontend/
├── Dockerfile
├── nginx.conf               # serves the SPA and proxies /api to the backend
├── src/
│   ├── App.tsx
│   ├── main.tsx
│   ├── auth/
│   │   ├── RegisterForm.tsx
│   │   ├── RegisterForm.test.tsx
│   │   ├── LoginForm.tsx
│   │   ├── LoginForm.test.tsx
│   │   ├── SessionStatus.tsx
│   │   ├── SessionStatus.test.tsx
│   │   ├── api.ts
│   │   ├── session.ts
│   │   └── types.ts
│   └── styles.css           # imports Tailwind CSS and registers daisyUI
├── index.html
├── package.json
└── vite.config.ts

compose.yaml                 # database, migration, backend and frontend services
```

**Structure Decision**: `identity` owns registration, credentials, token issuance and
authentication. Controllers adapt the contract; service normalizes inputs, enforces rules, hashes
or compares passwords, and signs or verifies JWTs; repository executes two parameterized
operations. `dto` owns existing HTTP request and response shapes. `dao` owns existing query input
and scanned persistence shapes. Service-specific inputs and results remain in `service`, avoiding
an extra domain model and redundant conversions. The authentication middleware extracts a Bearer
token, delegates cryptographic validation to the identity service and writes only verified user ID
and username into a private, typed request-context value. Controllers and services never parse the
token again. Concrete
dependencies are wired in `main.go`; no service/repository interfaces, DI framework, generic
repository, ORM or extra domain model are introduced. Platform packages contain only
configuration, the shared PostgreSQL pool and the typed request metadata helper. Frontend forms
live together under `auth` and use local state; `App` switches between unauthenticated forms and a
minimal authenticated identity view without routing or global state. `session.ts` is a small
`sessionStorage` helper, not a session framework. Forms compose daisyUI `fieldset`, `input`,
`button`, `alert` and `loading` patterns directly. Registration keeps password, confirmation and
two independent visibility booleans in local state; confirmation is validated before calling
`api.ts` and is never included in the backend request. Each visibility button has an explicit
accessible name that changes between `Ver contraseña` and `Ocultar contraseña`. The default
daisyUI theme is used; custom CSS is limited to layout gaps unsupported by those patterns.

## JWT Session Design

- Login signs HS256 tokens with `github.com/golang-jwt/jwt/v5` 5.3.1. Accepted validation methods
  are pinned to HS256 to prevent algorithm substitution.
- `JWT_SECRET` is mandatory at startup, remains outside version control and must contain at least
  32 random bytes. Token values and secret material never enter logs.
- Claims are limited to `sub` with the decimal user ID, `username`, `iat` and `exp`. Login sets
  `exp = iat + 30 minutes` and responds with `accessToken`, `tokenType: "Bearer"` and
  `expiresIn: 1800`.
- `POST /api/auth/register` and `POST /api/auth/login` remain public. `GET /api/auth/me` is protected
  by middleware and returns only verified `id` and `username`.
- Missing, malformed, expired, incorrectly signed or non-HS256 tokens all return the same `401`
  `invalid_token` body. Expiration is strict: the token is invalid at `now >= exp`.
- Frontend stores only the access token in `sessionStorage`, attaches it as
  `Authorization: Bearer <token>`, restores authenticated state through `/auth/me`, and removes the
  token after local expiration or any `401` from a protected request.
- No token or session table, refresh token, cookie, revocation list or backend logout endpoint is
  introduced. A stolen token therefore remains usable until expiry; the accepted academic control
  is its short fixed lifetime.

## Verification Strategy

- Backend service tests cover 30-minute claims, minimal claim content, HS256 enforcement, altered
  signatures and expiration boundary behavior with a controlled clock.
- Middleware/controller tests cover missing and malformed Bearer headers, uniform `401` responses,
  verified identity propagation and prevention of protected-handler execution after rejection.
- Integration tests prove login returns a usable token, `/auth/me` returns its identity and neither
  operation writes session data to PostgreSQL.
- Frontend tests cover token storage after login, Bearer attachment, restoration through `/auth/me`,
  and cleanup after expiration or `401`, alongside existing form behaviors.

## Complexity Tracking

No constitution violations or exceptions.
