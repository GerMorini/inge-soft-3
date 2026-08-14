# Implementation Plan: Rutinas de ejercicios

**Branch**: `002-exercise-routines` | **Date**: 2026-08-13 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-exercise-routines/spec.md`

## Summary

Conservar las reglas backend y el modelo relacional, ampliar `SessionSummary` con el conteo derivado
de ejercicios requerido por sus cards, y rediseñar el frontend como FitPro. La vista no autenticada
usará una composición visual con la imagen de gimnasio
suministrada y una card que alterna login y registro. El workspace autenticado incorporará navbar,
hero por apartado, cards progresivas y detalles modales. Creación y edición compartirán un wizard
modal accesible con tabs navegables libremente, búsqueda local y resumen final. Se mantendrán estado
local, daisyUI y Tailwind; `lucide-react` será la única dependencia nueva y no se añadirá router,
store global, librería de formularios, drag-and-drop ni modal externo.

## Technical Context

**Language/Version**: Go 1.26.5 (backend), TypeScript 7.0 + React 19.2 (frontend)

**Primary Dependencies**: Go standard library and existing `github.com/jackc/pgx/v5` 5.10;
React 19.2, daisyUI 5.7 and Tailwind CSS 4.3; add pinned `lucide-react` 1.31.0 for the explicitly
required icon set, imported by named icon for tree shaking

**Storage**: PostgreSQL 18.4; existing five relational tables from migration 002; no new migration

**Testing**: Go `testing`, `httptest` and PostgreSQL integration tests with the existing
`integration` build tag; Vitest 4, React Testing Library and user-event

**Target Platform**: Existing Linux containers and evergreen browsers supported by Vite 8

**Project Type**: Modular monolith web application

**Constraints**: Every data operation keeps the existing JWT and ownership rules; no schema change
and only one additive summary field; no pagination, partial update, edit history, optimistic versioning, file upload,
backend search, RLS, ORM, router, form library, drag-and-drop library or global state. Wizard steps
must remain freely navigable; backdrop exit confirms; save never confirms; external media needs
lazy loading, safe fallback and no backend proxy. Global CSS remains theme/document-only.

**Scale/Scope**: Existing backend module, five tables and fifteen operations remain; session list
adds one grouped count without endpoint or migration. One auth shell, one authenticated shell, three
catalog views, three concrete wizards and four small shared UI components serve hundreds of local
catalog items per user.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Pre-design gate**: PASS. No constitutional violations require justification.

- [x] Scope is deliberately small and contains no speculative capabilities.
- [x] Design stays within one React + Go + PostgreSQL modular monolith.
- [x] Backend remains grouped in one cohesive `routines` module and follows
      `controller -> service -> repository` without forbidden direct dependencies.
- [x] DTO and DAO packages contain concrete HTTP and persistence shapes only; no additional layer,
      generic mapper or duplicate domain model is introduced.
- [x] Data design is relational, minimal, constrained where reasonable, and uses no JSON columns.
- [x] Existing environment configuration and secret handling remain unchanged.
- [x] Twenty backend and twenty-three frontend behaviors have explicit unit or integration coverage.
- [x] Frontend uses daisyUI, Tailwind utilities and local state. `ModalDialog`, `WizardDialog`,
      `SearchableCatalog` and `MediaPreview` remove behavior repeated by current flows.
- [x] Global CSS remains limited to Tailwind/daisyUI setup, theme tokens and document-wide rules;
      feature-specific appearance stays beside routines components or in their utility classes.
- [x] Dark surfaces and semantic primary, secondary, accent and error colors reuse the palette
      fixed by constitution 1.3.0, with no alternate palette.
- [x] `lucide-react` is the sole new dependency, required explicitly and limited to named SVG icon
      imports. It has no runtime service requirement and replaces hand-authored icon duplication.
- [x] Existing build, test, migration and Docker Compose workflows remain reproducible after adding
      migration 002 to the migration service.

**Post-design re-check**: PASS. Schema, ownership and write contracts remain intact. Session list
adds only the derived `exerciseCount` required by current cards, through the existing layered flow.
Frontend
adds only abstractions demanded by three repeated modal workflows. Native `<dialog>`, independent
`<details>`, search inputs, checkboxes and buttons retain browser semantics. The image becomes a
local optimized asset, avoiding runtime hotlink dependency. Lucide is justified by the explicit
requirement and uses `currentColor`; no second component library or design system is introduced.

## Project Structure

### Documentation (this feature)

```text
specs/002-exercise-routines/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── openapi.yaml         # unchanged backend contract
│   └── ui.md                # new visual and interaction contract
└── tasks.md                 # generated by /speckit-tasks
```

### Source Code (repository root)

```text
backend/
├── cmd/api/main.go
├── internal/
│   ├── identity/controller/
│   │   ├── controller.go    # exposes existing authentication wrapper
│   │   └── middleware.go
│   ├── platform/requestctx/identity.go
│   └── routines/
│       ├── types.go         # minimal module results shared across layer boundaries
│       ├── controller/
│       │   ├── controller.go
│       │   └── controller_test.go
│       ├── service/
│       │   ├── service.go
│       │   ├── validation.go
│       │   └── service_test.go
│       ├── repository/
│       │   ├── repository.go
│       │   └── repository_test.go
│       ├── dto/dto.go
│       ├── dao/dao.go
│       ├── errors/errors.go
│       └── tests/
│           ├── integration_test.go
│           └── testdb_test.go
└── migrations/
    ├── 002_create_exercise_routines.up.sql
    └── 002_create_exercise_routines.down.sql

