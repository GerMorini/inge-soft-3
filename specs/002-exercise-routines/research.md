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
`window.confirm`. Keep navigation as local state; add only the explicitly requested Lucide icons.

**Rationale**: Local state matches the small interface. Buttons are accessible, deterministic and
easy to test with current tools. Native confirmation meets the requirement without another modal.

**Alternatives considered**: React Router and global state solve no current navigation problem.
Drag-and-drop adds accessibility, event and dependency complexity. A custom confirmation modal adds
state solely to replace a sufficient browser control.

## Frontend editing state

**Decision**: Move each concrete create/edit form into an entity-specific wizard displayed by one
shared native-dialog shell. Entering edit preloads complete detail and save sends PUT. Backdrop click
always asks whether to discard; other exit paths and workspace changes retain dirty comparison.
Pending saves disable submission and closure; validation and server errors preserve entered data.

**Rationale**: The same wizard shell now has three current consumers, while concrete draft rules
remain understandable. Session array order is meaningful; routine assignment order is canonicalized
because only `(sessionId, day)` has meaning. The complete response creates or replaces the matching
card, then the wizard closes and announces success without GET or automatic detail opening.

**Alternatives considered**: Separate create/edit screens drift. A generic `useWizardForm<T>`, form
framework, persisted drafts or global wizard context hide small domain-specific rules. A modal
library duplicates native `<dialog>` and daisyUI behavior. `beforeunload` remains excluded.

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
maps existing errors and clears authentication on `401`. The redesigned interface may load external
images lazily and direct video URLs through native media elements only when details expand. Failed,
unsupported or page/platform URLs always retain a safe external link; no iframe is created.

**Rationale**: The helper addresses repeated authenticated calls inside one feature without creating
a global HTTP framework. Progressive browser-native previews satisfy the visual requirement without
adding a backend proxy, host allowlist, content inspection or third-party player SDK.

**Alternatives considered**: A third-party HTTP client is unnecessary. Embedding or probing media
through arbitrary iframes allows untrusted pages and often fails frame policies. A backend proxy or
media service changes security, storage and deployment scope. Links alone no longer satisfy the
requested thumbnails and video previews.

## Session card exercise count

**Decision**: Add `exerciseCount` to `SessionSummary` and derive it in the existing owner-scoped
session list query with `LEFT JOIN` and `COUNT`. Return zero for empty sessions. Detail responses keep
their current shape; DAO/DTO expose PostgreSQL `COUNT` as int64 and frontend derives summary count
from `exercises.length` after POST/PUT.

**Rationale**: Session cards require the count before expansion. One aggregate query provides it
without loading full details or issuing one request per card. The value is derived and needs no
column, trigger or migration.

**Alternatives considered**: N detail requests add latency and complexity. Hiding the count until
expansion contradicts FR-057. Persisting a counter introduces synchronization without need. Adding
the count to every nested detail duplicates information already represented by the array.

## Media CSP and browser boundary

**Decision**: Extend nginx CSP narrowly with HTTP/HTTPS sources for `img-src` and a new `media-src`;
keep all script/style/object/frame restrictions. Mount external media only after disclosure, use
native `<video controls playsInline preload="metadata">`, keep a safe link always visible and show
textual fallback on `error`. Do not infer support from filename or make HEAD/fetch probes.

**Rationale**: Current `img-src 'self' data:` and implicit `default-src 'self'` block the requested
remote previews in Docker. The browser is the correct codec/MIME authority. Deferred mounting limits
unrequested transfers, while a permanent link covers slow, blocked, mixed-content and unsupported
resources.

**Alternatives considered**: `default-src *` weakens unrelated protections. Proxying or probing media
adds SSRF/content handling. Host allowlists contradict arbitrary valid URLs. Iframes and platform
SDKs violate FR-061.

## FitPro wizard composition

**Decision**: Add `ModalDialog`, `WizardDialog`, `SearchableCatalog` and `MediaPreview` as four
concrete shared components. Add `ExerciseWizard`, `SessionWizard` and `RoutineWizard` for domain
drafts and payloads. Views retain data fetching, list/detail state, deletion and result application.

**Rationale**: Modal lifecycle, step tabs and searchable catalogs now repeat across three flows;
media preview/fallback repeats in exercise and routine detail. The catalog component owns filtering
only, allowing exercise checkboxes and repeatable routine Add actions without a generic selector
state. Domain validation, summaries and association rules remain explicit in each wizard.

**Alternatives considered**: One generic schema-driven wizard requires configuration objects,
generic error paths and mappers larger than the three forms. Duplicating complete modal logic risks
inconsistent focus and discard behavior. Form, modal and drag-and-drop libraries add dependencies
without solving an unmet browser capability.

## Wizard navigation and dismissal

**Decision**: Present wizard steps as an accessible tablist with bottom-border selection. Tabs,
Previous and Next can move to any step without validating incomplete fields. Full validation occurs
only on Save from `Resumen` and sends the user to the first erroneous step. Earlier steps never show
Save. Backdrop click always uses the supplied discard confirmation; inside clicks never close.
Escape/Close/Cancel use dirty state. Saving blocks all closing and never asks for confirmation.

