# Phase 0 Research: Rutinas de ejercicios

## Cohesive backend module

**Decision**: Implement exercises, workout sessions and routines in one `routines` module with
controller, service, repository, DTO, DAO, errors and integration-test packages.

**Rationale**: The entities form one capability, share ownership rules and participate in the same
transactions. One module keeps related SQL, validation and nested reads together.

**Alternatives considered**: One module per entity would require cross-module coordination and
contracts without independent domains. A generic CRUD module would obscure the association rules.

## Reusing authentication

**Decision**: Expose the existing identity authentication wrapper and compose it around routines
routes in `cmd/api/main.go`. Continue reading verified user ID from `platform/requestctx`.

**Rationale**: JWT parsing, HS256 validation and identity propagation already exist. Composition at
the application root avoids duplicate token logic and keeps routines independent from cryptography.

**Alternatives considered**: Moving token validation into `platform` would be a broad refactor.
Creating a token-validator interface has one implementation and no current testing necessity.
Accepting user ID from the request would violate the ownership boundary.

## Tenant-safe relational ownership

**Decision**: Give each owner entity a composite primary key `(user_id, id)` and reference both
columns from association tables. Scope every repository query by verified user ID.

**Rationale**: Composite foreign keys make cross-user associations persistently impossible while
tenant-scoped queries prevent disclosure. They use native PostgreSQL primary-key and foreign-key
constraints without RLS configuration.

**Alternatives considered**: A global ID primary key plus `(user_id, id)` unique constraint adds an
extra index without a current requirement. Application-only ownership allows an omitted predicate to
persist invalid associations. RLS adds roles, policies and per-transaction connection context beyond
the academic scope.

**Source**: PostgreSQL 18 documents composite primary and foreign keys and notes that foreign-key
columns are not indexed automatically: https://www.postgresql.org/docs/18/ddl-constraints.html

## Minimal persistence model

**Decision**: Use five tables: `exercises`, `workout_sessions`, `session_exercises`, `routines` and
`routine_sessions`. Use relational columns and no JSON, timestamps, statuses or soft deletion.

**Rationale**: Two many-to-many associations have their own required values. Each table represents
a real multiplicity or lifecycle; all other proposed metadata is outside current requirements.

**Alternatives considered**: JSON arrays prevent foreign-key ownership and constraint enforcement.
Separate day or ordering catalogs add no behavior. Snapshotting nested data would defeat reuse.

## Persistent constraints

**Decision**: PostgreSQL enforces required values, bounded technical lengths, text whitespace shape,
non-negative quantities, execution order greater than zero, day 1..7, unique session exercise,
unique session/day assignment and composite ownership. Full URL parsing and exact order
consecutiveness remain service rules, with database guards for URL scheme/whitespace and order
uniqueness.

**Rationale**: Row-local and relational invariants belong in native constraints. PostgreSQL `CHECK`
cannot safely express a sequence across multiple rows, and a database URL regex should not pretend
to implement the complete application parsing policy.

**Alternatives considered**: Triggers for consecutiveness and URL logic hide application behavior
and complicate deletion. Application-only constraints weaken integrity for alternate writers.

**Source**: PostgreSQL recommends `UNIQUE` and `FOREIGN KEY` for cross-row/table restrictions and
does not support cross-row `CHECK` guarantees:
https://www.postgresql.org/docs/18/ddl-constraints.html

## Composite creation transactions

**Decision**: Service opens a pgx transaction for session and routine creation, validates the full
input first, locks all selected tenant-owned rows with `FOR KEY SHARE`, inserts the parent and then
its associations, and commits. Repository methods accept `pgx.Tx` only for these operations.

**Rationale**: Selection ownership must still hold at confirmation time. Multiple inserts must
commit or roll back together, while service owns case-use atomicity.

**Alternatives considered**: Repository-owned transactions place a use-case decision in persistence.
A UnitOfWork or transaction manager is unnecessary. One transaction for every CRUD call adds noise.
Bulk JSON parameters and ORM cascades are more complex than short explicit loops at this scale.