frontend/src/
├── App.tsx
├── styles.css               # global theme, tokens and document-wide rules only
├── assets/
│   └── fitpro-gym.webp      # optimized local copy of the supplied image
├── auth/
│   ├── AuthShell.tsx
│   ├── AuthShell.test.tsx
│   ├── LoginForm.tsx
│   ├── RegisterForm.tsx
│   ├── SessionStatus.tsx
│   └── session.ts
└── routines/
    ├── api.ts
    ├── types.ts
    ├── components/
    │   ├── ModalDialog.tsx
    │   ├── ModalDialog.test.tsx
    │   ├── WizardDialog.tsx
    │   ├── WizardDialog.test.tsx
    │   ├── SearchableCatalog.tsx
    │   ├── SearchableCatalog.test.tsx
    │   └── MediaPreview.tsx
    ├── ExerciseWizard.tsx
    ├── SessionWizard.tsx
    ├── RoutineWizard.tsx
    ├── ExercisesView.tsx
    ├── ExercisesView.test.tsx
    ├── SessionsView.tsx
    ├── SessionsView.test.tsx
    ├── RoutinesView.tsx
    └── RoutinesView.test.tsx

frontend/nginx.conf          # CSP permits requested HTTP/HTTPS image and media previews

compose.yaml
```

**Structure Decision**: `routines` owns exercises, workout sessions, routines and both associations
because they form one cohesive capability and share transactional rules. Controller maps HTTP DTOs,
extracts trusted identity from `requestctx` and maps errors. Service validates business rules,
authorizes selections and owns transactions. Repository scopes every query by user ID and performs
parameterized SQL over DAO shapes that remain private to repository. Repository converts those DAO
rows into the minimal module result structures declared in `internal/routines/types.go` before
returning. Service accepts and returns only module/use-case structures, and controller maps them
explicitly into HTTP DTOs. This single small boundary prevents persistence shapes from reaching
controller without creating a chain of model/entity abstractions. No service or repository
interfaces, generic CRUD, unit-of-work, mapper, ORM or additional domain layer are planned.

Backend write behavior remains unchanged. Session listing uses one owner-scoped `LEFT JOIN` plus
`COUNT(session_exercises.exercise_id)` grouped by session to expose `exerciseCount`; repository maps
the aggregate DAO into a module session summary and controller maps the service result to the
additive DTO field. No per-card detail requests, interface or migration
are introduced. Frontend keeps each view responsible for loading,
list state, detail retrieval, deletion and API submission. Every concrete entity wizard owns its
normalized immutable baseline, current draft, dirty comparison, step validation and payload
conversion. Exercise compares scalar fields, session compares its ordered exercise composition and
routine compares canonical `(sessionId, day)` pairs. Shared `WizardDialog` owns tabs and footer but knows
nothing about entities or HTTP. `ModalDialog` wraps native `<dialog>` lifecycle and focus return;
`SearchableCatalog` owns local filtering, result status and labelled result layout. Session renders
unique exercise checkboxes; routine renders repeatable Add actions so a session remains available
after every assignment. `MediaPreview` centralizes safe image/video loading and fallback required by
exercise cards and nested routine details without interpreting URLs in the backend.

The supplied 5177×3410 JPEG is fetched once during implementation, resized/compressed into the
versioned `fitpro-gym.webp` asset and referenced through Vite imports. No remote request is required
at runtime. Authentication and heroes render it as an actual decorative `<img>` beneath overlays so
load failure can hide it and preserve a dark fallback. Component appearance remains in daisyUI and
Tailwind classes beside its owner; no feature rules enter `styles.css`.

## Persistence and Transaction Design

- `exercises`, `workout_sessions` and `routines` use `(user_id, id)` primary keys.
- `session_exercises` uses `(user_id, session_id, exercise_id)` as primary key and a deferrable
  unique constraint on `(user_id, session_id, execution_order)`.
- `routine_sessions` uses `(user_id, routine_id, day_of_week, session_id)` as primary key.
- Composite foreign keys prevent cross-user associations even if application validation fails.
- Association foreign keys use `ON DELETE CASCADE`; deleting a user remains outside this feature.
- Session and routine creation validate selected IDs with tenant-scoped locking, insert the parent
  and associations, then commit as one operation.
- Exercise update is one tenant-scoped `UPDATE ... RETURNING` and needs no explicit transaction.
- Session update validates all pure rules, begins a transaction and performs an unlocked,
  tenant-scoped target existence probe. It then locks selected exercises by ascending ID with
  `FOR KEY SHARE`, locks the target session with `FOR UPDATE`, resolves target absence before any
  recorded child-availability error, updates fields, deletes old associations and inserts the
  complete new composition before commit.
- Routine update follows the same sequence: pure validation, scoped target probe, selected-session
  locks by ascending ID, target routine lock, error-priority resolution, field update and complete
  association replacement.
- Global lock order remains exercises, sessions, routines. `READ COMMITTED` is sufficient. Any
  failure rolls back fields and associations to the previous complete state.
- Empty association lists are valid. No migration, timestamp, version, audit history, soft delete,
  retry framework or optimistic-lock mechanism is introduced. Concurrent writers use the last
  committed complete replacement.
- Exercise deletion locks its target and affected sessions, defers the execution-order uniqueness
  constraint, deletes associations through cascade and applies `row_number()` to compact remaining
  orders before commit.
- Session and routine details each use one tenant-scoped flat `LEFT JOIN` statement and aggregate
  nullable rows in repository. Under `READ COMMITTED`, one statement receives one committed MVCC
  snapshot, so a response cannot combine parent fields or associations from opposite sides of a
  concurrent edit. Empty containers still produce a parent row. PUT reads its response through the
  same detail SQL inside the write transaction before commit, then commits before writing HTTP.
  No read transaction, executor interface, JSON aggregation, view or materialization is introduced.
  Exercise detail remains one statement.
- Details continue joining referenced exercise and session rows rather than snapshots, so editing a
  reusable child appears in every subsequent containing detail while association values remain.

## HTTP and Interface Design

- Authenticated `/api/exercises`, `/api/sessions` and `/api/routines` resources support POST, GET
  collection, GET detail, PUT detail and DELETE detail.
- Each PUT uses the same complete writable shape as creation and returns `200` with `Exercise`,
  `SessionDetail` or `RoutineDetail`. Optional empty/omitted fields are cleared; required association
  arrays may be empty. IDs and owner are never accepted from the body.
- Collection responses are direct arrays because pagination and response metadata are out of scope.
- Details expose nested reusable content but never user IDs, secrets or persistence-only fields.
- `400 validation_failed` reports every detectable field or association error; indexed keys identify
  nested inputs. `404 not_found` is identical for absent and foreign resources. Existing
  `401 invalid_token` behavior is reused.
- Non-numeric, non-positive and out-of-range path IDs return `400 invalid_request`. For a
  structurally and semantically valid PUT, the service resolves target absence/ownership as
  `404 not_found` before returning unavailable selected references. Pure validation remains first.
- Lists use ID order; session exercises use execution order; routine assignments use day and session
  ID. No unsupported user-defined ordering is implied.
- Existing write/detail contracts remain unchanged; OpenAPI receives only the additive session
  summary count required by cards.
- `GET /api/sessions` adds required non-negative `exerciseCount` to each `SessionSummary`; this is
  computed and transported as int64 in the existing owner-scoped list query, avoiding N+1 details.
- Unauthenticated `App` renders `AuthShell` without navbar. Login is initial; its exact registration
  link swaps the card locally. Registration retains every spec 001 field and offers a reverse link.
- Authenticated `App` renders a responsive FitPro navbar. `Dumbbell` brands the app; centered controls
  select routines, sessions and exercises through existing local state and keyboard behavior.
- Each view renders the shared image as a hero with title and fixed copy: “Visualiza y ajusta tus
  rutinas”, “Organiza tus sesiones de entrenamiento” and “Administra tu catálogo de ejercicios”.
- Routine card main content is a labelled button that opens `ModalDialog`; Edit/Delete remain
  separate footer actions aligned right, avoiding nested buttons. Routine details use nested native
  `<details>` for sessions and exercises. Session and exercise cards use `<details>` for progressive
  disclosure, with Edit/Delete outside their `<summary>` and aligned lower right. Disclosures remain
  independent: opening one does not close siblings or nested items.
- `WizardDialog` uses native `<dialog>` plus daisyUI. Its tablist implements roving focus, arrow/Home/
  End keys, `aria-selected` and tabpanels. Any tab is selectable without intermediate validation.
- Exercise wizard has `Datos básicos` (name, description, image URL, video URL) and `Resumen`.
  Session wizard has `Datos básicos`, `Ejercicios` (searchable multi-select, selected ordered list,
  series/repetition spinboxes and move controls) and `Resumen`. Routine wizard has `Datos básicos`,
  `Sesiones` (searchable catalog with a repeatable Add action and selected assignment rows with day
  selects) and `Resumen`. Every repeated selection appends a row with unset day; Save rejects unset
  days or duplicate `(sessionId, day)` pairs.
- Full client validation runs only on Save. Invalid submissions move to the first affected step,
  focus the error summary and preserve all values. `Guardar` exists only in `Resumen`; previous
  steps expose tabs, Previous and Next without writing. Server remains authoritative.
- A backdrop click always calls `window.confirm('Los cambios se perderán, ¿seguro deseas salir?')`.
  Reject keeps dialog, active step and draft; accept closes without POST/PUT. Clicks inside never
  reach that path. Escape, Close, Cancel and workspace change reuse the existing dirty guard for
  exercise, session and routine wizards; rejecting discard preserves each normalized draft and its
  active step, while accepting closes without a write.
- Saving never asks confirmation, disables duplicate submit and blocks backdrop/Escape closure.
  Success applies the POST/PUT response to the card, closes the wizard, restores logical focus and
  announces success without GET or automatic detail modal. Failure keeps the wizard open.
- Image previews use `<img loading="lazy" referrerPolicy="no-referrer">` with a dark placeholder.
  Direct videos use native `<video controls playsInline preload="metadata">` only after expansion. Unsupported,
  failed, YouTube/Vimeo or HTML page URLs retain a safe external link with `target="_blank"`
  `rel="noopener noreferrer"`; no iframe or platform integration is attempted.
- Production nginx CSP adds `http:` and `https:` to `img-src` and defines
  `media-src 'self' http: https:`. Existing default, script, style, object and framing restrictions
  remain unchanged; global `Referrer-Policy: no-referrer` stays active. Media mounts only after
  disclosure expansion.
- UI uses `#1c1d1e` for main background, `#252728` for surfaces, `#b6ff57` for primary actions,
  `#5d58f3` for selection and links, `#ff7cff` for non-critical accent, and `#ee0000` only for
  errors or destructive actions, exposed through daisyUI semantic classes rather than component hex.
