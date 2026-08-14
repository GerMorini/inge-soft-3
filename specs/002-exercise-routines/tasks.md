# Tasks: Rutinas de ejercicios

**Input**: Design documents from `/specs/002-exercise-routines/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`,
`contracts/openapi.yaml`, `quickstart.md`

**Tests**: Required. The specification declares 20 backend and 10 frontend behavior groups. Tests
are written before their corresponding implementation and protect validation, ownership,
transactions, nested details, ordering, deletion and authentication behavior.

**Organization**: Tasks are grouped by user story. Each phase ends with an independently verifiable
increment. Paths follow the single `routines` backend module and the existing React application.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it changes different files and has no unfinished dependency.
- **[Story]**: Maps work to `US1`, `US2`, `US3`, `US4` or `US5` from `spec.md`.
- Every implementation task names its concrete target path.

## Phase 1: Setup (Shared Structure)

**Purpose**: Add the relational schema and reproducible migration/test foundations shared by every
story without introducing new dependencies.

- [X] T001 Create reversible migration 002 with `exercises`, `workout_sessions`, `session_exercises`, `routines`, and `routine_sessions`; tenant-safe composite keys; named text, URL, quantity, order, and weekday constraints; association cascades; and only planned indexes in `backend/migrations/002_create_exercise_routines.up.sql` and `backend/migrations/002_create_exercise_routines.down.sql`
- [X] T002 Update the migration service to apply migrations 001 and 002 in deterministic order while preserving `ON_ERROR_STOP` in `compose.yaml`
- [X] T003 Add routines integration database setup that requires `TEST_DATABASE_URL`, rejects databases without the `_test` suffix, applies migrations 001 and 002, and truncates the five routines tables plus users safely in `backend/internal/routines/tests/testdb_test.go` and `backend/internal/routines/repository/repository_test.go`

**Checkpoint**: PostgreSQL schema, rollback order, Docker migration flow and isolated test setup are
ready for story implementation.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Reuse authentication and establish the smallest shared routines errors/HTTP foundation.

**⚠️ CRITICAL**: Complete this phase before any user story.

- [X] T004 [P] Write identity controller tests proving the reusable authentication wrapper rejects missing, malformed, altered and expired Bearer tokens and propagates only verified request identity in `backend/internal/identity/controller/controller_test.go`
- [X] T005 Expose the existing JWT authentication middleware as a concrete wrapper usable from the composition root without exporting token parsing or adding an interface in `backend/internal/identity/controller/controller.go` and `backend/internal/identity/controller/middleware.go`
- [X] T006 [P] Define only shared routines `ErrNotFound` and field-keyed `ValidationError` classifications required by controllers and services in `backend/internal/routines/errors/errors.go`

**Checkpoint**: New routes can reuse one trusted identity source and one minimal module error vocabulary.

---

## Phase 3: User Story 1 - Crear ejercicios propios (Priority: P1) 🎯 MVP

**Goal**: Authenticated users create, list and inspect exercises with optional reference fields while
all data remains private to its owner.

**Independent Test**: Register two users, authenticate both, create exercises with and without
optional fields as the first user, verify only that user can list or inspect them, and confirm a
known foreign ID and an unused ID produce identical public results for the second user.

### Tests for User Story 1 *(write first and confirm meaningful failure)*

- [X] T007 [P] [US1] Write table-driven service tests for valid names/descriptions, omitted optionals, empty optionals, whitespace violations, 100/500/2048 boundaries, and absolute HTTP/HTTPS URL acceptance/rejection in `backend/internal/routines/service/service_test.go`
- [X] T008 [P] [US1] Write PostgreSQL repository tests for exercise creation, optional NULL persistence, ID-ordered owner lists, scoped detail, duplicate names, and indistinguishable foreign/missing lookup in `backend/internal/routines/repository/repository_test.go`
- [X] T009 [P] [US1] Write HTTP integration tests for authenticated exercise POST/list/detail contracts, aggregated field errors, unknown JSON fields, request limits, absent token rejection, and cross-user non-disclosure in `backend/internal/routines/tests/integration_test.go`
- [X] T010 [P] [US1] Write frontend tests for exercise loading/empty/error states, valid creation, field-level text/URL feedback, optional values, safe external links, and expired-authentication fallback in `frontend/src/routines/ExercisesView.test.tsx`

