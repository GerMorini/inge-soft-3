# Tasks: Registro e inicio de sesión

**Input**: Design documents from `/specs/001-user-auth/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml`,
`quickstart.md`

**Tests**: Required. Tasks follow test-first order and cover the 14 backend and 5 frontend behavior
groups declared in `spec.md`.

**Organization**: Tasks are grouped by user story. Every story ends with an independently
verifiable increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it uses different files and has no dependency on unfinished
  tasks in the same phase.
- **[Story]**: Maps work to `US1`, `US2`, `US3` or `US4` from `spec.md`.
- Every task names its concrete target path.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the smallest reproducible Go, React and PostgreSQL project skeleton.

- [X] T001 Create backend directories and entry files from the planned structure in `backend/cmd/api/main.go`, `backend/internal/platform/`, `backend/internal/identity/`, and `backend/migrations/`
- [X] T002 Initialize Go module and pin `pgx/v5` 5.10, `x/crypto` 0.54, and `golang-jwt/jwt/v5` 5.3.1 in `backend/go.mod` and `backend/go.sum`
- [X] T003 [P] Initialize React 19.2, TypeScript 7.0, Vite 8.1, Vitest 4, Testing Library, Tailwind CSS 4.3, `@tailwindcss/vite`, and daisyUI 5 in `frontend/package.json` and `frontend/package-lock.json`
- [X] T004 Configure TypeScript, Vite `/api` proxy, Vitest `jsdom`, and test setup in `frontend/tsconfig.json`, `frontend/vite.config.ts`, and `frontend/src/test/setup.ts`
- [X] T005 Configure Tailwind CSS and daisyUI without a custom theme in `frontend/src/styles.css` and import styles from `frontend/src/main.tsx`
- [X] T006 [P] Define application and isolated test PostgreSQL services in `compose.yaml`

**Checkpoint**: Backend, frontend, and PostgreSQL toolchains install and start without feature code.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add configuration, database connectivity, shared errors, and safe HTTP foundations
required by every story.

**⚠️ CRITICAL**: Complete this phase before any user story.

- [X] T007 [P] Write configuration tests for required `DATABASE_URL`, `JWT_SECRET` minimum length, optional `TEST_DATABASE_URL`, and default `HTTP_ADDR` in `backend/internal/platform/config/config_test.go`
- [X] T008 Implement environment-only configuration with startup failure for missing database URL or JWT secret in `backend/internal/platform/config/config.go`
- [X] T009 [P] Implement pgx pool creation, startup ping, and context propagation in `backend/internal/platform/database/database.go`
- [X] T010 [P] Define explicit identity validation, conflict, credential, and token errors without generic frameworks in `backend/internal/identity/errors/errors.go`
- [X] T011 Implement JSON response helpers, 16 KiB request limits, unknown-field rejection, and generic internal errors in `backend/internal/identity/controller/controller.go`
- [X] T012 Configure `net/http` routing, graceful shutdown, read/write/idle timeouts, header limit, and safe structured logging in `backend/cmd/api/main.go`
- [X] T013 Add test-database guard and shared PostgreSQL cleanup helpers that require `TEST_DATABASE_URL` and reject the application database in `backend/internal/identity/tests/testdb_test.go`

**Checkpoint**: Foundation is ready. Story implementation may begin.

---

## Phase 3: User Story 1 - Registrar una cuenta (Priority: P1) 🎯 MVP

**Goal**: Register one user with personal data, one residence, canonical username/email, optional
apartment, E.164 phone, and a non-plaintext password.

**Independent Test**: Submit valid data with and without apartment, receive one canonical user,
then log in through repository/service test setup to prove a usable Argon2id credential exists.

### Tests for User Story 1

> Write these tests first. Confirm they fail before implementation.

- [X] T014 [P] [US1] Write backend tests for valid normalization, optional apartment, E.164 acceptance, password-rule acceptance, and Argon2id hash round-trip in `backend/internal/identity/service/service_test.go`
- [X] T015 [P] [US1] Write PostgreSQL tests for atomic user insertion, canonical returned fields, and absence of plaintext passwords in `backend/internal/identity/repository/repository_test.go`
- [X] T016 [P] [US1] Write HTTP integration tests for valid registration with and without apartment and response exclusion of password/hash in `backend/internal/identity/tests/integration_test.go`
- [X] T017 [P] [US1] Write frontend tests for all registration fields, required/optional labels, valid submission, pending state, and canonical success rendering in `frontend/src/auth/RegisterForm.test.tsx`