**Source**: pgx transactions support explicit begin, rollback and commit; rollback after a committed
transaction is safe: https://pkg.go.dev/github.com/jackc/pgx/v5

## Deletion and association cleanup

**Decision**: Association foreign keys use `ON DELETE CASCADE`. Routine and session deletion remain
single scoped delete statements. Exercise deletion uses a service-owned transaction that locks the
exercise and affected sessions, deletes it, and compacts remaining orders with `row_number()` while
the unique order constraint is deferred.

**Rationale**: Associations have no independent lifecycle, so cascades implement their removal as
part of the parent statement. Exercise deletion additionally changes surviving order values and
therefore needs explicit orchestration and locking.

**Alternatives considered**: Cascading from routine to sessions would delete reusable content.
A renumbering trigger hides a workflow. Updating each row outside a transaction permits partial or
conflicting orders.

**Source**: PostgreSQL describes `CASCADE` as appropriate when referencing rows are components that
cannot exist independently: https://www.postgresql.org/docs/18/ddl-constraints.html

## HTTP contract shape

**Decision**: Expose POST, GET collection, GET detail, PUT detail and DELETE detail for all three
resources. PUT replaces the complete editable representation, returns the complete updated detail
with `200`, and reuses the creation payload shape. Association arrays are required and may be empty;
optional empty or omitted fields clear their stored values.

**Rationale**: These fifteen operations cover create, list, detail, edit and delete. PUT is
idempotent and avoids ambiguous merge rules for nested arrays. Direct arrays avoid pagination
metadata that the specification excludes.

**Alternatives considered**: PATCH needs nullable DTO fields plus absent/null/empty merge semantics,
especially for associations. Separate association endpoints expose persistence details. Search,
pagination and batch APIs exceed scope.

## Atomic full replacement

**Decision**: Update an exercise with one tenant-scoped `UPDATE ... RETURNING`. For session and
routine, validate pure rules first, begin a transaction, probe target existence without a lock, lock
selected children by ascending ID, re-lock/recheck the owned target, and only then expose any
unavailable-child result. A surviving valid target is updated by deleting old associations and
inserting the complete new set.

**Rationale**: The non-locking probe makes a statically absent or foreign target win before child
inspection. Delaying child-error classification until the target lock makes a concurrently deleted
target win too. Real lock order remains exercises, sessions, routines, preventing a cycle with
deletion. Delete-and-insert keeps replacement and rollback simple.

**Alternatives considered**: Updating associations in place complicates uniqueness and diff logic.
Deleting and recreating the parent changes identity and references. Repository-owned transactions,
generic replacement helpers and unit-of-work abstractions move use-case rules or add indirection.
Locking the target before children makes error priority trivial but inverts the shared lock order.

## Editing concurrency

**Decision**: Use PostgreSQL `READ COMMITTED` row locks and last-committed-writer semantics. Every
successful composite edit is one complete state; no ETag, version column, retry framework, history
or conflict-resolution UI is added.

**Rationale**: Atomicity and ownership are required, while conflict detection is not. Existing
schema constraints and locks already prevent partial or cross-owner associations.

**Alternatives considered**: Optimistic locking requires schema, API and UI conflict flows without a
current academic requirement. Serializable isolation and automatic retries add operational logic.

## Consistent nested details

**Decision**: Read each session or routine detail with one flat tenant-scoped SQL statement using
`LEFT JOIN` through associations and reusable children. Aggregate nullable rows in repository. Read
the PUT response with the same SQL inside its write transaction before commit, and commit before
writing the HTTP response.

**Rationale**: PostgreSQL gives one `READ COMMITTED` statement a single MVCC snapshot. A detail is
therefore entirely before or after a concurrent commit, including empty containers, without a
read-only transaction. Reading the PUT response before commit guarantees that response represents
that replacement rather than a later writer.

**Alternatives considered**: Current parent/detail queries in separate autocommit statements can
mix snapshots and violate FR-042. A read-only `REPEATABLE READ` transaction is correct but adds
transaction control and round trips. JSON aggregation, views and materialization are unnecessary.

## Live reference propagation

**Decision**: Keep only keys and association-specific values in join tables. Detail queries join
current exercise and session rows, so queries started after commit expose edited reusable data in
every container. A query already running may return the previous complete snapshot.