### Implementation for User Story 1

- [X] T011 [P] [US1] Define exercise HTTP request/response and shared routines error payload structures with exact JSON fields from OpenAPI in `backend/internal/routines/dto/dto.go`
- [X] T012 [P] [US1] Define exercise insert and scanned-row persistence structures without JSON tags or business rules in `backend/internal/routines/dao/dao.go`
- [X] T013 [US1] Implement parameterized owner-scoped exercise create, list and detail queries with explicit columns and deterministic ordering in `backend/internal/routines/repository/repository.go`
- [X] T014 [US1] Implement exercise inputs/results, exact text validation, optional normalization to absence, HTTP/HTTPS URL validation, aggregated field errors, and owner-scoped use cases in `backend/internal/routines/service/validation.go` and `backend/internal/routines/service/service.go`
- [X] T015 [US1] Implement bounded JSON handling, trusted identity extraction, exercise POST/list/detail response mapping, generic foreign/missing `404`, logging of technical failures only, protected route registration, and composition wiring in `backend/internal/routines/controller/controller.go` and `backend/cmd/api/main.go`
- [X] T016 [P] [US1] Add typed authenticated exercise API calls, token attachment, existing error-shape parsing, and `401` token cleanup in `frontend/src/routines/types.ts` and `frontend/src/routines/api.ts`
- [X] T017 [US1] Implement the daisyUI exercise catalog/form/detail view with explicit states and connect authenticated local navigation without a router or global store in `frontend/src/routines/ExercisesView.tsx`, `frontend/src/auth/SessionStatus.tsx`, and `frontend/src/App.tsx`

**Checkpoint**: Exercise creation and private consultation work independently as the feature MVP.

---

## Phase 4: User Story 2 - Crear sesiones reutilizables (Priority: P2)

**Goal**: Authenticated users create empty or composed sessions from their own exercises, preserving
non-negative quantities and an explicit consecutive execution order.

**Independent Test**: Seed owned exercises, reorder a selection, create a session, inspect its full
ordered exercise information, reuse one exercise with different quantities elsewhere, and prove a
duplicate, invalid order or foreign selection rolls back the whole creation.

### Tests for User Story 2 *(write first and confirm meaningful failure)*

- [X] T018 [P] [US2] Extend service tests for empty sessions, non-negative integer boundaries, duplicate exercise IDs, order values exactly `1..N`, aggregated indexed errors, and foreign/unavailable selections in `backend/internal/routines/service/service_test.go`
- [X] T019 [P] [US2] Extend PostgreSQL tests for tenant-safe session associations, atomic parent/association insertion, empty sessions, ordered detail, independent reuse values, composite FK rejection, and rollback when a selected exercise disappears in `backend/internal/routines/repository/repository_test.go`
- [X] T020 [P] [US2] Extend HTTP integration tests for session POST/list/full-detail contracts, decimal/overflow JSON rejection, duplicate/gapped order validation, empty composition, unavailable exercise non-disclosure, and no partial rows in `backend/internal/routines/tests/integration_test.go`
- [X] T021 [P] [US2] Write frontend tests for loading owned exercises, single selection per exercise, series/repetition inputs, `Subir`/`Bajar` keyboard operation, derived visible order `1..N`, exact submission payload, and full session detail in `frontend/src/routines/SessionsView.test.tsx`

### Implementation for User Story 2

