# Phase 0 Research: Registro e inicio de sesión

## Version baseline

**Decision**: Use Go 1.26.5, PostgreSQL 18.4, React 19.2, TypeScript 7.0, Vite 8.1, Tailwind CSS
4.3 and daisyUI 5. Node.js 26.7 is the local frontend toolchain.

**Rationale**: These stable versions match the available local environment or the current stable
documentation. Vite 8 supports the installed Node version. Pinning major/minor versions keeps local
and CI builds reproducible without adopting preview releases.

**Alternatives considered**: Older supported Go, PostgreSQL, React and Vite versions. They provide no
project-specific advantage for a new codebase.

**Sources**:

- https://go.dev/doc/devel/release
- https://www.postgresql.org/docs/18/release.html
- https://react.dev/versions
- https://www.typescriptlang.org/docs/handbook/release-notes/typescript-6-0.html
- https://vite.dev/releases
- https://tailwindcss.com/blog/tailwindcss-v4-3
- https://daisyui.com/docs/v5/

## Backend HTTP stack

**Decision**: Use Go `net/http` route patterns and `encoding/json` directly.

**Rationale**: Two POST operations need no web framework. The standard library handles routing,
request limits, JSON, context cancellation and `httptest` with fewer dependencies and less hidden
behavior. The same router can wrap one protected route with a focused middleware.

**Alternatives considered**: Chi, Gin and Echo. Each would add a dependency without solving a
current routing problem.

## PostgreSQL access

**Decision**: Use `pgx/v5` and `pgxpool` with `DATABASE_URL`; verify connectivity during startup.

**Rationale**: pgx provides direct PostgreSQL access, parameterized queries, SQLSTATE inspection and
a concurrency-safe pool. Registration is one atomic `INSERT`; login is one `SELECT`, so neither case
needs an explicit transaction abstraction.

**Alternatives considered**: `database/sql`, an ORM and a generic repository. `database/sql` is
viable but loses some PostgreSQL-specific ergonomics; the other options introduce unnecessary
mapping and abstraction.

**Source**: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool

## Password storage and comparison

**Decision**: Hash passwords with Argon2id using a fresh 16-byte random salt, parameters
`memory=19 MiB`, `time=2`, `threads=1`, and a 32-byte key. Store a PHC-style encoded string that
includes algorithm version, parameters, salt and hash. Compare derived keys in constant time.

**Rationale**: Password plaintext must never be stored. Argon2id is recommended for password
hashing, accepts the specified password rules without bcrypt's 72-byte limit and requires only the
official `x/crypto` module plus standard-library randomness and constant-time comparison.

**Alternatives considered**: bcrypt is simpler to call but imposes a 72-byte password limit absent
from the specification. PBKDF2 is reserved for FIPS-driven requirements. Plain SHA-256 is too fast
for password storage.

**Sources**:

- https://pkg.go.dev/golang.org/x/crypto/argon2
- https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

## Credential-enumeration resistance

**Decision**: Return one `invalid_credentials` response for an unknown username and a wrong
password. When the username is absent, execute one Argon2id comparison against a fixed valid dummy
hash before returning.

**Rationale**: Equal public responses satisfy the specification. Performing comparable password
work reduces timing differences without sessions, account state or an external component.

**Alternatives considered**: Returning different messages violates the spec. Adding account locks
or rate limiting contradicts the clarified out-of-scope decision.

## Identity normalization and uniqueness

**Decision**: Trim outer whitespace from names, email and address fields. Lowercase email and
username before persistence. Keep phone and password literal. Enforce ordinary unique constraints
on the normalized email and username columns and translate their named constraint violations to
field-specific registration conflicts.

**Rationale**: Storing one canonical value makes case-insensitive lookup and uniqueness explicit,
avoids duplicate normalized columns and lets PostgreSQL serialize concurrent conflicts.

**Alternatives considered**: Expression indexes on `lower(column)` preserve original casing but
make every lookup repeat the expression. `citext` adds an extension. Application-only prechecks
remain race-prone.

**Sources**:

- https://www.postgresql.org/docs/18/ddl-constraints.html
- https://www.postgresql.org/docs/current/sql-createindex.html

## Relational shape

**Decision**: Store one account and its single required residence in one `users` table. Treat the
address as a value group, not an independently identified row. Do not persist authentication
attempts.

**Rationale**: The feature defines exactly one address per account and no independent address
lifecycle. A second table would add a one-to-one join without enforcing a new rule. Login attempts
have no required history or state.

**Alternatives considered**: Separate `addresses`, `credentials` and `login_attempts` tables. Those
are justified only by multiple addresses, multiple authentication methods or audit/lockout
requirements, all absent here.

## Error contract

**Decision**: Use `400` for malformed or invalid input, `409` for duplicate email or username,
`401` with one generic body for invalid login and `500` for unexpected errors. Validation responses
contain field-keyed messages; password values and internal/database errors never appear.

**Rationale**: The UI can render actionable registration feedback while login preserves account
confidentiality. A small explicit error shape avoids a generic error framework.

**Alternatives considered**: `422` for validation is defensible but adds no value for two internal
forms. Different authentication errors violate the specification.