### Implementation for User Story 1

- [X] T018 [P] [US1] Create reversible `users` table migration with identity primary key, relational personal/address columns, named uniqueness constraints, E.164 and canonical-value checks, and Argon2id prefix check in `backend/migrations/001_create_users.up.sql` and `backend/migrations/001_create_users.down.sql`
- [X] T019 [P] [US1] Implement Argon2id PHC encoding and constant-time verification using fresh random salts in `backend/internal/identity/service/password.go`
- [X] T020 [US1] Implement registration DTOs, normalization, required-field validation, E.164 validation, username rules, password rules, and registration orchestration in `backend/internal/identity/service/validation.go` and `backend/internal/identity/service/service.go`
- [X] T021 [US1] Implement parameterized user insertion and safe returned fields with pgx in `backend/internal/identity/repository/repository.go`
- [X] T022 [US1] Implement `POST /api/auth/register` request/response mapping and register its route in `backend/internal/identity/controller/controller.go` and `backend/cmd/api/main.go`
- [X] T023 [US1] Implement registration request types, fetch call, controlled form, keyboard-accessible feedback, and daisyUI fieldset/input/button/alert/loading composition in `frontend/src/auth/types.ts`, `frontend/src/auth/api.ts`, `frontend/src/auth/RegisterForm.tsx`, and `frontend/src/App.tsx`

**Checkpoint**: Registration works independently and stores exactly one valid account row.

---

## Phase 4: User Story 2 - Iniciar sesión (Priority: P2)

**Goal**: Authenticate with username and password, return a signed HS256 JWT containing user ID in
`sub`, and store it in frontend `sessionStorage`.

**Independent Test**: Prepare one account, log in using differently cased username and exact
password, verify a 30-minute token containing only `sub`, `username`, `iat`, and `exp`, then compare
unknown-user and wrong-password public failures.

### Tests for User Story 2

> Write these tests first. Confirm they fail before implementation.

- [X] T024 [P] [US2] Extend password and login tests for case-insensitive username, case-sensitive password, dummy-hash work, identical invalid-credential errors, minimal JWT claims, and exact 30-minute lifetime in `backend/internal/identity/service/service_test.go`
- [X] T025 [P] [US2] Extend repository tests for canonical username credential lookup and not-found behavior in `backend/internal/identity/repository/repository_test.go`
- [X] T026 [P] [US2] Extend HTTP integration tests for login success contract, unknown username, wrong password, repeated failures without lockout, and no database mutation in `backend/internal/identity/tests/integration_test.go`
- [X] T027 [P] [US2] Write frontend tests for login input, pending/success/failure states, JWT storage, password clearing, and identical invalid-credential presentation in `frontend/src/auth/LoginForm.test.tsx`

### Implementation for User Story 2

- [X] T028 [P] [US2] Implement HS256 token issuance with decimal user ID in `sub`, canonical username, `iat`, `exp`, fixed 30-minute lifetime, and no personal-data claims in `backend/internal/identity/service/token.go`
- [X] T029 [US2] Implement parameterized credential lookup returning only ID, username, and password hash in `backend/internal/identity/repository/repository.go`
- [X] T030 [US2] Implement login orchestration, canonical username lookup, Argon2id comparison, dummy-hash comparison for unknown users, and generic credential rejection in `backend/internal/identity/service/service.go`
- [X] T031 [US2] Implement `POST /api/auth/login` mapping with `accessToken`, `tokenType`, `expiresIn`, generic `401`, and route registration in `backend/internal/identity/controller/controller.go` and `backend/cmd/api/main.go`
- [X] T032 [US2] Implement token storage helper, login API call, controlled daisyUI login form, and authenticated transition in `frontend/src/auth/session.ts`, `frontend/src/auth/api.ts`, `frontend/src/auth/LoginForm.tsx`, `frontend/src/auth/types.ts`, and `frontend/src/App.tsx`

**Checkpoint**: Login issues a usable JWT without creating persistent session state.

---

## Phase 5: User Story 3 - Mantener acceso autenticado temporal (Priority: P3)

**Goal**: Validate JWTs through middleware, expose current user identity from verified claims, and
restore or clear frontend authentication state.