- [X] T022 [P] [US2] Extend HTTP DTOs with session create input, nested exercise association, summary and complete ordered detail structures in `backend/internal/routines/dto/dto.go`
- [X] T023 [P] [US2] Extend DAO shapes with workout-session rows, selected-exercise inputs and joined session-exercise detail rows in `backend/internal/routines/dao/dao.go`
- [X] T024 [US2] Implement tenant-scoped exercise selection with `FOR KEY SHARE`, session insertion through `pgx.Tx`, association insertion, owner list and full ordered-detail queries in `backend/internal/routines/repository/repository.go`
- [X] T025 [US2] Implement session validation and service-owned transaction orchestration with rollback on any unavailable exercise or association failure in `backend/internal/routines/service/validation.go` and `backend/internal/routines/service/service.go`
- [X] T026 [US2] Implement protected session POST/list/detail handlers with indexed validation errors and nested response mapping in `backend/internal/routines/controller/controller.go`
- [X] T027 [P] [US2] Extend typed authenticated API calls and frontend contracts for session summaries, composition inputs and complete details in `frontend/src/routines/types.ts` and `frontend/src/routines/api.ts`
- [X] T028 [US2] Implement the daisyUI sessions view with owned-exercise selection, non-negative numeric inputs, accessible move buttons, automatic order derivation, empty composition support and nested detail in `frontend/src/routines/SessionsView.tsx` and `frontend/src/App.tsx`

**Checkpoint**: Sessions are independently creatable, reusable, private and fully inspectable.

---

## Phase 5: User Story 3 - Crear y consultar rutinas (Priority: P3)

**Goal**: Authenticated users create empty or composed routines from their own sessions, assign ISO
weekday numbers and inspect the full nested routine through every exercise.

**Independent Test**: Seed sessions, assign one session to different days and another to the same
day, create the routine, inspect all nested values, and prove a duplicate pair, invalid day or
foreign session prevents every routine row from being created.

### Tests for User Story 3 *(write first and confirm meaningful failure)*

- [X] T029 [P] [US3] Extend service tests for days 1 and 7, days outside range, empty routines, repeated session on different days, duplicate session/day pairs, aggregated indexed errors, and unavailable session selections in `backend/internal/routines/service/service_test.go`
- [X] T030 [P] [US3] Extend PostgreSQL tests for tenant-safe routine associations, atomic insertion, empty routines, same-day different sessions, same-session different days, duplicate pair rejection, independent reuse and full nested read ordering in `backend/internal/routines/repository/repository_test.go`
- [X] T031 [P] [US3] Extend HTTP integration tests for routine POST/list/full-detail contracts, nested session/exercise information, day ordering, foreign/missing session non-disclosure and rollback without partial routine data in `backend/internal/routines/tests/integration_test.go`
- [X] T032 [P] [US3] Write frontend tests for owned-session loading, Spanish weekday display, same-session different-day reuse, duplicate pair blocking, exact request payload, routine list/empty state and complete nested detail in `frontend/src/routines/RoutinesView.test.tsx`

### Implementation for User Story 3

- [X] T033 [P] [US3] Extend HTTP DTOs with routine create assignments, summaries and complete nested routine/session/exercise detail structures in `backend/internal/routines/dto/dto.go`
- [X] T034 [P] [US3] Extend DAO shapes with routine rows, selected-session inputs and joined nested-detail rows carrying day and execution values in `backend/internal/routines/dao/dao.go`
- [X] T035 [US3] Implement tenant-scoped session selection with `FOR KEY SHARE`, routine insertion through `pgx.Tx`, assignment insertion, owner list and complete joined detail ordered by day/session/exercise in `backend/internal/routines/repository/repository.go`
- [X] T036 [US3] Implement routine validation and service-owned transaction orchestration that permits reuse only across distinct days and rolls back unavailable selections in `backend/internal/routines/service/validation.go` and `backend/internal/routines/service/service.go`
- [X] T037 [US3] Implement protected routine POST/list/detail handlers and explicit complete nested response mapping in `backend/internal/routines/controller/controller.go`
- [X] T038 [P] [US3] Extend typed authenticated API calls and frontend contracts for routine assignments, summaries and complete nested details in `frontend/src/routines/types.ts` and `frontend/src/routines/api.ts`
- [X] T039 [US3] Implement the daisyUI routines view with session/day assignment rows, ISO weekday labels, duplicate-pair prevention, list and complete nested detail states in `frontend/src/routines/RoutinesView.tsx` and `frontend/src/App.tsx`