**Rationale**: Normalized live references satisfy propagation without fan-out writes while
preserving series, repetitions, order and day on their respective associations.

**Alternatives considered**: Copied names, descriptions or nested snapshots can diverge and require
bulk updates, triggers, events or cache invalidation.

## Public errors

**Decision**: Keep existing error shape. Invalid path IDs and malformed bodies use
`invalid_request`; pure business validation uses `validation_failed`; authentication uses
`invalid_token`; valid foreign/missing targets share `not_found`. For PUT, pure validation precedes
target lookup, and target lookup precedes public child-availability errors.

**Rationale**: The contract gives actionable creation feedback while preserving non-disclosure.
Known errors remain mapped only by controller.

**Alternatives considered**: `403` for foreign IDs reveals existence. A global error registry or
large error hierarchy adds abstraction without a second consumer.

## Frontend ordering and navigation

**Decision**: Add authenticated local navigation for routines, sessions and exercises. Use buttons
to move selected exercises up or down, derive order as `index + 1`, and confirm deletion with
`window.confirm`.

**Rationale**: Local state matches the small interface. Buttons are accessible, deterministic and
easy to test with current tools. Native confirmation meets the requirement without another modal.

**Alternatives considered**: React Router and global state solve no current navigation problem.
Drag-and-drop adds accessibility, event and dependency complexity. A custom modal adds state solely
to replace a sufficient browser control.

## Frontend editing state

**Decision**: Each existing view keeps one form with a discriminated `create | edit` mode, editing
identifier and normalized initial draft. Entering edit preloads complete detail and save sends PUT.
If the current draft differs, cancel or an internal workspace-section change uses `window.confirm`;
rejecting keeps form and section, accepting discards without PUT. Unchanged forms leave directly.
Pending saves disable submission; validation and server errors preserve entered data.

**Rationale**: One form reuses existing validation, ordering and selection controls without a
second page hierarchy. Explicit per-view comparison is clearer than a form framework. Session array
order is meaningful; routine assignment order is canonicalized because only `(sessionId, day)` has
meaning. The complete response updates list and selected detail without an extra GET.

**Alternatives considered**: Duplicate edit components drift from creation behavior. A custom modal
requires dialog focus management when native confirmation already exists. Persisted drafts and a
generic form framework exceed scope. `beforeunload` for closing or reloading the browser is excluded
because the clarified scenarios cover cancel and internal section changes only.

## Visual constitution alignment

**Decision**: Use daisyUI semantic components and Tailwind utilities in each routines view. Define
the constitutional dark palette once in the global daisyUI theme; keep global CSS limited to theme
tokens and document-wide behavior. Links use the secondary semantic color and destructive actions
use error. No feature-specific rule is added to the global stylesheet.

**Rationale**: Semantic classes preserve consistent contrast and make affected styling traceable to
the component while satisfying constitution 1.3.0.

**Alternatives considered**: Hard-coded component hex values duplicate palette knowledge. A new
design system or monolithic stylesheet adds abstraction and makes feature styles harder to locate.

## Frontend API and media

**Decision**: Add a routines-specific authorized request helper that attaches the stored token,
maps existing errors and clears authentication on `401`. Render image/video URLs as external links
with safe link attributes; do not fetch or embed their contents.

**Rationale**: The helper addresses repeated authenticated calls inside one feature without creating
a global HTTP framework. External media validity is limited to URL syntax by specification.

**Alternatives considered**: A third-party HTTP client is unnecessary. Embedding or probing media
would introduce network behavior and content validation outside scope.

## Testing levels

**Decision**: Test pure validation in service, PostgreSQL invariants and transactions against the
existing test database, HTTP/auth flows through integration tests, and stateful UI behavior with
existing Vitest/Testing Library tooling.

**Rationale**: Each rule is tested near its authority. PostgreSQL semantics and cross-user isolation
cannot be proven with string-based SQL mocks.

**Alternatives considered**: Mock frameworks would add dependency and duplicate query details.
Status-only endpoint tests do not protect meaningful behavior.