**Independent Test**: Call `/api/auth/me` with a valid token, then with missing, malformed, altered,
expired, and non-HS256 tokens; confirm only valid identity reaches the handler and no session data is
persisted.

### Tests for User Story 3

> Write these tests first. Confirm they fail before implementation.

- [X] T033 [P] [US3] Extend token tests for HS256 pinning, altered signatures, malformed claims, invalid subject IDs, and validity immediately before versus exactly at expiration in `backend/internal/identity/service/service_test.go`
- [X] T034 [P] [US3] Write middleware/controller tests for Bearer parsing, uniform `invalid_token` responses, protected-handler blocking, and verified request identity propagation in `backend/internal/identity/controller/controller_test.go`
- [X] T035 [P] [US3] Extend HTTP integration tests for login-to-`/auth/me`, matching user ID/username, expiration rejection, and unchanged PostgreSQL state in `backend/internal/identity/tests/integration_test.go`
- [X] T036 [P] [US3] Write frontend tests for Bearer attachment, reload restoration, expired-token cleanup, protected-request `401` cleanup, and unauthenticated fallback in `frontend/src/auth/SessionStatus.test.tsx`

### Implementation for User Story 3

- [X] T037 [P] [US3] Implement private typed request-context helpers carrying only verified user ID and username in `backend/internal/platform/requestctx/identity.go`
- [X] T038 [US3] Implement JWT parsing with HS256 allow-list, signature verification, required claim validation, strict expiration, and safe subject conversion in `backend/internal/identity/service/token.go`
- [X] T039 [US3] Implement authentication middleware that extracts Bearer tokens, delegates validation, writes verified identity to request context, and emits one generic `401` without logging tokens in `backend/internal/identity/controller/middleware.go`
- [X] T040 [US3] Implement protected `GET /api/auth/me` response from trusted context and wrap only that route with middleware in `backend/internal/identity/controller/controller.go` and `backend/cmd/api/main.go`
- [X] T041 [US3] Implement authenticated fetch behavior, `/auth/me` restoration, expiration cleanup, and protected-request `401` cleanup in `frontend/src/auth/api.ts` and `frontend/src/auth/session.ts`
- [X] T042 [US3] Implement minimal authenticated identity state with daisyUI status/alert patterns and no global store in `frontend/src/auth/SessionStatus.tsx` and `frontend/src/App.tsx`

**Checkpoint**: Valid JWTs restore identity for 30 minutes; invalid JWTs never reach protected code.

---

## Phase 6: User Story 4 - Corregir datos inválidos (Priority: P4)

**Goal**: Report every relevant registration error, preserve correctable input, clear passwords,
and keep PostgreSQL uniqueness authoritative under concurrency.

**Independent Test**: Submit combined invalid fields and concurrent duplicate registrations;
confirm field-level feedback, no partial rows, cleared password, preserved safe input, and exactly one
winner for each canonical username/email conflict.

### Tests for User Story 4

> Write these tests first. Confirm they fail before implementation.

- [X] T043 [P] [US4] Extend backend table tests for every username boundary, whitespace position, E.164 rejection, exact password thresholds, repeated symbols/digits, trimmed empty fields, and 20-character street number in `backend/internal/identity/service/service_test.go`
- [X] T044 [P] [US4] Extend repository tests for named email/username conflict mapping and synchronized concurrent inserts with exactly one persisted winner in `backend/internal/identity/repository/repository_test.go`
- [X] T045 [P] [US4] Extend HTTP integration tests for combined field errors, duplicate casing, malformed JSON, unknown fields, oversized bodies, no partial account, and password omission in `backend/internal/identity/tests/integration_test.go`
- [X] T046 [P] [US4] Extend registration UI tests for field-level messages, error summary focus, preserved correctable values, cleared password, duplicate conflicts, and retry behavior in `frontend/src/auth/RegisterForm.test.tsx`

### Implementation for User Story 4