**Checkpoint**: Users can build and inspect the complete private routine hierarchy.

---

## Phase 6: User Story 4 - Eliminar contenido propio (Priority: P4)

**Goal**: Authenticated users confirm deletion of owned content; associations disappear atomically,
reusable containers survive and affected session exercise orders remain consecutive.

**Independent Test**: Reuse content across several containers, cancel then confirm each deletion
type, verify cascades preserve reusable entities, delete exercises from first/middle/last positions,
and prove foreign and absent deletes are identical and harmless.

### Tests for User Story 4 *(write first and confirm meaningful failure)*

- [X] T040 [P] [US4] Extend PostgreSQL tests for scoped routine/session deletion, association-only cascades, exercise deletion across several sessions, first/middle/last compaction, stable relative order, transaction rollback and unchanged foreign content in `backend/internal/routines/repository/repository_test.go`
- [X] T041 [P] [US4] Extend HTTP integration tests for all three DELETE contracts, `204` empty success bodies, identical foreign/missing `404`, absent-auth rejection, preserved containers and observable compacted details in `backend/internal/routines/tests/integration_test.go`
- [X] T042 [P] [US4] Extend all three frontend view tests for cancellation without API calls, confirmed success refresh, rejection state, removed selections/details and authentication loss during delete in `frontend/src/routines/ExercisesView.test.tsx`, `frontend/src/routines/SessionsView.test.tsx`, and `frontend/src/routines/RoutinesView.test.tsx`

### Implementation for User Story 4

- [X] T043 [US4] Implement owner-scoped single-statement routine/session deletes and association cascades with affected-row not-found detection in `backend/internal/routines/repository/repository.go`
- [X] T044 [US4] Implement exercise lock, affected-session discovery/ordered locks, deferred order constraint, scoped delete and set-based `row_number()` compaction methods over one `pgx.Tx` in `backend/internal/routines/repository/repository.go`
- [X] T045 [US4] Implement delete use cases with service-owned transaction boundaries for exercise compaction and generic unavailable classification for every entity type in `backend/internal/routines/service/service.go`
- [X] T046 [US4] Implement protected DELETE handlers with positive path-ID parsing, `204` responses and identical foreign/missing error mapping in `backend/internal/routines/controller/controller.go`
- [X] T047 [P] [US4] Extend authenticated API helpers with typed delete operations and `401` cleanup in `frontend/src/routines/api.ts`
- [X] T048 [US4] Add `window.confirm` deletion flows, cancellation behavior, success refresh and explicit error states to `frontend/src/routines/ExercisesView.tsx`, `frontend/src/routines/SessionsView.tsx`, and `frontend/src/routines/RoutinesView.tsx`

**Checkpoint**: Deletion fulfills ownership, atomic cleanup and exercise-order guarantees.

---

## Phase 7: User Story 5 - Editar contenido propio (Priority: P5)

**Goal**: Authenticated users edit owned exercises, sessions and routines through complete PUT
replacement, with atomic associations, deterministic errors, consistent nested details and guarded
discard of modified forms.

**Independent Test**: Reuse an exercise and session in several containers, edit each entity, verify
the same IDs and propagated nested details, replace and empty associations, reject invalid or foreign
targets without partial changes, observe one complete state under concurrency, and confirm that
cancel/internal navigation preserves or discards dirty drafts according to the chosen response.

### Tests for User Story 5 *(write first and confirm meaningful failure)*

