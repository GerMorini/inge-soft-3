# Data Model: Rutinas de ejercicios

## Overview

The feature adds three user-owned reusable entities and two associations. PostgreSQL 18.4 is the
selected engine. The model uses native relational constraints, no extensions and no JSON. Its write
surface is normal application DML; service controls transactions for composite use cases. Editing
and the FitPro visual redesign reuse this schema and require no additional migration.

## Transient frontend state

The wizard introduces client-only state. None of it is persisted outside existing POST/PUT payloads.

```text
WizardState
├── mode: create | edit
├── activeStep: integer index
├── saving: boolean
├── baseline: normalized initial draft
├── draft: ExerciseInput | SessionInput | RoutineInput
└── fieldErrors: map keyed by existing API field paths
```

- Exercise draft contains scalar writable fields.
- Session draft contains selected exercises in execution order; each selection contains exercise
  reference, series and repetitions. `order` is derived as position plus one.
- Routine draft contains selected `(sessionId, day)` assignments. The normalized baseline sorts by
  day and session ID because visual row position has no persistent meaning.
- Search queries and current result visibility are ephemeral. Filtering never mutates selections.
- Active disclosure cards, detail modal and media fallback states are ephemeral presentation state.
- Routine catalog search may append the same session repeatedly. Every draft assignment receives a
  local row key and unset day until the user chooses one; only `(sessionId, day)` reaches the API.
- Disclosure open state is independent per stable entity/association key and is never persisted.
- Closing after accepted discard removes the draft without any API request. Successful Save sends
  the same complete write contract, updates the card from its response and then removes transient
  state without a follow-up GET.

## Derived response field: Session exercise count

`exerciseCount` is not stored. `GET /api/sessions` derives it with an owner-scoped aggregate over
`session_exercises` and returns `0` for an empty session. It belongs to `SessionSummary` only;
session details already expose `exercises` and clients derive their length. Creation/update detail
responses are projected to summary cards with `exerciseCount = exercises.length`.

## Logical model

```text
User 1 ── N Exercise
User 1 ── N WorkoutSession
User 1 ── N Routine

WorkoutSession N ── N Exercise       through SessionExercise
Routine N ── N WorkoutSession        through RoutineSession
```

Every relationship carries `user_id`. Associations can only reference entities owned by the same
user.

## Entity: Exercise

| Field | PostgreSQL type | Null | Rules |
|---|---|---:|---|
| `user_id` | `bigint` | No | Owner; references `users(id)` |
| `id` | `bigint GENERATED ALWAYS AS IDENTITY` | No | Identifier within owner |
| `name` | `varchar(100)` | No | Valid single-line spaced text |
| `description` | `varchar(500)` | Yes | Same text rule when present |
| `image_url` | `varchar(2048)` | Yes | HTTP/HTTPS absolute URL when present |
| `video_url` | `varchar(2048)` | Yes | HTTP/HTTPS absolute URL when present |

Primary key: `(user_id, id)`. Names are deliberately not unique.

## Entity: Workout session

| Field | PostgreSQL type | Null | Rules |
|---|---|---:|---|
| `user_id` | `bigint` | No | Owner; references `users(id)` |
| `id` | `bigint GENERATED ALWAYS AS IDENTITY` | No | Identifier within owner |
| `name` | `varchar(100)` | No | Valid single-line spaced text |
| `description` | `varchar(500)` | Yes | Same text rule when present |

Primary key: `(user_id, id)`. A session may have zero associated exercises.

## Association: Session exercise

| Field | PostgreSQL type | Null | Rules |
|---|---|---:|---|
| `user_id` | `bigint` | No | Shared owner |
| `session_id` | `bigint` | No | Tenant-safe FK to workout session |
| `exercise_id` | `bigint` | No | Tenant-safe FK to exercise |
| `series_count` | `integer` | No | `0..2147483647` |
| `repetition_count` | `integer` | No | `0..2147483647` |
| `execution_order` | `integer` | No | At least 1; unique and consecutive in session |

Primary key `(user_id, session_id, exercise_id)` prevents exercise duplication. A named,
deferrable unique constraint on `(user_id, session_id, execution_order)` protects order uniqueness.
Consecutiveness `1..N` is validated by service on creation and replacement, and restored
transactionally after an exercise deletion.

## Entity: Routine

| Field | PostgreSQL type | Null | Rules |
|---|---|---:|---|
| `user_id` | `bigint` | No | Owner; references `users(id)` |
| `id` | `bigint GENERATED ALWAYS AS IDENTITY` | No | Identifier within owner |
| `name` | `varchar(100)` | No | Valid single-line spaced text |
| `description` | `varchar(500)` | Yes | Same text rule when present |

