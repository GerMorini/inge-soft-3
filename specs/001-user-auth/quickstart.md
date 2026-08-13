# Quickstart Validation: Registro e inicio de sesión

This guide describes the commands and end-to-end checks the implementation must support. Contract
details live in [contracts/openapi.yaml](contracts/openapi.yaml); persistence rules live in
[data-model.md](data-model.md).

## Prerequisites

- Go 1.26.5
- Node.js 26.7 and npm 12
- Docker with Compose
- `psql` 18.x for direct migration commands

## Run the complete stack

Copy `.env.example` to the ignored `.env` file. Generate the two values below once, place them in
`.env`, and keep the PostgreSQL password unchanged while the persisted database volume exists:

```bash
openssl rand -hex 24
openssl rand -base64 32
docker compose up --build
```

Open `http://localhost:3000`. Compose waits for PostgreSQL, applies the migration, starts the Go API
and serves the React build through nginx. Requests under `/api` are proxied to the backend. Set
`APP_PORT` before starting Compose to use another host port.

Stop the stack without deleting PostgreSQL data:

```bash
docker compose down
```

## Environment

Use local-only credentials. Do not commit either URL.

```bash
export POSTGRES_PASSWORD="$(openssl rand -hex 24)"
export DATABASE_URL="postgres://app:${POSTGRES_PASSWORD}@localhost:5432/inge_soft_3?sslmode=disable"
export TEST_DATABASE_URL="postgres://app:${POSTGRES_PASSWORD}@localhost:5433/inge_soft_3_test?sslmode=disable"
export HTTP_ADDR='127.0.0.1:8080'
export JWT_SECRET="$(openssl rand -base64 32)"
```

Generate a new local value for `JWT_SECRET`. Never commit it or print it in logs.

## Start databases

The planned `compose.yaml` exposes an application database on port 5432 and an isolated test
database on port 5433.

```bash
docker compose --profile test up -d db test-db
psql "$DATABASE_URL" -f backend/migrations/001_create_users.up.sql
psql "$TEST_DATABASE_URL" -f backend/migrations/001_create_users.up.sql
```

## Install and verify

```bash
cd backend
go mod download
go test ./...
go vet ./...

cd ../frontend
npm ci
npm run test -- --run
npm run build
```

Expected result: at least 8 meaningful backend behaviors and 4 frontend behavior groups pass. The
backend integration suite must use only `TEST_DATABASE_URL` and must refuse an unidentified
database.

## Run locally

Terminal 1:

```bash
cd backend
DATABASE_URL="$DATABASE_URL" HTTP_ADDR="$HTTP_ADDR" JWT_SECRET="$JWT_SECRET" go run ./cmd/api
```

Terminal 2:

```bash
cd frontend
npm run dev
```

The Vite development server proxies `/api` to `http://127.0.0.1:8080`, avoiding a separate CORS
policy for local development.

## Validate registration

1. Open the registration form.
2. Enter a password satisfying all four password rules and repeat it exactly in the confirmation
   field.
3. Use `Ver contraseña` independently in both fields; confirm each button reveals only its own
   value, changes to `Ocultar contraseña`, restores masking, and never alters either value.
4. Change one confirmation character and submit; confirm no request is sent and the mismatch is
   identified beside the confirmation field.
5. Restore the matching confirmation, submit valid personal data, an E.164 phone, an address,
   username `Alumno_01` and a unique email.
6. Confirm success shows canonical lowercase username/email and never shows either password value.
7. Repeat with the same username in different casing and confirm a username conflict.
8. Repeat with the same email in different casing and confirm an email conflict.
9. Submit invalid phone, username and password values together; confirm every affected field shows
   feedback, other values remain, and password plus confirmation are empty.
10. Submit without apartment and confirm registration succeeds.
11. Confirm fields, buttons, alerts and pending indicators use daisyUI components with the default
   theme and remain operable by keyboard.

Expected HTTP behavior is `201`, `400` or `409` as specified in the OpenAPI contract.

## Validate login

1. Open the login form.
2. Enter the registered username with different casing and its exact password.
3. Confirm authentication returns a Bearer JWT with `expiresIn` equal to `1800` and the frontend
   stores it in `sessionStorage`.
4. Reload the page and confirm `/api/auth/me` restores the authenticated user while the token is
   valid.
5. Enter a wrong password; record the message and status.
6. Enter an unknown username; confirm the same message and status appear.
7. Perform several failed attempts, then use valid credentials; confirm no lockout or delay was
   introduced.

Expected HTTP behavior is `200`, `400` or `401` as specified in the OpenAPI contract.

## Validate token middleware

1. Call `GET /api/auth/me` with `Authorization: Bearer <token>` from a successful login and confirm
   the returned ID and username match that account.
2. Repeat without the header, with a malformed token, with an altered signature and with an
   unsupported signing algorithm; confirm every request returns the same `401 invalid_token` body.
3. Run the controlled-clock backend test at one instant before expiry and exactly at expiry; confirm
   the first request succeeds and the second is rejected.
4. Present an expired token to the frontend or return `401` from `/api/auth/me`; confirm the token is
   removed from `sessionStorage` and the unauthenticated forms return.
5. Query PostgreSQL before and after login and protected calls; confirm only the `users` table exists
   for this feature and no user row changes.

## Validate concurrent uniqueness

Run the backend integration test that synchronizes two registration requests with the same
canonical username or email. Exactly one request must create a row; the other must receive a
field-specific conflict. The database must contain exactly one matching user afterward.

## Security checks

- Inspect application logs: no password, hash, request body or full personal profile appears.
- Inspect the user row: only a PHC-style Argon2id hash is stored, never plaintext.
- Confirm unknown username and wrong password return identical public error bodies.
- Confirm logs contain neither JWT values nor `JWT_SECRET`.
- Confirm login JWT claims contain only `sub`, `username`, `iat` and `exp`.
- Confirm middleware accepts only HS256 and rejects tokens at `now >= exp`.
- Confirm the production frontend build includes daisyUI styles and no second component library.
- Run `go vet ./...`; when available in CI, also run `govulncheck ./...`.