- [X] T049 [P] [US5] Extend service tests for update validation reuse, optional clearing, empty replacement sets, target-before-child error priority, same-ID results and complete rollback classifications in `backend/internal/routines/service/service_test.go`
- [X] T050 [P] [US5] Extend PostgreSQL tests for single-statement session/routine details including empty containers, live reusable-reference propagation, complete association replacement, rollback, target revalidation, concurrent PUT serialization and before/after-only detail snapshots in `backend/internal/routines/repository/repository_test.go`
- [X] T051 [P] [US5] Extend HTTP integration tests for all three PUT contracts, strict body decoding, `abc`/zero/negative/overflow path IDs, body-before-target and target-before-child error priority, identical foreign/missing targets, complete response bodies, JWT rejection and atomic failure in `backend/internal/routines/tests/integration_test.go`
- [X] T052 [P] [US5] Extend exercise UI tests for edit prefill, optional clearing, exact PUT payload, pending-submit protection, response-driven list/detail refresh, preserved rejection input and dirty cancel acceptance/rejection in `frontend/src/routines/ExercisesView.test.tsx`
- [X] T053 [P] [US5] Extend session UI tests for complete composition prefill, add/remove/reorder, derived `1..N`, empty replacement, exact PUT payload, dirty detection and generic unavailable errors in `frontend/src/routines/SessionsView.test.tsx`
- [X] T054 [P] [US5] Extend routine UI tests for complete assignment prefill, empty replacement, same-session distinct days, duplicate-pair rejection, canonical dirty comparison independent of visual pair order and exact PUT payload in `frontend/src/routines/RoutinesView.test.tsx`
- [X] T055 [P] [US5] Add application navigation tests proving unchanged edits switch directly, dirty cancel/tab changes call confirmation, rejected discard preserves view and draft, accepted discard switches without PUT, and authentication loss clears private state in `frontend/src/App.test.tsx`

### Implementation for User Story 5

- [X] T056 [P] [US5] Extend persistence-specific DAO shapes with update parameters and nullable flat session/routine detail rows without HTTP tags or business rules in `backend/internal/routines/dao/dao.go`
- [X] T057 [US5] Replace multi-query session and routine reads with one owner-scoped `LEFT JOIN` statement each, aggregate empty and nested rows deterministically, and make the same detail SQL callable inside a concrete `pgx.Tx` in `backend/internal/routines/repository/repository.go`
- [X] T058 [US5] Implement owner-scoped exercise `UPDATE ... RETURNING` with optional NULL clearing and unchanged identity in `backend/internal/routines/repository/repository.go`
- [X] T059 [US5] Implement session target probe, sorted selected-exercise locks, target recheck/lock, scalar update, association delete-and-insert replacement and in-transaction detail read in `backend/internal/routines/repository/repository.go`
- [X] T060 [US5] Implement routine target probe, sorted selected-session locks, target recheck/lock, scalar update, assignment delete-and-insert replacement and in-transaction nested detail read in `backend/internal/routines/repository/repository.go`
- [X] T061 [US5] Implement exercise update use case using existing text/URL validation and owner-scoped not-found behavior in `backend/internal/routines/service/service.go` and `backend/internal/routines/service/validation.go`
- [X] T062 [US5] Implement session update transaction orchestration with pure validation first, target-before-child public priority, rollback on unavailable references or persistence errors and commit-before-response semantics in `backend/internal/routines/service/service.go`
- [X] T063 [US5] Implement routine update transaction orchestration with pure validation first, target-before-child public priority, full assignment replacement and last-committed-writer behavior in `backend/internal/routines/service/service.go`
- [X] T064 [US5] Register protected PUT routes and implement positive-int64 path parsing, strict shared request DTO decoding, trusted owner extraction, complete response mapping and public error precedence in `backend/internal/routines/controller/controller.go`
- [X] T065 [P] [US5] Add typed exercise/session/routine PUT operations and complete write payloads while reusing current auth/error handling in `frontend/src/routines/types.ts` and `frontend/src/routines/api.ts`
- [X] T066 [US5] Add create/edit mode, immutable normalized baseline, prefill, save/cancel controls, focus transitions, pending state and response-driven refresh to `frontend/src/routines/ExercisesView.tsx`
- [X] T067 [US5] Add create/edit mode with complete ordered composition prefill, accessible add/remove/move controls, dirty comparison, empty replacement and preserved errors to `frontend/src/routines/SessionsView.tsx`
- [X] T068 [US5] Add create/edit mode with complete assignment prefill, canonical `(sessionId, day)` dirty comparison, empty replacement and preserved errors to `frontend/src/routines/RoutinesView.tsx`
- [X] T069 [US5] Coordinate per-view dirty state in local workspace navigation, use native discard confirmation for cancel/tab changes, preserve rejected navigation and bypass drafts on authentication loss in `frontend/src/App.tsx`, `frontend/src/routines/ExercisesView.tsx`, `frontend/src/routines/SessionsView.tsx`, and `frontend/src/routines/RoutinesView.tsx`