Primary key: `(user_id, id)`. A routine may have zero assigned sessions.

## Association: Routine session

| Field | PostgreSQL type | Null | Rules |
|---|---|---:|---|
| `user_id` | `bigint` | No | Shared owner |
| `routine_id` | `bigint` | No | Tenant-safe FK to routine |
| `session_id` | `bigint` | No | Tenant-safe FK to workout session |
| `day_of_week` | `smallint` | No | 1 Monday through 7 Sunday |

Primary key `(user_id, routine_id, day_of_week, session_id)` rejects the same session/day pair while
allowing multiple sessions on one day and one session on several different days.

## Text and URL normalization

- Names must be present exactly as submitted and have no leading/trailing whitespace, repeated
  separators, tab or newline.
- Present descriptions use the same rule. Empty descriptions become `NULL` before persistence.
- Empty optional URLs become `NULL`.
- Service parses each present URL, requires an absolute HTTP or HTTPS URL with a host, and rejects
  any whitespace. PostgreSQL adds a simple scheme/no-whitespace check as defense, not a full parser.
- Input is rejected rather than silently trimmed or collapsed.

## Foreign keys and cascades

- Owner tables reference `users(id)` with default `NO ACTION`; account deletion is outside scope.
- `session_exercises` references `(user_id, session_id)` and `(user_id, exercise_id)` with
  `ON DELETE CASCADE`.
- `routine_sessions` references `(user_id, routine_id)` and `(user_id, session_id)` with
  `ON DELETE CASCADE`.
- Cascades delete only associations. Reusable exercises, sessions and routines do not cascade into
  each other.

## Indexes

- Primary and unique constraints supply owner/detail, session composition, routine composition and
  execution-order indexes.
- Add `(user_id, exercise_id, session_id)` on `session_exercises` for exercise deletion and affected
  session discovery.
- Add `(user_id, session_id, routine_id)` on `routine_sessions` for session deletion cleanup.
- Do not index names or URLs because search, filtering and uniqueness are excluded.

## Transactional operations

### Create exercise

Validate and execute one tenant-owned `INSERT ... RETURNING`. No explicit transaction is needed.

### Create session

1. Validate all fields, quantities, duplicates and order before database work.
2. Begin transaction in service.
3. Verify and lock all selected exercises for `user_id` in one query.
4. Insert the workout session.
5. Insert each session-exercise row.
6. Commit. Any missing/foreign selection or persistence error rolls back everything.

### Create routine

Use the same pattern, locking selected workout sessions before inserting routine assignments.
Duplicate `(session_id, day_of_week)` pairs fail validation before persistence and remain protected
by the primary key.

### Update exercise

Execute one `UPDATE ... WHERE user_id = $1 AND id = $2 RETURNING ...`. Zero rows maps to the same
result for absent and foreign resources. Empty optional values persist as `NULL`. Associations keep
the same exercise ID and expose its new values in subsequent nested details.

### Replace session

1. Validate the complete input before opening a transaction.
2. Probe the tenant-owned target without locking; absence returns not found before child inspection.
3. Lock selected tenant-owned exercises with `FOR KEY SHARE`, ordered by ID, retaining any missing
   result internally.
4. Lock and recheck the target tenant-owned session with `FOR UPDATE`.
5. If target disappeared return not found; otherwise report missing children only after target
   confirmation.
6. Update fields, delete prior `session_exercises` rows and insert the complete new set.
7. Read canonical detail through one joined statement inside the transaction, then commit; any
   failure restores all previous fields and associations.

The required association array may be empty. Delete-before-insert avoids diff logic and conflicts
with prior execution orders.

### Replace routine

1. Validate the complete input before opening a transaction.
2. Probe the tenant-owned target without a lock.
3. Lock selected tenant-owned sessions with `FOR KEY SHARE`, ordered by ID, retaining missing IDs.
4. Lock and recheck the target tenant-owned routine with `FOR UPDATE`.
5. Resolve disappearing target before selected-session errors, replace fields and associations,
   read the joined detail inside the transaction, then commit.

All writers follow lock order exercise, workout session, routine, with ascending IDs inside each
group. Default `READ COMMITTED` is sufficient. Concurrent replacements use the last committed
complete state and do not require version columns.

### Read consistent detail

- Session detail uses one statement from `workout_sessions` through `LEFT JOIN session_exercises`
  and `LEFT JOIN exercises`, ordered by execution order.
- Routine detail uses one statement from `routines` through both associations and reusable entities,
  ordered by day, session ID and exercise order.
- `LEFT JOIN` keeps one nullable parent row for an empty container. Repository aggregates rows by
  session assignment `(session_id, day)` where necessary.