- Lucide named imports cover brand, navigation, create/view/edit/delete, close/save/search,
  Eye/EyeOff password toggles, move up/down and media links. Icons beside text are `aria-hidden`;
  icon-only close buttons receive contextual `aria-label`. SVGs inherit `currentColor` and never
  carry semantic state alone.

## Verification Strategy

- Pure service tests cover text, URL, integer, day, duplicate and consecutive-order rules.
- PostgreSQL tests cover composite ownership constraints, tenant-scoped queries, atomic creation,
  atomic full replacement, rollback, association cascades, exercise-order compaction and reuse with
  independent values. Concurrent replacement and read tests prove complete, non-mixed final states
  and snapshot-consistent nested details.
- HTTP integration tests obtain real JWTs through spec 001 and prove authenticated contracts plus
  indistinguishable foreign/missing behavior, invalid path handling and target-before-reference
  error priority. Repository/controller tests verify owner-scoped session counts for empty and
  populated sessions without multiplying list rows.
- Shared component tests cover native dialog open/close, focus restoration, free tab navigation,
  backdrop confirmation, pending closure blocking, local catalog filtering and empty results. A
  jsdom setup shim supplies only missing `showModal`/`close` behavior.
- Entity tests cover exact steps, prefill, normalized dirty comparison, accepted/rejected discard
  through Close, Cancel, Escape and workspace change, summaries, complete POST/PUT payloads, empty associations,
  ordering, repeat-selection for distinct routine days, duplicate-pair rejection, Save only on
  Summary, response-driven close without GET, invalid-step focus, direct/failed video fallback,
  independent disclosures and server rejection preservation.
- Auth/App tests cover no-navbar unauthenticated shell, login/register link swaps, image fallback,
  FitPro navigation, heroes, dirty workspace changes and authentication loss.
- Accessibility tests query controls by accessible name rather than SVG, run axe on auth, cards,
  detail modals and every wizard, and verify focus entry/return. Browser validation covers native
  dialog focus containment, contrast, 320/768/1280 widths and 200% zoom because jsdom cannot.
- Docker Compose validation applies migrations 001 then 002 and exercises the flows through the
  frontend proxy.

## Complexity Tracking

`lucide-react@1.31.0` is an intentional dependency justified by the explicit icon requirement. It
provides maintained, typed, tree-shakeable React SVG components and avoids local duplicated SVGs.
Only named imports are allowed. No other exception or constitutional violation exists.