**Rationale**: This directly preserves exploratory step navigation and prevents accidental loss or
duplicate writes. A single close path keeps draft, focus and pending behavior consistent.

**Alternatives considered**: Gating each step contradicts the requirement. Auto-save changes API and
draft persistence. Nested confirmation dialogs create focus traps; `window.confirm` already provides
a blocking accessible decision for this academic scope.

## Searchable selection

**Decision**: Implement `SearchableCatalog` as a labelled `input type="search"` plus result list.
Filtering is case-insensitive and in-memory over the already loaded owner catalog. Session wizard
results use native checkboxes and preserve hidden selections. Routine results use an `Agregar`
button; every activation appends a fresh row with no day selected and leaves that session available
for another activation. Save validates missing days and duplicate `(sessionId, day)` pairs.

**Rationale**: This meets both unique exercise selection and explicitly repeatable session selection
without a custom combobox keyboard protocol. An unset day avoids creating an invalid default
duplicate before the user chooses. Academic catalog size makes local filtering appropriate.

**Alternatives considered**: A custom ARIA combobox requires active-descendant, popup and keyboard
state. Backend search and pagination add endpoints excluded by scope. Native `<select multiple>` has
poor discoverability and cannot append the same session repeatedly. A single session row with many
day checkboxes contradicts the selected repeated-search behavior.

## Successful wizard completion

**Decision**: A successful POST/PUT applies its returned representation directly to the matching
card, closes the wizard, restores focus to a logical view control and announces success. It performs
no GET and opens no detail modal. Multiple independent `<details>` may remain open simultaneously.

**Rationale**: The mutation response already contains canonical data. Closing completes the task
without a redundant request or context switch. Independent native disclosures require no accordion
coordinator and let users compare content.

**Alternatives considered**: Keeping a saved wizard open leaves stale editable state. Opening detail
adds an unrequested transition. Refetching duplicates current data. Accordion state adds coordination
and closes information the user may be comparing.

## Supplied gym image

**Decision**: Fetch the user-supplied 5177×3410 JPEG once during implementation, optimize it into a
versioned local WebP asset and render it as a decorative absolute `<img>` for auth and heroes. Use
dark base fallback, `object-cover` and overlays; auth receives a stronger left gradient.

**Rationale**: A local asset fulfills the selected visual while keeping Docker builds reproducible,
avoiding runtime hotlink failure, privacy leakage and third-party availability. An `<img>` permits
load-error handling unlike CSS background alone.

**Alternatives considered**: Runtime DuckDuckGo/Pinterest hotlinking exposes every user request and
can disappear. Keeping the original 2.37 MB JPEG increases transfer unnecessarily. A generated
replacement would not be the image explicitly supplied.

**Source asset**:
`https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Fi.pinimg.com%2Foriginals%2F3f%2F5c%2Fe4%2F3f5ce46ff01c0d0bc64a2652c34b5c54.jpg&f=1&nofb=1&ipt=e2731386957c5511a51be9c5a1a411bce361d44cd142e9e0bf3ccdfadaebe346`

## Lucide icon dependency

**Decision**: Add exact `lucide-react@1.31.0`, verified against the npm registry on 2026-08-13. Use
named imports only for brand, navigation and user actions. Icons inherit `currentColor`; beside-text
icons are decorative, while icon-only buttons receive contextual accessible names.

**Rationale**: Lucide is explicitly required, supports React 19, includes TypeScript declarations,
uses tree-shakeable SVG components and has no runtime service. One maintained package is clearer
than copying many SVG paths into local components.

**Sources**: https://lucide.dev/ and https://www.npmjs.com/package/lucide-react

**Alternatives considered**: Hand-authored SVGs duplicate icon data and violate the requested icon
source. Dynamic icon registries defeat straightforward tree shaking. A second icon package or generic
Icon wrapper adds no value.

## Responsive and accessibility strategy

**Decision**: Auth uses one column on small screens and split composition at large widths. Navbar
wraps without hiding destinations. Cards use one/two-column grids. Dialogs become viewport-width and
scrollable on mobile. Native `<dialog>`, `<details>`, inputs, checkboxes and buttons carry keyboard
semantics; opening and closing restore focus. Every touch action remains at least 44 CSS pixels.

**Rationale**: Tailwind breakpoints and native elements satisfy current layouts without JavaScript
media queries or focus-trap code. Text and opaque/gradient overlays preserve contrast independently
of the photograph.

**Alternatives considered**: Mobile-specific component trees duplicate behavior. Manual focus traps
and disclosure widgets are error-prone. Hiding navigation behind hover is inaccessible.

## Testing levels

**Decision**: Test pure validation in service, PostgreSQL invariants and transactions against the
existing test database, HTTP/auth flows through integration tests, and stateful UI behavior with
existing Vitest/Testing Library tooling.

**Rationale**: Each rule is tested near its authority. PostgreSQL semantics and cross-user isolation
cannot be proven with string-based SQL mocks.

**Alternatives considered**: Mock frameworks would add dependency and duplicate query details.
Status-only endpoint tests do not protect meaningful behavior.