- [X] T047 [US4] Complete aggregated registration validation so every violated field rule is returned in one result without hashing or persistence in `backend/internal/identity/service/validation.go` and `backend/internal/identity/service/service.go`
- [X] T048 [US4] Map only named PostgreSQL email and username constraint violations to field conflicts while keeping other database errors internal in `backend/internal/identity/repository/repository.go`
- [X] T049 [US4] Map validation and conflict errors to `400` and `409` field-keyed responses without echoing passwords or internal details in `backend/internal/identity/controller/controller.go`
- [X] T050 [US4] Render aggregated field errors and summary, retain non-password form values, clear password after submission, and restore focus for correction in `frontend/src/auth/RegisterForm.tsx` and `frontend/src/auth/types.ts`

**Checkpoint**: Invalid and concurrent registration paths are actionable, atomic, and deterministic.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Verify integration, security boundaries, reproducibility, and constitutional simplicity.

- [X] T051 [P] Add `.env`, private-key, and certificate exclusions without adding committed secrets in `.gitignore`
- [X] T052 Audit request and error logging to exclude passwords, hashes, JWTs, signing secrets, request bodies, and full profiles in `backend/cmd/api/main.go` and `backend/internal/identity/controller/controller.go`
- [X] T053 Verify frontend avoids unsafe HTML, unnecessary third-party scripts, custom daisyUI wrappers, extra component libraries, and token logging in `frontend/src/App.tsx`, `frontend/src/auth/`, `frontend/src/main.tsx`, and `frontend/src/styles.css`
- [X] T054 Run backend formatting, `go test ./...`, `go vet ./...`, frontend tests, and production build using `backend/` and `frontend/package.json`; resolve failures only in files introduced by this feature
- [X] T055 Execute every registration, login, JWT middleware, concurrency, database, and security validation from `specs/001-user-auth/quickstart.md` and record unresolved deviations in `specs/001-user-auth/tasks.md`

---

## Phase 8: Confirmación y visibilidad de contraseña

**Purpose**: Incorporate the added registration interaction without changing backend persistence or
the registration API contract.

- [X] T056 [US1] Extend registration UI tests for required password confirmation, exact-match blocking without an API call, independent visibility controls, changing `Ver contraseña` to `Ocultar contraseña`, preserved values while toggling, clearing both secret fields after a submitted attempt, and keyboard operation in `frontend/src/auth/RegisterForm.test.tsx`
- [X] T057 [US1] Add controlled password confirmation, field-level mismatch feedback, independent accessible daisyUI visibility buttons for both password fields, and coordinated clearing after a submitted attempt while keeping the API type unchanged so confirmation cannot enter its payload in `frontend/src/auth/RegisterForm.tsx`
- [X] T058 [US1] Run the focused registration tests, complete frontend suite, production build, and updated registration checks from `specs/001-user-auth/quickstart.md`

**Checkpoint**: Registration requires two matching entries and lets users independently inspect or
mask each value without changing the backend contract.

---

## Phase 9: Separación explícita de DTO y DAO

**Purpose**: Apply constitution 1.2.0 to existing identity structures without changing behavior or
adding architectural layers.

- [X] T059 Move HTTP request, response, and error payload structures into `backend/internal/identity/dto/dto.go` and update controller mappings
- [X] T060 Move PostgreSQL query input and scanned row structures into `backend/internal/identity/dao/dao.go` and update repository/service usage
- [X] T061 Update feature plan and data-model decisions for justified DTO/DAO ownership in `specs/001-user-auth/plan.md` and `specs/001-user-auth/data-model.md`
- [X] T062 Run backend formatting, unit tests, integration compilation, and vet after the structural refactor

**Checkpoint**: HTTP contracts and persistence shapes have explicit owners while the three-layer
flow and observable behavior remain unchanged.

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Phase 1 Setup
    ↓
Phase 2 Foundation
    ├──→ US1 Registration (MVP)
    │        ↓
    ├──→ US2 Login
    │        ↓
    ├──→ US3 Temporary authenticated access
    │
    └──→ US4 Invalid-data correction

US1 + US2 + US3 + US4
    ↓
Phase 7 Polish
    ↓
Phase 8 Password confirmation follow-up
    ↓