- One PostgreSQL statement observes one committed MVCC snapshot under `READ COMMITTED`; no read
  transaction is required and a detail cannot mix states around a concurrent commit.
- Association tables store no copied descriptive data. Joined reads propagate committed exercise
  or session edits to every containing detail while preserving association-specific values.

### Delete routine or session

A scoped `DELETE WHERE user_id = $1 AND id = $2` is one atomic statement. Foreign-key cascades
remove only dependent associations. Zero affected rows becomes the same not-found result for absent
and foreign entities.

### Delete exercise and compact order

1. Begin transaction and lock the scoped exercise.
2. Discover affected sessions and lock them in ascending ID order.
3. Defer the named execution-order unique constraint.
4. Delete the exercise; cascade removes its associations.
5. Reassign remaining orders with `row_number()` partitioned by owner/session and ordered by prior
   execution order plus exercise ID.
6. Commit all removal and compaction together.

## Rule authority matrix

| ID | Rule | Persistence | Authority | Capability | Database mechanism | Outside database |
|---|---|---|---|---|---|---|
| BR-001 | Every entity has one owner | `user_id` | BDD | CAP_FOREIGN_KEY | NOT NULL FK to users | JWT identifies actor |
| BR-002 | Associations cannot cross owners | composite FKs | BDD | CAP_COMPOSITE_FOREIGN_KEY | `(user_id,id)` references | Service returns generic unavailable error |
| BR-003 | Quantities are non-negative integers | association columns | BDD | CAP_CHECK_CONSTRAINT | `integer` and named CHECK | DTO rejects decimals/overflow |
| BR-004 | Day is 1..7 | `day_of_week` | BDD | CAP_CHECK_CONSTRAINT | `smallint` and named CHECK | UI displays weekday names |
| BR-005 | Exercise appears once per session | association PK | BDD | CAP_UNIQUE_CONSTRAINT | composite primary key | Service aggregates validation |
| BR-006 | Session/day pair appears once per routine | association PK | BDD | CAP_UNIQUE_CONSTRAINT | composite primary key | Service aggregates validation |
| BR-007 | Execution order is unique | unique columns | BDD | CAP_UNIQUE_CONSTRAINT | deferrable named UNIQUE | Service validates sequence |
| BR-008 | Execution order is consecutive | all rows in session | APLICACIÓN | APPLICATION_REQUIRED | unique positive pieces only | Service validates and compacts in transaction |
| BR-009 | URLs are absolute HTTP/HTTPS | URL columns | COMPARTIDA | CAP_CHECK_CONSTRAINT | basic scheme/whitespace CHECK | Go URL parser is semantic authority |
| BR-010 | Composite changes are atomic | several tables | COMPARTIDA | transaction | PostgreSQL transaction/cascade | Service owns begin/commit/rollback |
| BR-011 | Foreign and absent IDs are indistinguishable | no extra state | APLICACIÓN | NOT_APPLICABLE | tenant-scoped no-row result | Controller maps one public error |
| BR-012 | Session/routine edits replace all associations atomically | parent and association rows | COMPARTIDA | transaction | row locks, DELETE and INSERT before commit | Service validates and owns transaction |
| BR-013 | Editing preserves resource ID and owner | composite parent key | BDD | CAP_PRIMARY_KEY | scoped UPDATE never changes key | DTO omits ID and user ID |
| BR-014 | Missing target takes priority over unavailable selected children | no extra state | APLICACIÓN | NOT_APPLICABLE | scoped probe and target recheck | Service delays child error classification |
| BR-015 | Each nested detail represents one committed state | joined rows | BDD | transaction snapshot | one SQL statement under READ COMMITTED | Repository aggregates one result set |
| BR-016 | Reusable edits propagate to all containers | normalized foreign keys | BDD | CAP_FOREIGN_KEY | joins read current referenced rows | No snapshot copies or fan-out updates |

## Capability resolution

PostgreSQL 18.4 natively supplies all selected database capabilities: NOT NULL, CHECK, UNIQUE,
composite primary/foreign keys, deferrable unique constraints, DML transactions, row locks,
`ON DELETE CASCADE` and window functions. No extension, stored routine, trigger, RLS policy or
accepted degradation is required. Integration tests must prove the effective schema behavior on the
project's PostgreSQL 18.4 test container. Updates need no migration beyond migration 002.

## Excluded persistence

- execution history, progress, weights and rest periods;
- editing history, soft deletion or audit timestamps;
- sharing, templates, visibility and permission records;
- files or downloaded media;
- ordering of catalog lists or routine sessions beyond day;
- search indexes, pagination cursors or counters.