## Stateless JWT session

**Decision**: On successful login, issue an HS256 JWT with `sub`, `username`, `iat` and `exp`, using
`github.com/golang-jwt/jwt/v5` 5.3.1. Set `exp` exactly 30 minutes after `iat`. Load a minimum
32-byte signing secret from mandatory `JWT_SECRET`, pin validation to HS256 and reject a token at
`now >= exp`. A middleware validates Bearer tokens and places only verified user ID and username in
a private typed request context. `GET /api/auth/me` demonstrates the protected flow.

**Rationale**: JWT satisfies the required stateless session without a table or extra service. A
maintained library reduces cryptographic and standards mistakes compared with hand-written JWT
code. One HMAC key is proportionate for a single monolith that both issues and validates tokens.
The narrow request context avoids repeated parsing and does not become a dependency container.

**Alternatives considered**: Opaque server-side sessions require persistence; refresh tokens and
revocation lists add lifecycle state outside scope; asymmetric signing adds key-management cost
without a separate issuer or verifier; a custom JWT implementation is security-sensitive and
unnecessary. Applying middleware without a protected operation would create unused infrastructure.

**Sources**:

- https://github.com/golang-jwt/jwt
- https://pkg.go.dev/github.com/golang-jwt/jwt/v5
- https://datatracker.ietf.org/doc/html/rfc7519

## Frontend token storage

**Decision**: Store only the JWT in browser `sessionStorage`. After login, attach it as
`Authorization: Bearer <token>`. On application load, call `/api/auth/me` to restore verified
identity. Remove the token when its expiration is reached locally or a protected request returns
`401`.

**Rationale**: `sessionStorage` meets the frontend-storage requirement and survives reloads while
remaining limited to one tab. Backend validation remains authoritative; decoded client claims do
not establish identity. This is simpler than a global store and leaves no long-lived browser
credential after the tab closes.

**Alternatives considered**: `localStorage` persists beyond the tab and lengthens exposure; an
HttpOnly cookie contradicts the explicit frontend token-storage and Bearer-token requirement;
memory-only storage loses the session on refresh. Browser storage remains accessible to injected
scripts, so avoiding unsafe HTML and third-party scripts is a required frontend constraint.

## Frontend composition

**Decision**: Use two controlled React forms and local component state. `App` switches the visible
form; a small `api.ts` owns fetch calls. Integrate Tailwind through `@tailwindcss/vite`, import
Tailwind and register daisyUI in one CSS entrypoint, then compose daisyUI `fieldset`, `input`,
`button`, `alert` and `loading` patterns. Use semantic labels, field-level errors, an error summary,
disabled submit while pending and focus movement to feedback. Registration adds a required local
password confirmation and one independent `Ver contraseña`/`Ocultar contraseña` button beside each
password field. A mismatch blocks the API call, and confirmation never enters the request DTO.

**Rationale**: Local state is sufficient for two independent screens. daisyUI provides the required
visual vocabulary without JavaScript runtime components, and Tailwind's Vite integration needs
minimal configuration. User-centered queries in React Testing Library verify behavior and
accessibility without coupling tests to class names or component internals.
Keeping confirmation and both visibility states inside `RegisterForm` satisfies the interaction
without a form library, shared store or backend contract change.

**Alternatives considered**: React Router, a global store, a form library, another component library
and a custom daisyUI theme. None solves a current navigation, state-sharing or visual requirement.

**Sources**:

- https://testing-library.com/docs/react-testing-library/intro/
- https://vitest.dev/guide/
- https://tailwindcss.com/docs/installation/using-vite
- https://daisyui.com/docs/v5/

## Testing and reproducibility

**Decision**: Use Go unit tests for normalization, validation and password helpers; use real
PostgreSQL integration tests for queries, unique constraints, concurrent registration and full HTTP
flows. Use Vitest, jsdom, React Testing Library and user-event for four UI behavior groups. Run a
PostgreSQL 18.4 container through Compose; use separate `DATABASE_URL` and `TEST_DATABASE_URL`.

**Rationale**: Business rules remain testable without HTTP, while PostgreSQL-specific guarantees are
verified by PostgreSQL rather than SQL mocks. Compose adds no runtime service beyond the mandated
database and makes local validation repeatable.

**Alternatives considered**: Driver mocks cannot prove constraints or concurrent uniqueness.
Testcontainers adds a dependency when Compose already provides the required database.

## Explicit security tradeoff

**Decision**: Do not implement rate limiting, account lockout, persisted sessions, refresh tokens,
revocation, recovery, email verification, MFA, roles or permissions in this feature. Never log
request bodies, passwords, password hashes, JWTs, signing secrets or full personal profiles.

**Rationale**: The omissions are explicit product decisions. Logging only operation name, outcome,
safe error category and request ID preserves basic diagnosis without expanding persistent identity
state or leaking credentials. A stolen JWT cannot be revoked before its fixed 30-minute expiry;
this is an accepted non-production limitation.

**Alternatives considered**: Adding these controls now would contradict scope and constitution.
Their absence, especially login throttling, remains a known non-production limitation.