**Checkpoint**: All owned resources support complete, private and atomic editing; details are
snapshot-consistent and modified drafts cannot be discarded through covered internal actions without
confirmation.

---

## Phase 8: Polish & Cross-Cutting Validation

**Purpose**: Revalidate architecture, styling, accessibility, security and the complete stack after
adding editing behavior.

- [X] T070 Apply the constitutional dark daisyUI theme tokens, move the skip-link presentation from global component CSS to local Tailwind utilities, change feature links to secondary semantics and keep only document-wide rules in `frontend/src/styles.css`, `frontend/src/App.tsx`, and `frontend/src/routines/ExercisesView.tsx`
- [X] T071 [P] Extend axe and keyboard coverage for edit headings, dynamic controls, alerts, status announcements, discard flows and focus restoration across all three edit modes in `frontend/src/App.accessibility.test.tsx`
- [X] T072 Audit owner scoping, SQL parameterization, target/child error priority, lock hierarchy, commit/rollback handling, flat detail snapshots, DTO/DAO boundaries and absence of unnecessary interfaces in `backend/internal/routines/repository/repository.go`, `backend/internal/routines/service/service.go`, and `backend/internal/routines/controller/controller.go`
- [X] T073 Run `gofmt`, `go test ./...`, routines PostgreSQL tests with `-tags=integration`, and `go vet ./...` from `backend/`; resolve only feature regressions in `backend/internal/routines/`
- [X] T074 Run `npm test -- --run` and `npm run build` from `frontend/`, then execute editing, error-priority, propagation, concurrency, discard and visual scenarios from `specs/002-exercise-routines/quickstart.md` through Docker Compose and record unresolved deviations in `specs/002-exercise-routines/tasks.md`

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Phase 1 Setup
    ↓
Phase 2 Foundation
    ↓
US1 Exercises (MVP)
    ↓
US2 Sessions
    ↓
US3 Routines
    ↓
US4 Deletion
    ↓
US5 Editing
    ↓