Phase 9 DTO/DAO separation
```

- **Phase 1** has no dependencies.
- **Phase 2** depends on Phase 1 and blocks all stories.
- **US1** depends only on Phase 2.
- **US2** depends on Phase 2 plus persisted accounts from US1 for end-to-end verification.
- **US3** depends on US2 because middleware validates tokens issued by login.
- **US4** depends on US1 registration behavior but not on US2 or US3.
- **Phase 7** depends on every story selected for delivery.
- **Phase 8** depends on the existing US1 registration form and its completed baseline tests.
- **Phase 9** depends on the completed backend and preserves all existing public contracts.

### Within Each User Story

1. Write listed tests.
2. Confirm tests fail meaningfully.
3. Implement persistence prerequisites.
4. Implement service behavior.
5. Implement controller contract.
6. Implement frontend behavior.
7. Run story-specific tests.
8. Validate independent checkpoint.

### Parallel Opportunities

- T003 and T006 can run beside backend setup work after T001.
- T007, T009, and T010 use separate foundational files.
- Each story's backend unit, repository/integration, and frontend test files can be prepared in
  parallel where marked `[P]`.
- T018 and T019 can run in parallel after US1 tests exist.
- T028 can run while T029 prepares credential lookup.
- T037 can run while US3 tests are written.
- After US1 completes, US2 and US4 can proceed in parallel; US3 waits for US2.
- T051 can run independently during final verification.
- T056 must fail meaningfully before T057; T058 validates their completed behavior.

## Parallel Examples

### User Story 1

```text
T014: Service tests in backend/internal/identity/service/service_test.go
T015: Repository tests in backend/internal/identity/repository/repository_test.go
T016: HTTP integration tests in backend/internal/identity/tests/integration_test.go
T017: Registration UI tests in frontend/src/auth/RegisterForm.test.tsx
```

### User Story 2

```text
T024: Login service tests in backend/internal/identity/service/service_test.go
T025: Credential repository tests in backend/internal/identity/repository/repository_test.go
T026: Login integration tests in backend/internal/identity/tests/integration_test.go
T027: Login UI tests in frontend/src/auth/LoginForm.test.tsx
```

### User Story 3

```text
T033: JWT validation tests in backend/internal/identity/service/service_test.go
T034: Middleware tests in backend/internal/identity/controller/controller_test.go
T035: Protected-flow integration tests in backend/internal/identity/tests/integration_test.go
T036: Frontend restoration tests in frontend/src/auth/SessionStatus.test.tsx
```

### User Story 4

```text
T043: Validation boundary tests in backend/internal/identity/service/service_test.go
T044: Concurrent uniqueness tests in backend/internal/identity/repository/repository_test.go
T045: Invalid HTTP flow tests in backend/internal/identity/tests/integration_test.go
T046: Registration correction tests in frontend/src/auth/RegisterForm.test.tsx
```

## Implementation Strategy

### MVP First

1. Complete T001–T006.
2. Complete T007–T013.
3. Complete T014–T023.
4. Stop and validate US1.
5. Demonstrate valid registration.

This MVP persists a safe account. It does not claim authentication completion.

### Incremental Delivery

1. **Foundation**: T001–T013.
2. **Registration MVP**: T014–T023.
3. **Login and JWT issuance**: T024–T032.
4. **JWT middleware and restoration**: T033–T042.
5. **Complete invalid-data experience**: T043–T050.
6. **Cross-cutting verification**: T051–T055.
7. **Password confirmation follow-up**: T056–T058.

### Suggested Commit Boundaries

- Setup and foundational infrastructure.
- Registration tests and migration.
- Registration backend.
- Registration frontend.
- Login and JWT issuance.
- JWT middleware and authenticated frontend state.
- Validation and concurrent uniqueness.
- Security and quickstart verification.
- Registration password confirmation and visibility controls.

## Notes

- `[P]` means file-level parallelism, not permission to skip dependencies.
- Story labels provide traceability to `spec.md`.
- No task creates session or token tables.
- No task introduces refresh tokens, revocation, roles, MFA, recovery, or throttling.
- Controllers never access repositories directly.
- JWT `sub` remains authoritative user ID only after middleware verification.
- Password confirmation remains frontend-only and never changes persistence or HTTP contracts.
- Do not edit `README.md` during implementation.
- Automated axe checks passed in jsdom, but live screen-reader evidence and pixel-level contrast or
  focus measurements remain unverified because this Linux workspace has no NVDA, VoiceOver, or
  browser measurement harness. This blocks a strict production accessibility sign-off, not the
  academic feature behavior.
- `govulncheck` was unavailable locally. `npm audit` reported zero vulnerabilities; Go tests,
  `go vet`, race tests, PostgreSQL integration tests, and real-process HTTP checks passed.