Phase 8 Validation
```

- **Phase 1** has no feature dependency and establishes schema/test migration support.
- **Phase 2** depends on Phase 1 and blocks every authenticated story.
- **US1** depends on Phase 2 and delivers the exercise catalog independently.
- **US2** depends on US1 because session creation selects existing exercises.
- **US3** depends on US2 because routine creation selects existing sessions.
- **US4** depends on US1-US3 because it verifies cleanup across every established association.
- **US5** depends on US1-US4 because it edits all existing resource types and reuses their detail,
  validation, ownership and deletion/concurrency behavior.
- **Phase 8** depends on every story selected for delivery.

### User Story Dependencies

- **US1 (P1)**: No other story dependency after foundation; suggested MVP.
- **US2 (P2)**: Requires US1 exercise persistence and catalog API; independently testable with seeded exercises.
- **US3 (P3)**: Requires US2 session persistence and API; independently testable with seeded sessions.
- **US4 (P4)**: Requires existing reusable entities and associations; independently testable with a prepared hierarchy.
- **US5 (P5)**: Requires current exercise, session and routine contracts; independently testable
  against a prepared reusable hierarchy without changing creation or deletion semantics.

### Within Each User Story

1. Write all listed tests and confirm they fail for the intended missing behavior.
2. Add only concrete DTO and DAO shapes used by that story.
3. Implement owner-scoped PostgreSQL operations.
4. Implement service validation, authorization and transactions.
5. Implement controller mappings and protected routes.
6. Implement typed frontend API operations and UI behavior.
7. Run focused backend/frontend tests and validate the checkpoint.

## Parallel Opportunities

- T004 and T006 can run in parallel after setup because they modify separate identity/routines files.
- Within each story, service, repository, HTTP integration and frontend tests marked `[P]` use separate files.
- DTO and DAO tasks marked `[P]` use separate packages and can start together after tests fail.
- Frontend API/type work can run beside backend implementation after its story contracts are fixed.
- US2 and US3 are sequential by product data dependency, but their test design can be prepared after earlier contracts stabilize.
- US5 backend service, repository, HTTP and three view test suites can be written in parallel before
  production changes; repository and service implementations then follow dependency order.
- Styling, security and architecture audits in Phase 8 can be prepared before final command execution.

## Parallel Examples

### User Story 1

```text
T007: Service validation tests in backend/internal/routines/service/service_test.go
T008: Exercise PostgreSQL tests in backend/internal/routines/repository/repository_test.go
T009: Exercise HTTP integration tests in backend/internal/routines/tests/integration_test.go
T010: Exercise UI tests in frontend/src/routines/ExercisesView.test.tsx
```

### User Story 2

```text
T018: Session service tests in backend/internal/routines/service/service_test.go
T019: Session PostgreSQL tests in backend/internal/routines/repository/repository_test.go
T020: Session HTTP integration tests in backend/internal/routines/tests/integration_test.go
T021: Session ordering UI tests in frontend/src/routines/SessionsView.test.tsx
```

### User Story 3

```text
T029: Routine service tests in backend/internal/routines/service/service_test.go
T030: Routine PostgreSQL tests in backend/internal/routines/repository/repository_test.go
T031: Routine HTTP integration tests in backend/internal/routines/tests/integration_test.go
T032: Routine assignment UI tests in frontend/src/routines/RoutinesView.test.tsx
```

### User Story 4

```text
T040: PostgreSQL cascade and compaction tests in backend/internal/routines/repository/repository_test.go
T041: DELETE HTTP integration tests in backend/internal/routines/tests/integration_test.go
T042: Confirmation UI tests across frontend/src/routines/*View.test.tsx
```

### User Story 5

```text
T049: Update service/error-priority tests in backend/internal/routines/service/service_test.go
T050: Snapshot/replacement PostgreSQL tests in backend/internal/routines/repository/repository_test.go
T051: PUT contract tests in backend/internal/routines/tests/integration_test.go
T052-T055: Exercise, session, routine and workspace editing tests in separate frontend test files
T056: DAO shapes in backend/internal/routines/dao/dao.go
T065: Frontend PUT types and API in frontend/src/routines/types.ts and frontend/src/routines/api.ts
```

## Implementation Strategy

### MVP First

1. Complete Phase 1 schema setup.
2. Complete Phase 2 authentication foundation.
3. Complete US1 exercise catalog.
4. Stop and validate two-user privacy, optional fields and invalid text/URL behavior.
5. Demonstrate exercise creation/list/detail before adding composed entities.

### Incremental Delivery

1. Setup + foundation establish trusted ownership and schema.
2. US1 delivers reusable private exercises.
3. US2 adds ordered sessions without changing exercise behavior.
4. US3 adds complete routines without changing reusable session values.
5. US4 completes lifecycle cleanup and order compaction.
6. US5 adds full replacement editing and guarded dirty drafts without changing identifiers.
7. Phase 8 validates the entire Docker Compose stack.

## Notes

- `[P]` means different files and no dependency on an unfinished task.
- Story labels provide traceability to spec scenarios and testable behaviors.
- Tests precede corresponding production code and must fail for the expected reason.
- Every repository method receives verified user ID; no request body supplies ownership.
- DTO and DAO are organizational types, not additional layers.
- Commit after each task or coherent task group.
- Do not edit README files, add speculative features or introduce new dependencies.
